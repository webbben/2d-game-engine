// Package npc defines NPC management logic
package npc

import (
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/combat"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/screen"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/worldgraph"
	"golang.org/x/image/font"
)

type WorldContext interface {
	GetActiveMapID() defs.MapID // this is here instead of activeMapContext so that NPCs can check it without being in the active world
	FindWorldPath(from, to defs.MapID) (pathToGoal worldgraph.WorldPath, foundPath bool)
	FindClosestMapType(fromMapID defs.MapID, mapType defs.MapType) (defs.MapID, bool)
	ChangeMapOccupancyEvent(charStateID id.CharacterStateID, from, to defs.MapID, toSpawn int)
	GetPlayerPosition() model.Coords
	GetCurrentGameTime() clock.GameTime

	// Combat. An NPC can't tell whether being attacked starts a new fight or pulls them into an
	// existing one, so it hands that decision to the World rather than tracking fights itself.
	StartCombat(initiator, target id.CharacterStateID, initiatorIntent, targetIntent defs.CombatIntent) combat.SessionView
	LeaveCombat(charStateID id.CharacterStateID)
	SessionFor(charStateID id.CharacterStateID) combat.SessionView
	IsInCombat(charStateID id.CharacterStateID) bool
	// CombatProfile returns a combatant's resolved profile within a session. The cache lives on the
	// session, so NPCs looking at the same character share one resolution. See World.CombatProfile.
	CombatProfile(sessionID combat.CombatSessionID, charStateID id.CharacterStateID) defs.CombatProfile
	EntityFor(charStateID id.CharacterStateID) *entity.Entity
	// IsFleeing asks whether a character is currently running from their enemies. Stance isn't
	// recorded anywhere in a session -- it's live state that changes tick to tick -- so anything that
	// needs it asks the World, which is also where the sessions get it from.
	IsFleeing(charID id.CharacterStateID) bool
	// WonCelebratableFight asks whether a fight this character just finished was theirs to celebrate.
	// Asked as they stand down, because that is the last moment the answer can be had: they are leaving
	// the session on their way out, and once they're gone nothing can say which side they were on.
	WonCelebratableFight(charID id.CharacterStateID) bool
}

type ActiveMapContext interface {
	GetMapInfo() defs.MapInfo
	GetOverlayManager() *overlay.OverlayManager
	RemoveNPCFromActiveMap(charStateID id.CharacterStateID, toMap defs.MapID)

	FindNPCAtPosition(c model.Coords) (NPC, bool)
	FindObjectsAtPosition(c model.Coords) []*object.Object
	GetValidMapPosition(n NPC) model.Coords
	IsTileCollision(c model.Coords) bool
	IsTileEntityCollision(c model.Coords, excludeEntID string) bool
	GetAllObjects() []*object.Object
	GetAllNPCs() []*NPC
	GetCostMap() [][]int // gets cost map (rebuilds it, so not entirely free)
	IsDialogActive() bool
	GetDialogNPC() id.CharacterStateID // empty if no dialog is active
	GetEntityInfo(charStateID id.CharacterStateID) entity.EntityInfo
	GetLightIntensity(pos model.Coords) float32
	IsLineOfSightBlocked(from, to model.Coords) bool

	StartDialog(dialogProfileID defs.DialogProfileID, npcID string)

	GetPathfindingSnapshot() [][]int
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

	visibleEntities               map[id.CharacterStateID]VisibleEntityInfo // maps which characters are visible
	initialPlayerSightingThisTick bool                                      // if true, the NPC just saw the player for the first time this update tick
	hasSeenPlayerYet              bool                                      // if true, this NPC has seen the player at some point already (in the current map)
	lastPlayerSightingTime        time.Time                                 // the last time the player was seen

	// lastContestedLog rate-limits waypoint contention reporting (see contention.go). The check runs
	// every frame for every active NPC, and a standoff persists for as long as it takes to resolve, so
	// an unthrottled log would bury everything else. Main-loop only.
	lastContestedLog time.Time

	// recentAttackers is who has hit this NPC lately, most recent first. FightTask reads it to decide
	// who to turn on -- see targetCandidate.AttackedMe.
	//
	// This lives on the NPC rather than on the task because a hit usually arrives before any FightTask
	// exists: OnAttacked runs whatever task the NPC happened to be doing (a schedule task, nothing), and
	// StartCombat only opens the session. The FightTask that responds to the hit is built afterwards.
	// Task-scoped state would be discarded at exactly the moment it matters, and the responding task
	// would never learn who struck first -- which is the bug this data exists to fix. Main-loop only.
	recentAttackers []recentAttacker

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

	// subscription IDs for events only listened to when NPC is in active map.
	// mapped booleans in this map are always expected to be true; to unsubscribe, the value must be removed from the map entirely.
	//
	// Note: avoid directly manipulating this, and instead use the subscribe and unsubscribe functions
	activeMapSubscriptionIDs map[string]bool

	// combat flags

	// surrenderOfferSpent is a one-way latch: once this npc has offered to surrender, it never will again,
	// whether the player accepted or turned it down. Set on accept too, not just rejection -- otherwise
	// an npc whose surrender was accepted could offer again if it later ended up fleeing.
	surrenderOfferSpent bool
	// set when the player accepts, and means this npc will no longer fight
	surrendered bool

	// victoryCelebration is a celebration this npc has earned and not yet got round to, or nil.
	//
	// Queued rather than started, because the fight it belongs to has only just ended and this npc may
	// still be finishing the swing that ended it. Starting a task over the top of that would leave the
	// attack queued with nothing to release it -- the same hazard FightTask.standDown cancels for. The
	// task loop starts it instead, once this npc is free. See TaskMGMT.Update.
	victoryCelebration *pendingCelebration
}

