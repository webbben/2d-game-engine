package npc

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/utils"
)

const (
	// fleeSearchThreshold is how far (in tiles, manhattan) from the threat we bother looking for an
	// escape route at all. Past this the npc holds still rather than paying for a search it doesn't
	// need, and re-checks the distance once the repath interval expires.
	//
	// This doubles as the safety bound for the search radius handed to FleeFromPosition: we only ever
	// search when dist <= fleeSearchThreshold, and the radius is derived from the distance, so
	// BuildDistanceMap can never be asked to chart distances from a threat outside its own radius
	// (which it treats as a caller error and panics on).
	fleeSearchThreshold int = 10

	// fleeRepathInterval is how long to wait after acting on a search result before looking again. It
	// also paces retries while the npc is cornered.
	fleeRepathInterval = 2 * time.Second
)

// fleeRequest is the input to a background-assist escape search. The main loop fills it in, since only
// the main loop is allowed to read entity state.
type fleeRequest struct {
	Me, Them model.Coords
	// Closed gates this npc has no way to open, treated as walls for this search only. Collected by the
	// main loop, since object open/close state is main-loop state and would race if read here.
	BlockedTiles []model.Coords
}

// fleeResult is what background assist produced for a fleeRequest.
type fleeResult struct {
	Path []model.Coords
	// Reachable is false when the threat can't reach the npc at all (e.g. walled off), which means
	// there's nothing to run from.
	Reachable bool
	// CannotFlee is true when the threat can reach the npc but no route moves strictly further away.
	CannotFlee bool
}

var _ Task = (*FleeTask)(nil)

type FleeTask struct {
	TaskBase

	targetEntity *entity.Entity

	// pathRequest carries a pending search to the background goroutine and pathResult carries the
	// answer back. Both are atomic pointers, so a nil store is legal. Both follow the same contract as
	// FollowTask's mailboxes: the main loop writes, the background goroutine reads and writes back.
	// A request left unclaimed is simply dropped -- a search is cheap to redo and always better with
	// fresh positions than a half-stale one.
	pathRequest atomic.Pointer[fleeRequest]
	pathResult  atomic.Pointer[fleeResult]

	// awaitingPick means "a search has been requested and not yet answered". It is set only by
	// requestPick (which also posts the coordinates) and cleared only by the main loop when it claims a
	// result. Background assist deliberately leaves it alone, so a result that hasn't been picked up
	// yet can't be followed by a second overlapping search.
	awaitingPick atomic.Bool

	// nextRequestAt throttles how often we look for a new route, and paces retries while cornered.
	// Main loop only, so it doesn't need to be atomic.
	nextRequestAt time.Time

	// ownsPath records that we handed the entity a flee route and haven't cancelled it yet. Main loop
	// only. Cleanup on task end can't rely on a hook alone (see Finish), so we track this too.
	ownsPath bool
}

type FleeTaskParams struct {
	TargetEntity *entity.Entity // the entity to flee from
}

func init() {
	registerTask(TaskFlee, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			params, ok := def.Params.(FleeTaskParams)
			if !ok {
				logz.Println("FleeTask", def.Params)
				logz.Panicln("FleeTask", "tried to run a flee task, but the params could not be converted into FleeTaskParams. make sure you are using the right struct")
			}
			return NewFleeTask(params.TargetEntity, owner, def)
		},
		validateParams: func(def defs.TaskDef) error {
			params, ok := def.Params.(FleeTaskParams)
			if !ok {
				return fmt.Errorf("FleeTask params must be FleeTaskParams, got %T", def.Params)
			}
			if params.TargetEntity == nil {
				return fmt.Errorf("FleeTask params has a nil TargetEntity")
			}
			return nil
		},
	})
}

func NewFleeTask(targetEnt *entity.Entity, owner *NPC, def defs.TaskDef) *FleeTask {
	if def.TaskID != TaskFlee {
		logz.Panicln("NewFleeTask", "task def has wrong ID", string(def.TaskID))
	}
	if targetEnt == nil {
		logz.Panicln("NewFleeTask", "target was nil")
	}
	if owner == nil {
		logz.Panicln("NewFleeTask", "owner was nil")
	}
	return &FleeTask{
		TaskBase:     NewTaskBase(def, "Flee", "Flee from another entity", owner),
		targetEntity: targetEnt,
	}
}

