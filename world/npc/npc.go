// Package npc defines NPC management logic
package npc

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/pubsub"
)

type WorldContext interface {
	// general map functions

	FindNPCAtPosition(c model.Coords) (NPC, bool)
	FindObjectsAtPosition(c model.Coords) []*object.Object
	GetValidMapPosition(n NPC) model.Coords
	IsTileCollision(c model.Coords) bool
	IsTileEntityCollision(c model.Coords, excludeEntID string) bool
	GetAllObjects() []*object.Object
	GetAllNPCs() []*NPC
	GetCostMap() [][]int

	// start screens and UI

	StartDialog(dialogProfileID defs.DialogProfileID, npcID string)
}

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

func (n NPC) WhoAmI() string {
	return fmt.Sprintf("%s [%s]", n.DisplayName(), n.ID())
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

func (n NPC) X() float64 {
	return n.Entity.X
}

type NPCParams struct {
	CharStateID state.CharacterStateID
}

// NewNPC instantiates an NPC for use in a game world or scenario.
//
// Only do this ONCE for a non-temporary NPC, in a given game session/world instantiation.
//
// Once this has been instantiated (for non-temp characters) it should live in the World struct and be reused from there.
// The reason you don't want to reinstantiate is that this function does things like subscribe to NPC events, so calling twice
// for a single character state ID would cause duplicate subscriptions, resulting in a panic.
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

// SetupStarterScheduleTask puts an NPC into its "active state" in a map.
// It is expected that the NPC is first officially "added to the map" which means it is part of the NPC list,
// and is placed at some position that is guaranteed to be open (like a spawn point).
// Then, this will handle actually moving the NPC to its correct place; wherever in the map it should be while doing its task.
func (n *NPC) SetupStarterScheduleTask(gameTime clock.GameTime, customStartLocation *defs.TaskStartLocation) {
	hour := gameTime.Hour
	scheduleTask := n.Schedule.Hourly[hour]
	if customStartLocation != nil {
		scheduleTask.StartLocation = customStartLocation
	}

	// TODO: I don't think we're handling Routing tasks properly as of now. that's okay, since that's a bit more advanced,
	// but I think we need to decide exactly how that should work. Currently, in RunTask, it checks if the next task is in a new map,
	// and if so, it tells the NPC to do a routing task first. This brings us to a question:
	// - Should Routing tasks be explicitly set in the schedule?
	// - Or, should they just be something that is automatically applied, when the next task requires moving to a new map?
	n.RunTask(scheduleTask, n)

	if n.CurrentTask == nil {
		if scheduleTask.TaskID == TaskDoNothing {
			// as expected, no task was assigned.
			return
		}
		panic("current task is nil")
	}

	// we also need to setup the "initial active state" of the task.
	// we want the NPC to already be "actively in progress" on their task.
	n.CurrentTask.SetupActiveState()
}

// gets the nearest open, unobstructed tile to the given position.
// useful for trying to place an NPC somewhere, but handles moving around objects, collisions, or other NPCs.
// tileDistLimit must not be 0.
func (n NPC) getNearestOpenTile(c model.Coords, tileDistLimit int, allowEntityCollision bool) (model.Coords, bool) {
	if tileDistLimit <= 0 {
		panic("tileDistLimit was <= 0")
	}

	if !n.World.IsTileCollision(c) {
		if allowEntityCollision || !n.World.IsTileEntityCollision(c, string(n.Entity.ID())) {
			return c, true // turns out the given position was open
		}
	}

	costMap := n.World.GetCostMap()
	if !allowEntityCollision {
		for _, n := range n.World.GetAllNPCs() {
			tilePos := n.Entity.TilePos()
			costMap[tilePos.Y][tilePos.X] += path_finding.BlockThreshold
		}
	}
	return path_finding.FindNearestOpenPosition(c, tileDistLimit, costMap)
}
