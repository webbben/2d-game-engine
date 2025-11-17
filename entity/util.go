package entity

import (
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/model"
)

func (e Entity) DistFromEntity(otherEnt Entity) float64 {
	return general_util.EuclideanDistCenter(e.CollisionRect(), otherEnt.CollisionRect())
}

// Warning: not a trivial calculation (uses path finding algorithm)
func (e Entity) GetPathToEntity(otherEnt Entity) (path []model.Coords, found bool) {
	return e.World.FindPath(e.TilePos, otherEnt.TilePos)
}

func (e *Entity) TryMoveTowardsEntity(otherEnt Entity, dist, speed float64) MoveError {
	currentPosition := model.NewVec2(e.X, e.Y)
	targetPosition := model.NewVec2(otherEnt.X, otherEnt.Y)
	v := targetPosition.Sub(currentPosition)
	scaled := v.Normalize().Scale(dist)

	return e.TryMoveMaxPx(int(scaled.X), int(scaled.Y), speed)
}

func (e *Entity) FaceTowardsEntity(otherEnt Entity) {
	currentPosition := model.NewVec2(e.X, e.Y)
	targetPosition := model.NewVec2(otherEnt.X, otherEnt.Y)
	v := targetPosition.Sub(currentPosition)
	e.FaceTowards(v.X, v.Y)
}
