package camera

import (
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
)

// the user's viewport/camera
type Camera struct {
	X, Y float64 // camera position
}

func (c *Camera) MoveCamera(x float64, y float64) {
	// newX := (x - c.X) * 0.1
	// newY := (y - c.Y) * 0.1
	// if math.Abs(c.X-newX) < 0.01 || math.Abs(c.Y-newY) < 0.01 {
	// 	return
	// }
	// c.X += newX
	// c.Y += newY
	c.X = x / config.TileSize
	c.Y = y / config.TileSize
}

func (c *Camera) GetAbsPos() (float64, float64) {
	offsetX := (float64(display.SCREEN_WIDTH) / config.GameScale) / 2
	offsetY := (float64(display.SCREEN_HEIGHT) / config.GameScale) / 2
	return (c.X * config.TileSize) - float64(offsetX), (c.Y * config.TileSize) - float64(offsetY)
}
