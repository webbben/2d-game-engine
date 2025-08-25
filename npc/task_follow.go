package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

func (t *Task) startFollow() {
	// when following an entity, the NPC tries to go directly behind it.
	// so, if the entity as at position x/y and facing 'R', then this NPC will try to go to x-1/y
	// Start: go to the position behind the target entity
	target := getFollowPosition(*t.targetEntity, t.FollowTask.distance)
	if target.Equals(t.Owner.Entity.TilePos) {
		return
	}
	actualGoal, me := t.Owner.Entity.GoToPos(target, true)
	if !me.Success {
		// if GoToPos fails, that means we could not find any route at all to the goal (or even a partial route).
		// wait a second and retry
		logz.Println(t.Owner.DisplayName, "failed to find path to follow entity:", me)
		t.Owner.Wait(time.Millisecond * 500)
		t.Restart = true
		return
	}
	t.GotoTask.goalPos = actualGoal
}

func getFollowPosition(e entity.Entity, dist int) model.Coords {
	target := e.TilePos
	switch e.Movement.Direction {
	case entity.DIR_L:
		target.X += dist + 1
	case entity.DIR_R:
		target.X -= dist + 1
	case entity.DIR_U:
		target.Y += dist + 1
	case entity.DIR_D:
		target.Y -= dist + 1
	default:
		panic("entity has invalid direction!")
	}
	return target
}

func (t *Task) updateFollow() {
	t.handleNPCCollision(t.startFollow)

	// check if NPC is going the wrong direction. this can happen if the target entity suddenly turns around.
	if t.Owner.Entity.Movement.Direction == entity.GetOppositeDirection(t.FollowTask.targetEntity.Movement.Direction) {
		// TODO: improve logic for identifying if NPC should cancel path.
		// sometimes it needs to move one tile in the wrong direction in order to start moving in the right direction, so this logic here
		// is too simple and needs improvement.

		// entity is moving in the opposite direction of the target; recalculate path
		t.Owner.Entity.CancelCurrentPath()
		if t.Owner.Entity.Movement.IsMoving {
			t.Owner.Wait(time.Second)
		}
		t.Restart = true
		return
	}

	// while following, there is no defined end goal. the NPC keeps trying to follow the entity until specifically told to stop.
	// so, check for when the NPC has reached its current goal and then try to recalculate a new one.
	if t.Owner.Entity.TilePos.Equals(t.GotoTask.goalPos) {
		// we've met our first goal, now time to try and recalculate.
		t.startFollow()
		return
	}
}
