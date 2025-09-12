package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
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
	if config.ShowNPCPaths {
		g.drawPaths(screen, offsetX, offsetY)
	}

	// draw objects, entities, and the player in order of Y position (higher renders first)
	for _, thing := range g.MapInfo.sortedRenderables {
		thing.Draw(screen, offsetX, offsetY)
	}
}
