package activemap

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
)

type debugData struct {
	positionDot    *ebiten.Image
	rectImg        *ebiten.Image
	activateRect   *ebiten.Image
	costTile       *ebiten.Image
	collisionRects map[string]*ebiten.Image

	pathTile1, pathTile2 *ebiten.Image
}

func (m ActiveMap) drawGridLines(screen *ebiten.Image, offsetX float64, offsetY float64) {
	offsetX = offsetX * config.GameScale
	offsetY = offsetY * config.GameScale
	lineColor := color.RGBA{255, 0, 0, 255}
	maxWidth := float64(m.Map.Width*m.Map.TileWidth) * config.GameScale
	maxHeight := float64(m.Map.Height*m.Map.TileHeight) * config.GameScale

	// vertical lines
	for x := 0; x < m.Map.Width; x++ {
		drawX := (float64(x*config.TileSize) * config.GameScale) - offsetX
		drawY := -offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(drawX), float32(maxHeight-offsetY), 1, lineColor, true)
	}
	// horizontal lines
	for y := 0; y < m.Map.Height; y++ {
		drawX := -offsetX
		drawY := (float64(y*config.TileSize) * config.GameScale) - offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(maxWidth-offsetX), float32(drawY), 1, lineColor, true)
	}
}

func (m *ActiveMap) drawPaths(screen *ebiten.Image, offsetX, offsetY float64) {
	if m.debugData.pathTile1 == nil {
		color1 := color.RGBA{255, 0, 0, 50}
		m.debugData.pathTile1 = ebiten.NewImage(config.TileSize, config.TileSize)
		m.debugData.pathTile1.Fill(color1)
	}
	if m.debugData.pathTile2 == nil {
		m.debugData.pathTile2 = ebiten.NewImage(config.TileSize, config.TileSize)
		color2 := color.RGBA{0, 0, 255, 50}
		m.debugData.pathTile2.Fill(color2)
	}

	for _, n := range m.NPCs {
		if len(n.Entity.Movement.TargetPath) > 0 {
			for _, c := range n.Entity.Movement.TargetPath {
				op := &ebiten.DrawImageOptions{}
				drawX, drawY := rendering.GetImageDrawPos(m.debugData.pathTile1, float64(c.X)*config.TileSize, float64(c.Y)*config.TileSize, offsetX, offsetY)
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(m.debugData.pathTile1, op)
			}
		}
		if len(n.Entity.Movement.SuggestedTargetPath) > 0 {
			for _, c := range n.Entity.Movement.SuggestedTargetPath {
				op := &ebiten.DrawImageOptions{}
				drawX, drawY := rendering.GetImageDrawPos(m.debugData.pathTile2, float64(c.X)*config.TileSize, float64(c.Y)*config.TileSize, offsetX, offsetY)
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(m.debugData.pathTile2, op)
			}
		}
	}
}

func (m *ActiveMap) drawCollisions(screen *ebiten.Image, offsetX, offsetY float64) {
	if m.debugData.costTile == nil {
		m.debugData.costTile = ebiten.NewImage(config.TileSize, config.TileSize)
		color1 := color.RGBA{150, 0, 0, 50}
		m.debugData.costTile.Fill(color1)
	}
	if m.debugData.collisionRects == nil {
		m.debugData.collisionRects = make(map[string]*ebiten.Image)
	}

	color2 := color.RGBA{0, 0, 150, 50}

	for y, row := range m.CostMap() {
		for x, cost := range row {
			if cost >= 10 {
				op := &ebiten.DrawImageOptions{}
				drawX, drawY := rendering.GetImageDrawPos(m.debugData.costTile, float64(x)*config.TileSize, float64(y)*config.TileSize, offsetX, offsetY)
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(m.debugData.costTile, op)
			}

			r := m.Map.CollisionRects[y][x]
			if r.IsCollision {
				key := fmt.Sprintf("w:%v/h:%v", r.Rect.W, r.Rect.H)
				rectImg, exists := m.debugData.collisionRects[key]
				if !exists {
					rectImg = ebiten.NewImage(int(r.Rect.W), int(r.Rect.H))
					rectImg.Fill(color2)
					m.debugData.collisionRects[key] = rectImg
				}

				op := &ebiten.DrawImageOptions{}
				drawX, drawY := (float64(x*config.TileSize)+r.Rect.X)-offsetX, (float64(y*config.TileSize)+r.Rect.Y)-offsetY
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(rectImg, op)
			}
		}
	}

	for _, obj := range m.Objects {
		if !obj.IsCollidable() {
			continue
		}
		objRect := obj.GetRect()
		key := fmt.Sprintf("obj:%s", objRect)
		rectImg, exists := m.debugData.collisionRects[key]
		if !exists {
			rectImg = ebiten.NewImage(int(objRect.W), int(objRect.H))
			rectImg.Fill(color2)
			m.debugData.collisionRects[key] = rectImg
		}
		op := &ebiten.DrawImageOptions{}
		drawX, drawY := objRect.X-offsetX, objRect.Y-offsetY
		op.GeoM.Translate(drawX, drawY)
		op.GeoM.Scale(config.GameScale, config.GameScale)
		screen.DrawImage(rectImg, op)
	}
}

