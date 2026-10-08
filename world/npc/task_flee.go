package npc

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/utils"
)

const (
	// fleeSearchThreshold is how far (in tiles, manhattan) from the nearest threat we bother looking for
	// an escape route at all. Past this the npc holds still rather than paying for a search it doesn't
	// need, and re-checks the distance once the repath interval expires.
	//
	// This also bounds the search area handed to FleeFromPositions, in two ways. It's the floor for the
	// radius, so a group of close threats still gets a full-size search. And collectThreats drops any
	// enemy past this distance, so the radius can never be pushed arbitrarily far out by a distant
	// enemy -- without that, one enemy on the other side of the map would grow the BFS for nothing, and
	// BuildDistanceMap panics if asked to chart from a threat outside the radius it's given.
	//
	// Judged on the *nearest* threat, not all of them: a distant extra enemy doesn't make the situation
	// more urgent, and it can't change the route either since it isn't considered one.
	fleeSearchThreshold int = 10

	// fleeRepathInterval is how long to wait after acting on a search result before looking again. It
	// also paces retries while the npc is cornered.
	fleeRepathInterval = 2 * time.Second
)

// fleeRequest is the input to a background-assist escape search. The main loop fills it in, since only
// the main loop is allowed to read entity state.
type fleeRequest struct {
	Me model.Coords
	// Them is every position worth running from, not just one. The search treats them as a single
	// multi-source distance map, so it optimizes distance from the *nearest* of them -- which is the
	// correct shape for escaping a group. Running one search per enemy would produce a route that
	// dodges the last bandit while walking straight into the first. See issue #191.
	Them []model.Coords
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

	// targetEntity is who we started running from. It is NOT what the route is built around -- that
	// comes from the whole threat list, see collectThreats. It's kept because it's the entity this task
	// was constructed with, so it survives to answer "who started this?" and to stand in for the
	// nearest threat in the rare case where no threat can be resolved. Prefer collectThreats for
	// anything that needs an enemy.
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
	// TargetEntity is whoever provoked the escape. It seeds the task and answers "who started this?",
	// but it does not bound what we run from: every enemy in the fight is considered, and the route is
	// chosen against all of them at once. See FleeTask.collectThreats.
	TargetEntity *entity.Entity
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

// distFromNearestThreat returns how far away the closest enemy we're running from is, or
// fleeSearchThreshold+1 when there is no such enemy.
//
// That sentinel sits one past the threshold so the "too far to bother searching" comparison in Update
// reports true without needing a separate empty-threat case at the call site. A threat standing right
// on top of us has distance 0, so the loop can't use a zero-valued best to detect its first iteration.
func (t *FleeTask) distFromNearestThreat() int {
	me := t.Owner.Entity.TilePos()
	best, found := 0, false

	for _, enemy := range t.threatsWithinReach() {
		dist := utils.ManhattanDistCoords(me, enemy.TilePos())
		if !found || dist < best {
			best, found = dist, true
		}
	}
	if !found {
		return fleeSearchThreshold + 1
	}
	return best
}

// requestPick asks background assist for a fresh escape route. Main loop only (it reads entity state).
func (t *FleeTask) requestPick() {
	threats, _ := t.collectThreats()
	t.pathRequest.Store(&fleeRequest{
		Me:           t.Owner.Entity.TilePos(),
		Them:         threats,
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

// threatsWithinReach returns every enemy we're still running from: in the fight, alive, present on the
// map, and not fleeing themselves.
//
// A nil result covers both "we're not in a fight any more" and "everyone left is already beaten or
// running away themselves", and both mean the same thing to the caller: stop fleeing.
//
// This is the single definition of "a real enemy" for fleeing, and it deliberately matches the rules
// target selection and the flee willingness check apply. Presence has to be checked rather than assumed
// from the session, because an enemy can be unloaded from the map without ever leaving the fight.
//
// Deliberately not distance-limited. An enemy too far away to route around is still someone closing in
// on us, and treating them as gone would end the task while they're still a threat.
func (t *FleeTask) threatsWithinReach() []*entity.Entity {
	selfID := t.Owner.CharacterStateRef.ID
	session := t.Owner.WorldCtx.SessionFor(selfID)
	if session == nil {
		return nil
	}

	var found []*entity.Entity
	for _, enemyID := range session.EnemiesOf(selfID) {
		enemy := t.Owner.WorldCtx.EntityFor(enemyID)
		if enemy == nil || enemy.IsDead() {
			continue
		}
		if t.Owner.WorldCtx.IsFleeing(enemyID) {
			continue
		}
		found = append(found, enemy)
	}
	return found
}

// collectThreats returns the positions worth routing around, and identifies the nearest threat.
//
// Same set as threatsWithinReach, minus anyone past fleeSearchThreshold. There's no reason to route
// around an enemy we can't feel yet, and including one would inflate the search radius to cover them,
// growing the BFS for no benefit. It also means the radius is bounded no matter how many enemies are
// involved, which is what keeps the search cheap in a crowded fight.
//
// Returning nearest separately is for the consumers that genuinely need one specific enemy -- the
// surrender handoff. It isn't used for pathing, and it's derived here rather than stored as a field so
// nothing can mistake it for the route's basis.
func (t *FleeTask) collectThreats() (threats []model.Coords, nearest *entity.Entity) {
	me := t.Owner.Entity.TilePos()
	bestDist := 0

	for _, enemy := range t.threatsWithinReach() {
		pos := enemy.TilePos()
		dist := utils.ManhattanDistCoords(me, pos)
		if dist > fleeSearchThreshold {
			continue
		}
		threats = append(threats, pos)
		// nearest == nil rather than bestDist == 0 guards the comparison: a threat standing on top of
		// us has distance 0 and must still be selectable as the nearest
		if nearest == nil || dist < bestDist {
			nearest = enemy
			bestDist = dist
		}
	}

	return threats, nearest
}

// stillWorthFleeing reports whether there's any reason left to keep running.
//
// Running from someone who's already beaten -- dead, out of the fight, or broken and running themselves
// -- isn't escape, it's just running away from nothing, and it keeps the npc from ever getting back to
// whatever they were doing. So this is a question about the whole fight, not about one enemy: being
// chased by a second bandit is very much still worth fleeing, so it keeps running as long as any
// enemy qualifies.
func (t *FleeTask) stillWorthFleeing() bool {
	return len(t.threatsWithinReach()) > 0
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
	if !t.stillWorthFleeing() {
		t.FinishSuccess()
		return
	}

	// Last legs: try to end this by giving ourselves up rather than keep running. One-shot -- once
	// offered, we've spent it whether the player accepted or not.
	if !t.Owner.surrenderOfferSpent {
		if cs := t.Owner.CharacterStateRef; float64(cs.Health) <= float64(cs.MaxHealth)*config.NearDeathHealthPercent {
			// Surrender names a single person, so hand over whoever's closest. stillWorthFleeing already
			// guaranteed there's at least one threat, so this can't come back empty in practice.
			_, nearest := t.collectThreats()
			if nearest == nil {
				nearest = t.targetEntity
			}
			t.Owner.RunTask(defs.TaskDef{
				TaskID:   TaskSurrender,
				Priority: Emergency,
				Params:   SurrenderTaskParams{TargetEntity: nearest},
			}, t.Owner)
			// We don't want to be resumed if preempted later; the npc has moved on. Must come after
			// RunTask, which is what records us as the interrupted task.
			t.Owner.clearInterruptedTask()
			return
		}
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

	// Judged on the *nearest* threat, since that's the one whose distance decides whether running is
	// currently urgent. A distant extra enemy doesn't make the situation more pressing, and it can't
	// affect the route either -- collectThreats leaves those out.
	if t.distFromNearestThreat() > fleeSearchThreshold {
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

	// The radius must cover every threat we're asking about, since BuildDistanceMap panics on any
	// single `from` outside it. The +1 keeps each threat itself inside the radius rather than on its
	// edge, and we start from fleeSearchThreshold so a group that's all close together still gets the
	// full search area rather than collapsing to a tiny circle around the nearest one.
	//
	// collectThreats already drops anything past fleeSearchThreshold, so this is bounded by
	// threshold+1 and can't grow with the number of enemies.
	searchRadius := fleeSearchThreshold
	for _, them := range req.Them {
		searchRadius = max(searchRadius, utils.ManhattanDistCoords(req.Me, them)+1)
	}

	costMap := t.Owner.ActiveMapCtx.GetPathfindingSnapshot()
	// Gates we can't open are walls as far as this search is concerned, otherwise the BFS happily
	// routes straight through one and the npc stalls against it.
	if len(req.BlockedTiles) > 0 {
		costMap = costMapWithBlockedTiles(costMap, req.BlockedTiles)
	}
	path, reachable, cannotFlee := path_finding.FleeFromPositions(req.Them, req.Me, searchRadius, costMap)

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
