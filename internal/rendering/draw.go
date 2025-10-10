package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
)

/*
c_src: source RGB values
c_dst: destination RGB values
c_out: result RGB values
α_src: source alpha values
α_dst: destination alpha values
α_out: result alpha values

c_out = BlendOperationRGB((BlendFactorSourceRGB) × c_src, (BlendFactorDestinationRGB) × c_dst)
α_out = BlendOperationAlpha((BlendFactorSourceAlpha) × α_src, (BlendFactorDestinationAlpha) × α_dst)
A blend factor is a factor for source and color destination color values. The default is source-over (regular alpha blending).

A blend operation is a binary operator of a source color and a destination color. The default is adding.
*/

var BlendMultiply ebiten.Blend = ebiten.Blend{
	// BlendFactorSourceRGB:        ebiten.BlendFactorSourceColor,
	// BlendFactorSourceAlpha:      ebiten.BlendFactorOne,
	// BlendFactorDestinationRGB:   ebiten.BlendFactorDestinationColor,
	// BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
	// BlendOperationRGB:           ebiten.BlendOperationAdd,
	// BlendOperationAlpha:         ebiten.BlendOperationAdd,
	BlendFactorSourceRGB:        ebiten.BlendFactorOne,
	BlendFactorSourceAlpha:      ebiten.BlendFactorOne,
	BlendFactorDestinationRGB:   ebiten.BlendFactorDestinationColor,
	BlendFactorDestinationAlpha: ebiten.BlendFactorOneMinusSourceAlpha,
	BlendOperationRGB:           ebiten.BlendOperationAdd,
	BlendOperationAlpha:         ebiten.BlendOperationAdd,
}

// draws an image in the map, using the standard settings to do so.
// can optionally pass in special options, but they should not mess with the following:
//
// # GeoM.Translate - already using this to place the object at its position
//
// GeoM.Scale - already setting this to the game scale set in config.
func DrawWorldImage(screen *ebiten.Image, img *ebiten.Image, x, y, offsetX, offsetY float64, op *ebiten.DrawImageOptions) {
	if screen == nil {
		panic("DrawImage: screen is nil!")
	}
	if img == nil {
		panic("DrawImage: image to draw is nil!")
	}
	if op == nil {
		op = &ebiten.DrawImageOptions{}
	}
	drawX, drawY := GetImageDrawPos(img, x, y, offsetX, offsetY)
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)

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

	if scale > 0 {
		op.GeoM.Scale(scale, scale)
	}
	// important: if translate is above scale, it will come out weird
	// I guess this is because these effects are applied in order
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func DrawImageWithOps(screen *ebiten.Image, img *ebiten.Image, x, y, scale float64, op *ebiten.DrawImageOptions) {
	if screen == nil {
		panic("DrawImage: screen is nil!")
	}
	if img == nil {
		panic("DrawImage: image to draw is nil!")
	}

	if op == nil {
		op = &ebiten.DrawImageOptions{}
	}
	if scale > 0 {
		op.GeoM.Scale(scale, scale)
	}
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

// scales the image by the given scale factor
func ScaleImage(img *ebiten.Image, scale float64) *ebiten.Image {
	bounds := img.Bounds()

	dx := float64(bounds.Dx()) * scale
	dy := float64(bounds.Dy()) * scale
	frame := ebiten.NewImage(int(dx), int(dy))

	ops := ebiten.DrawImageOptions{}
	ops.GeoM.Scale(scale, scale)
	frame.DrawImage(img, &ops)

	return frame
}
