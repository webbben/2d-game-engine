package tileset

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type SpriteComponentPaths struct {
	Skin   string
	Head   string
	Body   string
	Legs   string
	Shadow string
}

func BuildSpriteFrameImage(componentPaths SpriteComponentPaths) (*ebiten.Image, error) {
	// load component images
	skinImage, _, err := ebitenutil.NewImageFromFile(componentPaths.Skin)
	if err != nil {
		return nil, fmt.Errorf("error while loading file at %s: %s", componentPaths.Skin, err)
	}
	headImage, _, err := ebitenutil.NewImageFromFile(componentPaths.Head)
	if err != nil {
		return nil, fmt.Errorf("error while loading file at %s: %s", componentPaths.Head, err)
	}
	bodyImage, _, err := ebitenutil.NewImageFromFile(componentPaths.Body)
	if err != nil {
		return nil, fmt.Errorf("error while loading file at %s: %s", componentPaths.Body, err)
	}
	legsImage, _, err := ebitenutil.NewImageFromFile(componentPaths.Legs)
	if err != nil {
		return nil, fmt.Errorf("error while loading file at %s: %s", componentPaths.Legs, err)
	}
	shadowImage, _, err := ebitenutil.NewImageFromFile(componentPaths.Shadow)
	if err != nil {
		return nil, fmt.Errorf("error while loading file at %s: %s", componentPaths.Shadow, err)
	}

	baseWidth := skinImage.Bounds().Dx() * 2
	baseHeight := skinImage.Bounds().Dy() * 2

	// combine all components into result image
	result := ebiten.NewImage(baseWidth, baseHeight)
	result.Fill(color.RGBA{0, 255, 0, 255}) // Green background

	baseX := float64(baseWidth / 4)
	x := baseX
	baseY := float64((baseHeight / 4) * 3)
	y := baseY

	y = float64(baseHeight - skinImage.Bounds().Dy())

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	result.DrawImage(skinImage, op)

	y = float64(baseHeight - legsImage.Bounds().Dy())
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	result.DrawImage(legsImage, op)

	y -= float64(bodyImage.Bounds().Dy())
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	result.DrawImage(bodyImage, op)

	y -= float64(headImage.Bounds().Dy())
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	result.DrawImage(headImage, op)

	// put shadow on last, starting from the bottom again
	y = float64(baseHeight - shadowImage.Bounds().Dy())
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	result.DrawImage(shadowImage, op)

	return result, nil
}
