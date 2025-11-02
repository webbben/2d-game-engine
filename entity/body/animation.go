package body

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
)

const (
	ANIM_WALK  = "walk"
	ANIM_RUN   = "run"
	ANIM_SLASH = "slash"
)

type Animation struct {
	Name         string
	Skip         bool            // if true, this animation does not get defined
	L            []*ebiten.Image `json:"-"`
	R            []*ebiten.Image `json:"-"`
	U            []*ebiten.Image `json:"-"`
	D            []*ebiten.Image `json:"-"`
	TileSteps    []int
	StepsOffsetY []int
}

func (a *Animation) reset() {
	a.L = make([]*ebiten.Image, 0)
	a.R = make([]*ebiten.Image, 0)
	a.U = make([]*ebiten.Image, 0)
	a.D = make([]*ebiten.Image, 0)
}

func (a Animation) getFrame(dir byte, animationIndex int) *ebiten.Image {
	switch dir {
	case 'L':
		if len(a.L) == 0 {
			return nil
		}
		if animationIndex >= len(a.L) {
			logz.Println(a.Name, "past left frames; returning last frame", "animIndex:", animationIndex)
			return a.L[len(a.L)-1]
		}
		return a.L[animationIndex]
	case 'R':
		if len(a.R) == 0 {
			return nil
		}
		if animationIndex >= len(a.R) {
			logz.Println(a.Name, "past right frames; returning last frame", "animIndex:", animationIndex)
			return a.R[len(a.R)-1]
		}
		return a.R[animationIndex]
	case 'U':
		if len(a.U) == 0 {
			return nil
		}
		if animationIndex >= len(a.U) {
			logz.Println(a.Name, "past up frames; returning last frame", "animIndex:", animationIndex)
			return a.U[len(a.U)-1]
		}
		return a.U[animationIndex]
	case 'D':
		if len(a.D) == 0 {
			return nil
		}
		if animationIndex >= len(a.D) {
			logz.Println(a.Name, "past down frames; returning last frame", "animIndex:", animationIndex)
			return a.D[len(a.D)-1]
		}
		return a.D[animationIndex]
	}
	panic("unrecognized direction")
}