// pendingCelebration is a queued victory celebration and the deadlines around it.
type pendingCelebration struct {
	// notBefore is when this npc is allowed to start celebrating. Rolled at queue time and different
	// for every winner, so a crowd doesn't switch to the victory animation in lockstep.
	notBefore time.Time
	// queuedAt is when the fight ended. Used only to give up: see CelebrationTimeout.
	queuedAt time.Time
}

// QueueVictoryCelebration notes that this npc won a fight worth celebrating.
//
// It starts nothing; see victoryCelebration for why. Called by the world when a fight ends, which is a
// point where this npc is by definition still in the middle of something.
func (n *NPC) QueueVictoryCelebration() {
	notBefore := time.Now().Add(config.CelebrationDelay)
	if config.CelebrationDelayJitter > 0 {
		notBefore = notBefore.Add(time.Duration(rand.Int63n(int64(config.CelebrationDelayJitter))))
	}

	if n.victoryCelebration != nil {
		// Already owed one that hasn't happened. Keep whichever is due sooner rather than pushing this
		// one out behind it; two fights ending back to back shouldn't push the second celebration late.
		if notBefore.Before(n.victoryCelebration.notBefore) {
			n.victoryCelebration.notBefore = notBefore
		}
		return
	}

	n.victoryCelebration = &pendingCelebration{notBefore: notBefore, queuedAt: time.Now()}
	logz.Println(n.WhoAmI(), "queued a victory celebration, starting at", notBefore.Sub(time.Now()))
}

// hasPendingVictoryCelebration reports whether this npc is owed a celebration it hasn't started yet.
func (n *NPC) hasPendingVictoryCelebration() bool {
	return n.victoryCelebration != nil
}

// victoryCelebrationReady reports whether a queued celebration can start on this tick.
//
// False while waiting out this npc's stagger, so winners don't celebrate in unison. Also false mid-swing:
// a charged attack is released by the fight task rather than on a timer of its own, and a finish
// animation outlives the task that threw it, so an npc can be free while still finishing an attack.
//
// Drops the queue if it's gone stale, since it holds up the task loop.
func (n *NPC) victoryCelebrationReady() bool {
	if n.victoryCelebration == nil {
		return false
	}
	now := time.Now()
	if now.Sub(n.victoryCelebration.queuedAt) > config.CelebrationTimeout {
		logz.Println(n.WhoAmI(), "dropping a stale queued victory celebration")
		n.victoryCelebration = nil
		return false
	}
	if now.Before(n.victoryCelebration.notBefore) {
		return false
	}
	return !n.Entity.Body.IsAttacking()
}

