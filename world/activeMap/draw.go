package activemap

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/ui/overlay"
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
	if config.ShowEntityPositions {
		m.drawEntityPositions(screen, offsetX, offsetY)
	}
	if config.ShowCollisions {
		m.drawCollisions(screen, offsetX, offsetY)
	}

	// draw NPCs and the player in order of Y position (higher renders first)
	for _, thing := range m.sortedRenderables {
		if thing == nil {
			continue
		}
		thing.Draw(screen, offsetX, offsetY)
	}

	// draw roof tops
	m.Map.DrawRooftopLayer(screen, offsetX, offsetY)
}

func (m *ActiveMap) Draw(screen *ebiten.Image, om *overlay.OverlayManager) {
	m.worldScene.Clear()

	offsetX, offsetY := m.Camera.GetAbsPos()
	m.drawWorldScene(m.worldScene, offsetX, offsetY)

	// add lighting

	objectLights := []*lights.Light{}
	for _, lightObj := range m.LightObjects {
		if lightObj.Light.On {
			objectLights = append(objectLights, lightObj.Light.Light)
		}
	}
	m.daylightFader.SetOverallFactor(float32(m.Map.DaylightFactor))
	lights.DrawMapLighting(
		screen,
		m.worldScene,
		m.Lights,
		objectLights,
		m.daylightFader.GetCurrentColor(),
		m.daylightFader.GetDarknessFactor(),
		offsetX,
		offsetY,
	)
}
