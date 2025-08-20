package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/debug"
)

func (g *Game) Draw(screen *ebiten.Image) {

	// draw a screen, if present
	if g.CurrentScreen != nil {
		g.CurrentScreen.DrawScreen(screen)
		return
	}

	// ~~~
	// below is in-game rendering; stop here if we are not in the game world
	// ~~~

	offsetX, offsetY := g.Camera.GetAbsPos()

	// draw the terrain of the room
	g.Map.DrawLayers(screen, offsetX, offsetY)
	if config.DrawGridLines {
		g.drawGridLines(screen, offsetX, offsetY)
	}

	// draw objects, entities, and the player in order of Y position (higher renders first)
	g.Player.Entity.Draw(screen, offsetX, offsetY)

	// draw lighting shade (e.g. for night) here
	// blue := color.RGBA{R: 0, G: 0, B: 50, A: 10}
	// ebitenutil.DrawRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, blue)

	// draw dialog
	if g.Conversation != nil {
		g.Conversation.DrawConversation(screen)
	}

	if config.ShowPlayerCoords {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("Player pos: [%v, %v]", g.Player.Entity.X, g.Player.Entity.Y))
	}
	if config.TrackMemoryUsage {
		ebitenutil.DebugPrint(screen, debug.GetLog())
	}
}

func (g *Game) drawGridLines(screen *ebiten.Image, offsetX float64, offsetY float64) {
	offsetX = offsetX * config.GameScale
	offsetY = offsetY * config.GameScale
	lineColor := color.RGBA{255, 0, 0, 255}
	maxWidth := float64(g.Map.Width*g.Map.TileWidth) * config.GameScale
	maxHeight := float64(g.Map.Height*g.Map.TileHeight) * config.GameScale

	// vertical lines
	for x := 0; x < g.Map.Width; x++ {
		drawX := (float64(x*config.TileSize) * config.GameScale) - offsetX
		drawY := -offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(drawX), float32(maxHeight-offsetY), 1, lineColor, true)
	}
	// horizontal lines
	for y := 0; y < g.Map.Height; y++ {
		drawX := -offsetX
		drawY := (float64(y*config.TileSize) * config.GameScale) - offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(maxWidth-offsetX), float32(drawY), 1, lineColor, true)
	}
}
