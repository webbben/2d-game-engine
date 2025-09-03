package game

import (
	"fmt"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

func (g Game) drawGridLines(screen *ebiten.Image, offsetX float64, offsetY float64) {
	if g.MapInfo == nil {
		panic("GAME: tried to draw grid lines on a nil map!")
	}
	offsetX = offsetX * config.GameScale
	offsetY = offsetY * config.GameScale
	lineColor := color.RGBA{255, 0, 0, 255}
	maxWidth := float64(g.MapInfo.Map.Width*g.MapInfo.Map.TileWidth) * config.GameScale
	maxHeight := float64(g.MapInfo.Map.Height*g.MapInfo.Map.TileHeight) * config.GameScale

	// vertical lines
	for x := 0; x < g.MapInfo.Map.Width; x++ {
		drawX := (float64(x*config.TileSize) * config.GameScale) - offsetX
		drawY := -offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(drawX), float32(maxHeight-offsetY), 1, lineColor, true)
	}
	// horizontal lines
	for y := 0; y < g.MapInfo.Map.Height; y++ {
		drawX := -offsetX
		drawY := (float64(y*config.TileSize) * config.GameScale) - offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(maxWidth-offsetX), float32(drawY), 1, lineColor, true)
	}
}

func (g Game) drawPaths(screen *ebiten.Image, offsetX, offsetY float64) {
	if g.MapInfo == nil {
		panic("tried to draw paths on a nil map!")
	}
	color1 := color.RGBA{255, 0, 0, 50}
	tile := ebiten.NewImage(config.TileSize, config.TileSize)
	tile.Fill(color1)
	tile2 := ebiten.NewImage(config.TileSize, config.TileSize)
	color2 := color.RGBA{0, 0, 255, 50}
	tile2.Fill(color2)

	for _, n := range g.MapInfo.NPCManager.NPCs {
		if len(n.Entity.Movement.TargetPath) > 0 {
			for _, c := range n.Entity.Movement.TargetPath {
				op := &ebiten.DrawImageOptions{}
				drawX, drawY := rendering.GetImageDrawPos(tile, float64(c.X)*config.TileSize, float64(c.Y)*config.TileSize, offsetX, offsetY)
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(tile, op)
			}
		}
		if len(n.Entity.Movement.SuggestedTargetPath) > 0 {
			for _, c := range n.Entity.Movement.SuggestedTargetPath {
				op := &ebiten.DrawImageOptions{}
				drawX, drawY := rendering.GetImageDrawPos(tile2, float64(c.X)*config.TileSize, float64(c.Y)*config.TileSize, offsetX, offsetY)
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(tile2, op)
			}
		}
	}
}

func (g Game) drawEntityCoords(screen *ebiten.Image) {
	s := ""
	s += fmt.Sprintf(
		"Player pos: [%v, %v] (%v, %v)\n",
		g.Player.Entity.TilePos.X,
		g.Player.Entity.TilePos.Y,
		g.Player.Entity.X,
		g.Player.Entity.Y)
	for _, n := range g.MapInfo.NPCManager.NPCs {
		s += fmt.Sprintf(
			"%s: [%v, %v] (%v, %v)\n",
			n.DisplayName,
			n.Entity.TilePos.X,
			n.Entity.TilePos.Y,
			n.Entity.X,
			n.Entity.Y,
		)
	}
	ebitenutil.DebugPrint(screen, s)
}

func (g Game) showGameDebugInfo(screen *ebiten.Image) {
	var s strings.Builder

	// show screen size information
	scaleX := math.Round(float64(g.outsideWidth) / float64(display.SCREEN_WIDTH) * 100)
	scaleY := math.Round(float64(g.outsideHeight) / float64(display.SCREEN_HEIGHT) * 100)
	scale := math.Min(scaleX, scaleY)
	s.WriteString(
		fmt.Sprintf(
			"SCREEN SIZE\nvirtual: %v x %v\nreal: %v x %v\nscale: %v%%",
			display.SCREEN_WIDTH, display.SCREEN_HEIGHT, g.outsideWidth, g.outsideHeight, scale,
		))

	ebitenutil.DebugPrint(screen, s.String())
}
