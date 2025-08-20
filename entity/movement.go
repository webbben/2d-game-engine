package entity

import (
	"fmt"
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

// set the next movement target and start the move process.
// not meant to be a main entry point for movement control.
func (e *Entity) move(c model.Coords) MoveError {
	if e.World == nil {
		panic("entity does not have world context set")
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

// TryMove is for handling an attempt to move to an adjacent tile.
func (e *Entity) TryMove(c model.Coords) MoveError {
	if e.Movement.IsMoving {
		return MoveError{
			AlreadyMoving: true,
		}
	}
	if general_util.EuclideanDistCoords(e.TilePos, c) > 1 {
		panic("TryMove: entity tried to move to a non-adjacent tile")
	}

	return e.move(c)
}

// Almost the same as TryMove, but for queuing up the next move right before the current one ends.
// The purpose of this is to slightly improve movement performance and avoid wasting a tick.
// Only allowed for moving to a "mostly adjacent" tile, i.e. a tile that is not further than 1.5 tiles away
//
// Only should be used in specific scenarios:
//
// 1. entity is following a path and its next move is the same direction
//
// 2. user is moving again in the same direction
func (e *Entity) tryQueueNextMove(c model.Coords) MoveError {
	// since we are already moving, we expect to have IsMoving be true
	if !e.Movement.IsMoving {
		panic("tryQueueNextMove called when entity is not moving. this means some logic somewhere must be mixed up!")
	}
	if general_util.EuclideanDistCoords(e.Movement.TargetTile, c) > 1.5 {
		fmt.Println("from:", e.Movement.TargetTile, "to:", c)
		fmt.Println("dist:", general_util.EuclideanDistCoords(e.Movement.TargetTile, c))
		panic("tryQueueNextMove: entity tried to queue move to a tile that is not 'semi-adjacent' (< 1.5 tiles away)")
	}

	return e.move(c)
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

func (e *Entity) GoToPos(c model.Coords) {
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
