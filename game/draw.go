package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/debug"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
)

func (g *Game) Draw(screen *ebiten.Image) {
	offsetX, offsetY := g.Camera.GetAbsPos()

	// draw the terrain of the room
	g.Room.DrawFloor(screen, offsetX, offsetY)
	g.Room.DrawCliffs(screen, offsetX, offsetY)
	if config.DrawGridLines {
		g.drawGridLines(screen, offsetX, offsetY)
	}

	// draw objects, entities, and the player in order of Y position (higher renders first)
	o, e := 0, 0
	obj, ent := g.Objects[0], g.Entities[0]
	playerDrawn := false
	for o < len(g.Objects) || e < len(g.Entities) {
		if !playerDrawn && g.Player.Y <= ent.Y && g.Player.Y <= float64(obj.Y) {
			g.Player.Draw(screen, offsetX, offsetY)
			playerDrawn = true
		}
		if o < len(g.Objects) && float64(obj.Y) < ent.Y {
			// if there are objects to draw still and this one is higher than the next entity
			obj.Draw(screen, offsetX, offsetY, g.ImageMap)
			o++
			if o < len(g.Objects) {
				obj = g.Objects[o]
			} else {
				obj = object.Object{Y: 99999} // no more objects; set to high Y value so it won't stop other things from being drawn
			}
		} else {
			// else - if there are no objects to draw anymore, or this entity is same or higher than the next object
			ent.Draw(screen, offsetX, offsetY)
			e++
			if e < len(g.Entities) {
				ent = g.Entities[e]
			} else {
				ent = &entity.Entity{Position: model.Position{Y: 99999}} // no more entities
			}
		}
	}
	if !playerDrawn {
		g.Player.Draw(screen, offsetX, offsetY)
	}

	// draw lighting shade (e.g. for night) here
	// blue := color.RGBA{R: 0, G: 0, B: 50, A: 10}
	// ebitenutil.DrawRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, blue)

	// draw dialog
	if g.Dialog != nil {
		g.Dialog.DrawDialog(screen)
	}

	if config.ShowPlayerCoords {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("Player pos: [%v, %v]", g.Player.X, g.Player.Y))
	}
	if config.TrackMemoryUsage {
		ebitenutil.DebugPrint(screen, debug.GetMemoryUsageStats())
	}
}

func (g *Game) drawGridLines(screen *ebiten.Image, offsetX float64, offsetY float64) {
	offsetX = offsetX * config.GameScale
	offsetY = offsetY * config.GameScale
	lineColor := color.RGBA{255, 0, 0, 255}
	maxWidth := float64(len(g.Room.TileLayout[0]) * config.TileSize * config.GameScale)
	maxHeight := float64(len(g.Room.TileLayout) * config.TileSize * config.GameScale)

	// vertical lines
	for x := 0; x < len(g.Room.TileLayout[0]); x++ {
		drawX := float64(x*config.TileSize*config.GameScale) - offsetX
		drawY := -offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(drawX), float32(maxHeight-offsetY), 1, lineColor, true)
	}
	// horizontal lines
	for y := 0; y < len(g.Room.TileLayout); y++ {
		drawX := -offsetX
		drawY := float64(y*config.TileSize*config.GameScale) - offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(maxWidth-offsetX), float32(drawY), 1, lineColor, true)
	}
}