// clearVictoryCelebration forgets a queued celebration, called once one has been started (or dropped).
func (n *NPC) clearVictoryCelebration() {
	n.victoryCelebration = nil
}

// HasSurrendered reports whether the player accepted this npc's surrender, meaning it won't fight again.
func (n *NPC) HasSurrendered() bool {
	return n.surrendered
}

// IsOfferingSurrender reports whether this npc has an offer on the table right now, which is what a
// dialog needs to know to show a surrender greeting. Distinct from HasSurrendered: this is the pending
// offer, that one is a deal already struck.
func (n *NPC) IsOfferingSurrender() bool {
	_, ok := n.getCurrentTask().(*SurrenderTask)
	return ok
}

// IsFleeing reports whether this npc is currently running from whoever they're fighting. Used by
// combat, which needs to know that a fleeing combatant is still in the fight but is often a lower
// priority target than one who's standing their ground.
func (n *NPC) IsFleeing() bool {
	_, ok := n.getCurrentTask().(*FleeTask)
	return ok
}

// MarkSurrendered records that the player accepted this npc's surrender, so it won't fight again and
// won't offer to surrender a second time.
func (n *NPC) MarkSurrendered() {
	n.surrendered = true
	n.surrenderOfferSpent = true
}

type VisibleEntityInfo struct {
	FirstSeen  time.Time
	Visibility float64
}

func (n *NPC) unsubscribe(subID string) {
	if subID == "" {
		panic("subID was empty")
	}
	n.eventBus.Unsubscribe(subID)

	if _, exists := n.activeMapSubscriptionIDs[subID]; !exists {
		logz.PanicCtx("NPC unsubscribe", "tried to unsubscribe from event, but subID doesn't exist in active map subscriptions", subID, n.WhoAmI())
	}
	delete(n.activeMapSubscriptionIDs, subID)
}

func (n *NPC) unsubscribeAll() {
	for subID := range n.activeMapSubscriptionIDs {
		n.eventBus.Unsubscribe(subID)
	}
	n.activeMapSubscriptionIDs = make(map[string]bool)
}

func (n *NPC) subscribeToEvent(subID string, eventType defs.EventType, subFn func(e defs.Event)) {
	if n.activeMapSubscriptionIDs[subID] {
		logz.PanicCtx("NPC subscribeToEvent", "NPC already has subID in subscription map", subID, n.WhoAmI())
	}
	if !n.inActiveMap() {
		logz.PanicCtx("NPC subscribeToEvent", "NPC is not in active map", subID, n.WhoAmI())
	}
	n.activeMapSubscriptionIDs[subID] = true
	n.eventBus.Subscribe(subID, eventType, subFn)
}

func (n *NPC) inActiveMap() bool {
	if n.WorldCtx == nil {
		logz.Panic("WorldCtx was nil")
	}
	return n.WorldCtx.GetActiveMapID() == n.CharacterStateRef.CurrentMap
}

// GetCurrentTaskForBgAssist returns the NPC's current task if set, for use by the
// background jobs goroutine. Safe for concurrent use: it reads the pointer under taskStateMu.
// Deliberately does NOT check completion status: task.Status is written by the main loop without
// synchronization, so reading it here would just create another race. The caller must only invoke
// BackgroundAssist on the returned task (never its other methods, which are main-loop-owned).
//
// This is the only task accessor that takes taskStateMu. Main-loop code uses getCurrentTask(), which
// is unlocked and is safe to call from inside a task's Start()/Finish() -- see the note there.
func (n *NPC) GetCurrentTaskForBgAssist() Task {
	n.TaskMGMT.taskStateMu.RLock()
	defer n.TaskMGMT.taskStateMu.RUnlock()
	return n.TaskMGMT.CurrentTask
}

