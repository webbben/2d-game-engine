package npc

import (
	"sync/atomic"
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/model"
)

type FollowTask struct {
	TaskBase

	targetEntity *entity.Entity
	distance     int // number of tiles behind the target entity to stand. default, 0, means the tile directly behind.

	// needPath is set by the main loop when a new path to the target is needed, and consumed by the
	// background jobs goroutine on each pathfinding attempt. Written by both, hence atomic.
	needPath atomic.Bool

	// lastGoal is the "behind the target" position the background goroutine last computed a path for.
	// Written only by the background goroutine; used to skip recomputes when the target hasn't moved.
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

	// pathfinding is owned by background assist now; this task's Update loop picks up a computed path
	// once the background goroutine produces one.
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

	if len(t.Owner.Entity.Movement.TargetPath) > 0 {
		// the entity is following a path. redirection when the target moves is handled by background
		// assist + the entity's suggestion merge in updateMovement, so there's nothing to do here.
		if !t.Owner.Entity.Movement.IsMoving {
			// the entity stopped unexpectedly while still having a path (e.g. a collision).
			// clear the stale path and request a fresh one, so a background suggestion can be adopted.
			t.Owner.Entity.CancelCurrentPath()
			t.requestPath()
		}
		return
	}

	// no active path; figure out what the entity should be doing.
	target := _followGetTargetPosition(*t.targetEntity, t.distance)
	if target.Equals(t.Owner.Entity.TilePos()) {
		// already standing behind the target; nothing to do until the target moves.
		return
	}
	if t.Owner.Entity.Movement.IsMoving {
		// still finishing the current move; re-evaluate next frame.
		return
	}

	// not at the target and not moving: request a path from background assist.
	t.requestPath()
}

// requestPath asks background assist to compute a new path, throttled to at most once per second so
// that a permanently unreachable goal doesn't trigger an A* search every background pass.
func (t *FollowTask) requestPath() {
	if time.Since(t.lastPathRequest) > time.Second {
		t.needPath.Store(true)
		t.lastPathRequest = time.Now()
	}
}

// BackgroundAssist computes a fresh path for the NPC when the target has moved or a new path was
// requested.
//
// NOTE: runs on the background jobs goroutine. must NOT directly modify crucial state of NPC or
// entity (e.g. setting its position, movement, etc directly). can only suggest changes that will
// then be picked up in the normal game update loop and handled there.
func (t *FollowTask) BackgroundAssist() {
	goal := _followGetTargetPosition(*t.targetEntity, t.distance)
	if !t.needPath.Load() && goal.Equals(t.lastGoal) {
		// target hasn't moved and no new path was requested; the current path is still valid.
		return
	}
	// consume the request. note that lastGoal is set even on failure, so a permanently blocked goal
	// doesn't trigger an A* search every background pass (the main loop re-requests via requestPath).
	t.needPath.Store(false)
	t.lastGoal = goal

	var start model.Coords
	switch {
	case len(t.Owner.Entity.Movement.TargetPath) >= 3:
		// mid-path: anchor the suggestion at a position ahead on the current path so it can be merged in
		start = t.Owner.Entity.Movement.TargetPath[2]
	case len(t.Owner.Entity.Movement.TargetPath) == 0 && !t.Owner.Entity.Movement.IsMoving:
		// idle: compute a fresh path from the entity's current tile
		start = t.Owner.Entity.TilePos()
	default:
		// too few tiles left on the current path to anchor a merge; let the entity finish and
		// re-request once it stops.
		return
	}
	if start.Equals(goal) {
		return
	}

	newPath, _ := path_finding.FindPath(start, goal, t.Owner.ActiveMapCtx.GetPathfindingSnapshot())
	if len(newPath) < 2 {
		return
	}

	t.Owner.Entity.Movement.SuggestedTargetPath = newPath
}

func (t *FollowTask) SimulationUpdate() {}

func (t *FollowTask) SetupActiveState() {
	panic("not yet implemented!")
}
