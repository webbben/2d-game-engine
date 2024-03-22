// utility functions for rendering images
package rendering

import (
	"github.com/webbben/2d-game-engine/config"

	"github.com/hajimehoshi/ebiten/v2"
)

// gets the absolute position an image should be drawn at if it is to be centered correctly in the given tile-based coordinates
func GetImageDrawPos(image *ebiten.Image, tileX float64, tileY float64, offsetX float64, offsetY float64) (float64, float64) {
	imgWidth := image.Bounds().Dx()
	imgHeight := image.Bounds().Dy()
	drawX := (tileX * float64(config.TileSize)) - offsetX - ((float64(imgWidth) - config.TileSize) / 2)
	drawY := (tileY * float64(config.TileSize)) - offsetY - (float64(imgHeight) - config.TileSize)
	return drawX, drawY
}

// determines if the given tile-based coordinates are within the camera view
func ObjectInsideCameraView(tileX float64, tileY float64, widthAdj, heightAdj float64, offsetX float64, offsetY float64) bool {
	xMin := offsetX
	yMin := offsetY
	xMax := offsetX + (config.ScreenWidth / config.GameScale)
	yMax := offsetY + (config.ScreenHeight / config.GameScale)
	x := tileX * config.TileSize
	y := tileY * config.TileSize
	return x+widthAdj >= xMin && x-widthAdj <= xMax && y+heightAdj >= yMin && y-heightAdj <= yMax
}

// determines if the given tile-based y coordinate (i.e. row) is above the camera view
// if it's above, then that row can skip rendering but the next rows need to continue to be checked
func RowAboveCameraView(tileY float64, offsetY float64) bool {
	y := tileY * config.TileSize
	// offset by one tile above, so we don't see it disappearing
	return y+config.TileSize < offsetY
}

// determines if the given tile-based y coordinate (i.e. row) is above the camera view
// if it's below, then this and all remaining rows can skip rendering
func RowBelowCameraView(tileY float64, offsetY float64) bool {
	yMax := offsetY + (config.ScreenHeight / config.GameScale)
	y := tileY * config.TileSize
	return y > yMax
}

// determines if the given tile-based y coordinate (i.e. row) is within the camera view
func ColBeforeCameraView(tileX float64, offsetX float64) bool {
	x := tileX * config.TileSize
	return x+config.TileSize < offsetX
}

func ColAfterCameraView(tileX float64, offsetX float64) bool {
	x := tileX * config.TileSize
	xMax := offsetX + (config.ScreenWidth / config.GameScale)
	return x > xMax
}
