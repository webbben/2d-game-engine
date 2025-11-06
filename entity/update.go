package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

func (e *Entity) Draw(screen *ebiten.Image, offsetX float64, offsetY float64) {
	if !e.Loaded {
		return
	}

	drawX, drawY := e.DrawPos(offsetX, offsetY)
	// apparently, when drawing the body, gameScale isn't automatically factored in so we need to multiply it in here
	// however, I see another place that relies on DrawPos, so I'm leaving this out of that function for now. (TODO?)
	drawX *= config.GameScale
	drawY *= config.GameScale
	e.Body.Draw(screen, drawX, drawY, config.GameScale)
	e.Position.drawX = drawX
	e.Position.drawY = drawY
}

// returns the actual absolute position where the entity will be drawn
func (e Entity) DrawPos(offsetX, offsetY float64) (drawX, drawY float64) {
	dx, dy := e.Body.Dimensions()
	rect := model.NewRect(0, 0, float64(dx), float64(dy))
	drawX, drawY = rendering.GetRectDrawPos(rect, e.X, e.Y, offsetX, offsetY)
	drawY -= 6 // move up a little, since we want the entity to look like its standing in the middle of the tile
	return drawX, drawY
}

// returns the absolute position where the "extent" of the entity's image lies.
// by "extent", we basically mean just the position of the end of the actual image rectangle, when the image is positioned for drawing.
// you would use this (along with DrawPos) when checking if an entity is actually touching or overlapping physically with something
func (e Entity) ExtentPos(offsetX, offsetY float64) (extentX, extentY float64) {
	extentX, extentY = e.DrawPos(offsetX, offsetY)
	dx, dy := e.Body.Dimensions()
	extentX += float64(dx)
	extentY += float64(dy)
	return extentX, extentY
}

func (e *Entity) Update() {
	if !e.Loaded {
		panic("entity not loaded yet!")
	}
	dx, dy := e.Body.Dimensions()
	dx = int(float64(dx) * config.GameScale)
	dy = int(float64(dy) * config.GameScale)
	e.MouseBehavior.Update(int(e.Position.drawX), int(e.Position.drawY), dx, dy, false)

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

	e.Body.Update()
}