func (m *ActiveMap) drawEntityPositions(screen *ebiten.Image, offsetX, offsetY float64) {
	if m.debugData.positionDot == nil {
		yellow := color.RGBA{0, 255, 255, 50}
		m.debugData.positionDot = ebiten.NewImage(1, 1)
		m.debugData.positionDot.Fill(yellow)
	}
	if m.debugData.rectImg == nil {
		rect := m.PlayerRef.Entity.CollisionRect()
		m.debugData.rectImg = ebiten.NewImage(int(rect.W), int(rect.H))
		m.debugData.rectImg.Fill(color.RGBA{0, 0, 255, 50})
	}
	if m.debugData.activateRect == nil {
		rect := m.PlayerRef.Entity.GetFrontRect()
		m.debugData.activateRect = ebiten.NewImage(int(rect.W), int(rect.H))
		m.debugData.activateRect.Fill(color.RGBA{255, 0, 255, 50})
	}

	// tile2 := ebiten.NewImage(config.TileSize, config.TileSize)
	// color2 := color.RGBA{0, 0, 255, 50}
	// tile2.Fill(color2)

	op := &ebiten.DrawImageOptions{}
	drawX, drawY := m.PlayerRef.Entity.X-offsetX, m.PlayerRef.Entity.Y-offsetY
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(m.debugData.positionDot, op)

	op = &ebiten.DrawImageOptions{}
	rect := m.PlayerRef.Entity.CollisionRect()
	drawX, drawY = rect.X-offsetX, rect.Y-offsetY
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(m.debugData.rectImg, op)

	op = &ebiten.DrawImageOptions{}
	rect = m.PlayerRef.Entity.GetFrontRect()
	drawX, drawY = rect.X-offsetX, rect.Y-offsetY
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(m.debugData.activateRect, op)

	for _, n := range m.NPCs {
		op := &ebiten.DrawImageOptions{}
		// drawX, drawY := rendering.GetImageDrawPos(g.debugData.positionDot, n.Entity.X, n.Entity.Y, offsetX, offsetY)
		drawX, drawY := n.Entity.X-offsetX, n.Entity.Y-offsetY
		op.GeoM.Translate(drawX, drawY)
		op.GeoM.Scale(config.GameScale, config.GameScale)
		screen.DrawImage(m.debugData.positionDot, op)
	}
}

func (m ActiveMap) ShowEntityCoords() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		"Player pos: [%v, %v] (%v, %v)\n",
		m.PlayerRef.Entity.TilePos().X,
		m.PlayerRef.Entity.TilePos().Y,
		m.PlayerRef.Entity.X,
		m.PlayerRef.Entity.Y))
	for _, n := range m.NPCs {
		sb.WriteString(fmt.Sprintf(
			"%s: [%v, %v] (%v, %v)\n",
			n.DisplayName(),
			n.Entity.TilePos().X,
			n.Entity.TilePos().Y,
			n.Entity.X,
			n.Entity.Y,
		))
	}
	return sb.String()
}

func (m ActiveMap) GetDaylightData(s *strings.Builder) {
	lightColor := m.daylightFader.GetCurrentColor()
	fmt.Fprintf(s, "daylight (RGB scales): [%v %v %v]\n", lightColor[0], lightColor[1], lightColor[2])
	fmt.Fprintf(s, "darkness factor: %v", m.daylightFader.GetDarknessFactor())
}
