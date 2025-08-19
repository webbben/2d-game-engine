package entity

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
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
	if general_util.EuclideanDistCoords(e.TilePos, c) > 1 {
		panic("entity tried to move to a non-adjacent tile")
	}
	if e.World.Collides(c) {
		return MoveError{
			Collision: true,
		}
	}

	e.Movement.TargetTile = c
	e.Movement.IsMoving = true
	return MoveError{
		Success: true,
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

	path := e.World.FindPath(c)
	e.Movement.TargetPath = path
}

// gets the correct set of animations frames, based on the direction the entity is facing
func (e Entity) getMovementAnimationFrames() []*ebiten.Image {
	var animationFrames []*ebiten.Image
	switch e.Movement.Direction {
	case 'L':
		if e.Movement.IsRunning {
			animationFrames = e.Movement.LeftRun
		} else {
			animationFrames = e.Movement.Left
		}
	case 'R':
		if e.Movement.IsRunning {
			animationFrames = e.Movement.RightRun
		} else {
			animationFrames = e.Movement.Right
		}
	case 'U':
		if e.Movement.IsRunning {
			animationFrames = e.Movement.UpRun
		} else {
			animationFrames = e.Movement.Up
		}
	case 'D':
		if e.Movement.IsRunning {
			animationFrames = e.Movement.DownRun
		} else {
			animationFrames = e.Movement.Down
		}
	default:
		panic("incorrect direction value found during UpdateMovement")
	}

	return animationFrames
}