// PrepareLeaveActiveMap does all things necessary to prepare an NPC to leave the active map; undoes entity active state, unsubscribes
// active map event subscriptions, etc.
func (n *NPC) PrepareLeaveActiveMap() {
	n.Entity.ResetActiveMapRuntimeState()
	n.unsubscribeAll()
	n.ActiveMapCtx = nil
}

func (n NPC) GetInfo() defs.NPCInfo {
	activateText := "Talk"
	if n.Entity.IsDead() {
		activateText = "Loot"
	}
	return defs.NPCInfo{
		CharID:       n.CharacterStateRef.ID,
		DisplayName:  n.DisplayName(),
		ActivateText: activateText,
		InCombat:     n.InCombat(),
	}
}

// InCombat reports whether this npc is currently involved in a fight, whatever they're doing about it.
//
// This used to mean "is running a FightTask", which wrongly reported false for an npc that was
// fleeing or offering to surrender -- both of which are still mid-fight. That was visible: the HUD
// offers to talk to any npc not in combat, so a surrendering guard showed as approachable.
func (n NPC) InCombat() bool {
	return n.WorldCtx.IsInCombat(n.CharacterStateRef.ID)
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
	logz.Println("NPC", "NPC", n.DisplayName(), "was activated")
	if n.Entity.IsDead() {
		n.eventBus.Publish(defs.Event{
			Type: pubsub.SysShowScreen,
			Data: map[string]any{
				"screen_id": config.LootNPCScreen,
				"params": screen.LootScreenParams{
					CharStateID: n.Entity.ID(),
					DisplayName: n.DisplayName(),
				},
			},
		})
		return
	}

	// a dialog is already open, so a click here would try to start a second one (which StartDialog
	// panics on). This can happen if the player clicks an npc mid-conversation.
	if n.ActiveMapCtx.IsDialogActive() {
		return
	}

	// an NPC can't be talked to while they are in combat
	if !n.surrendered {
		switch n.getCurrentTask().(type) {
		case *FightTask:
			return
		case *FleeTask:
			return
		}
	}

	if n.dialogProfileID != "" {
		n.ActiveMapCtx.StartDialog(n.dialogProfileID, n.ID())
		return
	}
	logz.Println(n.DisplayName(), "nothing happened on activation")
}

// Y is used for renderables sorting
func (n NPC) Y() float64 {
	return n.Entity.RenderOrderY()
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

	ent := entity.LoadCharacterStateIntoEntity(params.CharStateID, dataman, audioMgr, eventBus)

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
			taskStateMu: &sync.RWMutex{},
			Schedule:    scheduleDef,
			dataman:     dataman,
		},
		speechBubbleTileset:      params.SpeechBubbleTileset,
		speechBubbleOriginIndex:  params.SpeechBubbleOriginIndex,
		speechBubbleFont:         params.SpeechBubbleFont,
		activeMapSubscriptionIDs: make(map[string]bool),
		visibleEntities:          make(map[id.CharacterStateID]VisibleEntityInfo),
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
			logz.Println(n.ID(), "recieved assign task event, but data was missing expected key")
			logz.Panicln("NPC", "recieved assign task event, but data was missing expected key")
		}
		taskDef, ok := taskDefData.(defs.TaskDef)
		if !ok {
			logz.Println(n.ID(), "failed to type event data as taskDef")
			logz.Panicln("NPC", "failed to type event data as taskDef")
		}
		n.RunTask(taskDef, n)
	}
}

