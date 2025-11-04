package text

import (
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type TextInput struct {
	txt                    string
	allowSpecialCharacters bool
	holdBackspaceTicks     int
}

func NewTextInput(allowSpecialCharacters bool) TextInput {
	t := TextInput{
		allowSpecialCharacters: allowSpecialCharacters,
	}
	return t
}

func (ti TextInput) GetCurrentText() string {
	return ti.txt
}

func (ti *TextInput) Clear() {
	ti.txt = ""
}

func (ti *TextInput) SetText(s string) {
	ti.txt = s
}

func (f *TextInput) Update() {
	shift := ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)

	// letters
	for k, ch := range letters {
		if inpututil.IsKeyJustPressed(k) {
			if shift {
				ch = rune(strings.ToUpper(string(ch))[0])
			}
			f.txt += string(ch)
		}
	}

	// numbers and symbols
	for k, ch := range digitsEtc {
		if inpututil.IsKeyJustPressed(k) {
			i := 0
			if shift {
				i = 1
			}
			f.txt += string(ch[i])
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		f.txt += " "
	}
	if len(f.txt) > 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			f.backspace()
		} else if ebiten.IsKeyPressed(ebiten.KeyBackspace) {
			f.holdBackspaceTicks++
			if f.holdBackspaceTicks > 8 {
				f.backspace()
				f.holdBackspaceTicks = 0
			}
		} else if inpututil.IsKeyJustReleased(ebiten.KeyBackspace) {
			f.holdBackspaceTicks = 0
		}
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

var digitsEtc = map[ebiten.Key][2]rune{
	ebiten.Key0: {'0', ')'}, ebiten.Key1: {'1', '!'}, ebiten.Key2: {'2', '@'},
	ebiten.Key3: {'3', '#'}, ebiten.Key4: {'4', '$'}, ebiten.Key5: {'5', '%'},
	ebiten.Key6: {'6', '^'}, ebiten.Key7: {'7', '&'}, ebiten.Key8: {'8', '*'},
	ebiten.Key9: {'9', '('}, ebiten.KeyComma: {',', '<'}, ebiten.KeyPeriod: {'.', '>'},
	ebiten.KeySlash: {'/', '?'},
}
