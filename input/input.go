// Package input controls user keyboard input, and how that input is translated into actions in the game engine.
// Note that this package is not used for actual text input for typing data, but instead is used for things like moving in a map, navigating menus, etc.
package input

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Key int

type Action int

const (
	ActionPause Action = iota

	ActionMoveUp
	ActionMoveDown
	ActionMoveLeft
	ActionMoveRight
)

const (
	KeyUnknown Key = iota

	KeyEscape
	KeyEnter
	KeySpace
	KeyShift

	KeyUp
	KeyDown
	KeyLeft
	KeyRight

	KeyA
	KeyB
	KeyC
	KeyD
	KeyE
	KeyF
	KeyG
	KeyH
	KeyI
	KeyJ
	KeyK
	KeyL
	KeyM
	KeyN
	KeyO
	KeyP
	KeyQ
	KeyR
	KeyS
	KeyT
	KeyU
	KeyV
	KeyW
	KeyX
	KeyY
	KeyZ

	Key1
	Key2
	Key3
	Key4
	Key5
	Key6
	Key7
	Key8
	Key9
	Key0

	KeyPeriod
	KeyComma
	KeySlash
	KeyBackslash
	KeyOpenBracket
	KeyCloseBracket
	KeyMinus
	KeyEqual
	KeyTick
	KeySingleQuote
	KeySemiColon
)

func toEngineKey(k ebiten.Key) Key {
	switch k {
	case ebiten.KeyEscape:
		return KeyEscape
	case ebiten.KeyEnter:
		return KeyEnter
	case ebiten.KeySpace:
		return KeySpace
	case ebiten.KeyShift:
		return KeyShift
	case ebiten.KeyUp:
		return KeyUp
	case ebiten.KeyDown:
		return KeyDown
	case ebiten.KeyLeft:
		return KeyLeft
	case ebiten.KeyRight:
		return KeyRight
	case ebiten.KeyA:
		return KeyA
	case ebiten.KeyB:
		return KeyB
	case ebiten.KeyC:
		return KeyC
	case ebiten.KeyD:
		return KeyD
	case ebiten.KeyE:
		return KeyE
	case ebiten.KeyF:
		return KeyF
	case ebiten.KeyG:
		return KeyG
	case ebiten.KeyH:
		return KeyH
	case ebiten.KeyI:
		return KeyI
	case ebiten.KeyJ:
		return KeyJ
	case ebiten.KeyK:
		return KeyK
	case ebiten.KeyL:
		return KeyL
	case ebiten.KeyM:
		return KeyM
	case ebiten.KeyN:
		return KeyN
	case ebiten.KeyO:
		return KeyO
	case ebiten.KeyP:
		return KeyP
	case ebiten.KeyQ:
		return KeyQ
	case ebiten.KeyR:
		return KeyR
	case ebiten.KeyS:
		return KeyS
	case ebiten.KeyT:
		return KeyT
	case ebiten.KeyU:
		return KeyU
	case ebiten.KeyV:
		return KeyV
	case ebiten.KeyW:
		return KeyW
	case ebiten.KeyX:
		return KeyX
	case ebiten.KeyY:
		return KeyY
	case ebiten.KeyZ:
		return KeyZ
	case ebiten.Key1:
		return Key1
	case ebiten.Key2:
		return Key2
	case ebiten.Key3:
		return Key3
	case ebiten.Key4:
		return Key4
	case ebiten.Key5:
		return Key5
	case ebiten.Key6:
		return Key6
	case ebiten.Key7:
		return Key7
	case ebiten.Key8:
		return Key8
	case ebiten.Key9:
		return Key9
	case ebiten.Key0:
		return Key0
	default:
		return KeyUnknown
	}
}

type Input struct {
	// raw physical state; what keys are actually being pressed
	keyDown map[Key]bool
	// user intent mapping; maps actions to which keys can trigger them.
	// we don't map key to action because it's possible in different situations, keys could be used in different ways.
	bindings map[Action][]Key
	// what the player is trying to do right now; the resolved "intent state".
	actionDown map[Action]bool
}

// TODO: start using this to get user keyboard input, for everything from player movement to opening menus
func NewInput() *Input {
	return &Input{
		keyDown:    make(map[Key]bool),
		bindings:   make(map[Action][]Key),
		actionDown: make(map[Action]bool),
	}
}

func (i *Input) Bind(a Action, keys ...Key) {
	i.bindings[a] = keys
}

func (i Input) IsActionPressed(a Action) bool {
	return i.actionDown[a]
}

func (i *Input) Update() {
	// clear previous key state
	for k := range i.keyDown {
		i.keyDown[k] = false
	}

	// poll ebiten keys
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		if ebiten.IsKeyPressed(k) {
			engineKey := toEngineKey(k)
			if engineKey != KeyUnknown {
				i.keyDown[engineKey] = true
			}
		}
	}

	// clear previous action state
	for action := range i.actionDown {
		i.actionDown[action] = false
	}

	// resolve bindings
	for action, keys := range i.bindings {
		for _, k := range keys {
			if i.keyDown[k] {
				i.actionDown[action] = true
				break
			}
		}
	}
}
