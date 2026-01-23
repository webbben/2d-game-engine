// Package image holds core functions for loading images, fonts, etc.
package image

import (
	"fmt"
	"image"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/webbben/2d-game-engine/internal/config"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// LoadImage is the core function for loading an image file into ebiten.
// If the image is fully transparent or has no dimensions (i.e. "empty") then we return nil instead of an ebiten.Image, but no error.
func LoadImage(imgFilePath string) (*ebiten.Image, error) {
	ebitenImg, img, err := ebitenutil.NewImageFromFile(imgFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}
	if IsImageEmpty(img) {
		return nil, nil // no error, but return nil for the image since there's nothing to show
	}
	return ebitenImg, nil
}

func IsImageEmpty(img image.Image) bool {
	if img == nil {
		panic("passed in nil image")
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return true
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			// RGBA() returns values in range [0, 65535]
			if a != 0 {
				return false
			}
		}
	}

	return true
}

func LoadFont(relPath string, size float64, dpi float64) font.Face {
	fullPath := config.ResolveFontPath(relPath)
	fontFile, err := os.ReadFile(fullPath)
	if err != nil {
		panic(err)
	}

	// parse font file
	ttf, err := opentype.Parse(fontFile)
	if err != nil {
		panic(err)
	}

	// default font settings
	op := &opentype.FaceOptions{
		Size:    20,
		DPI:     72,
		Hinting: font.HintingFull, // defaults to none, which apparently makes it smoother
	}
	if size > 0 {
		op.Size = size
	}
	if dpi > 0 {
		op.DPI = dpi
	}

	// create font face
	customFont, err := opentype.NewFace(ttf, op)
	if err != nil {
		panic(err)
	}
	return customFont
}
