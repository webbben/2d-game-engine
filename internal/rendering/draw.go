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

// DrawWorldImage draws an image in the map, using the standard settings to do so.
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

	drawImage(screen, img, op)
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
	if x != 0 || y != 0 {
		op.GeoM.Translate(x, y)
	}
	drawImage(screen, img, op)
}

// DrawImageOnlyOps allows you to draw an image only providing the ops data - just a wrapper around the core drawImage function.
// TODO: should we delete this? Only made this for one use case in map.go which we need to investigate more.
func DrawImageOnlyOps(screen *ebiten.Image, img *ebiten.Image, ops *ebiten.DrawImageOptions) {
	drawImage(screen, img, ops)
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
	drawImage(screen, img, op)
}

// drawImage is the core drawing image function. this should always be used when directly drawing something, since it includes the basic guardrails.
func drawImage(screen *ebiten.Image, img *ebiten.Image, op *ebiten.DrawImageOptions) {
	if screen == nil {
		panic("screen is nil")
	}
	if img == nil {
		panic("img to draw on screen is nil")
	}
	screen.DrawImage(img, op)
}

// ScaleImage scales the image by the given scale factor
func ScaleImage(img *ebiten.Image, scaleX, scaleY float64) *ebiten.Image {
	if img == nil {
		// suspicious, but since animation frames can sometimes be empty, we'll allow it for now
		return nil
	}
	bounds := img.Bounds()

	dx := float64(bounds.Dx()) * scaleX
	dy := float64(bounds.Dy()) * scaleY
	frame := ebiten.NewImage(int(dx), int(dy))

	ops := ebiten.DrawImageOptions{}
	ops.GeoM.Scale(scaleX, scaleY)
	drawImage(frame, img, &ops)

	return frame
}

func FlipHoriz(img *ebiten.Image) *ebiten.Image {
	if img == nil {
		// if img is nil, that means the loaded image was empty or fully transparent.
		// we could panic here since it seems like a suspicious situation, but for now let's just return nil as the "flipped" image.
		return nil
	}
	bounds := img.Bounds()
	frame := ebiten.NewImage(bounds.Dx(), bounds.Dy())

	ops := ebiten.DrawImageOptions{}
	ops.GeoM.Scale(-1, 1)
	ops.GeoM.Translate(float64(bounds.Dx()), 0)

	drawImage(frame, img, &ops)

	return frame
}