func (t *FleeTask) Start() {
	t.Status = TaskInProg
	t.reset()
	// Deliberately no search request here. Update asks for one as soon as it finds no result and no
	// search in flight, which puts the distance check ahead of the first search. Hand-setting
	// awaitingPick instead would wedge BackgroundAssist, since it would swap a nil request and return.
}

// reset clears the per-run search state, for start and for re-entering the active map (where the world
// may have changed enough that the previous answer isn't worth trusting).
func (t *FleeTask) reset() {
	t.pathRequest.Store(nil)
	t.pathResult.Store(nil)
	t.awaitingPick.Store(false)
	t.nextRequestAt = time.Time{}
}

func (t *FleeTask) distFromThreat() int {
	return utils.ManhattanDistCoords(t.Owner.Entity.TilePos(), t.targetEntity.TilePos())
}

// requestPick asks background assist for a fresh escape route. Main loop only (it reads entity state).
func (t *FleeTask) requestPick() {
	t.pathRequest.Store(&fleeRequest{
		Me:           t.Owner.Entity.TilePos(),
		Them:         t.targetEntity.TilePos(),
		BlockedTiles: unopenableGateTiles(t.Owner),
	})
	t.awaitingPick.Store(true)
}

// cancelFleePath stops the npc walking a route we handed it. Needed because a task can be dropped
// without Finish ever running (RunScheduleTask -> clearTask), which would leave the npc walking an
// abandoned flee route under whatever task replaced it.
func (t *FleeTask) cancelFleePath() {
	if !t.ownsPath {
		return
	}
	if t.Owner.Entity.HasPath() {
		t.Owner.Entity.CancelCurrentPath()
	}
	t.Owner.Entity.Movement.SuggestedTargetPath = nil
	t.ownsPath = false
}

// applyResult acts on a completed search and schedules the next look.
func (t *FleeTask) applyResult(res *fleeResult) {
	if res.Reachable && !res.CannotFlee && len(res.Path) > 0 {
		t.Owner.Entity.SuggestPath(res.Path)
		t.ownsPath = true
	} else {
		// Cornered, walled off from the threat, or handed an empty route. Hold still and try again
		// after the interval -- the threat is likely to move, which may open an escape.
		t.cancelFleePath()
	}
	t.nextRequestAt = time.Now().Add(fleeRepathInterval)
}

func (t *FleeTask) Update() {
	if t.IsDone() {
		return
	}
	if t.targetEntity.IsDead() {
		t.FinishSuccess()
		return
	}

	// While stunned we can't act at all, so stop here. This has to come before the collision handling
	// below: a stun aborts movement via markInterrupted, which looks exactly like an obstacle
	// collision, and without this we'd misreport it as one and churn the path for no reason.
	if t.Owner.Entity.IsStunned() {
		return
	}

	// Gates are deliberately excluded from the cost map so npcs path through them and open them on
	// contact, so a route can still lead straight into one. Resolve that collision here, before the
	// recovery block below -- that block fires on exactly this state (holding a path, not moving) and
	// would otherwise cancel the path out from under us.
	//
	// HasPath is required because HandleNPCCollision panics if the movement was interrupted with no
	// next path step.
	if t.Owner.Entity.HasStoppedUnexpectedly() && t.Owner.Entity.HasPath() {
		result := t.HandleNPCCollision()
		if result.Wait {
			// a gate is opening, or already animating. Owner.Wait pauses npcUpdates, so we won't
			// re-enter here until it settles.
			return
		}
		if result.ReRoute || result.UnknownCollision {
			// we can't get through, or nothing at that tile explains the interruption. Drop the route
			// and let the repath interval below pick a new one. Deliberately leaving nextRequestAt
			// alone: that keeps re-routes on the same cadence as any other repath, so a permanent
			// obstruction can't turn into a per-frame request loop. Once a gate we've opened is no
			// longer collidable, the fresh search is free to route through it again.
			t.cancelFleePath()
			return
		}
		// NoneDetected can't happen behind the guard above; fall through to the recovery block rather
		// than assuming.
	}

	// Safety net for the case where we're holding a path but not moving, with no unresolved collision
	// above to explain it -- e.g. an interruption whose flag was already cleared by a later movement
	// start. Drop the stale route so a fresh suggestion can be adopted from where we actually are; a
	// bump-back can leave the old route's next step behind us. Same recovery FollowTask uses (see
	// task_follow.go), for the same reason: updateMovement can't adopt a suggestion while a path is
	// already set but the entity isn't moving.
	if t.ownsPath && t.Owner.Entity.HasPath() && !t.Owner.Entity.IsMoving() {
		t.cancelFleePath()
	}

	if time.Now().Before(t.nextRequestAt) {
		return
	}

	if res := t.pathResult.Swap(nil); res != nil {
		// main loop is the only party that clears this, so an unanswered request can never be
		// followed by a second overlapping search
		t.awaitingPick.Store(false)
		t.applyResult(res)
		return
	}

	if t.awaitingPick.Load() {
		// background assist is still working on it
		return
	}

	if t.distFromThreat() > fleeSearchThreshold {
		// far enough away that searching isn't worth it. Keep walking whatever route we already have
		// and re-check the distance once the interval expires.
		t.nextRequestAt = time.Now().Add(fleeRepathInterval)
		return
	}

	t.requestPick()
}

