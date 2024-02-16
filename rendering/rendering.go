package rendering

import (
	"ancient-rome/config"

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
