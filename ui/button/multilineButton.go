package button

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/ui/text"
	"golang.org/x/image/font"
)

type MultilineButton struct {
	btn *Button

	ml text.Multiline
}

func NewMultilineButton(buttonText string, maxWidthPx int, f font.Face, mlParams text.MultilineParams) *MultilineButton {
	if f == nil {
		f = config.DefaultFont
		if f == nil {
			panic("default font not loaded!")
		}
	}
	// get the multiline text that will go inside the button.
	ml := text.NewMultiline(buttonText, maxWidthPx, f, mlParams)
	dx, dy := ml.Dimensions()

	_, dsc := text.GetRealisticFontMetrics(f)
	dy += dsc
	btn := NewButton("", nil, dx, dy)

	mb := MultilineButton{
		btn: btn,
		ml:  ml,
	}

	return &mb
}

func (mb *MultilineButton) Update() ButtonUpdateResult {
	return mb.btn.Update()
}

func (mb MultilineButton) Dimensions() (dx, dy int) {
	return mb.btn.Width, mb.btn.Height
}

func (mb *MultilineButton) Draw(screen *ebiten.Image, x, y float64) {
	mb.btn.Draw(screen, int(x), int(y))
	mb.ml.Draw(screen, x, y)
}