// TaskMGMT manages all task related stuff, such as schedules and whatnot.
type TaskMGMT struct {
	// protects CurrentTask, since it's used in both the main Update thread and bg assist
	taskStateMu *sync.RWMutex

	CurrentTask         Task
	waitUntil           time.Time
	waitUntilDoneMoving bool // if set, will wait until entity has stopped moving before processing next update
	// A default task that will run whenever no other task is active.
	// Useful for if this NPC should always just continuously do one task.
	Schedule defs.ScheduleDef

	// interruptedTask holds the (reconstructible) def of the task that was preempted by a higher-priority task.
	// When the current task finishes, if the NPC has no scheduled task for the current hour, this def is resumed
	// (re-created from the def) so the NPC returns to what it was doing before an interruption.
	interruptedTask *defs.TaskDef

	dataman *datamanager.DataManager
}

func (tm *TaskMGMT) ClearCurrentTask() {
	tm.clearTask()
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

// scheduledMapAt returns the mapID where the NPC is scheduled to be at the given hour, according to their schedule.
// Static tasks (a declared start location, UseHomeMap, sleep, do nothing) resolve directly. Dynamic tasks (no declared
// start location) resolve their start map via ResolveStartMap, anchored on the map the schedule placed the NPC in the
// previous hour. This makes placement schedule-authoritative, so NPCs end up where their schedule says even across a
// player timeLapse (e.g. an afternoon GO_TO_TAVERN following a harbour task resolves to the harbour tavern, not home).
func (n *NPC) scheduledMapAt(hour int) defs.MapID {
	if hour < 0 || hour > 23 {
		// wrapped before the start of day: everyone was asleep at home the prior day
		return n.CharacterStateRef.HomeMapID
	}
	scheduleTask := n.Schedule.Hourly[hour]
	if scheduleTask.TaskID == "" {
		logz.Println("scheduledMapAt", hour, n.WhoAmI())
		logz.Panicln("scheduledMapAt", "task at the given hour has no task ID")
	}
	if scheduleTask.StartLocation != nil && !scheduleTask.StartLocation.UseHomeMap {
		return scheduleTask.StartLocation.MapID
	}
	if scheduleTask.StartLocation != nil {
		// UseHomeMap
		return n.CharacterStateRef.HomeMapID
	}
	if scheduleTask.TaskID == TaskSleep || scheduleTask.TaskID == TaskDoNothing {
		return n.CharacterStateRef.HomeMapID
	}
	// dynamic task with no declared start location: ask the task to resolve its start map, anchored on where the
	// schedule had this NPC the previous hour.
	t := n.buildTask(scheduleTask, n)
	return t.ResolveStartMap(n.scheduledMapAt(hour - 1))
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

// SetupScheduledTaskForPlacement is the placement counterpart to SetupTaskState: it builds the task scheduled for the
// given hour, resolves the map that task places the NPC in (schedule-authoritative — dynamic tasks resolve against the
// prior hour's scheduled location as their anchor), runs that same built task as the NPC's current task, and returns
// the map the NPC should be placed in. Callers must clear the NPC's current task before calling. DO_NOTHING hours run
// no task and return the NPC's scheduled do-nothing location (home).
func (n *NPC) SetupScheduledTaskForPlacement(gameTime clock.GameTime) defs.MapID {
	hour := gameTime.Hour
	if hour < 0 || hour > 23 {
		logz.Println("SetupScheduledTaskForPlacement", hour)
		logz.Panicln("SetupScheduledTaskForPlacement", "hour was invalid")
	}
	if n.CurrentTask != nil {
		logz.Panic("called SetupScheduledTaskForPlacement, but NPC already has a task set. Make sure to clear the current task first.")
	}
	def := n.Schedule.Hourly[hour]
	if def.TaskID == "" {
		logz.Println("SetupScheduledTaskForPlacement", hour, n.WhoAmI())
		logz.Panicln("SetupScheduledTaskForPlacement", "scheduled task for the hour has no task ID")
	}

	if def.TaskID == TaskDoNothing {
		return n.scheduledMapAt(hour)
	}

	t := n.buildTask(def, n)
	startMap := t.ResolveStartMap(n.scheduledMapAt(hour - 1))
	if !n.switchTask(t) {
		logz.Println("SetupScheduledTaskForPlacement", n.WhoAmI())
		logz.Panicln("SetupScheduledTaskForPlacement", "failed to switch to scheduled task; current task was cleared.")
	}
	return startMap
}

// gets the nearest open, unobstructed tile to the given position.
// useful for trying to place an NPC somewhere, but handles moving around objects, collisions, or other NPCs.
// tileDistLimit must not be 0.
//
// args:
//   - c: position where we start the search from for an open tile
//   - tileDistLimit: the maximum (tile) distance we can search from c
//   - allowEntityCollision: if false, tiles where entities exist are considered blocked
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
			n.subscribeToEvent(subID, eventType, n.OnSpeechBubbleEvent)
			i++
		}
	}
}

