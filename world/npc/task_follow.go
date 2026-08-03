package npc

import (
	"sync/atomic"
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/model"
)

// followPathRequest carries the inputs for one background path computation. It is populated by the
// main game loop (which is the only goroutine allowed to read entity state) and consumed by the
// background jobs goroutine, which performs the actual A* search.
type followPathRequest struct {
	goal  model.Coords // the tile to walk toward (behind the target)
	start model.Coords // the anchor tile the path should begin from
}

type FollowTask struct {
	TaskBase

	targetEntity *entity.Entity
	distance     int // number of tiles behind the target entity to stand. default, 0, means the tile directly behind.

	// pathRequest carries a pending path computation from the main loop to the background goroutine.
	// Stores overwrite (the freshest request wins); the bg goroutine claims it with Swap.
	pathRequest atomic.Pointer[followPathRequest]

	// pathResult carries the computed path back from the background goroutine to the main loop, which
	// claims it with Swap and hands it to the entity's SuggestedTargetPath.
	pathResult atomic.Pointer[[]model.Coords]

	// lastGoal is the goal of the last path request, so the main loop can detect when the target moved
	// and request a redirect. Written only by the main loop.
	lastGoal model.Coords

	// lastPathRequest throttles how often the main loop re-requests a path when the goal is
	// unreachable (i.e. background assist keeps failing to find a path). Written only by the main loop.
	lastPathRequest time.Time
}

func (t *FollowTask) ZzCompileCheck() {
	_ = append([]Task{}, t)
}

func NewFollowTask(target *entity.Entity, distance int, owner *NPC, p defs.TaskPriority, nextTask *defs.TaskDef) *FollowTask {
	if target == nil {
		panic("target is nil")
	}
	t := defs.TaskDef{
		TaskID:   TaskFollow,
		Priority: p,
		NextTask: nextTask,
	}
	return &FollowTask{
		TaskBase: NewTaskBase(
			t,
			"follow task",
			"NPC follows a target entity",
			owner,
		),
		targetEntity: target,
		distance:     distance,
	}
}

func (t *FollowTask) End() {
	if len(t.Owner.Entity.Movement.TargetPath) > 0 {
		t.Owner.Entity.CancelCurrentPath()
	}
	// clear any pending request/result so an in-flight path computation can't leak into a later run.
	t.pathRequest.Store(nil)
	t.pathResult.Swap(nil)

	t.Status = TaskEnded
}

func (t *FollowTask) IsComplete() bool {
	return false
}

func (t *FollowTask) IsFailure() bool {
	return false
}

func (t *FollowTask) Start() {
	if t.Owner == nil {
		panic("no owner set")
	}

	t.Status = TaskInProg

	// reset per-run request state so a reused task requests a fresh path immediately.
	t.lastPathRequest = time.Time{}
	t.lastGoal = model.Coords{}
}

func _followGetTargetPosition(e entity.Entity, dist int) model.Coords {
	target := e.TilePos()
	switch e.Movement.Direction {
	case model.Directions.Left:
		target.X += dist + 1
	case model.Directions.Right:
		target.X -= dist + 1
	case model.Directions.Up:
		target.Y += dist + 1
	case model.Directions.Down:
		target.Y -= dist + 1
	default:
		panic("entity has invalid direction!")
	}
	return target
}

func (t *FollowTask) Update() {
	t.Status = TaskInProg

	// claim any path the background goroutine computed and hand it to the entity. this is the only
	// place SuggestedTargetPath gets written (main loop only).
	if p := t.pathResult.Swap(nil); p != nil {
		t.Owner.Entity.Movement.SuggestedTargetPath = *p
	}

	goal := _followGetTargetPosition(*t.targetEntity, t.distance)

	if len(t.Owner.Entity.Movement.TargetPath) > 0 {
		// the entity is following a path. redirection when the target moves is handled by requesting
		// a new path here and letting the entity's suggestion merge in updateMovement.
		if !t.Owner.Entity.Movement.IsMoving {
			// the entity stopped unexpectedly while still having a path (e.g. a collision).
			// clear the stale path and request a fresh one, so a background suggestion can be adopted.
			t.Owner.Entity.CancelCurrentPath()
			t.requestPath(goal, true)
			return
		}
		// mid-path: if the target moved, request a redirect (unthrottled, so tracking is responsive).
		if !goal.Equals(t.lastGoal) {
			t.requestPath(goal, false)
		}
		return
	}

	// no active path; figure out what the entity should be doing.
	if goal.Equals(t.Owner.Entity.TilePos()) {
		// already standing behind the target; nothing to do until the target moves.
		return
	}
	if t.Owner.Entity.Movement.IsMoving {
		// still finishing the current move; re-evaluate next frame.
		return
	}

	// not at the target and not moving: request a path from background assist.
	t.requestPath(goal, true)
}

// requestPath asks background assist to compute a path to goal. When throttle is set, the request is
// limited to once per second so that a permanently unreachable goal doesn't trigger an A* search
// every frame; mid-path redirects pass throttle=false for responsive tracking.
func (t *FollowTask) requestPath(goal model.Coords, throttle bool) {
	if throttle && time.Since(t.lastPathRequest) < time.Second {
		return
	}

	// anchor the suggestion: a position ahead on the current path (so it can be merged in), or the
	// current tile if idle. if the path is too short to anchor a merge, let the entity finish and
	// re-request once it stops.
	var start model.Coords
	switch {
	case len(t.Owner.Entity.Movement.TargetPath) >= 3:
		start = t.Owner.Entity.Movement.TargetPath[2]
	case len(t.Owner.Entity.Movement.TargetPath) == 0 && !t.Owner.Entity.Movement.IsMoving:
		start = t.Owner.Entity.TilePos()
	default:
		return
	}
	if start.Equals(goal) {
		return
	}

	t.lastPathRequest = time.Now()
	t.lastGoal = goal
	t.pathRequest.Store(&followPathRequest{goal: goal, start: start})
}

// BackgroundAssist runs on the background jobs goroutine and performs the A* search for a requested
// path. It never reads or writes entity/NPC state directly: it receives the goal and anchor tile from
// the main loop via pathRequest, and returns the computed path via pathResult, which the main loop
// picks up in Update.
func (t *FollowTask) BackgroundAssist() {
	req := t.pathRequest.Swap(nil)
	if req == nil {
		return
	}

	newPath, _ := path_finding.FindPath(req.start, req.goal, t.Owner.ActiveMapCtx.GetPathfindingSnapshot())
	if len(newPath) < 2 {
		return
	}

	t.pathResult.Store(&newPath)
}

func (t *FollowTask) SimulationUpdate() {}

func (t *FollowTask) SetupActiveState() {
	panic("not yet implemented!")
}
