// utility functions for rendering images
package rendering

import (
	"bytes"
	"image"
	"math"

	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"golang.org/x/image/font"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

// gets the absolute position an image should be drawn at if it is to be centered correctly in the given tile-based coordinates
// TODO - is this being used right? description indicates x and y should be tile coords I think, but I'm pretty sure we are using abs coords.
func GetImageDrawPos(image *ebiten.Image, x float64, y float64, offsetX float64, offsetY float64) (float64, float64) {
	imgWidth := image.Bounds().Dx()
	imgHeight := image.Bounds().Dy()
	drawX := x - offsetX - ((float64(imgWidth) - config.TileSize) / 2)
	drawY := y - offsetY - (float64(imgHeight) - config.TileSize)
	return drawX, drawY
}

func GetRectDrawPos(rect model.Rect, x, y float64, offsetX, offsetY float64) (float64, float64) {
	drawX := x - offsetX - ((float64(rect.W) - config.TileSize) / 2)
	drawY := y - offsetY - (float64(rect.H) - config.TileSize)
	return drawX, drawY
}

// determines if the given tile-based coordinates are within the camera view
func ObjectInsideCameraView(tileX float64, tileY float64, widthAdj, heightAdj float64, offsetX float64, offsetY float64) bool {
	xMin := offsetX
	yMin := offsetY
	xMax := offsetX + (float64(display.SCREEN_WIDTH) / config.GameScale)
	yMax := offsetY + (float64(display.SCREEN_HEIGHT) / config.GameScale)
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
	yMax := offsetY + (float64(display.SCREEN_HEIGHT) / config.GameScale)
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
	xMax := offsetX + (float64(display.SCREEN_WIDTH) / config.GameScale)
	return x > xMax
}

// CenterTextOnImage returns the x and y (offset) coordinates to center the given text on the given image
func CenterTextOnImage(img *ebiten.Image, text string, font font.Face) (int, int) {
	textWidth, textHeight := TextDimensions(text, font)
	x := (img.Bounds().Dx() - textWidth) / 2
	y := (img.Bounds().Dy()-textHeight)/2 + textHeight
	return x, y
}

// TextWidth returns the width of the given text when rendered with the given font
func TextWidth(text string, font font.Face) int {
	width := 0
	for _, r := range text {
		_, advance, _ := font.GlyphBounds(r)
		width += advance.Ceil()
	}
	return width
}

func TextHeight(text string, font font.Face) int {
	_, h := font.Metrics().Height.Ceil(), font.Metrics().Descent.Ceil()
	return h
}

func TextDimensions(text string, font font.Face) (int, int) {
	return TextWidth(text, font), TextHeight(text, font)
}

// CenterImageOnImage returns the x and y (offset) coordinates to center the given image on the given background image
func CenterImageOnImage(bg *ebiten.Image, img *ebiten.Image) (int, int) {
	x := (bg.Bounds().Dx() - img.Bounds().Dx()) / 2
	y := (bg.Bounds().Dy() - img.Bounds().Dy()) / 2
	return x, y
}

// Blending:
// https://ebitengine.org/en/examples/blend.html

func CropImageByOtherImage(img, otherImage *ebiten.Image) *ebiten.Image {
	result := ebiten.NewImage(img.Bounds().Dx(), img.Bounds().Dy())
	result.DrawImage(img, nil)

	ops := ebiten.DrawImageOptions{}
	ops.Blend = ebiten.BlendDestinationIn
	result.DrawImage(otherImage, &ops)

	return result
}

func SubtractImageByOtherImage(img, otherImage *ebiten.Image, imgOffsetY, otherOffsetY int) *ebiten.Image {
	if img == nil {
		panic("img is nil")
	}
	if otherImage == nil {
		panic("otherImage is nil")
	}
	if IsImageEmpty(img) {
		logz.Panicln("SubtractImageByOtherImage", "tried to subtract from an empty image!")
	}
	if IsImageEmpty(otherImage) {
		logz.Panicln("SubtractImageByOtherImage", "tried to subtract by an empty image!")
	}
	result := ebiten.NewImage(img.Bounds().Dx(), img.Bounds().Dy())
	imgOps := ebiten.DrawImageOptions{}
	if imgOffsetY != 0 {
		imgOps.GeoM.Translate(0, float64(imgOffsetY))
	}
	result.DrawImage(img, &imgOps)

	otherOps := ebiten.DrawImageOptions{}
	if otherOffsetY != 0 {
		otherOps.GeoM.Translate(0, float64(otherOffsetY))
	}
	otherOps.Blend = ebiten.BlendDestinationOut
	result.DrawImage(otherImage, &otherOps)

	if ImagesEqual(img, result) {
		logz.Panicln("SubtractImageByOtherImage", "original image subtracted from appears to be the same as the result image (subtraction doesn't seem to have worked)")
	}

	return result
}

// ImagesEqual checks if the two images have equal image data (i.e. are the same)
func ImagesEqual(a, b *ebiten.Image) bool {
	if a == nil || b == nil {
		logz.Panicln("ImagesEqual", "one of the images passed in was nil")
	}

	wA, hA := a.Bounds().Dx(), a.Bounds().Dy()
	wB, hB := b.Bounds().Dx(), b.Bounds().Dy()
	if wA != wB || hA != hB {
		return false
	}

	pixelsA := make([]byte, 4*wA*hA)
	pixelsB := make([]byte, 4*wB*hB)

	a.ReadPixels(pixelsA)
	b.ReadPixels(pixelsB)
	return bytes.Equal(pixelsA, pixelsB)
}

// IsImageEmpty checks if an image has any non-transparent pixels.
// WARNING: Not performance friendly. Do not use in draw or repeatedly in long lasting loops. Just for short term validation.
func IsImageEmpty(img *ebiten.Image) bool {
	if img == nil {
		panic("passed in nil image")
	}

	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	if w == 0 || h == 0 {
		return true
	}

	pixels := make([]byte, 4*w*h)
	img.ReadPixels(pixels)

	for i := 3; i < len(pixels); i += 4 {
		if pixels[i] != 0 {
			return false
		}
	}

	return true
}

func DrawHueRotatedImage(screen, img *ebiten.Image, sliderValue float64, x, y, scale float64) {
	hueShift := sliderValue * 2 * math.Pi

	op := &colorm.DrawImageOptions{}

	if scale > 0 {
		op.GeoM.Scale(scale, scale)
	}
	// important: if translate is above scale, it will come out weird
	// I guess this is because these effects are applied in order
	op.GeoM.Translate(x, y)

	var c colorm.ColorM
	c.RotateHue(hueShift)
	colorm.DrawImage(screen, img, c, op)
}

func DrawHSVImage(screen, img *ebiten.Image, h, s, v float64, x, y, scale float64) {
	hue := (h - 0.5) * 2 * math.Pi // rotate -180° to +180°
	sat := s * 2.0                 // allow up to 2× saturation
	val := v * 2.0                 // allow up to 2× brightness

	op := &colorm.DrawImageOptions{}
	if scale > 0 {
		op.GeoM.Scale(scale, scale)
	}
	// important: if translate is above scale, it will come out weird
	// I guess this is because these effects are applied in order
	op.GeoM.Translate(x, y)

	var c colorm.ColorM
	c.ChangeHSV(hue, sat, val)
	colorm.DrawImage(screen, img, c, op)
}

func StretchImage(img *ebiten.Image, dx, dy int) *ebiten.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	newWidth := width + dx
	newHeight := height + dy

	scaleX := float64(newWidth) / float64(width)
	scaleY := float64(newHeight) / float64(height)

	return ScaleImage(img, scaleX, scaleY)
}

func StretchMiddle(src *ebiten.Image) *ebiten.Image {
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	dst := ebiten.NewImage(w, h)

	// Left half: columns 0–7 → destination 0–7
	leftRect := image.Rect(0, 0, w/2, h)
	DrawImage(dst, src.SubImage(leftRect).(*ebiten.Image), -1, 0, 0)

	// Right half: columns 8–15 → destination 9–16
	rightRect := image.Rect(w/2, 0, w, h)
	DrawImage(dst, src.SubImage(rightRect).(*ebiten.Image), float64(w/2)+1, 0, 0)

	// Fill the middle 2 columns (original cols 7 & 8)
	middleRect := image.Rect(w/2-1, 0, w/2+1, h)
	DrawImage(dst, src.SubImage(middleRect).(*ebiten.Image), float64(w/2)-1, 0, 0)

	return dst
}
