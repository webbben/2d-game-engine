package exptextfield

/*
Unfortunately, it sounds like this implementation (using ebiten/v2/exp/textinput) doesn't work on linux yet:
https://github.com/hajimehoshi/ebiten/issues/2736

would be cool to try to work on that myself, but I don't want to get bogged down with months+ of extra work just to get my textfield working for linux.
So, I'll leave this here in case exp/textinput starts supporting linux, but until then I'll make my own text input component.
*/

import (
	"image"
	"image/color"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/mouse"
	internaltext "github.com/webbben/2d-game-engine/internal/text"
	"golang.org/x/image/font"
)

type TextField struct {
	bounds image.Rectangle
	field  textinput.Field
	offset int // horizontal scroll offset in pixels

	fontFace text.Face

	mouseBehavior mouse.MouseBehavior

	textBox *ebiten.Image // image that holds the text - so it can be clipped off as it exceeds the bounds

	showCursor  bool
	cursorTicks int
}

func NewTextField(widthPx int, fontFace font.Face) *TextField {
	h, _ := internaltext.GetRealisticFontMetrics(fontFace)

	t := TextField{
		bounds:   image.Rect(0, 0, widthPx, h*2),
		fontFace: text.NewGoXFace(fontFace),
		textBox:  ebiten.NewImage(widthPx, h*2),
	}
	t.SetText("Hello")

	return &t
}

func (t *TextField) SetText(s string) {
	t.field.SetTextAndSelection(s, len(s), len(s))
	t.offset = 0
}

func (t *TextField) Contains(x, y int) bool {
	return image.Pt(x, y).In(t.bounds)
}

func (t *TextField) Focus() {
	t.cursorTicks = 0
	t.showCursor = true
	t.field.Focus()
}
func (t *TextField) Blur() { t.field.Blur() }

func (t *TextField) Update() {
	t.mouseBehavior.Update(t.bounds.Min.X, t.bounds.Min.Y, t.bounds.Dx(), t.bounds.Dy(), false)
	if t.mouseBehavior.LeftClick.ClickReleased {
		t.Focus()
	} else if t.mouseBehavior.LeftClickOutside.ClickReleased {
		t.Blur()
	}

	if !t.field.IsFocused() {
		return
	}

	t.cursorTicks++
	if t.cursorTicks > 30 {
		t.cursorTicks = 0
		t.showCursor = !t.showCursor
	}

	// handle normal text editing
	handled, err := t.field.HandleInputWithBounds(t.bounds)
	if err != nil {
		logz.Panicf("error while handling text field input: %s", err.Error())
	}
	if handled {
		return
	}

	// left/right navigation
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		text := t.field.Text()
		selStart, _ := t.field.Selection()
		if selStart > 0 {
			_, size := utf8.DecodeLastRuneInString(text[:selStart])
			selStart -= size
		}
		t.field.SetSelection(selStart, selStart)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		text := t.field.Text()
		_, selEnd := t.field.Selection()
		if selEnd < len(text) {
			_, size := utf8.DecodeLastRuneInString(text[selEnd:])
			selEnd += size
		}
		t.field.SetSelection(selEnd, selEnd)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		text := t.field.Text()
		start, end := t.field.Selection()
		if start != end {
			text = text[:start] + text[end:]
		} else if start > 0 {
			_, size := utf8.DecodeLastRuneInString(text[:start])
			text = text[:start-size] + text[end:]
			start -= size
		}
		t.field.SetTextAndSelection(text, start, start)
	}

	// prevent newlines (single line only)
	txt := t.field.Text()
	if newTxt := removeNewlines(txt); newTxt != txt {
		start, end := t.field.Selection()
		start -= stringsCountNewlines(txt[:start])
		end -= stringsCountNewlines(txt[:end])
		t.field.SetTextAndSelection(newTxt, start, end)
	}

	// manage horizontal scrolling (hide overflow)
	totalWidth := int(text.Advance(t.field.Text(), t.fontFace))
	selStart, _ := t.field.Selection()
	cursorX := int(text.Advance(t.field.Text()[:selStart], t.fontFace))
	padding := 5
	fieldWidth := t.bounds.Dx() - 2*padding
	if cursorX-t.offset > fieldWidth {
		t.offset = cursorX - fieldWidth + 5
	} else if cursorX < t.offset {
		t.offset = cursorX - 5
		if t.offset < 0 {
			t.offset = 0
		}
	}
	if totalWidth < fieldWidth {
		t.offset = 0
	}
}

func removeNewlines(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r != '\n' && r != '\r' {
			out = append(out, r)
		}
	}
	return string(out)
}

func stringsCountNewlines(s string) int {
	count := 0
	for _, r := range s {
		if r == '\n' || r == '\r' {
			count++
		}
	}
	return count
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

	text.Draw(screen, t.field.TextForRendering(), t.fontFace, &op)

	// Draw cursor
	if t.field.IsFocused() && t.showCursor {
		selStart, _ := t.field.Selection()
		cursorX := int(text.Advance(t.field.TextForRendering()[:selStart], t.fontFace))
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
