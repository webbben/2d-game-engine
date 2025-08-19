package entity

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

func (e Entity) Draw(screen *ebiten.Image, offsetX float64, offsetY float64) {
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
	targetPx := float64(e.Movement.TargetTile.X * config.TileSize)
	targetPy := float64(e.Movement.TargetTile.Y * config.TileSize)
	dx := targetPx - e.X
	dy := targetPy - e.Y

	dist := math.Hypot(dx, dy)
	if dist < e.Movement.Speed {
		// snap to target since one more tick would overshoot
		e.X, e.Y = targetPx, targetPy
		e.TilePos = e.Movement.TargetTile.Copy()
		e.Movement.IsMoving = false
	} else {
		e.X += e.Movement.Speed * dx / dist
		e.Y += e.Movement.Speed * dy / dist
	}

	// update animation
	e.Movement.AnimationTimer++
	if e.Movement.AnimationTimer > 10 {
		animationFrames := e.getMovementAnimationFrames()
		e.Movement.AnimationFrame = (e.Movement.AnimationFrame + 1) % len(animationFrames)
		e.Movement.AnimationTimer = 0
	}
}

func (e *Entity) updateCurrentFrame() {
	if e.Movement.IsMoving {
		e.CurrentFrame = e.getMovementAnimationFrames()[e.Movement.AnimationFrame]
		return
	}
}
