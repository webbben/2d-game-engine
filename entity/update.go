package entity

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

func (e Entity) Draw(screen *ebiten.Image, offsetX float64, offsetY float64) {
	if !e.Loaded {
		return
	}
	if e.CurrentFrame == nil {
		panic("tried to draw entity with no set frame")
	}

	op := &ebiten.DrawImageOptions{}
	drawX, drawY := rendering.GetImageDrawPos(e.CurrentFrame, e.X, e.Y, offsetX, offsetY)
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(e.CurrentFrame, op)
}

func (e *Entity) Update() {
	if !e.Movement.IsMoving {
		if len(e.Movement.TargetPath) > 0 {
			moveError := e.TryMove(e.Movement.TargetPath[0])
			if moveError.Success {
				// shift target path
				e.Movement.TargetPath = e.Movement.TargetPath[1:]
			} else {
				log.Println("TryMove failed:", moveError)
			}
		}
	}
	if e.Movement.IsMoving {
		e.updateMovement()
	}

	e.updateCurrentFrame()
}

func (e *Entity) updateMovement() {
	if e.Movement.Speed == 0 {
		panic("updateMovement called when speed is 0; speed was not set wherever entity movement was started")
	}
	if e.Movement.TargetTile.Equals(e.TilePos) && len(e.Movement.TargetPath) == 0 {
		panic("updateMovement called when entity has no target tile or target path")
	}
	targetPx := float64(e.Movement.TargetTile.X * config.TileSize)
	targetPy := float64(e.Movement.TargetTile.Y * config.TileSize)
	dx := targetPx - e.X
	dy := targetPy - e.Y

	dist := math.Hypot(dx, dy)

	finishMove := false

	if dist <= e.Movement.Speed {
		// check if there is a next move ready that is the same direction as the current move
		// if so, then keep on trucking. this helps to avoid a little blip and makes walking go smoother

		if len(e.Movement.TargetPath) > 0 {
			// Case: entity has path it is following
			if getRelativeDirection(e.Movement.TargetTile, e.Movement.TargetPath[0]) == e.Movement.Direction {
				lastTarget := e.Movement.TargetTile.Copy()
				moveError := e.tryQueueNextMove(e.Movement.TargetPath[0])
				if moveError.Success {
					// update entity tile coords
					e.TilePos = lastTarget
					// shift target path
					e.Movement.TargetPath = e.Movement.TargetPath[1:]
				} else {
					log.Println("tryQueueNextMove failed:", moveError)
					finishMove = true
				}
			} else {
				finishMove = true
			}
		} else if e.IsPlayer && playerStillWalking(e.Movement.Direction) {
			// Case: player is still moving in the same direction
			nextTarget := e.Movement.TargetTile.GetAdj(e.Movement.Direction)
			lastTarget := e.Movement.TargetTile.Copy()
			moveError := e.tryQueueNextMove(nextTarget)
			if moveError.Success {
				e.TilePos = lastTarget
			} else {
				log.Println("tryQueueNextMove failed:", moveError)
				finishMove = true
			}
		} else {
			finishMove = true
		}
	}
	if finishMove {
		// finish movement
		e.X, e.Y = targetPx, targetPy
		e.TilePos = e.Movement.TargetTile.Copy()
		e.Movement.IsMoving = false
	} else {
		moveDx := e.Movement.Speed * dx / dist
		moveDy := e.Movement.Speed * dy / dist
		if moveDx == 0 && moveDy == 0 {
			panic("somehow, movement distance calculation is 0! entity is stuck and not moving towards its goal")
		}
		e.X += moveDx
		e.Y += moveDy
	}

	// update animation
	e.Movement.AnimationTimer++
	if e.Movement.AnimationTimer > 10 {
		_, frameCount := e.getMovementAnimationInfo()
		e.Movement.AnimationFrame = (e.Movement.AnimationFrame + 1) % frameCount
		e.Movement.AnimationTimer = 0
	}
}

func playerStillWalking(direction byte) bool {
	switch direction {
	case 'L':
		return ebiten.IsKeyPressed(ebiten.KeyA)
	case 'R':
		return ebiten.IsKeyPressed(ebiten.KeyD)
	case 'U':
		return ebiten.IsKeyPressed(ebiten.KeyW)
	case 'D':
		return ebiten.IsKeyPressed(ebiten.KeyS)
	default:
		panic("playerStillWalking: invalid direction passed")
	}
}

func (e *Entity) updateCurrentFrame() {
	if e.Movement.IsMoving {
		animationName, _ := e.getMovementAnimationInfo()
		e.CurrentFrame = e.getAnimationFrame(animationName, e.Movement.AnimationFrame)
		return
	}
	// idle
	animationName, _ := e.getMovementAnimationInfo()
	e.CurrentFrame = e.getAnimationFrame(animationName, 0)
}
