// Package npc defines NPC management logic
package npc

import (
	"fmt"
	"slices"
	"time"

	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/worldgraph"
	"golang.org/x/image/font"
)

type WorldContext interface {
	GetActiveMapID() defs.MapID // this is here instead of activeMapContext so that NPCs can check it without being in the active world
	FindWorldPath(from, to defs.MapID) (pathToGoal worldgraph.WorldPath, foundPath bool)
	ChangeMapOccupancyEvent(charStateID id.CharacterStateID, from, to defs.MapID, toSpawn int)
}

type ActiveMapContext interface {
	GetOverlayManager() *overlay.OverlayManager
	RemoveNPCFromActiveMap(charStateID id.CharacterStateID, toMap defs.MapID)

	FindNPCAtPosition(c model.Coords) (NPC, bool)
	FindObjectsAtPosition(c model.Coords) []*object.Object
	GetValidMapPosition(n NPC) model.Coords
	IsTileCollision(c model.Coords) bool
	IsTileEntityCollision(c model.Coords, excludeEntID string) bool
	GetAllObjects() []*object.Object
	GetAllNPCs() []*NPC
	GetCostMap() [][]int

	StartDialog(dialogProfileID defs.DialogProfileID, npcID string)
}

type NPC struct {
	// === Map-specific things ===

	debug debug
	// The entity is basically the NPC's "controller" for an in-game map. It is used to move the character around, check its vitals, etc.
	Entity *entity.Entity

	// comes from either an override set in character state, or from the character def.
	dialogProfileID defs.DialogProfileID

	ActiveMapCtx ActiveMapContext

	WorldCtx WorldContext

	// priority assigned to this NPC by the map it is added to. used for prioritizing which NPC moves first in a collision.
	Priority int

	// === World related things ===

	// The character state gives the NPC access to get data about the character's current state, as well as make changes to it.
	CharacterStateRef *state.CharacterState

	TaskMGMT

	// === Both? ===

	eventBus *pubsub.EventBus

	speechBubbleCtx         defs.SpeechBubbleContext
	speechBubbleTileset     string
	speechBubbleOriginIndex int
	speechBubbleFont        font.Face

	activeMapSubscriptionIDs map[string]bool // subscription IDs for events only listened to when NPC is in active map
}

// PrepareLeaveActiveMap does all things necessary to prepare an NPC to leave the active map; undoes entity active state, unsubscribes
// active map event subscriptions, etc.
func (n *NPC) PrepareLeaveActiveMap() {
	n.Entity.ResetActiveMapRuntimeState()

	for subID := range n.activeMapSubscriptionIDs {
		n.eventBus.Unsubscribe(subID)
	}

	n.activeMapSubscriptionIDs = make(map[string]bool)
}

func (n NPC) GetInfo() defs.NPCInfo {
	return defs.NPCInfo{
		CharID:       n.CharacterStateRef.ID,
		DisplayName:  n.DisplayName(),
		ActivateText: "Talk",
	}
}

func (n NPC) IsHovering(x, y int) bool {
	return n.Entity.GetDrawRect().Within(x, y)
}

func (n NPC) WhoAmI() string {
	return fmt.Sprintf("%s [%s] (currentMap: %s)", n.DisplayName(), n.ID(), n.CharacterStateRef.CurrentMap)
}

func (n NPC) ID() string {
	return string(n.Entity.ID())
}

func (n NPC) DisplayName() string {
	return n.Entity.DisplayName()
}