func (n NPC) defaultSpeechBubbleParams() entity.SpeechBubbleParams {
	return entity.SpeechBubbleParams{
		Font:          n.speechBubbleFont,
		BoxTileset:    n.speechBubbleTileset,
		BoxTileOrigin: n.speechBubbleOriginIndex,
		Duration:      time.Second * 5,
	}
}

func (n *NPC) OnSpeechBubbleEvent(e defs.Event) {
	dialogProfileDef := n.dataman.GetDialogProfile(n.dialogProfileID)
	for _, speechBubbleReaction := range dialogProfileDef.SpeechBubbles {
		if slices.Contains(speechBubbleReaction.SubscribeEvents, e.Type) {
			reactionString := speechBubbleReaction.Reaction.Reaction(e, n.speechBubbleCtx)
			if reactionString != "" {
				n.Entity.ShowSpeechBubble(reactionString, n.defaultSpeechBubbleParams())
				return
			}
		}
	}
}

// OnAttacked is called when an attack actually lands on this npc.
//
// It notes who struck before anything else decides what to do about the hit, because the fight task
// that responds reads that record while starting up. See recentAttackers.go for why the record lives
// on the NPC rather than on the task.
func (n *NPC) OnAttacked(attackedBy *entity.Entity) {
	if attackedBy == nil {
		return
	}

	n.recordAttacker(attackedBy.ID())

	n.respondToCombatant(attackedBy)
}

// OnDrawnIntoCombat tells the npc they've been pulled into a fight without anything having landed on
// them -- a scripted world effect, or a quest starting one.
//
// The response is identical to being attacked, minus recording an attacker: nobody swung, so there is
// nobody to prioritise. Without that distinction an npc dragged into a fight would treat the other
// combatant as someone who had just struck them, and favour them over whoever it was actually fighting.
func (n *NPC) OnDrawnIntoCombat(into *entity.Entity) {
	if into == nil {
		return
	}

	n.respondToCombatant(into)
}

