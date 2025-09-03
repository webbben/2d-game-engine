package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/debug"
)

func (g *Game) Draw(screen *ebiten.Image) {
	//screen.Clear()

	// draw a screen, if present
	if g.CurrentScreen != nil {
		g.CurrentScreen.DrawScreen(screen)
		return
	}

	// ~~~
	// below is in-game rendering; stop here if we are not in the game world
	// ~~~

	if g.MapInfo == nil {
		return
	}

	offsetX, offsetY := g.Camera.GetAbsPos()

	// draw the terrain of the room
	g.MapInfo.Map.DrawLayers(screen, offsetX, offsetY)
	if config.DrawGridLines {
		g.drawGridLines(screen, offsetX, offsetY)
	}
	g.drawPaths(screen, offsetX, offsetY)

	// draw objects, entities, and the player in order of Y position (higher renders first)
	// should be sorted in the update loop
	playerDrawn := false
	for _, n := range g.MapInfo.NPCs {
		if !playerDrawn && g.Player.Entity.Position.TilePos.Y <= n.Entity.Position.TilePos.Y {
			g.Player.Entity.Draw(screen, offsetX, offsetY)
			playerDrawn = true
		}
		n.Entity.Draw(screen, offsetX, offsetY)
	}
	if !playerDrawn {
		g.Player.Entity.Draw(screen, offsetX, offsetY)
	}

	// draw lighting shade (e.g. for night) here
	// blue := color.RGBA{R: 0, G: 0, B: 50, A: 10}
	// ebitenutil.DrawRect(screen, 0, 0, config.ScreenWidth, config.ScreenHeight, blue)

	// draw dialog
	if g.Dialog != nil {
		g.Dialog.Draw(screen)
	}

	if config.ShowPlayerCoords {
		g.drawEntityCoords(screen)
	}
	if config.TrackMemoryUsage {
		ebitenutil.DebugPrint(screen, debug.GetLog())
	}

	// show game debug info, including this scale info
	if config.ShowGameDebugInfo {
		g.showGameDebugInfo(screen)
	}
}
