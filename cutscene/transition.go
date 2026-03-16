// Package cutscene provides tools and logic for handling cutscenes, transitions, etc
package cutscene

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
)

type FadeToBlack struct {
	fadeValue   float32 // once this reaches 0, fade to black will be complete
	fadeFactor  float32
	done        bool
	blackScreen *ebiten.Image
}

func (t *FadeToBlack) ZzInterfaceCheck() {
	_ = append([]defs.Transition{}, t)
}

func NewFadeToBlackTransition(fadeFactor float32) *FadeToBlack {
	if fadeFactor <= 0 || fadeFactor >= 1 {
		logz.Panicln("FadeToBlack", "fade factor invalid:", fadeFactor)
	}

	t := FadeToBlack{
		fadeFactor: fadeFactor,
		fadeValue:  1,
	}
	t.blackScreen = ebiten.NewImage(display.SCREEN_WIDTH, display.SCREEN_HEIGHT)
	t.blackScreen.Fill(color.Black)

	return &t
}

func (t *FadeToBlack) Update() {
	if t.done {
		return
	}
	t.fadeValue *= t.fadeFactor
	if t.fadeValue < 0.01 {
		t.fadeValue = 0
		t.done = true
	}
}

func (t FadeToBlack) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1, 1, 1, 1-t.fadeValue)
	rendering.DrawImageWithOps(screen, t.blackScreen, 0, 0, 0, op)
}

func (t FadeToBlack) IsDone() bool {
	return t.done
}

type FadeFromBlack struct {
	fadeValue   float32 // once this reaches 0, fade to black will be complete
	fadeFactor  float32
	done        bool
	blackScreen *ebiten.Image
}

func (t *FadeFromBlack) ZzInterfaceCheck() {
	_ = append([]defs.Transition{}, t)
}

func NewFadeFromBlackTransition(fadeFactor float32) *FadeFromBlack {
	if fadeFactor <= 0 || fadeFactor >= 1 {
		logz.Panicln("FadeFromBlack", "fade factor invalid:", fadeFactor)
	}

	t := FadeFromBlack{
		fadeFactor: fadeFactor,
		fadeValue:  1,
	}
	t.blackScreen = ebiten.NewImage(display.SCREEN_WIDTH, display.SCREEN_HEIGHT)
	t.blackScreen.Fill(color.Black)

	return &t
}

func (t *FadeFromBlack) Update() {
	if t.done {
		return
	}
	t.fadeValue *= t.fadeFactor
	if t.fadeValue < 0.01 {
		t.fadeValue = 0
		t.done = true
	}
}

func (t FadeFromBlack) Draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.Scale(1, 1, 1, t.fadeValue)
	rendering.DrawImageWithOps(screen, t.blackScreen, 0, 0, 0, op)
}

func (t FadeFromBlack) IsDone() bool {
	return t.done
}
