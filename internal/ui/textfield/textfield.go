package textfield

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/webbben/2d-game-engine/internal/mouse"
	internaltext "github.com/webbben/2d-game-engine/internal/text"
	"golang.org/x/image/font"
)

type TextField struct {
	bounds    image.Rectangle
	textInput internaltext.TextInput
	offset    int // horizontal scroll offset in pixels

	fontFace text.Face

	mouseBehavior mouse.MouseBehavior

	textBox *ebiten.Image // image that holds the text - so it can be clipped off as it exceeds the bounds

	showCursor  bool
	cursorTicks int
	isFocused   bool
}

func NewTextField(widthPx int, fontFace font.Face, allowSpecial bool) *TextField {
	h, _ := internaltext.GetRealisticFontMetrics(fontFace)

	t := TextField{
		bounds:    image.Rect(0, 0, widthPx, h*2),
		fontFace:  text.NewGoXFace(fontFace),
		textBox:   ebiten.NewImage(widthPx, h*2),
		textInput: internaltext.NewTextInput(allowSpecial),
	}
	t.SetText("Hello")

	return &t
}

func (t *TextField) SetText(s string) {
	t.textInput.SetText(s)
	t.offset = 0
}

func (t *TextField) Contains(x, y int) bool {
	return image.Pt(x, y).In(t.bounds)
}

func (t *TextField) Focus() {
	t.cursorTicks = 0
	t.showCursor = true
	t.isFocused = true
}
func (t *TextField) Blur() { t.isFocused = false }

func (t *TextField) Update() {
	t.mouseBehavior.Update(t.bounds.Min.X, t.bounds.Min.Y, t.bounds.Dx(), t.bounds.Dy(), false)
	if t.mouseBehavior.LeftClick.ClickReleased {
		t.Focus()
	} else if t.mouseBehavior.LeftClickOutside.ClickReleased {
		t.Blur()
	}

	if !t.isFocused {
		return
	}

	t.cursorTicks++
	if t.cursorTicks > 30 {
		t.cursorTicks = 0
		t.showCursor = !t.showCursor
	}

	// handle normal text editing
	t.textInput.Update()
}

func (t *TextField) Draw(screen *ebiten.Image, x, y float64) {
	if t.textBox == nil {
		panic("no textbox created yet")
	}
	t.textBox.Clear()

	// Update bounds if position changed
	width := t.bounds.Dx()
	height := t.bounds.Dy()
	t.bounds = image.Rect(int(x), int(y), int(x)+width, int(y)+height)

	// Draw border
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(height), 1, color.White, false)

	// Draw background
	vector.FillRect(screen, float32(x), float32(y), float32(width), float32(height), color.RGBA{0, 0, 0, 255}, false)

	op := text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = t.fontFace.Metrics().HLineGap + t.fontFace.Metrics().HAscent + t.fontFace.Metrics().HDescent

	text.Draw(screen, t.textInput.GetCurrentText(), t.fontFace, &op)

	// Draw cursor
	if t.isFocused && t.showCursor {
		cursorX := int(text.Advance(t.textInput.GetCurrentText(), t.fontFace))
		cursorX += int(x) - t.offset
		cursorY := int(y) + 4
		vector.StrokeLine(
			screen,
			float32(cursorX),
			float32(cursorY),
			float32(cursorX),
			float32(cursorY+height),
			1,
			color.White,
			false,
		)
	}
}
