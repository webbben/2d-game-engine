package entity

import "github.com/hajimehoshi/ebiten/v2"

type Position struct {
	X        float64
	Y        float64
	Facing   string
	IsMoving bool
}

type Frames struct {
	IdleFrames      []*ebiten.Image
	MoveUpFrames    []*ebiten.Image
	MoveDownFrames  []*ebiten.Image
	MoveLeftFrames  []*ebiten.Image
	MoveRightFrames []*ebiten.Image
}

type Entity struct {
	ID             string
	Name           string
	IsHuman        bool
	IsInteractable bool
	Position
	Frames
}
