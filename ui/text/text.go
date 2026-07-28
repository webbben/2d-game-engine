package text

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	ebiten_text "github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

// DrawShadowText draws text with a shadow effect. leave colors nil to use defaults (black fg and gray bg).
// bgOffsets, which adjust the position of the "shadow" text, default to -2 if left at 0.
func DrawShadowText(screen *ebiten.Image, s string, f font.Face, x, y int, fg color.Color, bg color.Color, bgOffsetX, bgOffsetY int) {
	if fg == nil {
		fg = color.Black
	}
	if bg == nil {
		bg = color.RGBA{20, 20, 20, 75} // semi-transparent dark gray
	}
	if bgOffsetX == 0 {
		bgOffsetX = -2
	}
	if bgOffsetY == 0 {
		bgOffsetY = -2
	}
	DrawText(screen, s, f, x+bgOffsetX, y-bgOffsetY, bg)
	DrawText(screen, s, f, x, y, fg)
}

// DrawText is the main function to draw text
//
// IMPORTANT: the "y" coordinate is actually the position **BELOW** where the text is drawn.
// NOT the top left corner of the text image.
func DrawText(screen *ebiten.Image, s string, f font.Face, x, y int, c color.Color) {
	if s == "" {
		return
	}
	if c == nil {
		c = color.Black
	}
	if f == nil {
		panic("no font set!")
	}
	ebiten_text.Draw(screen, s, f, x, y, c)
}

func DrawOutlinedText(screen *ebiten.Image, s string, f font.Face, x, y int, fg color.Color, bg color.Color, bgOffsetX, bgOffsetY int) {
	if fg == nil {
		fg = color.Black
	}
	if bg == nil {
		bg = color.White
	}
	if bgOffsetX == 0 {
		bgOffsetX = 2
	}
	if bgOffsetY == 0 {
		bgOffsetY = 2
	}
	DrawText(screen, s, f, x-bgOffsetX, y, bg)
	DrawText(screen, s, f, x+bgOffsetX, y, bg)
	DrawText(screen, s, f, x, y-bgOffsetY, bg)
	DrawText(screen, s, f, x, y+bgOffsetY, bg)
	DrawText(screen, s, f, x, y, fg)
}

// TODO: not doing this now because there would be so many places to update, but...
// probably should just make one DrawText function and use this DrawTextParams struct as an argument.

type DrawTextParams struct {
	Font font.Face
	Fg   color.Color
	// for now, not implementing background color and stuff for this
}

func DrawTextWithOptions(screen *ebiten.Image, s string, x, y int, params DrawTextParams, ops *ebiten.DrawImageOptions) {
	if s == "" {
		panic("string is empty")
	}
	if params.Font == nil {
		panic("font was nil")
	}

	if params.Fg != nil {
		r, g, b, a := params.Fg.RGBA()
		ops.ColorScale.Scale(
			float32(r)/0xffff,
			float32(g)/0xffff,
			float32(b)/0xffff,
			float32(a)/0xffff,
		)
	}

	ops.GeoM.Translate(float64(x), float64(y))
	ebiten_text.DrawWithOptions(screen, s, params.Font, ops)
}