// respondToCombatant is the shared reaction to finding yourself in a fight, however it happened:
// record the session, then pick between fighting, fleeing, and doing neither.
//
// The param is the other combatant, who may or may not be an attacker -- OnAttacked and
// OnDrawnIntoCombat are the two callers, and only the first means somebody hit us.
func (n *NPC) respondToCombatant(opponent *entity.Entity) {
	n.WorldCtx.StartCombat(opponent.ID(), n.CharacterStateRef.ID, defs.IntentKill, defs.IntentSelfDefense)

	// attacking an NPC mid-surrender backs out of the offer, and they go to fleeing.
	// surrender cannot be offered a second time
	if _, ok := n.getCurrentTask().(*SurrenderTask); ok {
		n.surrenderOfferSpent = true
		// TODO: perhaps we should route to FightTask, and have FightTask handle starting the flee task. just so logic of "who flees and when" is completely consistent.
		n.RunTask(defs.TaskDef{
			TaskID:   TaskFlee,
			Priority: Emergency,
			Params:   FleeTaskParams{TargetEntity: opponent},
		}, n)
		return
	}

	// already surrendered, surrender was accepted, and then was betrayed and attacked again
	// TODO: at some point in the future we might need to track who the NPC surrendered to; I could see a scenario where multiple enemies are attackin
	// a single NPC, and this logic wouldn't be able to differentiate the person they surrendered to vs the person who attacked again.
	// go back to fleeing
	if n.surrendered {
		n.surrendered = false
		n.surrenderOfferSpent = true
		n.RunTask(defs.TaskDef{
			TaskID:   TaskFlee,
			Priority: Emergency,
			Params:   FleeTaskParams{TargetEntity: opponent},
		}, n)
		return
	}

	// If we're already fleeing, leave the flee task alone and let it handle this attacker.
	if _, ok := n.getCurrentTask().(*FleeTask); ok {
		return
	}

	// A fight task is the single decision point for "we're in a fight": it re-checks every tick
	// whether this npc still wants to be here, so an unarmed one runs and a coward breaks off before
	// landing a swing. All that's left to do here is start that task.
	//
	// Being unarmed is a reason to run no matter how hard someone is, while being outmatched is a
	// judgment the game data gets to make -- so try to arm first, and let the fight task ask about
	// bravery only once there's actually something to fight with. NeverFlee characters short-circuit
	// inside shouldFlee and always take the fighting branch.
	//
	// The target below is only a starting suggestion: the fight task re-derives its target from the
	// session every tick, so a second attacker doesn't force a switch -- and restarting the task
	// instead would discard its status, charge, and shield state partway through a swing.
	if ft, ok := n.getCurrentTask().(*FightTask); ok {
		if ft.targetEntity == nil {
			logz.PanicCtx("OnAttacked", "fight task has no target entity", n.ID())
		} else if ft.targetEntity.ID() != opponent.ID() {
			logz.Println("OnAttacked", "already fighting; new attacker noted.", n.ID(), opponent.ID())
		}
		return
	}

	n.RunTask(defs.TaskDef{
		TaskID:   TaskFight,
		Priority: Emergency,
		Params:   FightTaskParams{TargetEntity: opponent},
	}, n)
}

// ensureWeaponEquipped moves the first weapon found in this npc's inventory into its equipped slot.
// Returns true if the npc is holding a weapon when it finishes, whether it already was or not.
func (n *NPC) ensureWeaponEquipped() bool {
	cs := n.CharacterStateRef
	if cs == nil {
		return false
	}
	if cs.EquipedWeapon != nil && cs.EquipedWeapon.DefID != "" {
		return true
	}

	for _, invItem := range cs.InventoryItems {
		if invItem == nil || invItem.DefID == "" {
			continue
		}
		if n.dataman.GetItemDef(invItem.DefID).Type != defs.TypeWeapon {
			continue
		}

		// Go through RemoveItemFromStandardInventory rather than just assigning EquipedWeapon, so the
		// item comes out of the pack (and the coin purse, if it was stashed there) properly.
		if found, _ := item.RemoveItemFromStandardInventory(&cs.StandardInventory, *invItem, n.dataman); !found {
			continue
		}

		cs.EquipedWeapon = invItem
		// Re-sync the body immediately instead of waiting for the next Entity.Update(): task updates run
		// first, so anything checking IsWeaponEquiped() before then would otherwise see a body that
		// doesn't match its character state (which panics).
		n.Entity.SyncBodyToState()
		return true
	}
	return false
}

func (n *NPC) CanSeeEntity(charStateID id.CharacterStateID) (bool, VisibleEntityInfo) {
	info, seen := n.visibleEntities[charStateID]
	return seen, info
}

// shouldPursue asks the game's combat calculation whether this npc would keep fighting an enemy who has
// started running, rather than letting them go.
//
// otherStandingEnemies counts this npc's *other* enemies who are still standing, excluding enemyID --
// effectively "is there something else I'd rather be dealing with?" It deliberately excludes the enemy
// being asked about, since whether to chase them is the question, not part of the answer.
//
// Only ever called about a fleeing enemy. There's no pursuit decision for someone standing their
// ground, so the caller filters first rather than paying for profiles to ask a question with one
// obvious answer.
//
// Returns false when the npc is in no fight at all, which reads the same as declining to chase: either
// way there is nothing here to follow.
func (n *NPC) shouldPursue(sessionID combat.CombatSessionID, enemyID id.CharacterStateID, otherStandingEnemies int) bool {
	selfID := n.CharacterStateRef.ID
	enemyProfile := n.WorldCtx.CombatProfile(sessionID, enemyID)
	if !enemyProfile.Fleeing {
		return true
	}
	selfProfile := n.WorldCtx.CombatProfile(sessionID, selfID)
	return n.dataman.CombatSystemCalc.ShouldPursue(selfProfile, enemyProfile, otherStandingEnemies)
}

