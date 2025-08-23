package npc

import (
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

func (t *Task) startFollow() {
	// when following an entity, the NPC tries to go directly behind it.
	// so, if the entity as at position x/y and facing 'R', then this NPC will try to go to x-1/y
	// Start: go to the position behind the target entity
	target := getFollowPosition(*t.TargetEntity)
	actualGoal, me := t.Owner.Entity.GoToPos(target, true)
	if !me.Success {
		logz.Println(t.Owner.DisplayName, "failed to find path to follow entity")
	}
	t.GotoTask.GoalPos = actualGoal
}

func getFollowPosition(e entity.Entity) model.Coords {
	target := e.TilePos
	switch e.Movement.Direction {
	case entity.DIR_L:
		target.X++
	case entity.DIR_R:
		target.X--
	case entity.DIR_U:
		target.Y++
	case entity.DIR_D:
		target.Y--
	default:
		panic("entity has invalid direction!")
	}
	return target
}

func (t *Task) updateFollow() {

}
