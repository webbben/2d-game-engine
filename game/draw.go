package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/debug"
	"github.com/webbben/2d-game-engine/internal/lights"
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

	if g.MapInfo == nil {
		return
	}

	g.drawWorld(screen)

	// ~~~
	// Below is all config; any real world stuff should be rendered above
	// ~~~

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

func (g *Game) drawWorld(screen *ebiten.Image) {
	g.worldScene.Clear()
	g.drawWorldScene(g.worldScene)

	// draw lighting with shader
	lights.DrawMapLighting(screen, g.worldScene, g.MapInfo.Lights)

	// draw dialog
	if g.Dialog != nil {
		g.Dialog.Draw(screen)
	}
}

func (g *Game) drawWorldScene(screen *ebiten.Image) {
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
}
