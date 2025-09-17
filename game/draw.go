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

	var nightFx float32 = 0.0
	// draw lighting with shader
	if g.Hour <= 6 && g.Hour >= 18 {
		// scale nightFx as the night deepens
		switch g.Hour {
		case 18:
			nightFx = 0.1
		case 19:
			nightFx = 0.2
		case 20:
			nightFx = 0.3
		case 21:
			nightFx = 0.4
		case 22:
			nightFx = 0.5
		case 23:
			nightFx = 0.6
		case 0:
			nightFx = 0.7
		case 1:
			nightFx = 0.6
		case 2:
			nightFx = 0.5
		case 3:
			nightFx = 0.4
		case 4:
			nightFx = 0.3
		case 5:
			nightFx = 0.2
		case 6:
			nightFx = 0.1
		}
	}
	offsetX, offsetY := g.Camera.GetAbsPos()
	lights.DrawMapLighting(screen, g.worldScene, g.MapInfo.Lights, g.daylightFader.GetCurrentColor(), nightFx, offsetX, offsetY)

	// draw dialog
	if g.Dialog != nil {
		g.Dialog.Draw(screen)
	}
}

func (g *Game) drawWorldScene(screen *ebiten.Image) {
	offsetX, offsetY := g.Camera.GetAbsPos()

	// draw all layers that should be shown below entities
	g.MapInfo.Map.DrawGroundLayers(screen, offsetX, offsetY)

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

	// draw roof tops
	g.MapInfo.Map.DrawRooftopLayer(screen, offsetX, offsetY)
}
