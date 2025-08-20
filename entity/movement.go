package entity

import (
	"log"
	"math"

	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/model"
)

type MoveError struct {
	AlreadyMoving bool
	Collision     bool
	Success       bool
}

// returns MoveError which indicates if any error prevented the move from happening
func (e *Entity) TryMove(c model.Coords) MoveError {
	if e.Movement.IsMoving {
		return MoveError{
			AlreadyMoving: true,
		}
	}
	if e.World == nil {
		panic("entity does not have world context set")
	}
	if general_util.EuclideanDistCoords(e.TilePos, c) > 1 {
		panic("entity tried to move to a non-adjacent tile")
	}
	if e.World.Collides(c) {
		log.Println("collision")
		return MoveError{
			Collision: true,
		}
	}

	e.Movement.TargetTile = c

	e.Movement.Speed = e.Movement.WalkSpeed
	if e.Movement.Speed == 0 {
		panic("entity movement speed set to 0 in TryMove")
	}

	e.Movement.Direction = getRelativeDirection(e.TilePos, e.Movement.TargetTile)

	e.Movement.IsMoving = true
	return MoveError{
		Success: true,
	}
}

func getRelativeDirection(a, b model.Coords) byte {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if math.Abs(float64(dx)) > math.Abs(float64(dy)) {
		if dx < 0 {
			return 'L'
		} else {
			return 'R'
		}
	} else {
		if dy < 0 {
			return 'U'
		} else {
			return 'D'
		}
	}
}

func (e *Entity) GoToPos(c model.Coords, run bool) {
	if e.Movement.IsMoving {
		log.Println("GoToPos: entity is already moving")
		return
	}
	if len(e.Movement.TargetPath) > 0 {
		log.Println("GoToPos: entity already has a target path. Target path should be cancelled first.")
	}

	path := e.World.FindPath(e.TilePos, c)
	e.Movement.TargetPath = path
}

// gets the frameType and frameCount for the correct animation, based on the entity's direction
func (e Entity) getMovementAnimationInfo() (string, int) {
	var name string
	switch e.Movement.Direction {
	case 'L':
		if e.Movement.IsMoving {
			name = "left_walk"
		} else {
			name = "left_idle"
		}
	case 'R':
		if e.Movement.IsMoving {
			name = "right_walk"
		} else {
			name = "right_idle"
		}
	case 'U':
		if e.Movement.IsMoving {
			name = "up_walk"
		} else {
			name = "up_idle"
		}
	case 'D':
		if e.Movement.IsMoving {
			name = "down_walk"
		} else {
			name = "down_idle"
		}
	default:
		panic("incorrect direction value found during UpdateMovement")
	}

	count, exists := e.AnimationFrameCount[name]
	if !exists {
		panic("animation name has no record in AnimationFrameCount map")
	}

	return name, count
}
