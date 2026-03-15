package game

import (
	"fmt"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/internal/debug"
)

func (g Game) showGameDebugInfo(screen *ebiten.Image) {
	var s strings.Builder

	if config.TrackMemoryUsage {
		perfLogs := debug.GetLog()
		s.WriteString("PERFORMANCE\n")
		s.WriteString(perfLogs)
	}

	// show screen size information
	scaleX := math.Round(float64(g.outsideWidth) / float64(display.SCREEN_WIDTH) * 100)
	scaleY := math.Round(float64(g.outsideHeight) / float64(display.SCREEN_HEIGHT) * 100)
	scale := math.Min(scaleX, scaleY)
	s.WriteString(
		fmt.Sprintf(
			"SCREEN SIZE\nvirtual: %v x %v\nreal: %v x %v\nscale: %v%%\n",
			display.SCREEN_WIDTH, display.SCREEN_HEIGHT, g.outsideWidth, g.outsideHeight, scale,
		))

	mouseX, mouseY := ebiten.CursorPosition()
	s.WriteString(fmt.Sprintf("MOUSE\n%v, %v \n", mouseX, mouseY))

	if g.World != nil && g.World.ActiveMap != nil {
		if config.ShowPlayerCoords {
			s.WriteString("ENTITY POSITIONS\n")
			s.WriteString(g.World.ActiveMap.ShowEntityCoords())
		}

		s.WriteString(fmt.Sprintf("TIME: %s\n", g.World.Clock))
		g.World.ActiveMap.GetDaylightData(&s)
	}

	ebitenutil.DebugPrint(screen, s.String())
}
