package entity

import (
	"fmt"
	"math"

	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

type MoveError struct {
	AlreadyMoving bool
	Collision     bool
	Cancelled     bool
	Success       bool
}

func (me MoveError) String() string {
	if me.Success {
		return "Success"
	}
	if me.Collision {
		return "Collision"
	}
	if me.AlreadyMoving {
		return "Already Moving"
	}
	if me.Cancelled {
		return "Cancelled"
	}
	return "No value set"
}

// set the next movement target and start the move process.
// not meant to be a main entry point for movement control.
func (e *Entity) move(c model.Coords) MoveError {
	if e.World == nil {
		panic("entity does not have world context set")
	}
	if e.World.Collides(c) {
		logz.Println(e.DisplayName, "collision")
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
	e.Movement.Interrupted = false

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
		fmt.Println("from:", e.TilePos, "to:", c)
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

func (e *Entity) GoToPos(c model.Coords) MoveError {
	if e.Movement.IsMoving {
		logz.Println(e.DisplayName, "GoToPos: entity is already moving")
		return MoveError{AlreadyMoving: true}
	}
	if len(e.Movement.TargetPath) > 0 {
		logz.Println(e.DisplayName, "GoToPos: entity already has a target path. Target path should be cancelled first.")
		logz.Println(e.DisplayName, e.Movement.TargetPath)
	}
	if e.TilePos.Equals(c) {
		logz.Errorln(e.DisplayName, "entity attempted to GoToPos for position it is already in")
		return MoveError{Cancelled: true}
	}

	path := e.World.FindPath(e.TilePos, c)
	if len(path) == 0 {
		fmt.Println("tile pos:", e.TilePos, "goal:", c)
		logz.Warnln(e.DisplayName, "GoToPos: calculated path is empty. Is the entity blocked in? Why is the movement prevented?")
		return MoveError{Cancelled: true}
	}

	e.Movement.TargetPath = path
	return MoveError{Success: true}
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
