// Package camera contains logic for controlling the camera in a map
package camera

import (
	"math"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/display"
)

// Camera is the user's viewport/camera in a map
type Camera struct {
	X, Y       float64 // camera position. I believe this is tile position, not absolute pixels.
	MaxX, MaxY float64 // max positions the camera can go to. for keeping camera from going off edge of map.
	MinX, MinY float64 // min positions the camera can go to.
}

func (c *Camera) MoveCamera(x float64, y float64) {
	newX := x / config.TileSize
	newY := y / config.TileSize

	// ensure it's between [0, Max]
	newX = min(max(c.MinX, newX), c.MaxX)
	newY = min(max(c.MinY, newY), c.MaxY)

	dx := (newX - c.X) * 0.2
	dy := (newY - c.Y) * 0.2

	if math.Abs(dx) < 0.01 && math.Abs(dy) < 0.01 {
		return
	}

	c.X += dx
	c.Y += dy
}

func (c *Camera) SetMapLimits(mapTileWidth, mapTileHeight int) {
	if mapTileWidth < 1 {
		panic("map width was less than 1")
	}
	if mapTileHeight < 1 {
		panic("map height was less than 1")
	}

	cameraTileWidth := display.SCREEN_WIDTH / int(config.TileSize*config.GameScale)
	cameraTileHeight := display.SCREEN_HEIGHT / int(config.TileSize*config.GameScale)

	if mapTileWidth <= cameraTileWidth {
		// if map is too small, then we should just try to center the camera in the middle of the map
		c.MaxX = float64(mapTileWidth) / 2
		c.MinX = c.MaxX
	} else {
		// camera's position corresponds to the middle of the visible screen; so, we should set the max position to
		// the map max, minus half the camera's width and height
		c.MaxX = max(float64(mapTileWidth)-float64(cameraTileWidth/2), 0)
		c.MinX = min(float64(cameraTileWidth)/2, c.MaxX)
	}
	if mapTileHeight <= cameraTileHeight {
		c.MaxY = float64(mapTileHeight) / 2
		c.MinY = c.MaxY
	} else {
		c.MaxY = max(float64(mapTileHeight)-float64(cameraTileHeight/2), 0)
		c.MinY = min(float64(cameraTileHeight)/2, c.MaxY)
	}
}

func (c *Camera) SetCameraPosition(x, y float64) {
	c.X = x / config.TileSize
	c.Y = y / config.TileSize
	c.X = min(max(c.MinX, c.X), c.MaxX)
	c.Y = min(max(c.MinY, c.Y), c.MaxY)
}

func (c *Camera) GetAbsPos() (float64, float64) {
	offsetX := (float64(display.SCREEN_WIDTH) / config.GameScale) / 2
	offsetY := (float64(display.SCREEN_HEIGHT) / config.GameScale) / 2
	return (c.X * config.TileSize) - float64(offsetX), (c.Y * config.TileSize) - float64(offsetY)
}
