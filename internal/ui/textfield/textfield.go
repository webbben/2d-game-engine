package textfield

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/rendering"
	internaltext "github.com/webbben/2d-game-engine/internal/text"
	"golang.org/x/image/font"
)

type TextField struct {
	bounds    image.Rectangle
	textInput internaltext.TextInput

	fontFace    text.Face
	textColor   color.Color
	borderColor color.Color
	bgColor     color.Color

	mouseBehavior mouse.MouseBehavior

	textBox *ebiten.Image // image that holds the text - so it can be clipped off as it exceeds the bounds

	showCursor  bool
	cursorTicks int
	isFocused   bool
}

type TextFieldParams struct {
	WidthPx            int
	FontFace           font.Face
	AllowSpecial       bool
	TextColor          color.Color
	BorderColor        color.Color
	BgColor            color.Color
	MaxCharacterLength int
}

func NewTextField(params TextFieldParams) *TextField {
	if params.FontFace == nil {
		panic("no font set")
	}
	if params.TextColor == nil {
		params.TextColor = color.Black
	}
	if params.BorderColor == nil {
		params.BorderColor = params.TextColor
	}

	h, _ := internaltext.GetRealisticFontMetrics(params.FontFace)

	t := TextField{
		bounds:      image.Rect(0, 0, params.WidthPx, h*2),
		fontFace:    text.NewGoXFace(params.FontFace),
		textBox:     ebiten.NewImage(params.WidthPx, h*2),
		textInput:   internaltext.NewTextInput(params.AllowSpecial, params.MaxCharacterLength),
		textColor:   params.TextColor,
		borderColor: params.BorderColor,
		bgColor:     params.BgColor,
	}

	return &t
}

func (t *TextField) SetText(s string) {
	t.textInput.SetText(s)
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
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(height), 1, t.borderColor, false)

	// Draw background
	if t.bgColor != nil {
		vector.FillRect(screen, float32(x), float32(y), float32(width), float32(height), t.bgColor, false)
	}

	textStartX := 0
	// if the text has exceeded the width of the textbox, start pushing it back so we can see the last characters
	textWidth := int(text.Advance(t.textInput.GetCurrentText(), t.fontFace))
	if textWidth > t.bounds.Dx() {
		textStartX -= textWidth - t.bounds.Dx()
	}

	op := text.DrawOptions{}
	op.GeoM.Translate(float64(textStartX), 0)
	op.ColorScale.ScaleWithColor(t.textColor)
	op.LineSpacing = t.fontFace.Metrics().HLineGap + t.fontFace.Metrics().HAscent + t.fontFace.Metrics().HDescent

	text.Draw(t.textBox, t.textInput.GetCurrentText(), t.fontFace, &op)

	// Draw cursor
	if t.isFocused && t.showCursor {
		cursorX := int(x) + textStartX + textWidth
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

	rendering.DrawImage(screen, t.textBox, x, y, 0)
}
