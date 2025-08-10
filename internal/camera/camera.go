package camera

import (
	"math"

	"github.com/webbben/2d-game-engine/internal/config"
)

// the user's viewport/camera
type Camera struct {
	X, Y float64 // camera position
}

func (c *Camera) MoveCamera(x float64, y float64) {
	newX := (x - c.X) * 0.1
	newY := (y - c.Y) * 0.1
	if math.Abs(c.X-newX) < 0.01 || math.Abs(c.Y-newY) < 0.01 {
		return
	}
	c.X += (x - c.X) * 0.1
	c.Y += (y - c.Y) * 0.1
}

func (c *Camera) GetAbsPos() (float64, float64) {
	offsetX := ((config.ScreenWidth / config.GameScale) / 2)
	offsetY := ((config.ScreenHeight / config.GameScale) / 2)
	return (c.X * config.TileSize) - float64(offsetX), (c.Y * config.TileSize) - float64(offsetY)
}
