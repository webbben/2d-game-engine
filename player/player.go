package player

import (
	"fmt"
	"math"
	"time"

	"github.com/webbben/2d-game-engine/config"
	m "github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/rendering"

	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	m.Position
	CurrentFrame *ebiten.Image
	Frames       map[string]*ebiten.Image
}

const (
	movementSpeed = 0.025
	delay         = time.Millisecond * 6
)

var (
	direc_down  = []string{"down-1", "down-2", "down-3", "down-4", "down-5", "down-6", "down-7"}
	direc_up    = []string{"up-1", "up-2", "up-3", "up-4", "up-5", "up-6", "up-7"}
	direc_left  = []string{"left-1", "left-2", "left-3", "left-4", "left-5", "left-6", "left-7"}
	direc_right = []string{"right-1", "right-2", "right-3", "right-4", "right-5", "right-6", "right-7"}
)

func CreatePlayer(posX int, posY int, frames map[string]*ebiten.Image) Player {
	return Player{
		Position: m.Position{
			X:               float64(posX),
			Y:               float64(posY),
			Direction_Horiz: "X",
			Direction_Vert:  "X",
		},
		Frames:       frames,
		CurrentFrame: frames[direc_down[0]],
	}
}

func (p *Player) Draw(screen *ebiten.Image, offsetX float64, offsetY float64) {
	op := &ebiten.DrawImageOptions{}
	drawX, drawY := rendering.GetImageDrawPos(p.CurrentFrame, p.X, p.Y, offsetX, offsetY)
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(p.CurrentFrame, op)
}

func (p *Player) Update(barrierLayout [][]bool) {
	// handle player movement
	if p.IsMoving {
		p.move()
	}
	p.checkMovementInput(barrierLayout)
}

func (p *Player) checkMovementInput(barrierLayout [][]bool) {
	// UP
	if ebiten.IsKeyPressed(ebiten.KeyW) && p.Direction_Vert == "X" {
		if !ebiten.IsKeyPressed(ebiten.KeyS) {
			p.Direction_Vert = "U"
			p.IsMoving = true
			go p.continueWalking("U", barrierLayout)
		}
	}
	// DOWN
	if ebiten.IsKeyPressed(ebiten.KeyS) && p.Direction_Vert == "X" {
		if !ebiten.IsKeyPressed(ebiten.KeyW) {
			p.Direction_Vert = "D"
			p.IsMoving = true
			go p.continueWalking("D", barrierLayout)
		}
	}
	// LEFT
	if ebiten.IsKeyPressed(ebiten.KeyA) && p.Direction_Horiz == "X" {
		if !ebiten.IsKeyPressed(ebiten.KeyD) {
			p.Direction_Horiz = "L"
			p.IsMoving = true
			go p.continueWalking("L", barrierLayout)
		}
	}
	// RIGHT
	if ebiten.IsKeyPressed(ebiten.KeyD) && p.Direction_Horiz == "X" {
		if !ebiten.IsKeyPressed(ebiten.KeyA) {
			p.Direction_Horiz = "R"
			p.IsMoving = true
			go p.continueWalking("R", barrierLayout)
		}
	}
}

func (p *Player) continueWalking(direction string, barrierLayout [][]bool) {
	barrier := false
	if p.Facing != direction {
		time.Sleep(80 * time.Millisecond)
	}
	switch direction {
	case "U":
		// keep going until that key is released
		for ebiten.IsKeyPressed(ebiten.KeyW) {
			// check for barriers before continuing
			if p.movingTowardsBarrier(barrierLayout, direction) {
				barrier = true
				break
			}
			p.Y -= movementSpeed
			time.Sleep(delay)
		}
		// round the position out in case they overshot it
		p.easeToPosY(math.Floor(p.Y))
		p.Direction_Vert = "X"
	case "D":
		for ebiten.IsKeyPressed(ebiten.KeyS) {
			if p.movingTowardsBarrier(barrierLayout, direction) {
				barrier = true
				break
			}
			p.Y += movementSpeed
			time.Sleep(delay)
		}
		if barrier {
			p.easeToPosY(math.Floor(p.Y))
		} else {
			p.easeToPosY(math.Ceil(p.Y))
		}
		p.Direction_Vert = "X"
	case "L":
		for ebiten.IsKeyPressed(ebiten.KeyA) {
			if p.movingTowardsBarrier(barrierLayout, direction) {
				barrier = true
				break
			}
			p.X -= movementSpeed
			time.Sleep(delay)
		}
		p.easeToPosX(math.Floor(p.X))
		p.Direction_Horiz = "X"
	case "R":
		for ebiten.IsKeyPressed(ebiten.KeyD) {
			if p.movingTowardsBarrier(barrierLayout, direction) {
				barrier = true
				break
			}
			p.X += movementSpeed
			time.Sleep(delay)
		}
		if barrier {
			p.easeToPosX(math.Floor(p.X))
		} else {
			p.easeToPosX(math.Ceil(p.X))
		}
		p.Direction_Horiz = "X"
	}
}

func (p *Player) movingTowardsBarrier(barrierLayout [][]bool, dir string) bool {
	gridX := math.Floor(p.X) // current position
	gridY := math.Floor(p.Y)
	realX := p.X + 0.5 // player actually appears in the middle of the tile, not the top left corner
	realY := p.Y + 0.5
	nextX := int(gridX)
	nextY := int(gridY)
	dist := 1.0
	// For some reason I have to have special handling for U and L.
	// not sure why yet, but since it seems to work smoothly I won't question it!

	switch dir {
	case "U":
		nextY = int(realY - dist)
		if nextY < 0 {
			return true
		}
	case "D":
		nextY = int(math.Floor(gridY + dist))
		if nextY >= len(barrierLayout) {
			return true
		}
	case "L":
		nextX = int(realX - dist)
		if nextX < 0 {
			return true
		}
	case "R":
		nextX = int(math.Floor(gridX + dist))
		if nextX >= len(barrierLayout[0]) {
			return true
		}
	}
	if nextX < 0 || nextX >= len(barrierLayout[0]) {
		return true
	}
	if nextY < 0 || nextY >= len(barrierLayout) {
		return true
	}
	return barrierLayout[nextY][nextX]
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
	if p.IsMoving && (curTime.Sub(p.LastAnimStep) < 150*time.Millisecond) {
		return
	}
	p.LastAnimStep = curTime
	p.AnimStep++
	if p.AnimStep >= 7 || !p.IsMoving {
		p.AnimStep = 0
	}
	switch p.Facing {
	// UP
	case "U":
		p.setFrame(direc_up[p.AnimStep])
	// DOWN
	case "D":
		p.setFrame(direc_down[p.AnimStep])
	// LEFT
	case "L":
		p.setFrame(direc_left[p.AnimStep])
	// RIGHT
	case "R":
		p.setFrame(direc_right[p.AnimStep])
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
