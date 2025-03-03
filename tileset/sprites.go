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

	baseWidth := skinImage.Bounds().Dx() * 2
	baseHeight := skinImage.Bounds().Dy() * 2

	// combine all components into result image
	result := ebiten.NewImage(baseWidth, baseHeight+2) // +2: adds a little extra space at the bottom for the shadow
	result.Fill(color.RGBA{0, 255, 0, 255})            // Green background

	baseX := float64(baseWidth / 4)

	// the center x position that all component centers must match to
	centerX := baseX + (float64(maxWidth) / 2)

	// Skin
	y := float64(baseHeight-skinImage.Bounds().Dy()) + float64(components.Skin.Dy)
	dx := centerX - (float64(skinImage.Bounds().Dx()/2) + baseX) // difference in center of component vs global center
	x := baseX + dx + float64(components.Skin.Dx)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	result.DrawImage(skinImage, op)

	// Legs
	y = float64(baseHeight-legsImage.Bounds().Dy()) + float64(components.Legs.Dy)
	dx = centerX - (float64(legsImage.Bounds().Dx()/2) + baseX)
	x = baseX + dx + float64(components.Legs.Dx)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	result.DrawImage(legsImage, op)

	// Body
	y -= float64(bodyImage.Bounds().Dy()) + float64(components.Body.Dy)
	dx = centerX - (float64(bodyImage.Bounds().Dx()/2) + baseX)
	x = baseX + dx + float64(components.Body.Dx)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	result.DrawImage(bodyImage, op)

	y -= float64(headImage.Bounds().Dy()) + float64(components.Head.Dy)
	dx = centerX - (float64(headImage.Bounds().Dx()/2) + baseX)
	x = baseX + dx + float64(components.Head.Dx)
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	result.DrawImage(headImage, op)

	// put shadow on last, starting from the bottom again
	// putting on last since shadow may cover other things, like clothes too
	y = float64(baseHeight-shadowImage.Bounds().Dy()) + 2 // +2: shadow goes a little below the rest of the sprite
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	result.DrawImage(shadowImage, op)

	return result, nil
}
