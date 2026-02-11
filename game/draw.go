package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/internal/overlay"
)

func (g *Game) Draw(screen *ebiten.Image) {
	if g.MapInfo == nil {
		return
	}

	g.drawWorld(screen, g.OverlayManager)

	g.OverlayManager.Draw(screen)

	// show game debug info, including this scale info
	if config.ShowGameDebugInfo {
		g.showGameDebugInfo(screen)
	}
}

func (g *Game) drawWorld(screen *ebiten.Image, om *overlay.OverlayManager) {
	g.worldScene.Clear()
	g.drawWorldScene(g.worldScene)

	offsetX, offsetY := g.Camera.GetAbsPos()
	objectLights := []*lights.Light{}
	for _, lightObj := range g.MapInfo.LightObjects {
		if lightObj.Light.On {
			objectLights = append(objectLights, lightObj.Light.Light)
		}
	}
	g.daylightFader.SetOverallFactor(float32(g.MapInfo.Map.DaylightFactor))
	lights.DrawMapLighting(screen, g.worldScene, g.MapInfo.Lights, objectLights, g.daylightFader.GetCurrentColor(), g.daylightFader.GetDarknessFactor(), offsetX, offsetY)

	// draw dialog
	if g.dialogSession != nil {
		g.dialogSession.Draw(screen)
	} else if g.ShowPlayerMenu {
		g.PlayerMenu.Draw(screen, om)
	} else if g.ShowTradeScreen {
		g.TradeScreen.Draw(screen, om)
	} else {
		// if nothing else is showing (player menus, trade screens, etc) then draw HUD
		// TODO: should this be done after the overlay manager?
		if g.hud != nil {
			g.hud.Draw(screen)
		}
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
	if config.ShowEntityPositions {
		g.drawEntityPositions(screen, offsetX, offsetY)
	}

	if config.ShowCollisions {
		g.drawCollisions(screen, offsetX, offsetY)
	}

	// draw NPCs and the player in order of Y position (higher renders first)
	for _, thing := range g.MapInfo.sortedRenderables {
		if thing == nil {
			continue
		}
		thing.Draw(screen, offsetX, offsetY)
	}

	// draw roof tops
	g.MapInfo.Map.DrawRooftopLayer(screen, offsetX, offsetY)
}
