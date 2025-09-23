package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
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
	drawX, drawY := e.DrawPos(offsetX, offsetY)
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(e.CurrentFrame, op)
}

// returns the actual absolute position where the entity will be drawn
func (e Entity) DrawPos(offsetX, offsetY float64) (drawX, drawY float64) {
	if e.CurrentFrame == nil {
		panic("tried to get draw position for entity with no set frame")
	}
	drawX, drawY = rendering.GetImageDrawPos(e.CurrentFrame, e.X, e.Y, offsetX, offsetY)
	drawY -= 6 // move up a little, since we want the entity to look like its standing in the middle of the tile
	return drawX, drawY
}

// returns the absolute position where the "extent" of the entity's image lies.
// by "extent", we basically mean just the position of the end of the actual image rectangle, when the image is positioned for drawing.
// you would use this (along with DrawPos) when checking if an entity is actually touching or overlapping physically with something
func (e Entity) ExtentPos(offsetX, offsetY float64) (extentX, extentY float64) {
	if e.CurrentFrame == nil {
		panic("tried to get extent position for entity with no set frame")
	}
	extentX, extentY = e.DrawPos(offsetX, offsetY)
	extentX += float64(e.CurrentFrame.Bounds().Dx())
	extentY += float64(e.CurrentFrame.Bounds().Dy())
	return extentX, extentY
}

func (e *Entity) Update() {
	if !e.Movement.IsMoving {
		if len(e.Movement.TargetPath) > 0 {
			e.trySetNextTargetPath()
		}
	}

	if e.Movement.IsMoving {
		e.updateMovement()
	} else {
		if e.TargetX != e.X || e.TargetY != e.Y {
			logz.Println(e.DisplayName, "x:", e.X, "y:", e.Y, "targetX:", e.TargetX, "targetY:", e.TargetY)
			panic("entity is not moving but hasn't met its goal yet")
		}
	}

	e.updateCurrentFrame()
}

func (e *Entity) updateCurrentFrame() {
	// handle stopping movement
	// to prevent an awkward frame skip, we wait until one tick after movement stops to actually stop the movement animation.
	// this is so on the next tick, the player or npc logic has another chance to queue up a next movement before actually fully stopping.
	if !e.Movement.IsMoving {
		if e.Movement.movementStopped {
			// idle
			animationName, _ := e.getMovementAnimationInfo()
			e.CurrentFrame = e.getAnimationFrame(animationName, 0)
			return
		}
	}

	animationName, _ := e.getMovementAnimationInfo()
	e.CurrentFrame = e.getAnimationFrame(animationName, e.Movement.AnimationFrame)

	// need to set it here so that the last animation frame can be properly gotten
	if !e.Movement.IsMoving {
		e.Movement.movementStopped = true
	}
}
