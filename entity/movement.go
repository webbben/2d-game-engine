package entity

import (
	"fmt"
	"math"

	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

const (
	DIR_L byte = 'L'
	DIR_R byte = 'R'
	DIR_U byte = 'U'
	DIR_D byte = 'D'
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
		logz.Println(e.DisplayName, "move: collision")
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
			return DIR_L
		} else {
			return DIR_R
		}
	} else {
		if dy < 0 {
			return DIR_U
		} else {
			return DIR_D
		}
	}
}

func GetOppositeDirection(dir byte) byte {
	switch dir {
	case DIR_L:
		return DIR_R
	case DIR_R:
		return DIR_L
	case DIR_U:
		return DIR_D
	case DIR_D:
		return DIR_U
	default:
		panic("invalid direction given!")
	}
}

// Attempts to put the entity on a path to reach the given target.
// If the path to the target is blocked, you can conditionally go as close as possible with the "close enough" flag.
// Returns the actual goal target (in case it was changed due to a conflict and the "close enough" flag).
func (e *Entity) GoToPos(c model.Coords, closeEnough bool) (model.Coords, MoveError) {
	if e.Movement.IsMoving {
		logz.Println(e.DisplayName, "GoToPos: entity is already moving")
		return c, MoveError{AlreadyMoving: true}
	}
	if len(e.Movement.TargetPath) > 0 {
		logz.Warnln(e.DisplayName, "GoToPos: entity already has a target path. Target path should be cancelled first.")
	}
	if e.TilePos.Equals(c) {
		logz.Errorln(e.DisplayName, "entity attempted to GoToPos for position it is already in")
		return c, MoveError{Cancelled: true}
	}

	path, found := e.World.FindPath(e.TilePos, c)
	if !found {
		if !closeEnough {
			return c, MoveError{Cancelled: true}
		}
		logz.Warnln(e.DisplayName, "going a partial path since original path is blocked.", "start:", e.TilePos, "path:", path, "original goal:", c)
	}
	if len(path) == 0 {
		fmt.Println("tile pos:", e.TilePos, "goal:", c)
		logz.Warnln(e.DisplayName, "GoToPos: calculated path is empty. Is the entity completely blocked in?")
		return c, MoveError{Cancelled: true}
	}

	e.Movement.TargetPath = path
	return path[len(path)-1], MoveError{Success: true}
}

// gets the frameType and frameCount for the correct animation, based on the entity's direction
func (e Entity) getMovementAnimationInfo() (string, int) {
	var name string
	switch e.Movement.Direction {
	case DIR_L:
		if e.Movement.IsMoving {
			name = "left_walk"
		} else {
			name = "left_idle"
		}
	case DIR_R:
		if e.Movement.IsMoving {
			name = "right_walk"
		} else {
			name = "right_idle"
		}
	case DIR_U:
		if e.Movement.IsMoving {
			name = "up_walk"
		} else {
			name = "up_idle"
		}
	case DIR_D:
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

// Tells the entity to stop moving once it has finished its current tile movement.
// Meant for stopping entities that are currently following a path.
func (e *Entity) CancelCurrentPath() {
	if len(e.Movement.TargetPath) == 0 {
		logz.Warnln("tried to cancel path for an entity that has no path")
		return
	}
	e.Movement.TargetPath = []model.Coords{}
}
