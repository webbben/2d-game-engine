package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

func (t *Task) startFollow() {
	// GoToPos will fail if NPC is already moving
	if t.Owner.Entity.Movement.IsMoving {
		t.Owner.WaitUntilNotMoving()
		return
	}

	// when following an entity, the NPC tries to go directly behind it.
	// so, if the entity as at position x/y and facing 'R', then this NPC will try to go to x-1/y
	// Start: go to the position behind the target entity
	target := _followGetTargetPosition(*t.targetEntity, t.FollowTask.distance)
	if target.Equals(t.Owner.Entity.TilePos) {
		return
	}

	actualGoal, me := t.Owner.Entity.GoToPos(target, true)
	if !me.Success {
		// if GoToPos fails, that means we could not find any route at all to the goal (or even a partial route).
		// since the NPC should not be currently moving, that also means this should be cause by a real problem in the path finding.
		// wait a second and retry.
		logz.Println(t.Owner.DisplayName, "failed to find path to follow entity:", me)
		t.Owner.Wait(time.Second)
		t.Restart = true
		return
	}
	t.FollowTask.targetPosition = actualGoal
}

func _followGetTargetPosition(e entity.Entity, dist int) model.Coords {
	target := e.TilePos
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

func (t *Task) updateFollow() {
	t.handleNPCCollision(t.startFollow)

	// Check if we should recalculate a better path to the target entity
	if t.recalculatePath {
		logz.Println(t.Owner.DisplayName, "follow: NPC Manager told this NPC to recalculate path")
		t.Owner.Entity.CancelCurrentPath()
		t.Owner.Wait(time.Millisecond * 500)
		t.Restart = true
		t.recalculatePath = false
		return
	}

	// ensure the NPC's target position matches the actual target position it is pursuing
	if len(t.Owner.Entity.Movement.TargetPath) > 0 {
		entTargetTile := t.Owner.Entity.Movement.TargetPath[len(t.Owner.Entity.Movement.TargetPath)-1]
		if !t.FollowTask.targetPosition.Equals(entTargetTile) {
			logz.Println(t.Owner.DisplayName, "follow: updated NPC target position since it didn't match the last position of target path")
			t.FollowTask.targetPosition = entTargetTile.Copy()
		}
	}

	// while following, there is no defined end goal. the NPC keeps trying to follow the entity until specifically told to stop.
	// so, check for when the NPC has reached its current goal and then try to recalculate a new one.
	if t.Owner.Entity.TilePos.Equals(t.FollowTask.targetPosition) {
		// we've met our first goal, now time to try and recalculate.
		t.startFollow()
		return
	}

	// if, for some reason, the entity loses its target path on accident (but isn't at goal yet), then recalculate new path
	if len(t.Owner.Entity.Movement.TargetPath) == 0 {
		//logz.Println(t.Owner.DisplayName, "follow: NPC unexpected has empty path!")
		t.startFollow()
		return
	}
}

// must NOT directly modify crucial state of NPC or entity (e.g. setting its position, movement, etc directly).
// can only suggest changes that will then be picked up in the normal game update loop and handled there.
func (t *Task) followBackgroundAssist() {
	if t.FollowTask.recalculatePath {
		return // previous flag set hasn't been acted upon yet
	}

	// calculate a new path
	entDir := t.Owner.Entity.Movement.Direction
	start := t.Owner.Entity.TilePos.GetAdj(entDir).GetAdj(entDir)
	goal := _followGetTargetPosition(*t.targetEntity, t.FollowTask.distance)
	if start.Equals(goal) {
		return
	}
	newPath, _ := t.Owner.Entity.World.FindPath(start, goal)
	if len(newPath) < 2 {
		return
	}

	// current target path must be long enough to determine the actual direction its heading
	// NOTE: this was put before the FindPath call above, but an index out of range error happened.
	// since this is happening in a separate goroutine from the main update loop, it's possible for the target path to be changing
	// during the execution of this function. If the problem keeps happening, we may need to create a synchronized access system for this.
	if len(t.Owner.Entity.Movement.TargetPath) < 2 {
		return
	}

	// // check if the new path looks mergeable to the current one
	// newPathRelDir := model.GetRelativeDirection(newPath[0], newPath[1])
	// curPathRelDir := model.GetRelativeDirection(t.Owner.Entity.Movement.TargetPath[0], t.Owner.Entity.Movement.TargetPath[1])
	//logz.Println(t.Owner.DisplayName, "follow BG assist: new path dir: "+string(newPathRelDir)+" cur path dir: "+string(curPathRelDir))
	t.Owner.Entity.Movement.SuggestedTargetPath = newPath
	// if newPathRelDir == curPathRelDir {
	// 	// if the new path is moving in the same relative direction, we consider it as "looks mergeable"
	// 	// the entity logic will actual determine if it can be merged though, and handle the merging process
	// 	t.Owner.Entity.Movement.SuggestedTargetPath = newPath
	// } else {
	// 	// if not, then tell the NPC to fully recalculate its path
	// 	//t.FollowTask.recalculatePath = true
	// }
	t.lastBackgroundAssist = time.Now()
}
