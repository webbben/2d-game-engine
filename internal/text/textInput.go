package text

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type TextInput struct {
	txt                string
	allowSpecial       bool // if set, special characters can be input. if numericOnly is set, this is disabled.
	numericOnly        bool // if set, only numbers are able to be input (0 to 9, no decimal points as of now)
	maxCharLen         int
	holdBackspaceTicks int
}

func NewTextInput(allowSpecial bool, maxCharLen int, numericOnly bool) TextInput {
	if maxCharLen < 0 {
		panic("max char length must be 0 or greater")
	}
	t := TextInput{
		allowSpecial: allowSpecial,
		maxCharLen:   maxCharLen,
		numericOnly:  numericOnly,
	}
	return t
}

func (ti TextInput) GetCurrentText() string {
	return ti.txt
}

func (ti *TextInput) Clear() {
	if ti.numericOnly {
		ti.txt = "0"
		return
	}

	ti.txt = ""
}

func (ti *TextInput) SetText(s string) {
	if ti.numericOnly {
		if s == "" {
			panic("tried to set a numeric field to an empty string. set to 0 instead (or use the Clear function)")
		}

		// confirm the value is a valid integer
		_, err := strconv.Atoi(s)
		if err != nil {
			panic(err)
		}
	}
	ti.txt = s
}

func (ti *TextInput) Update() {
	ti.handleNewCharInput()

	shift := ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)

	// backspace
	if len(ti.txt) > 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			if shift {
				// shift+backspace deletes all the text - for convenience
				ti.Clear()
			} else {
				ti.backspace()
			}
		} else if ebiten.IsKeyPressed(ebiten.KeyBackspace) {
			ti.holdBackspaceTicks++
			if ti.holdBackspaceTicks > 8 {
				ti.backspace()
				ti.holdBackspaceTicks = 0
			}
		} else if inpututil.IsKeyJustReleased(ebiten.KeyBackspace) {
			ti.holdBackspaceTicks = 0
		}
	}
}

func (ti TextInput) atTextLimit() bool {
	return ti.maxCharLen > 0 && len(ti.txt) >= ti.maxCharLen
}

func (ti *TextInput) handleNewCharInput() {
	if ti.atTextLimit() {
		return
	}

	shift := ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)

	for k, ch := range numbers {
		if ti.atTextLimit() {
			break
		}
		if inpututil.IsKeyJustPressed(k) {
			i := 0
			if !ti.numericOnly && shift && ti.allowSpecial {
				i = 1
			}
			ti.txt += string(ch[i])
		}
	}

	if ti.numericOnly {
		return
	}

	// letters
	for k, ch := range letters {
		if ti.atTextLimit() {
			break
		}
		if inpututil.IsKeyJustPressed(k) {
			if shift {
				ch = rune(strings.ToUpper(string(ch))[0])
			}
			ti.txt += string(ch)
		}
	}

	// symbols
	for k, ch := range otherSymbols {
		if ti.atTextLimit() {
			break
		}
		if inpututil.IsKeyJustPressed(k) {
			i := 0
			if shift && ti.allowSpecial {
				i = 1
			}
			ti.txt += string(ch[i])
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) && !ti.atTextLimit() {
		ti.txt += " "
	}
}

func (ti *TextInput) backspace() {
	if len(ti.txt) > 0 {
		_, size := utf8.DecodeLastRuneInString(ti.txt)
		ti.txt = ti.txt[:len(ti.txt)-size]
	}
}

var letters = map[ebiten.Key]rune{
	ebiten.KeyA: 'a', ebiten.KeyB: 'b', ebiten.KeyC: 'c', ebiten.KeyD: 'd',
	ebiten.KeyE: 'e', ebiten.KeyF: 'f', ebiten.KeyG: 'g', ebiten.KeyH: 'h',
	ebiten.KeyI: 'i', ebiten.KeyJ: 'j', ebiten.KeyK: 'k', ebiten.KeyL: 'l',
	ebiten.KeyM: 'm', ebiten.KeyN: 'n', ebiten.KeyO: 'o', ebiten.KeyP: 'p',
	ebiten.KeyQ: 'q', ebiten.KeyR: 'r', ebiten.KeyS: 's', ebiten.KeyT: 't',
	ebiten.KeyU: 'u', ebiten.KeyV: 'v', ebiten.KeyW: 'w', ebiten.KeyX: 'x',
	ebiten.KeyY: 'y', ebiten.KeyZ: 'z',
}

var otherSymbols = map[ebiten.Key][2]rune{
	ebiten.KeyComma: {',', '<'}, ebiten.KeyPeriod: {'.', '>'},
	ebiten.KeySlash: {'/', '?'}, ebiten.KeyMinus: {'-', '_'}, ebiten.KeyEqual: {'=', '+'},
}

var numbers = map[ebiten.Key][2]rune{
	ebiten.Key0: {'0', ')'}, ebiten.Key1: {'1', '!'}, ebiten.Key2: {'2', '@'},
	ebiten.Key3: {'3', '#'}, ebiten.Key4: {'4', '$'}, ebiten.Key5: {'5', '%'},
	ebiten.Key6: {'6', '^'}, ebiten.Key7: {'7', '&'}, ebiten.Key8: {'8', '*'},
	ebiten.Key9: {'9', '('},
}
