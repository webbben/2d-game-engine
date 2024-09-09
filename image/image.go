package image

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// loads an individual image
func LoadImage(imagePath string) (*ebiten.Image, error) {
	img, _, err := ebitenutil.NewImageFromFile(imagePath)
	if err != nil {
		return nil, err
	}
	return img, nil
}