// countStandingEnemies returns how many of this npc's enemies are still on their feet, excluding
// excludeID. Used as the "is there something better to deal with first" input to pursuit: an enemy
// who's run doesn't pull an npc away from someone they're already fighting.
//
// Dead enemies are skipped rather than profiled, matching shouldFlee. A corpse is still a session member
// for as long as it lies on the map, and profiling one is both meaningless and rejected by CombatProfile.
// Skipping it here also keeps the tally honest: a body that was fleeing when it dropped is not standing,
// but it is also not a live enemy worth diverting attention away from.
func (n *NPC) countStandingEnemies(sessionID combat.CombatSessionID, excludeID id.CharacterStateID) int {
	selfID := n.CharacterStateRef.ID
	session := n.WorldCtx.SessionFor(selfID)
	if session == nil {
		return 0
	}

	count := 0
	for _, enemyID := range session.EnemiesOf(selfID) {
		if enemyID == excludeID {
			continue
		}
		if characterstate.IsDeadOrDying(enemyID, n.dataman) {
			continue
		}
		if !n.WorldCtx.CombatProfile(sessionID, enemyID).Fleeing {
			count++
		}
	}
	return count
}

// shouldFlee asks the game's combat calculation whether this npc wants to break off from the fight
// they're currently in.
//
// This is the single place that decision is made, so the tick loop and the moment an npc is first
// attacked can't drift apart in how they treat an enemy. Callers decide what to do about the answer.
//
// Willingness to flee at all is not part of the calculation: NeverFlee is authored data on the
// character's or class's def, and short-circuits here before the calculation is consulted at all.
//
// Returns false when the npc isn't in a fight, since there's nothing to flee from.
func (n *NPC) shouldFlee() bool {
	if characterstate.NeverFlees(n.CharacterStateRef.ID, n.dataman) {
		return false
	}
	if config.DisableFleeing {
		return false
	}

	session := n.WorldCtx.SessionFor(n.CharacterStateRef.ID)
	if session == nil {
		return false
	}

	selfID := n.CharacterStateRef.ID
	enemies := make([]defs.CombatProfile, 0, max(len(session.Sides())-1, 0))
	for _, enemyID := range session.EnemiesOf(selfID) {
		// Only count enemies that are genuinely still a threat. One can be unloaded from the map
		// without ever leaving the session, and a dead one is still present in the world as a corpse
		// for a while after the fight moved on. This deliberately matches the rules target selection
		// uses, so fleeing and fighting agree on who counts.
		enemy := n.WorldCtx.EntityFor(enemyID)
		if enemy == nil || enemy.IsDead() {
			continue
		}
		enemies = append(enemies, n.WorldCtx.CombatProfile(session.SessionID(), enemyID))
	}

	selfProfile := n.WorldCtx.CombatProfile(session.SessionID(), selfID)

	allies := make([]defs.CombatProfile, 0, len(session.Sides()))
	for _, allyID := range session.AlliesOf(selfID) {
		// Filter before building a profile, never after: CombatProfile panics on a corpse, and a
		// group fight's whole point is that allies die. The enemy loop above does the same in the same
		// order.
		ally := n.WorldCtx.EntityFor(allyID)
		if ally == nil || ally.IsDead() {
			continue
		}
		allies = append(allies, n.WorldCtx.CombatProfile(session.SessionID(), allyID))
	}

	return n.dataman.CombatSystemCalc.ShouldFlee(selfProfile, defs.CombatInfo{Enemies: enemies, Allies: allies})
}
