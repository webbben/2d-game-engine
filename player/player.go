package player

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Position struct {
	X         float64 // X position
	Y         float64 // Y position
	IsMoving  bool    // whether the player is actively moving
	Direction int     // the direction the player is facing
	TargetX   int     // the target X position if the player is moving
	TargetY   int     // the target Y position if the player is moving
}

type Player struct {
	Position
	CurrentFrame *ebiten.Image
	Frames       map[int]*ebiten.Image
}

const (
	movementSpeed = 0.2
)

func CreatePlayer(posX int, posY int, frames map[int]*ebiten.Image) Player {
	return Player{
		Position: Position{
			X: float64(posX),
			Y: float64(posY),
		},
		Frames:       frames,
		CurrentFrame: frames[0],
	}
}

func (p *Player) Update() {
	// handle player movement
	if p.IsMoving {
		p.move()
	} else {
		p.checkMovementInput()
	}
}

func (p *Player) checkMovementInput() {
	if p.IsMoving {
		return
	}
	// UP
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		p.TargetY = p.getY() - 1
		p.Direction = 1
		p.IsMoving = true
	}
	// DOWN
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		p.TargetY = p.getY() + 1
		p.Direction = 2
		p.IsMoving = true
	}
	// LEFT
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		p.TargetX = p.getX() - 1
		p.Direction = 3
		p.IsMoving = true
	}
	// RIGHT
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		p.TargetX = p.getX() + 1
		p.Direction = 4
		p.IsMoving = true
	}
}

func (p *Player) getY() int {
	return int(math.Round(p.Y))
}
func (p *Player) getX() int {
	return int(math.Round(p.X))
}

func (p *Player) snapToGrid() {
	p.X = math.Round(p.X)
	p.Y = math.Round(p.Y)
}

func (p *Player) move() {
	// handle change to p.x and p.y
	switch p.Direction {
	case 1:
		p.Y -= movementSpeed
	case 2:
		p.Y += movementSpeed
	case 3:
		p.X -= movementSpeed
	case 4:
		p.X += movementSpeed
	}

	// check if movement is done
	if p.playerAtTarget() {
		p.IsMoving = false
		p.snapToGrid()
	}
}

func (p *Player) playerAtTarget() bool {
	if p.X == float64(p.TargetX) && p.Y == float64(p.TargetY) {
		return true
	}
	switch p.Direction {
	// UP
	case 1:
		return p.Y <= float64(p.TargetY)
	// DOWN
	case 2:
		return p.Y >= float64(p.TargetY)
	// LEFT
	case 3:
		return p.X <= float64(p.TargetX)
	// RIGHT
	case 4:
		return p.X >= float64(p.TargetX)
	}
	return false
}
