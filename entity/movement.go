package entity

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/model"
)

func (e *Entity) TryMove(direction byte, run bool) {
	if e.Movement.IsMoving {
		return
	}

	c := e.TilePos.Copy()
	switch direction {
	case 'L':
		c.X--
	case 'R':
		c.X++
	case 'U':
		c.Y--
	case 'D':
		c.Y++
	default:
		panic("invalid direction given")
	}

	if e.World.Collides(c) {
		return
	}

	e.Movement.TargetTile = c
	e.Movement.IsMoving = true
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