func (n *NPC) Activate() {
	if n.dialogProfileID != "" {
		n.ActiveMapCtx.StartDialog(n.dialogProfileID, n.ID())
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
	CharStateID             id.CharacterStateID
	SpeechBubbleTileset     string
	SpeechBubbleOriginIndex int
	SpeechBubbleFont        font.Face
}

// NewNPC instantiates an NPC for use in a game world or scenario.
//
// Only do this ONCE for a non-temporary NPC, in a given game session/world instantiation.
//
// Once this has been instantiated (for non-temp characters) it should live in the World struct and be reused from there.
// The reason you don't want to reinstantiate is that this function does things like subscribe to NPC events, so calling twice
// for a single character state ID would cause duplicate subscriptions, resulting in a panic.
func NewNPC(params NPCParams, dataman *datamanager.DataManager, audioMgr *audio.AudioManager, eventBus *pubsub.EventBus, worldCtx WorldContext) *NPC {
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
		WorldCtx:          worldCtx,
		eventBus:          eventBus,
		Entity:            ent,
		CharacterStateRef: charState,
		dialogProfileID:   dialogProfileID,
		TaskMGMT: TaskMGMT{
			Schedule: scheduleDef,
			dataman:  dataman,
		},
		speechBubbleTileset:      params.SpeechBubbleTileset,
		speechBubbleOriginIndex:  params.SpeechBubbleOriginIndex,
		speechBubbleFont:         params.SpeechBubbleFont,
		activeMapSubscriptionIDs: make(map[string]bool),
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

// GetScheduledMap returns the mapID where the NPC is supposed to be, according to their schedule
func (n *NPC) GetScheduledMap(gameTime clock.GameTime) defs.MapID {
	hour := gameTime.Hour
	if hour < 0 || hour > 23 {
		logz.Panicln("GetScheduledMap", "hour was invalid:", hour)
	}
	scheduleTask := n.Schedule.Hourly[hour]
	if scheduleTask.TaskID == "" {
		logz.Panicln("GetScheduledMap", "task at the given hour has no task ID! hour:", gameTime.Hour, n.WhoAmI())
	}
	if scheduleTask.StartLocation == nil {
		if scheduleTask.TaskID == TaskSleep {
			// sleep task with no set start location => home bed location
			return n.CharacterStateRef.HomeMapID
		}
		logz.Warnln("GetScheduledMap", "NPC scheduled task doesn't have start map. taskID:", scheduleTask.TaskID)
		return ""
	}
	return scheduleTask.StartLocation.MapID
}

// SetupTaskState is for initializing a task for an NPC based on their schedule and the given hour.
// It does not require the NPC to be in the active map, and can be used for setting up tasks for NPC's in the simulation loop too.
// To actually prepare the "active map" state of a task, use task.SetupActiveState function.
//
// It's expected that current task is nil if this is called, and will panic if not; make sure to clear an existing task before calling this.
func (n *NPC) SetupTaskState(gameTime clock.GameTime, customStartLocation *defs.TaskStartLocation) {
	if n.CurrentTask != nil {
		logz.Panic("called SetupTaskState, but NPC already has a task set. Make sure to clear the current task if you really want to call this.")
	}

	// setup the task scheduled for the given hour
	hour := gameTime.Hour
	scheduleTask := n.Schedule.Hourly[hour]
	if customStartLocation != nil {
		scheduleTask.StartLocation = customStartLocation
	}
	logz.Println("SetupTaskState", "setting up schedule task:", scheduleTask.TaskID, n.WhoAmI())

	// Note: Routing is handled in individual task logic (through TaskBase, unless a task decides to someday have special handling for some reason)
	n.RunTask(scheduleTask, n)

	if n.CurrentTask == nil {
		if scheduleTask.TaskID == TaskDoNothing {
			// as expected, no task was assigned.
			return
		}
		panic("current task is nil")
	}
}

// gets the nearest open, unobstructed tile to the given position.
// useful for trying to place an NPC somewhere, but handles moving around objects, collisions, or other NPCs.
// tileDistLimit must not be 0.
func (n NPC) getNearestOpenTile(c model.Coords, tileDistLimit int, allowEntityCollision bool) (model.Coords, bool) {
	if tileDistLimit <= 0 {
		panic("tileDistLimit was <= 0")
	}

	if !n.ActiveMapCtx.IsTileCollision(c) {
		if allowEntityCollision || !n.ActiveMapCtx.IsTileEntityCollision(c, string(n.Entity.ID())) {
			return c, true // turns out the given position was open
		}
	}

	costMap := n.ActiveMapCtx.GetCostMap()
	if !allowEntityCollision {
		for _, n := range n.ActiveMapCtx.GetAllNPCs() {
			tilePos := n.Entity.TilePos()
			costMap[tilePos.Y][tilePos.X] += path_finding.BlockThreshold
		}
	}
	return path_finding.FindNearestOpenPosition(c, tileDistLimit, costMap)
}

func (n *NPC) SetupSpeechBubbleReactions(speechBubbleCtx defs.SpeechBubbleContext) {
	if n.eventBus == nil {
		logz.Panic("event bus was nil")
	}
	if speechBubbleCtx == nil {
		logz.Panic("speech bubble ctx is nil")
	}

	n.speechBubbleCtx = speechBubbleCtx

	dialogProfileDef := n.dataman.GetDialogProfile(n.dialogProfileID)

	i := 0
	for _, speechBubbleReaction := range dialogProfileDef.SpeechBubbles {
		for _, eventType := range speechBubbleReaction.SubscribeEvents {
			subID := fmt.Sprintf("%s_speech_bubble_reaction_%v", n.ID(), i)
			n.eventBus.Subscribe(subID, eventType, n.OnSpeechBubbleEvent)
			if n.activeMapSubscriptionIDs[subID] {
				logz.Panicln("NPC", "subscription is already mapped?", subID)
			}
			n.activeMapSubscriptionIDs[subID] = true
			i++
		}
	}
}

func (n *NPC) OnSpeechBubbleEvent(e defs.Event) {
	dialogProfileDef := n.dataman.GetDialogProfile(n.dialogProfileID)
	for _, speechBubbleReaction := range dialogProfileDef.SpeechBubbles {
		if slices.Contains(speechBubbleReaction.SubscribeEvents, e.Type) {
			reactionString := speechBubbleReaction.Reaction.Reaction(e, n.speechBubbleCtx)
			if reactionString != "" {
				n.Entity.ShowSpeechBubble(reactionString, entity.SpeechBubbleParams{
					Font:          n.speechBubbleFont,
					BoxTileset:    n.speechBubbleTileset,
					BoxTileOrigin: n.speechBubbleOriginIndex,
					Duration:      time.Second * 5,
				})
				return
			}
		}
	}
}
