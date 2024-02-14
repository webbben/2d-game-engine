package player

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

type Position struct {
	X                    float64 // X position
	Y                    float64 // Y position
	IsMoving             bool    // whether the player is actively moving
	Direction            int     // the direction the player is facing
	TargetX              int     // the target X position if the player is moving
	TargetY              int     // the target Y position if the player is moving
	animStep             int     // the step of the animation we are on
	sameDirectionCounter int     // counts how many frames the key for the same direction has been pressed. used to create a slight delay in movement
}

type Player struct {
	Position
	CurrentFrame *ebiten.Image
	Frames       map[string]*ebiten.Image
}

const (
	movementSpeed = 0.1
)

const (
	walk_anim_thresh = 3
	walk_frame_delay = 3 // number of frames to wait until walking once the player direction has changed
)

var (
	direc_down  = []string{"down-1", "down-2", "down-3", "down-4", "down-5", "down-6", "down-7"}
	direc_up    = []string{"up-1", "up-2", "up-3", "up-4", "up-5", "up-6", "up-7"}
	direc_left  = []string{"left-1", "left-2", "left-3", "left-4", "left-5", "left-6", "left-7"}
	direc_right = []string{"right-1", "right-2", "right-3", "right-4", "right-5", "right-6", "right-7"}
)

func CreatePlayer(posX int, posY int, frames map[string]*ebiten.Image) Player {
	return Player{
		Position: Position{
			X: float64(posX),
			Y: float64(posY),
		},
		Frames:       frames,
		CurrentFrame: frames[direc_down[0]],
	}
}

func (p *Player) Draw(screen *ebiten.Image, tileSize int, op *ebiten.DrawImageOptions) {
	op.GeoM.Translate(p.X*float64(tileSize), p.Y*float64(tileSize))
	screen.DrawImage(p.CurrentFrame, op)
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
		// first press should just turn the character
		if p.Direction == 1 {
			p.sameDirectionCounter++
		} else {
			p.sameDirectionCounter = 0
			p.Direction = 1
		}
		if p.sameDirectionCounter >= walk_frame_delay {
			p.IsMoving = true
		}
		p.setFrame(direc_up[0])
		p.animStep = 0
	}
	// DOWN
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		p.TargetY = p.getY() + 1
		if p.Direction == 2 {
			p.sameDirectionCounter++

		} else {
			p.sameDirectionCounter = 0
			p.Direction = 2
		}
		if p.sameDirectionCounter >= walk_frame_delay {
			p.IsMoving = true
		}
		p.setFrame(direc_down[0])
		p.animStep = 0
	}
	// LEFT
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		p.TargetX = p.getX() - 1
		if p.Direction == 3 {
			p.sameDirectionCounter++
		} else {
			p.sameDirectionCounter = 0
			p.Direction = 3
		}
		if p.sameDirectionCounter >= walk_frame_delay {
			p.IsMoving = true
		}
		p.setFrame(direc_left[0])
		p.animStep = 0
	}
	// RIGHT
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		p.TargetX = p.getX() + 1
		if p.Direction == 4 {
			p.sameDirectionCounter++
		} else {
			p.sameDirectionCounter = 0
			p.Direction = 4
		}
		if p.sameDirectionCounter >= walk_frame_delay {
			p.IsMoving = true
		}
		p.setFrame(direc_right[0])
		p.animStep = 0
	}
}

func (p *Player) walkAnimation() {
	p.animStep++
	if p.animStep%walk_anim_thresh != 0 {
		return
	}
	if p.animStep >= 7 {
		p.animStep = 0
	}
	switch p.Direction {
	// UP
	case 1:
		p.setFrame(direc_up[p.animStep])
	// DOWN
	case 2:
		p.setFrame(direc_down[p.animStep])
	// LEFT
	case 3:
		p.setFrame(direc_left[p.animStep])
	// RIGHT
	case 4:
		p.setFrame(direc_right[p.animStep])
	}
}

func (p *Player) setFrame(key string) {
	frame, ok := p.Frames[key]
	if !ok {
		fmt.Println("accessing unrecognized player frame:", key)
		return
	}
	p.CurrentFrame = frame
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
	// UP
	case 1:
		p.Y -= movementSpeed
	// DOWN
	case 2:
		p.Y += movementSpeed
	// LEFT
	case 3:
		p.X -= movementSpeed
	// RIGHT
	case 4:
		p.X += movementSpeed
	}

	p.walkAnimation()

	// check if movement is done
	if p.playerAtTarget() {
		p.IsMoving = false
		p.snapToGrid()
	}
}

// detects if the player has arrived at their target tile (or passed it)
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
