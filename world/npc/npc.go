// Package npc defines NPC management logic
package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/pubsub"
)

/*

NPC "State":
- current task, and task schedule

"Map NPC":
- a pointer to an entity
- carries out current task, with updates to entity body

So, how about...

CharacterState:
- has an "NPCState" section that tracks current tasks, and current task schedule.
- this NPC state is empty, of course, for the player

Map NPC remains what NPC is; it handles carrying out tasks in a map.

Now there's the problem of tasks:
- tasks run different when in a map than they do when not in a map.
- only certain specific tasks can run outside of a map, like traveling between maps around the world.

So, for all NPCs, we need to run their tasks. the difference is this:

- NPCs in a map are run directly in the update loop (for the most part)
- NPCs outside the map run their tasks in the background, and these are expected to be light weight tasks that only process every few seconds or so.
  Like, for example, updating an NPCs progress as he moves between different maps (where the player is not present).

What this means is, I think the World needs to keep track of which NPCs are in the world, and which ones aren't.

World:
- master list of all NPCs, possibly organized by map location
- list of NPCs currently in the game world, so they can be directly updated and use their entities.

*/

type NPC struct {
	// === Map-specific things ===

	debug debug
	// The entity is basically the NPC's "controller" for an in-game map. It is used to move the character around, check its vitals, etc.
	Entity *entity.Entity

	// comes from either an override set in character state, or from the character def.
	dialogProfileID defs.DialogProfileID

	World WorldContext

	// priority assigned to this NPC by the map it is added to. used for prioritizing which NPC moves first in a collision.
	Priority int

	// === World related things ===

	// The character state gives the NPC access to get data about the character's current state, as well as make changes to it.
	CharacterStateRef *state.CharacterState

	TaskMGMT

	// === Both? ===

	eventBus *pubsub.EventBus
}

func (n NPC) ID() string {
	return string(n.Entity.ID())
}

func (n NPC) DisplayName() string {
	return n.Entity.DisplayName()
}

func (n *NPC) Activate() {
	if n.dialogProfileID != "" {
		n.World.StartDialog(n.dialogProfileID, n.ID())
		return
	}
	logz.Println(n.DisplayName(), "nothing happened on activation")
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

func NewNPC(params NPCParams, dataman *datamanager.DataManager, audioMgr *audio.AudioManager, eventBus *pubsub.EventBus) *NPC {
	if params.CharStateID == "" {
		panic("CharStateID is empty")
	}
	if dataman == nil {
		panic("dataman was nil")
	}
	if audioMgr == nil {
		panic("audioMgr was nil")
	}
	if eventBus == nil {
		panic("eventBus was nil")
	}

	ent := entity.LoadCharacterStateIntoEntity(params.CharStateID, dataman, audioMgr)

	charState := dataman.GetCharacterState(params.CharStateID)
	charDef := dataman.GetCharacterDef(charState.DefID)

	// get dialog profile and schedule

	scheduleID := charDef.ScheduleID
	if charState.OverrideScheduleID != "" {
		scheduleID = charState.OverrideScheduleID
	}
	scheduleDef := dataman.GetScheduleDef(scheduleID)

	dialogProfileID := charDef.DialogProfileID
	if charState.OverrideDialogProfileID != "" {
		dialogProfileID = charState.OverrideDialogProfileID
	}

	n := NPC{
		eventBus:          eventBus,
		Entity:            ent,
		CharacterStateRef: charState,
		dialogProfileID:   dialogProfileID,
		TaskMGMT: TaskMGMT{
			Schedule: scheduleDef,
			dataman:  dataman,
		},
	}

	n.eventBus.SubscribeToNPCEvents(n.ID(), n.ID(), n.OnEvent)

	return &n
}

func (n *NPC) OnEvent(e defs.Event) {
	logz.Println(n.ID(), "received event:", e)
	assignTask := pubsub.NpcAssignTaskType(n.ID())
	switch e.Type {
	case assignTask:
		taskDefData, exists := e.Data["taskDef"]
		if !exists {
			logz.Panicln(n.ID(), "recieved assign task event, but data was missing expected key")
		}
		taskDef, ok := taskDefData.(defs.TaskDef)
		if !ok {
			logz.Panicln(n.ID(), "failed to type event data as taskDef")
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

	dataman *datamanager.DataManager
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
