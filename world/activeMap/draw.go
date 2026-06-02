package activemap

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/internal/camera"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/utils"
)

func (m *ActiveMap) drawWorldScene(screen *ebiten.Image, offsetX, offsetY float64) {
	// draw all layers that should be shown below entities
	m.Map.DrawGroundLayers(screen, offsetX, offsetY)

	if config.DrawGridLines {
		m.drawGridLines(screen, offsetX, offsetY)
	}
	if config.ShowNPCPaths {
		m.drawPaths(screen, offsetX, offsetY)
	}

	// draw NPCs and the player in order of Y position (higher renders first)
	for _, thing := range m.sortedRenderables {
		if thing == nil {
			continue
		}
		thing.Draw(screen, offsetX, offsetY)
	}

	if config.ShowEntityPositions {
		m.drawEntityPositions(screen, offsetX, offsetY)
	}
	if config.ShowCollisions {
		m.drawCollisions(screen, offsetX, offsetY)
	}

	// draw roof tops
	m.Map.DrawRooftopLayer(screen, offsetX, offsetY)
}

func (m *ActiveMap) Draw(screen *ebiten.Image, om *overlay.OverlayManager) {
	m.worldScene.Clear()

	offsetX, offsetY := m.Camera.GetAbsPos()
	m.drawWorldScene(m.worldScene, offsetX, offsetY)

	// add lighting

	// make sure that there are only MaxLights number of lights
	// we limit the number of lights for performance reasons, so we should only get lights that are close enough
	// to the player to influence the currently visible area
	drawLights := []*lights.Light{}
	skippedLight := false

	for _, l := range m.Lights {
		if len(drawLights) == lights.MaxLights {
			skippedLight = true
			break
		}
		if isLightVisible(l, m.Camera) {
			drawLights = append(drawLights, l)
		}
	}
	for _, lightObj := range m.LightObjects {
		if len(drawLights) == lights.MaxLights {
			skippedLight = true
			break
		}
		if !lightObj.Light.On {
			continue
		}
		if isLightVisible(lightObj.Light.Light, m.Camera) {
			drawLights = append(drawLights, lightObj.Light.Light)
		}
	}
	if skippedLight {
		logz.Warnln("Map Lights", "MaxLights reached; some lights were skipped")
		logz.Println("", drawLights)
	}

	m.daylightFader.SetOverallFactor(float32(m.Map.DaylightFactor))
	lights.DrawMapLighting(
		screen,
		m.worldScene,
		drawLights,
		m.daylightFader.GetCurrentColor(),
		m.daylightFader.GetDarknessFactor(),
		offsetX,
		offsetY,
	)

	if m.dialogSession != nil {
		m.dialogSession.Draw(screen)
	}

	if m.showPlayerMenu {
		m.playerMenuViewer.Draw(screen)
	}
	if m.showMiscScreen {
		m.miscScreenViewer.Draw(screen)
	}

	if m.bookSession != nil {
		bdx, bdy := m.bookSession.Dimensions()
		bsx, bsy := utils.CenterInScreen(bdx, bdy)
		m.bookSession.Draw(screen, bsx, bsy)
	}
}

func isLightVisible(l *lights.Light, camera camera.Camera) bool {
	screenRect := camera.GetVisibleScreenRect()
	lightRect := model.NewRect(float64(l.X-l.MaxRadius), float64(l.Y-l.MaxRadius), float64(l.MaxRadius*2), float64(l.MaxRadius*2))

	return lightRect.Intersects(screenRect)
}
