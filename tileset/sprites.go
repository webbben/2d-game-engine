package tileset

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type SpriteComponent struct {
	ImagePath string
	Dx        int // adjust placement on X axis
	Dy        int // adjust placement on Y axis. Negative moves down, positive moves up (opposite to how Y coordinates actually work in ebiten)
}

type SpriteComponentPaths struct {
	Skin   SpriteComponent
	Head   SpriteComponent
	Body   SpriteComponent
	Legs   SpriteComponent
	Shadow SpriteComponent
}

// BuildSpriteFrameImage builds a single frame image, given the component images.
//
// Note: This function is a bit oddly made as of now. It probably could be simplified, but for now I'll leave it as is since it works.
func BuildSpriteFrameImage(components SpriteComponentPaths) (*ebiten.Image, error) {
	// load component images
	skinImage, _, err := ebitenutil.NewImageFromFile(components.Skin.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("error while loading file at %s: %s", components.Skin.ImagePath, err)
	}
	headImage, _, err := ebitenutil.NewImageFromFile(components.Head.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("error while loading file at %s: %s", components.Head.ImagePath, err)
	}
	bodyImage, _, err := ebitenutil.NewImageFromFile(components.Body.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("error while loading file at %s: %s", components.Body.ImagePath, err)
	}
	legsImage, _, err := ebitenutil.NewImageFromFile(components.Legs.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("error while loading file at %s: %s", components.Legs.ImagePath, err)
	}
	shadowImage, _, err := ebitenutil.NewImageFromFile(components.Shadow.ImagePath)
	if err != nil {
		return nil, fmt.Errorf("error while loading file at %s: %s", components.Shadow.ImagePath, err)
	}

	// get the width of the widest component
	maxWidth := skinImage.Bounds().Dx()
	if shadowImage.Bounds().Dx() > maxWidth {
		maxWidth = shadowImage.Bounds().Dx()
	}
	if headImage.Bounds().Dx() > maxWidth {
		maxWidth = headImage.Bounds().Dx()
	}
	if bodyImage.Bounds().Dx() > maxWidth {
		maxWidth = bodyImage.Bounds().Dx()
	}
	if legsImage.Bounds().Dx() > maxWidth {
		maxWidth = legsImage.Bounds().Dx()
	}
	baseHeight := skinImage.Bounds().Dy() * 2

	// the center x position that all component centers must match to
	centerX := float64(maxWidth) / 2

	// Skin
	y := float64(baseHeight-skinImage.Bounds().Dy()) + float64(components.Skin.Dy)
	dx := centerX - float64(skinImage.Bounds().Dx()/2) // difference in center of component vs global center
	x := dx + float64(components.Skin.Dx) - 1
	skinX, skinY := x, y

	// Legs
	y = float64(baseHeight-legsImage.Bounds().Dy()) + float64(components.Legs.Dy)
	dx = centerX - float64(legsImage.Bounds().Dx()/2)
	x = dx + float64(components.Legs.Dx) - 1
	legsX, legsY := x, y

	// Body
	y -= float64(bodyImage.Bounds().Dy()) + float64(components.Body.Dy)
	dx = centerX - float64(bodyImage.Bounds().Dx()/2)
	x = dx + float64(components.Body.Dx) - 1
	bodyX, bodyY := x, y

	// Head
	y -= float64(headImage.Bounds().Dy()) + float64(components.Head.Dy)
	dx = centerX - float64(headImage.Bounds().Dx()/2)
	x = dx + float64(components.Head.Dx) - 1
	headX, headY := x, y

	// put shadow on last, starting from the bottom again
	// putting on last since shadow may cover other things, like clothes too
	y = float64(baseHeight-shadowImage.Bounds().Dy()) + 2 // +2: shadow goes a little below the rest of the sprite
	dx = centerX - float64(shadowImage.Bounds().Dx()/2)
	x = dx + float64(components.Shadow.Dx) - 1
	shadowX, shadowY := x, y

	// We've calculated all the dimensions, now actually put it all together
	totalHeight := float64(baseHeight+2) - headY
	result := ebiten.NewImage(maxWidth, int(totalHeight)) // +2: adds a little extra space at the bottom for the shadow
	result.Fill(color.RGBA{0, 255, 0, 255})               // Green background

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(skinX, skinY-headY)
	result.DrawImage(skinImage, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(legsX, legsY-headY)
	result.DrawImage(legsImage, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(bodyX, bodyY-headY)
	result.DrawImage(bodyImage, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(headX, headY-headY)
	result.DrawImage(headImage, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(shadowX, shadowY-headY)
	result.DrawImage(shadowImage, op)

	return result, nil
}
