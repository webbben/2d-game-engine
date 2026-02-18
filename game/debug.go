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
	"github.com/webbben/2d-game-engine/internal/debug"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

type debugData struct {
	positionDot    *ebiten.Image
	rectImg        *ebiten.Image
	costTile       *ebiten.Image
	collisionRects map[string]*ebiten.Image

	pathTile1, pathTile2 *ebiten.Image
}

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
	if g.debugData.pathTile1 == nil {
		color1 := color.RGBA{255, 0, 0, 50}
		g.debugData.pathTile1 = ebiten.NewImage(config.TileSize, config.TileSize)
		g.debugData.pathTile1.Fill(color1)
	}
	if g.debugData.pathTile2 == nil {
		g.debugData.pathTile2 = ebiten.NewImage(config.TileSize, config.TileSize)
		color2 := color.RGBA{0, 0, 255, 50}
		g.debugData.pathTile2.Fill(color2)
	}

	for _, n := range g.MapInfo.NPCs {
		if len(n.Entity.Movement.TargetPath) > 0 {
			for _, c := range n.Entity.Movement.TargetPath {
				op := &ebiten.DrawImageOptions{}
				drawX, drawY := rendering.GetImageDrawPos(g.debugData.pathTile1, float64(c.X)*config.TileSize, float64(c.Y)*config.TileSize, offsetX, offsetY)
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(g.debugData.pathTile1, op)
			}
		}
		if len(n.Entity.Movement.SuggestedTargetPath) > 0 {
			for _, c := range n.Entity.Movement.SuggestedTargetPath {
				op := &ebiten.DrawImageOptions{}
				drawX, drawY := rendering.GetImageDrawPos(g.debugData.pathTile2, float64(c.X)*config.TileSize, float64(c.Y)*config.TileSize, offsetX, offsetY)
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(g.debugData.pathTile2, op)
			}
		}
	}
}

func (g Game) drawCollisions(screen *ebiten.Image, offsetX, offsetY float64) {
	if g.MapInfo == nil {
		panic("tried to draw paths on a nil map!")
	}
	if g.debugData.costTile == nil {
		g.debugData.costTile = ebiten.NewImage(config.TileSize, config.TileSize)
		color1 := color.RGBA{150, 0, 0, 50}
		g.debugData.costTile.Fill(color1)
	}
	if g.debugData.collisionRects == nil {
		g.debugData.collisionRects = make(map[string]*ebiten.Image)
	}

	color2 := color.RGBA{0, 0, 150, 50}

	for y, row := range g.MapInfo.CostMap() {
		for x, cost := range row {
			if cost >= 10 {
				op := &ebiten.DrawImageOptions{}
				drawX, drawY := rendering.GetImageDrawPos(g.debugData.costTile, float64(x)*config.TileSize, float64(y)*config.TileSize, offsetX, offsetY)
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(g.debugData.costTile, op)
			}

			r := g.MapInfo.Map.CollisionRects[y][x]
			if r.IsCollision {
				key := fmt.Sprintf("w:%v/h:%v", r.Rect.W, r.Rect.H)
				rectImg, exists := g.debugData.collisionRects[key]
				if !exists {
					rectImg = ebiten.NewImage(int(r.Rect.W), int(r.Rect.H))
					rectImg.Fill(color2)
					g.debugData.collisionRects[key] = rectImg
				}

				op := &ebiten.DrawImageOptions{}
				drawX, drawY := (float64(x*config.TileSize)+r.Rect.X)-offsetX, (float64(y*config.TileSize)+r.Rect.Y)-offsetY
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(rectImg, op)
			}
		}
	}

	for _, obj := range g.MapInfo.Objects {
		if !obj.IsCollidable() {
			continue
		}
		objRect := obj.GetRect()
		key := fmt.Sprintf("obj:%s", objRect)
		rectImg, exists := g.debugData.collisionRects[key]
		if !exists {
			rectImg = ebiten.NewImage(int(objRect.W), int(objRect.H))
			rectImg.Fill(color2)
			g.debugData.collisionRects[key] = rectImg
		}
		op := &ebiten.DrawImageOptions{}
		drawX, drawY := objRect.X-offsetX, objRect.Y-offsetY
		op.GeoM.Translate(drawX, drawY)
		op.GeoM.Scale(config.GameScale, config.GameScale)
		screen.DrawImage(rectImg, op)
	}
}

func (g *Game) drawEntityPositions(screen *ebiten.Image, offsetX, offsetY float64) {
	if g.MapInfo == nil {
		panic("tried to draw paths on a nil map!")
	}
	if g.debugData.positionDot == nil {
		yellow := color.RGBA{0, 255, 255, 50}
		g.debugData.positionDot = ebiten.NewImage(1, 1)
		g.debugData.positionDot.Fill(yellow)
	}
	if g.debugData.rectImg == nil {
		rect := g.Player.Entity.CollisionRect()
		g.debugData.rectImg = ebiten.NewImage(int(rect.W), int(rect.H))
		g.debugData.rectImg.Fill(color.RGBA{0, 0, 255, 50})
	}

	// tile2 := ebiten.NewImage(config.TileSize, config.TileSize)
	// color2 := color.RGBA{0, 0, 255, 50}
	// tile2.Fill(color2)

	op := &ebiten.DrawImageOptions{}
	drawX, drawY := g.Player.Entity.X-offsetX, g.Player.Entity.Y-offsetY
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(g.debugData.positionDot, op)

	op = &ebiten.DrawImageOptions{}
	rect := g.Player.Entity.CollisionRect()
	drawX, drawY = rect.X-offsetX, rect.Y-offsetY
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(g.debugData.rectImg, op)

	for _, n := range g.MapInfo.NPCs {
		op := &ebiten.DrawImageOptions{}
		drawX, drawY := rendering.GetImageDrawPos(g.debugData.positionDot, n.Entity.X, n.Entity.Y, offsetX, offsetY)
		op.GeoM.Translate(drawX, drawY)
		op.GeoM.Scale(config.GameScale, config.GameScale)
		screen.DrawImage(g.debugData.positionDot, op)
	}
}

func (g Game) showEntityCoords() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"Player pos: [%v, %v] (%v, %v)\n",
		g.Player.Entity.TilePos().X,
		g.Player.Entity.TilePos().Y,
		g.Player.Entity.X,
		g.Player.Entity.Y))
	for _, n := range g.MapInfo.NPCs {
		sb.WriteString(fmt.Sprintf(
			"%s: [%v, %v] (%v, %v)\n",
			n.DisplayName,
			n.Entity.TilePos().X,
			n.Entity.TilePos().Y,
			n.Entity.X,
			n.Entity.Y,
		))
	}
	return sb.String()
}

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

	if config.ShowPlayerCoords {
		s.WriteString("ENTITY POSITIONS\n")
		s.WriteString(g.showEntityCoords())
	}

	s.WriteString(fmt.Sprintf("TIME: %s\n", g.Clock))
	lightColor := g.daylightFader.GetCurrentColor()
	s.WriteString(fmt.Sprintf("daylight (RGB scales): [%v %v %v]\n", lightColor[0], lightColor[1], lightColor[2]))
	s.WriteString(fmt.Sprintf("darkness factor: %v", g.daylightFader.GetDarknessFactor()))

	ebitenutil.DebugPrint(screen, s.String())
}
