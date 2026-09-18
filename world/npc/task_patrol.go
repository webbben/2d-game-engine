package npc

import (
	"fmt"
	"slices"
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/utils"
)

// confrontCooldown is how long after a confrontation session ends (dialog finished, or approach aborted)
// before the NPC will consider confronting the player again. Without this, a player standing next to the
// NPC would re-trigger the dialog the very tick after it ends, looping forever.
const confrontCooldown = 3 * time.Second

// patrolTaskState describes which part of the patrol task the NPC is in. All state is main-loop-owned.
type patrolTaskState int

const (
	// patrolState: cycling through the route waypoints.
	patrolState patrolTaskState = iota
	// approachState: walking over to the player to confront them.
	approachState
	// dialogState: the NPC's dialog profile is currently running against the player.
	dialogState
)

func (s patrolTaskState) String() string {
	switch s {
	case patrolState:
		return "patrol (0)"
	case approachState:
		return "approach (1)"
	case dialogState:
		return "dialog (2)"
	default:
		return "unregistered patrol state!"
	}
}

type PatrolTaskParams struct {
	ConfrontPlayer   bool // if true, the NPC will confront the player when they get too close
	ConfrontDistance int  // how close (in tiles) the player can get before the NPC confronts them; default 3
}

// PatrolTask is a deliberately dumb movement + player-proximity detector. The NPC cycles through the
// PATROL task areas in its start map (ordered by patrol_order, filtered by ownership), and when
// ConfrontPlayer is set it walks up to any player that gets too close and starts a dialog from the NPC's
// own dialog profile. It has no phases, no combat, and no timed state: the dialog profile owns all
// escalation logic (memory + combat effects). When the dialog ends, the patrol resumes.
type PatrolTask struct {
	TaskBase

	params PatrolTaskParams

	// waypoints are the PATROL task areas discovered in the active map, sorted by PatrolOrder.
	// waypointIndex points at the next waypoint to walk to, cycling on completion.
	waypoints           []*object.Object
	waypointIndex       int
	waypointsDiscovered bool

	// confrontation state: whether the NPC is patrolling, approaching the player, or in a dialog.
	state                 patrolTaskState
	confrontCooldownUntil time.Time // set when a dialog ends, so the NPC can't instantly re-trigger it
}

var _ Task = (*PatrolTask)(nil)

func NewPatrolTask(def defs.TaskDef, owner *NPC) *PatrolTask {
	if owner == nil {
		panic("owner was nil")
	}
	// nil params are allowed: they mean an all-default passive patrol.
	params := PatrolTaskParams{ConfrontDistance: 3}
	if def.Params != nil {
		patrolParams, ok := def.Params.(PatrolTaskParams)
		if !ok {
			logz.Println("PatrolTask", def.Params)
			logz.Panicln("PatrolTask", "tried to run a patrol task, but the params could not be converted into PatrolTaskParams. make sure you are using the right struct")
		}
		params = patrolParams
	}
	if params.ConfrontDistance <= 0 {
		params.ConfrontDistance = 3
	}
	return &PatrolTask{
		TaskBase: NewTaskBase(
			def,
			"Patrol",
			"Patrol an area, cycling through waypoints",
			owner,
		),
		params: params,
	}
}

func init() {
	registerTask(TaskPatrol, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			return NewPatrolTask(def, owner)
		},
		validateParams: func(def defs.TaskDef) error {
			if def.Params == nil {
				// nil params mean an all-default passive patrol; allowed.
				return nil
			}
			_, ok := def.Params.(PatrolTaskParams)
			if !ok {
				return fmt.Errorf("PatrolTask params must be PatrolTaskParams, got %T", def.Params)
			}
			return nil
		},
	})
}

func (t *PatrolTask) Update() {
	if t.IsDone() {
		return
	}
	t.Status = TaskInProg

	// route to the patrol map first. while routing, the child slot is owned by the RouteTask child; once
	// routing is done RouteToStartMap EndChild()s it, so our own children are always re-armed fresh below.
	if !t.RouteToStartMap(false) {
		return
	}

	// waypoint discovery needs the active map; if this map isn't active yet, wait for next tick.
	if !t.waypointsDiscovered {
		if !t.discoverWaypoints() {
			return
		}
	}

	switch t.state {
	case patrolState:
		t.handlePatrol()
	case approachState:
		t.handleApproach()
	case dialogState:
		t.handleDialog()
	default:
		logz.Panicln("PatrolTask", "unknown patrol state:", t.state)
	}
}

