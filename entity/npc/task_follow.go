package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/model"
)

type FollowTask struct {
	TaskBase

	targetEntity    *entity.Entity
	targetPosition  model.Coords // the last calculated target position
	distance        int          // number of tiles behind the target entity to stand. default, 0, means the tile directly behind.
	isFollowing     bool
	recalculatePath bool // if set, indicates NPC should recalculate the path to the target

	restart bool
}

func (t FollowTask) ZzCompileCheck() {
	_ = append([]Task{}, &t)
}

func NewFollowTask(target *entity.Entity, distance int, owner *NPC, p defs.TaskPriority, nextTask *defs.TaskDef) FollowTask {
	if target == nil {
		panic("target is nil")
	}
	t := FollowTask{
		TaskBase: NewTaskBase(
			TaskFollow,
			"follow task",
			"NPC follows a target entity",
			owner,
			p,
			nextTask,
		),
		targetEntity: target,
		distance:     distance,
	}

	return t
}

func (t *FollowTask) End() {
	if len(t.Owner.Entity.Movement.TargetPath) > 0 {
		t.Owner.Entity.CancelCurrentPath()
	}

	t.Status = TaskEnded
}

func (t FollowTask) IsComplete() bool {
	return false
}

func (t FollowTask) IsFailure() bool {
	return false
}

func (t *FollowTask) Start() {
	if t.Owner == nil {
		panic("no owner set")
	}

	t.Status = TaskInProg

	if t.Owner.Entity.Movement.IsMoving {
		t.restart = true
		t.Owner.WaitUntilNotMoving()
		return
	}

	// when following an entity, the NPC tries to go directly behind it.
	// so, if the entity as at position x/y and facing 'R', then this NPC will try to go to x-1/y
	// Start: go to the position behind the target entity
	target := _followGetTargetPosition(*t.targetEntity, t.distance)
	if target.Equals(t.Owner.Entity.TilePos()) {
		// logz.Println(t.Owner.DisplayName, "already at target position")
		return
	}

	actualGoal, me := t.Owner.Entity.GoToPos(target, true)
	if !me.Success {
		// if GoToPos fails, that means we could not find any route at all to the goal (or even a partial route).
		// since the NPC should not be currently moving, that also means this should be cause by a real problem in the path finding.
		// wait a second and retry.
		logz.Println(t.Owner.DisplayName, "failed to find path to follow entity:", me)
		t.Owner.Wait(time.Second)
		t.restart = true
		return
	}
	t.targetPosition = actualGoal
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

	// TODO: we need to redo this task probably. commenting this part out just to avoid errors
	// t.HandleNPCCollision(t.Start)

	if t.restart {
		t.restart = false
		t.Start()
		return
	}

	// Check if we should recalculate a better path to the target entity
	if t.recalculatePath {
		logz.Println(t.Owner.DisplayName, "follow: NPC Manager told this NPC to recalculate path")
		t.Owner.Entity.CancelCurrentPath()
		t.Owner.Wait(time.Millisecond * 500)
		t.restart = true
		t.recalculatePath = false
		return
	}

	// ensure the NPC's target position matches the actual target position it is pursuing
	if len(t.Owner.Entity.Movement.TargetPath) > 0 {
		entTargetTile := t.Owner.Entity.Movement.TargetPath[len(t.Owner.Entity.Movement.TargetPath)-1]
		if !t.targetPosition.Equals(entTargetTile) {
			logz.Println(t.Owner.DisplayName, "follow: updated NPC target position since it didn't match the last position of target path")
			t.targetPosition = entTargetTile.Copy()
		}
	}

	// while following, there is no defined end goal. the NPC keeps trying to follow the entity until specifically told to stop.
	// so, check for when the NPC has reached its current goal and then try to recalculate a new one.
	if t.Owner.Entity.TilePos().Equals(t.targetPosition) {
		// we've met our first goal, now time to try and recalculate.
		t.Start()
		return
	}

	// if, for some reason, the entity loses its target path on accident (but isn't at goal yet), then recalculate new path
	if len(t.Owner.Entity.Movement.TargetPath) == 0 {
		logz.Println(t.Owner.DisplayName, "follow: NPC unexpected has empty path!")
		t.Start()
		return
	}

	// not at target position and target path is not empty... but not moving for some reason? seems like a bad state to me
	if !t.Owner.Entity.Movement.IsMoving {
		t.Owner.Wait(time.Second)
		t.restart = true
	}
}

// must NOT directly modify crucial state of NPC or entity (e.g. setting its position, movement, etc directly).
// can only suggest changes that will then be picked up in the normal game update loop and handled there.
func (t *FollowTask) BackgroundAssist() {
	if t.recalculatePath {
		return // previous flag set hasn't been acted upon yet
	}
	if len(t.Owner.Entity.Movement.TargetPath) < 3 {
		return
	}

	// calculate a new path
	start := t.Owner.Entity.Movement.TargetPath[2]
	goal := _followGetTargetPosition(*t.targetEntity, t.distance)
	if start.Equals(goal) {
		return
	}
	newPath, _ := t.Owner.Entity.World.FindPath(start, goal)
	if len(newPath) < 2 {
		return
	}

	t.Owner.Entity.Movement.SuggestedTargetPath = newPath
}
