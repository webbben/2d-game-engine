package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// draws an image on the screen, without using any fancy options
func DrawWorldImage(screen *ebiten.Image, img *ebiten.Image, x, y, offsetX, offsetY float64, scale float64) {
	if screen == nil {
		panic("DrawImage: screen is nil!")
	}
	if img == nil {
		panic("DrawImage: image to draw is nil!")
	}
	op := &ebiten.DrawImageOptions{}
	drawX, drawY := GetImageDrawPos(img, x, y, offsetX, offsetY)
	op.GeoM.Translate(drawX, drawY)
	if scale > 0 {
		op.GeoM.Scale(scale, scale)
	}
	screen.DrawImage(img, op)
}

func DrawImage(screen *ebiten.Image, img *ebiten.Image, x, y, scale float64) {
	if screen == nil {
		panic("DrawImage: screen is nil!")
	}
	if img == nil {
		panic("DrawImage: image to draw is nil!")
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	if scale > 0 {
		op.GeoM.Scale(scale, scale)
	}
	screen.DrawImage(img, op)
}
