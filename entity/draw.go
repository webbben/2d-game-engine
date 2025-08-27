package entity

import (
	"errors"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type ImagePart struct {
	Image   *ebiten.Image
	OffsetX int // amount to offset the image from the center
	PosY    int // the Y position (top is 0)
}

// TODO use this to create individual entity frames from smaller, reusable pieces
func CreateImageFromParts(imageParts []ImagePart, widthPx, heightPx int) (*ebiten.Image, error) {
	// get the width of the widest component
	maxWidth := 0
	for _, imagePart := range imageParts {
		if imagePart.Image.Bounds().Dx() > maxWidth {
			maxWidth = imagePart.Image.Bounds().Dx()
		}
	}

	if maxWidth > widthPx {
		return nil, errors.New("width of image parts exceeds the set image width")
	}

	// the center x position that all component centers must match to
	centerX := float64(maxWidth) / 2

	bg := ebiten.NewImage(widthPx, heightPx) // +2: adds a little extra space at the bottom for the shadow
	// green = color.RGBA{0, 255, 0, 255}
	bg.Fill(color.RGBA{150, 150, 150, 255}) // gray background

	for _, imagePart := range imageParts {
		op := &ebiten.DrawImageOptions{}
		x := centerX - float64(imagePart.Image.Bounds().Dx()/2) + float64(imagePart.OffsetX)
		op.GeoM.Translate(x, float64(imagePart.PosY))
		bg.DrawImage(imagePart.Image, op)
	}

	return bg, nil
}