// handlePatrol keeps the NPC walking the route, and checks for an approaching player.
func (t *PatrolTask) handlePatrol() {
	if !t.ChildDone() {
		// an active waypoint goto is still underway; advance it.
		t.TaskBase.Update()
	} else {
		// no active child: either a waypoint goto just finished, or we just returned from a
		// confrontation. launch the next waypoint on the route.
		t.startNextWaypointGoto()
	}

	if t.params.ConfrontPlayer {
		t.maybeBeginConfront()
	}
}

// startNextWaypointGoto launches a goto child toward the next waypoint on the route, cycling back to the
// first after the last. The route always advances, even if this waypoint is skipped.
func (t *PatrolTask) startNextWaypointGoto() {
	wp := t.waypoints[t.waypointIndex]
	t.waypointIndex = (t.waypointIndex + 1) % len(t.waypoints)

	if wp.TilePos().Equals(t.Owner.Entity.TilePos()) {
		// already standing on this waypoint; GotoTask would panic, so just skip it.
		return
	}
	t.RunChild(NewGotoTask(
		GotoTaskParams{TileX: wp.TilePos().X, TileY: wp.TilePos().Y},
		t.Owner,
		defs.TaskDef{TaskID: TaskGoto, Priority: t.GetPriority()},
	))
	t.TaskBase.Update()
}

// maybeBeginConfront checks the player detection radius and, if the player got too close, moves from
// patrolling into the confront flow (approach the player, then start the NPC's dialog).
func (t *PatrolTask) maybeBeginConfront() {
	if !t.canConfrontPlayer() {
		return
	}
	playerPos := t.Owner.WorldCtx.GetPlayerPosition()
	dist := utils.EuclideanDistCoords(playerPos, t.Owner.Entity.TilePos())
	if dist > float64(t.params.ConfrontDistance) {
		return
	}

	if dist <= 2 {
		// close enough already; skip the approach and start the dialog directly.
		t.startDialogWithPlayer()
		return
	}

	// walk over to the player, then start the dialog once we arrive.
	t.cancelCurrentPath()
	t.RunChild(NewGotoTask(
		GotoTaskParams{TileX: playerPos.X, TileY: playerPos.Y},
		t.Owner,
		defs.TaskDef{TaskID: TaskGoto, Priority: t.GetPriority()},
	))
	t.TaskBase.Update()
	t.state = approachState
}

// handleApproach advances the approach goto child. If the player leaves the detection radius, the
// approach is cancelled and the NPC returns to patrolling. Once the approach reaches the player, the
// dialog starts.
func (t *PatrolTask) handleApproach() {
	if !t.InActiveMap() || !t.playerWithinDist(t.params.ConfrontDistance) {
		// the player got away; cancel the approach and go back to patrolling.
		t.cancelApproach()
		return
	}
	if !t.ChildDone() {
		t.TaskBase.Update()
		return
	}
	// reached where the player was standing; start the dialog.
	t.startDialogWithPlayer()
}

// handleDialog advances the dialog child until it ends, then returns to patrolling. The dialog profile
// owns all escalation (memory + combat effects); if a fight task is assigned mid-dialog at Emergency
// priority, the task manager preempts us, which is expected and needs no special handling here.
func (t *PatrolTask) handleDialog() {
	if !t.ChildDone() {
		t.TaskBase.Update()
		return
	}
	// the dialog ended; re-arm the route, but don't instantly re-confront the same spot.
	t.EndChild()
	t.state = patrolState
	t.confrontCooldownUntil = time.Now().Add(confrontCooldown)
}

// cancelApproach aborts an approach goto child and returns the task to patrolling.
func (t *PatrolTask) cancelApproach() {
	t.cancelCurrentPath()
	t.EndChild()
	t.state = patrolState
}

