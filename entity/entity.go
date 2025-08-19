package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/webbben/2d-game-engine/internal/model"
)

type Entity struct {
	EntityInfo
	Movement     Movement `json:"movement"`
	CurrentFrame *ebiten.Image
	Position

	World WorldContext
}

type WorldContext interface {
	Collides(c model.Coords) bool
	FindPath(c model.Coords) []model.Coords
	EntitiesNearby(x, y float64, radius float64) []*Entity
}

type EntityInfo struct {
	DisplayName string
	UID         string
	Source      string // JSON source file for this entity
}

type Position struct {
	X, Y                 float64      // the exact position the entity is at on the map
	TilePos              model.Coords // the tile the entity is technically inside of
	NextTileX, NextTileY int          // the next tile the entity is moving into. if -1, the entity is not moving.
}

type Movement struct {
	IdleLeft       *ebiten.Image
	Left           []*ebiten.Image
	LeftRun        []*ebiten.Image
	IdleRight      *ebiten.Image
	Right          []*ebiten.Image
	RightRun       []*ebiten.Image
	IdleUp         *ebiten.Image
	Up             []*ebiten.Image
	UpRun          []*ebiten.Image
	IdleDown       *ebiten.Image
	Down           []*ebiten.Image
	DownRun        []*ebiten.Image
	Direction      byte // L R U D
	AnimationTimer int  // counts the ticks until next animation frame
	AnimationFrame int  // the current movement animation frame index

	CanRun         bool           `json:"can_run"`
	MovementSource MovementSource `json:"movement_source"`

	IsMoving  bool
	IsRunning bool
	WalkSpeed float64 // value should be a TileSize / NumFrames calculation
	RunSpeed  float64 // value should be a TileSize / NumFrames calculation
	Speed     float64 // actual speed the entity is moving at

	TargetTile model.Coords   // next tile the entity is currently moving
	TargetPath []model.Coords // path the entity is currently trying to travel on
}

type MovementSource struct {
	IdleLeft  string   `json:"left_idle"`
	Left      []string `json:"left"`
	LeftRun   []string `json:"left_run"`
	IdleRight string   `json:"right_idle"`
	Right     []string `json:"right"`
	RightRun  []string `json:"right_run"`
	IdleUp    string   `json:"up_idle"`
	Up        []string `json:"up"`
	UpRun     []string `json:"up_run"`
	IdleDown  string   `json:"down_idle"`
	Down      []string `json:"down"`
	DownRun   []string `json:"down_run"`
}

func (e *Entity) LoadMovementFrames() error {
	idleLeft, _, err := ebitenutil.NewImageFromFile(e.Movement.MovementSource.IdleLeft)
	if err != nil {
		return err
	}
	idleRight, _, err := ebitenutil.NewImageFromFile(e.Movement.MovementSource.IdleRight)
	if err != nil {
		return err
	}
	idleUp, _, err := ebitenutil.NewImageFromFile(e.Movement.MovementSource.IdleUp)
	if err != nil {
		return err
	}
	idleDown, _, err := ebitenutil.NewImageFromFile(e.Movement.MovementSource.IdleDown)
	if err != nil {
		return err
	}
	e.Movement.IdleLeft = idleLeft
	e.Movement.IdleRight = idleRight
	e.Movement.IdleUp = idleUp
	e.Movement.IdleDown = idleDown

	for _, p := range e.Movement.MovementSource.Left {
		frame, _, err := ebitenutil.NewImageFromFile(p)
		if err != nil {
			return err
		}
		e.Movement.Left = append(e.Movement.Left, frame)
	}
	for _, p := range e.Movement.MovementSource.Right {
		frame, _, err := ebitenutil.NewImageFromFile(p)
		if err != nil {
			return err
		}
		e.Movement.Right = append(e.Movement.Right, frame)
	}
	for _, p := range e.Movement.MovementSource.Up {
		frame, _, err := ebitenutil.NewImageFromFile(p)
		if err != nil {
			return err
		}
		e.Movement.Up = append(e.Movement.Up, frame)
	}
	for _, p := range e.Movement.MovementSource.Down {
		frame, _, err := ebitenutil.NewImageFromFile(p)
		if err != nil {
			return err
		}
		e.Movement.Down = append(e.Movement.Down, frame)
	}

	if e.Movement.CanRun {
		for _, p := range e.Movement.MovementSource.LeftRun {
			frame, _, err := ebitenutil.NewImageFromFile(p)
			if err != nil {
				return err
			}
			e.Movement.LeftRun = append(e.Movement.LeftRun, frame)
		}
		for _, p := range e.Movement.MovementSource.RightRun {
			frame, _, err := ebitenutil.NewImageFromFile(p)
			if err != nil {
				return err
			}
			e.Movement.RightRun = append(e.Movement.RightRun, frame)
		}
		for _, p := range e.Movement.MovementSource.UpRun {
			frame, _, err := ebitenutil.NewImageFromFile(p)
			if err != nil {
				return err
			}
			e.Movement.UpRun = append(e.Movement.UpRun, frame)
		}
		for _, p := range e.Movement.MovementSource.DownRun {
			frame, _, err := ebitenutil.NewImageFromFile(p)
			if err != nil {
				return err
			}
			e.Movement.DownRun = append(e.Movement.DownRun, frame)
		}
	}

	return nil
}
