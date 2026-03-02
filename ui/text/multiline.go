package text

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/display"
	"golang.org/x/image/font"
)

// Multiline is essentially a simplified wrapper around LineWriter which makes it easier
// to put multiline text into UI screens. You set a width, and it handles all the internal
// logic, updates, settings, etc of the linewriter to get a simple UI component that can
// draw multiline text on a screen immediately, without any delays. There is no max height.
type Multiline struct {
	lw LineWriter
}

type MultilineParams struct {
	Fg, Bg    color.Color
	UseShadow bool
}

func NewMultiline(textToWrite string, lineWidthPx int, f font.Face, params MultilineParams) Multiline {
	if textToWrite == "" {
		panic("textToWrite was empty")
	}
	if lineWidthPx <= 0 {
		panic("linewidth was <= 0")
	}
	if f == nil {
		panic("font was nil")
	}
	ml := Multiline{
		// I set the max height as screen height since I'd never expect it to go beyond that...
		lw: NewLineWriter(lineWidthPx, display.SCREEN_HEIGHT, f, params.Fg, params.Bg, params.UseShadow, true),
	}
	ml.lw.SetSourceText(textToWrite)
	// only need to update once, to get the text to be processed and ready to draw.
	ml.lw.Update()

	return ml
}

func (ml Multiline) Dimensions() (dx, dy int) {
	return ml.lw.CurrentDimensions()
}

func (ml *Multiline) Draw(screen *ebiten.Image, x, y float64) {
	ml.lw.Draw(screen, int(x), int(y))
}