// startDialogWithPlayer launches the NPC's own dialog profile as a child task and enters the dialog
// state. The profile owns all escalation logic; this task only starts the dialog and waits for it to end.
func (t *PatrolTask) startDialogWithPlayer() {
	t.cancelCurrentPath()
	t.RunChild(NewStartDialogTask(
		StartDialogTaskParams{ProfileID: t.Owner.dialogProfileID},
		t.Owner,
		defs.TaskDef{TaskID: TaskStartDialog, Priority: t.GetPriority()},
	))
	t.state = dialogState
}

// canConfrontPlayer reports whether the confrontation check may run right now: we must be actively
// patrolling, in the active map, off the post-dialog cooldown, and have a dialog profile to start.
func (t *PatrolTask) canConfrontPlayer() bool {
	if t.state != patrolState {
		return false
	}
	if !t.InActiveMap() {
		return false
	}
	if time.Now().Before(t.confrontCooldownUntil) {
		return false
	}
	if t.Owner.dialogProfileID == "" {
		return false
	}
	return true
}

// playerWithinDist reports whether the player is within d tiles (euclidean) of the NPC, using the same
// distance helper as the player-sighting logic in update.go.
func (t *PatrolTask) playerWithinDist(d int) bool {
	return utils.EuclideanDistCoords(t.Owner.WorldCtx.GetPlayerPosition(), t.Owner.Entity.TilePos()) <= float64(d)
}

// cancelCurrentPath stops the entity from walking a stale patrol/approach path, if one exists.
func (t *PatrolTask) cancelCurrentPath() {
	if t.Owner.Entity.HasPath() {
		t.Owner.Entity.CancelCurrentPath()
	}
}

// discoverWaypoints finds the PATROL task areas this NPC may use in the active map, ordered by
// PatrolOrder. Returns false (without finding anything) if the NPC's map isn't the active map, since
// waypoints can only be read from the active map's objects.
func (t *PatrolTask) discoverWaypoints() bool {
	if !t.InActiveMap() {
		return false
	}
	t.waypoints = nil
	for _, obj := range t.Owner.ActiveMapCtx.GetAllObjects() {
		if obj.Type != object.TypeTaskArea {
			continue
		}
		if obj.TaskArea.TaskID != string(TaskPatrol) {
			continue
		}
		if !t.Owner.SatisfiesObjectOwnership(*obj) {
			continue
		}
		t.waypoints = append(t.waypoints, obj)
	}
	if len(t.waypoints) == 0 {
		logz.Println("PatrolTask", t.Owner.WhoAmI(), "map:", t.Owner.CharacterStateRef.CurrentMap)
		logz.Panicln("PatrolTask", "failed to find any PATROL task areas the NPC can use; at least one waypoint is required")
	}
	slices.SortStableFunc(t.waypoints, func(a, b *object.Object) int {
		return a.TaskArea.PatrolOrder - b.TaskArea.PatrolOrder
	})
	t.waypointIndex = 0
	t.waypointsDiscovered = true
	return true
}

// SetupActiveState resets the patrol so the next Update re-scans the map and resumes the route. It never
// forwards to a child: GotoTask/StartDialogTask/RouteTask SetupActiveState implementations panic, so the
// child slot is cleared here instead.
func (t *PatrolTask) SetupActiveState() {
	if !t.InStartMap() {
		// we must be mid-route to the start map; set up the routing child's active state instead.
		t.RouteToStartMapSetupActiveState()
		return
	}
	t.EndChild()
	t.waypoints = nil
	t.waypointIndex = 0
	t.waypointsDiscovered = false
	t.state = patrolState
	t.confrontCooldownUntil = time.Time{}
	t.Status = TaskInProg
}

// BackgroundAssist forwards to the active child (the RouteTask while routing, GotoTask otherwise) so its
// path can be computed on the background goroutine. Runs on the background goroutine; the child is read
// through the atomic child slot.
func (t *PatrolTask) BackgroundAssist() {
	t.TaskBase.BackgroundAssist()
}

// SimulationUpdate is a no-op: a patrol never moves an NPC between maps; the NPC stays in its start map.
func (t *PatrolTask) SimulationUpdate() {}

// Finish stops any stale movement before recording the result. Child cleanup (including unsubscribing a
// dialog child) is handled by TaskBase.Finish.
func (t *PatrolTask) Finish(result TaskResult) {
	t.cancelCurrentPath()
	t.TaskBase.Finish(result)
}