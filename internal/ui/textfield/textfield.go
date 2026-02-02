// Package textfield provides a textfield UI component for inputting text
package textfield

import (
	"image"
	"image/color"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/rendering"
	internaltext "github.com/webbben/2d-game-engine/internal/text"
	"golang.org/x/image/font"
)

type TextField struct {
	bounds    image.Rectangle
	textInput internaltext.TextInput

	numericOnly bool

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

func (tf TextField) GetText() string {
	return tf.textInput.GetCurrentText()
}

func (tf TextField) GetNumber() int {
	if !tf.numericOnly {
		panic("tried to get number from non-numeric textfield")
	}
	if tf.GetText() == "" {
		if tf.IsFocused() {
			return 0 // user is still editing, so no problem
		}
		panic("GetNumber: numeric textfield seems to be inactive, but has an empty string value. it should be 0 in this case")
	}
	v, err := strconv.Atoi(tf.GetText())
	if err != nil {
		logz.Panicf("error converting string to int: %s", err)
	}
	return v
}

func (tf TextField) Dimensions() (dx, dy int) {
	return tf.bounds.Dx(), tf.bounds.Dy()
}

type TextFieldParams struct {
	WidthPx            int
	FontFace           font.Face
	AllowSpecial       bool // if set, special characters can be input. NumericOnly disables this.
	NumericOnly        bool // if set, only numbers can be input. no decimals, just integers.
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
		textInput:   internaltext.NewTextInput(params.AllowSpecial, params.MaxCharacterLength, params.NumericOnly),
		textColor:   params.TextColor,
		borderColor: params.BorderColor,
		bgColor:     params.BgColor,
		numericOnly: params.NumericOnly,
	}

	if t.numericOnly {
		t.SetNumber(0)
	}

	return &t
}

func (t *TextField) Clear() {
	t.textInput.Clear()
}

func (t *TextField) SetText(s string) {
	t.textInput.SetText(s)
}

func (t *TextField) SetNumber(v int) {
	if v < 0 {
		panic("negative numbers not implemented yet")
	}
	if !t.numericOnly {
		panic("tried to set number on non-numeric field. if this field should be strictly for numbers, set numericOnly on this text field")
	}
	s := strconv.Itoa(v)
	t.SetText(s)
}

func (t *TextField) Contains(x, y int) bool {
	return image.Pt(x, y).In(t.bounds)
}

func (t *TextField) Focus() {
	t.cursorTicks = 0
	t.showCursor = true
	t.isFocused = true
}

func (t TextField) IsFocused() bool {
	return t.isFocused
}

func (t *TextField) Blur() {
	t.isFocused = false

	// ensure that numericOnly fields don't end up as an empty string
	if t.numericOnly && t.GetText() == "" {
		t.SetNumber(0)
	}
}

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
