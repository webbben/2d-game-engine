package text

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	ebiten_text "github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
)

// draw text with a shadow effect. leave colors nil to use defaults (black fg and gray bg).
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

// the main function to draw text
//
// IMPORTANT: the "y" coordinate is actually the position **BELOW** where the text is drawn.
// NOT the top left corner of the text image.
func DrawText(screen *ebiten.Image, s string, f font.Face, x, y int, c color.Color) {
	if c == nil {
		c = color.Black
	}
	ebiten_text.Draw(screen, s, f, x, y, c)
}
