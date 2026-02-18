// Package npc defines NPC management logic
package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/pubsub"
	"github.com/webbben/2d-game-engine/object"
)

type NPC struct {
	debug  debug
	Entity *entity.Entity

	dialogProfileID defs.DialogProfileID // retrieved from character state, but set here for convenience

	TaskMGMT
	eventBus   *pubsub.EventBus
	OnUpdateFn func(n *NPC)
	World      WorldContext

	ID          string // TODO: What is ID for? Should we just refer directly to CharacterStateID instead?
	DisplayName string

	// priority assigned to this NPC by the map it is added to. used for prioritizing which NPC moves first in a collision.
	Priority int
}

func (n *NPC) Activate() {
	if n.dialogProfileID != "" {
		n.World.StartDialog(n.dialogProfileID, n.ID)
		return
	}
	logz.Println(n.DisplayName, "nothing happened on activation")
}

// Y is used for renderables sorting
func (n NPC) Y() float64 {
	return n.Entity.Y
}

type WorldContext interface {
	FindNPCAtPosition(c model.Coords) (NPC, bool)
	FindObjectsAtPosition(c model.Coords) []*object.Object
	StartDialog(dialogProfileID defs.DialogProfileID, npcID string)
}

type NPCParams struct {
	CharStateID state.CharacterStateID
}

func NewNPC(params NPCParams, defMgr *definitions.DefinitionManager, audioMgr *audio.AudioManager, eventBus *pubsub.EventBus) *NPC {
	if params.CharStateID == "" {
		panic("CharStateID is empty")
	}
	if defMgr == nil {
		panic("defMgr was nil")
	}
	if audioMgr == nil {
		panic("audioMgr was nil")
	}
	if eventBus == nil {
		panic("eventBus was nil")
	}

	ent := entity.LoadCharacterStateIntoEntity(params.CharStateID, defMgr, audioMgr)

	// get dialog profile ID from character def
	charDef := defMgr.GetCharacterDef(ent.CharacterStateRef.DefID)
	scheduleDef := defMgr.GetScheduleDef(charDef.ScheduleID)

	n := NPC{
		eventBus:        eventBus,
		Entity:          ent,
		ID:              string(ent.ID()),
		DisplayName:     ent.DisplayName(),
		dialogProfileID: charDef.DialogProfileID,
		TaskMGMT: TaskMGMT{
			Schedule: scheduleDef,
			defMgr:   defMgr,
		},
	}

	n.eventBus.SubscribeToNPCEvents(n.ID, n.ID, n.OnEvent)

	return &n
}

func (n *NPC) OnEvent(e defs.Event) {
	logz.Println(n.ID, "received event:", e)
	assignTask := pubsub.NpcAssignTaskType(n.ID)
	switch e.Type {
	case assignTask:
		taskDefData, exists := e.Data["taskDef"]
		if !exists {
			logz.Panicln(n.ID, "recieved assign task event, but data was missing expected key")
		}
		taskDef, ok := taskDefData.(defs.TaskDef)
		if !ok {
			logz.Panicln(n.ID, "failed to type event data as taskDef")
		}
		n.RunTask(taskDef, n)
	}
}

// TaskMGMT manages all task related stuff, such as schedules and whatnot.
// TODO:
// - default to Idle task if no schedule or current task is set?
type TaskMGMT struct {
	CurrentTask         Task
	waitUntil           time.Time
	waitUntilDoneMoving bool // if set, will wait until entity has stopped moving before processing next update
	// A default task that will run whenever no other task is active.
	// Useful for if this NPC should always just continuously do one task.
	Schedule defs.ScheduleDef

	defMgr *definitions.DefinitionManager
}

// IsActive checks if the npc is currently working on a task
func (tm TaskMGMT) IsActive() bool {
	if tm.CurrentTask == nil {
		return false
	}
	return !tm.CurrentTask.IsDone()
}

// Wait interrupts regular NPC updates for a certain duration
func (n *NPC) Wait(d time.Duration) {
	n.waitUntil = time.Now().Add(d)
}

func (n *NPC) WaitUntilNotMoving() {
	// remove any existing wait
	if n.waitUntil.After(time.Now()) {
		n.waitUntil = time.Now()
	}
	n.waitUntilDoneMoving = true
}
