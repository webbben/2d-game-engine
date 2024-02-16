package player

import (
	"ancient-rome/config"
	"ancient-rome/rendering"
	"fmt"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Position struct {
	X               float64   // X position
	Y               float64   // Y position
	IsMoving        bool      // whether the player is actively moving
	Direction_Horiz string    // "L"/"R" - the direction the player is moving on the horizontal axis
	Direction_Vert  string    // "U"/"D" - the direction the player is moving on the vertical axis
	Facing          string    // "U"/"D"/"L"/"R" - the direction the player is facing (visually)
	animStep        int       // the step of the animation we are on
	lastAnimStep    time.Time // the last time the character changed frames
}

type Player struct {
	Position
	CurrentFrame *ebiten.Image
	Frames       map[string]*ebiten.Image
}

const (
	movementSpeed = 0.025
	delay         = time.Millisecond * 8
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
			X:               float64(posX),
			Y:               float64(posY),
			Direction_Horiz: "X",
			Direction_Vert:  "X",
		},
		Frames:       frames,
		CurrentFrame: frames[direc_down[0]],
	}
}

func (p *Player) Draw(screen *ebiten.Image, op *ebiten.DrawImageOptions, offsetX float64, offsetY float64) {
	drawX, drawY := rendering.GetImageDrawPos(p.CurrentFrame, p.X, p.Y, offsetX, offsetY)
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(p.CurrentFrame, op)
}

func (p *Player) Update() {
	// handle player movement
	if p.IsMoving {
		p.move()
	}
	p.checkMovementInput()
}

func (p *Player) checkMovementInput() {
	// UP
	if ebiten.IsKeyPressed(ebiten.KeyW) && p.Direction_Vert == "X" {
		if !ebiten.IsKeyPressed(ebiten.KeyS) {
			p.Direction_Vert = "U"
			p.IsMoving = true
			go p.continueWalking("U")
		}
	}
	// DOWN
	if ebiten.IsKeyPressed(ebiten.KeyS) && p.Direction_Vert == "X" {
		if !ebiten.IsKeyPressed(ebiten.KeyW) {
			p.Direction_Vert = "D"
			p.IsMoving = true
			go p.continueWalking("D")
		}
	}
	// LEFT
	if ebiten.IsKeyPressed(ebiten.KeyA) && p.Direction_Horiz == "X" {
		if !ebiten.IsKeyPressed(ebiten.KeyD) {
			p.Direction_Horiz = "L"
			p.IsMoving = true
			go p.continueWalking("L")
		}
	}
	// RIGHT
	if ebiten.IsKeyPressed(ebiten.KeyD) && p.Direction_Horiz == "X" {
		if !ebiten.IsKeyPressed(ebiten.KeyA) {
			p.Direction_Horiz = "R"
			p.IsMoving = true
			go p.continueWalking("R")
		}
	}
}

func (p *Player) continueWalking(direction string) {
	switch direction {
	case "U":
		// keep going until that key is released
		for ebiten.IsKeyPressed(ebiten.KeyW) {
			p.Y -= movementSpeed
			time.Sleep(delay)
		}
		// round the position out in case they overshot it
		p.easeToPosY(math.Floor(p.Y))
		p.Direction_Vert = "X"
	case "D":
		for ebiten.IsKeyPressed(ebiten.KeyS) {
			p.Y += movementSpeed
			time.Sleep(delay)
		}
		p.easeToPosY(math.Ceil(p.Y))
		p.Direction_Vert = "X"
	case "L":
		for ebiten.IsKeyPressed(ebiten.KeyA) {
			p.X -= movementSpeed
			time.Sleep(delay)
		}
		p.easeToPosX(math.Floor(p.X))
		p.Direction_Horiz = "X"
	case "R":
		for ebiten.IsKeyPressed(ebiten.KeyD) {
			p.X += movementSpeed
			time.Sleep(delay)
		}
		p.easeToPosX(math.Ceil(p.X))
		p.Direction_Horiz = "X"
	}
}

func (p *Player) easeToPosY(destY float64) {
	if destY > p.Y {
		// moving down
		for destY-p.Y >= movementSpeed {
			p.Y += movementSpeed
			time.Sleep(delay)
		}
	} else if p.Y > destY {
		// moving up
		for p.Y-destY >= movementSpeed {
			p.Y -= movementSpeed
			time.Sleep(delay)
		}
	}
	p.Y = destY
}
func (p *Player) easeToPosX(destX float64) {
	if destX > p.X {
		// moving right
		for destX-p.X >= movementSpeed {
			p.X += movementSpeed
			time.Sleep(delay)
		}
	} else if p.X > destX {
		// moving left
		for p.X-destX >= movementSpeed {
			p.X -= movementSpeed
			time.Sleep(delay)
		}
	}
	p.X = destX
}

func (p *Player) walkAnimation() {
	curTime := time.Now()
	if p.IsMoving && (curTime.Sub(p.lastAnimStep) < 150*time.Millisecond) {
		return
	}
	p.lastAnimStep = curTime
	p.animStep++
	if p.animStep >= 7 || !p.IsMoving {
		p.animStep = 0
	}
	switch p.Facing {
	// UP
	case "U":
		p.setFrame(direc_up[p.animStep])
	// DOWN
	case "D":
		p.setFrame(direc_down[p.animStep])
	// LEFT
	case "L":
		p.setFrame(direc_left[p.animStep])
	// RIGHT
	case "R":
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

func (p *Player) move() {
	// check if movements have finished
	if p.Direction_Horiz == "X" && p.Direction_Vert == "X" {
		p.IsMoving = false
	}

	if p.Direction_Horiz == "L" {
		p.Facing = "L"
	} else if p.Direction_Horiz == "R" {
		p.Facing = "R"
	} else if p.Direction_Vert == "U" {
		p.Facing = "U"
	} else if p.Direction_Vert == "D" {
		p.Facing = "D"
	}

	p.walkAnimation()
}
