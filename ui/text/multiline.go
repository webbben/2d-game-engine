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
	lw       LineWriter
	params   MultilineParams
	lwParams LineWriterParams
	txt      string
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
	lwParams := LineWriterParams{
		FontFace:         f,
		FgColor:          params.Fg,
		BgColor:          params.Bg,
		UseShadow:        params.UseShadow,
		WriteImmediately: true,
		LineWidthPx:      lineWidthPx,
		MaxHeightPx:      display.SCREEN_HEIGHT,
	}
	ml := Multiline{
		// I set the max height as screen height since I'd never expect it to go beyond that...
		lw: NewLineWriter(nil, lwParams),
		// setting these so its easy to recreate for resizing
		lwParams: lwParams,
		params:   params,
		txt:      textToWrite,
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

func (ml *Multiline) SetWidth(dx int) {
	curDx, _ := ml.Dimensions()
	if dx == curDx {
		return
	}
	*ml = NewMultiline(ml.txt, dx, ml.lwParams.FontFace, ml.params)
}