// BackgroundAssist runs on the background jobs goroutine, so it may only touch the atomic mailboxes.
// The positions it works from were captured by the main loop, and the cost map comes from the snapshot
// the main loop refreshes -- reading a live cost map would race with entity movement, since
// buildCostMap walks the NPC list without holding its mutex.
func (t *FleeTask) BackgroundAssist() {
	if !t.awaitingPick.Load() {
		return
	}
	req := t.pathRequest.Swap(nil)
	if req == nil {
		// Update asked for a pick but hasn't posted the coordinates yet; try again next pass.
		return
	}

	costMap := t.Owner.ActiveMapCtx.GetPathfindingSnapshot()
	// Gates we can't open are walls as far as this search is concerned, otherwise the BFS happily
	// routes straight through one and the npc stalls against it.
	if len(req.BlockedTiles) > 0 {
		costMap = costMapWithBlockedTiles(costMap, req.BlockedTiles)
	}
	// +1 keeps the threat inside the radius we ask about. Safe because requestPick is only reached
	// when dist <= fleeSearchThreshold, which bounds this.
	path, reachable, cannotFlee := path_finding.FleeFromPosition(req.Them, req.Me, fleeSearchThreshold, costMap)

	t.pathResult.Store(&fleeResult{Path: path, Reachable: reachable, CannotFlee: cannotFlee})
}

func (t *FleeTask) SetupActiveState() {
	// The player has entered (or re-entered) our map. We deliberately keep fleeing rather than ending
	// the task: if they stepped out and straight back in, the threat is still right there. But the
	// world may have shifted while we were away, so drop the old answer and re-decide from scratch.
	t.cancelFleePath()
	// reset() leaves awaitingPick false, so the next Update re-decides from scratch, distance check first
	t.reset()
}

func (t *FleeTask) Finish(result TaskResult) {
	// NOTE: this is the cleanup hook that actually runs. FollowTask has an End() method, but End() is
	// not part of the Task interface and nothing calls it -- TaskBase.Finish is what switchTask invokes
	// on preemption and what the decision loop runs on. See issue #182.
	//
	// It still isn't sufficient alone: RunScheduleTask -> clearTask drops a task without finishing it,
	// so cancelFleePath is also called from the branches that stop fleeing.
	t.cancelFleePath()
	t.pathRequest.Store(nil)
	t.pathResult.Store(nil)
	t.TaskBase.Finish(result)
}

func (t *FleeTask) SimulationUpdate() {
	// Intentionally does not end the task when the player leaves the active map. Nothing in the
	// framework closes active-map-only tasks on departure (see PrepareLeaveActiveMap / CloseMap), and
	// that matches what we want here: if the player steps out and straight back in, the npc should
	// still be running. The hourly schedule replaces this task at the next hour boundary via
	// OnHourChange -> RunScheduleTask.
	//
	// There's nothing useful to do while off-map regardless -- there's no cost map to search, and the
	// entity isn't being updated.
}

func (t *FleeTask) DisableDefaultSpeechBubbles() bool {
	// an npc mid-escape shouldn't be making small talk
	return true
}
