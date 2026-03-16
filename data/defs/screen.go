package defs

import "github.com/hajimehoshi/ebiten/v2"

type ScreenID string

type Transition interface {
	Update()
	Draw(screen *ebiten.Image)
	IsDone() bool
}
