package entity

import (
	"fmt"
	"sync"
	"time"

	m "github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/path_finding"
)

const (
	delay = time.Millisecond * 6
)

// tell an entity to move to the given position
func (e *Entity) TravelToPosition(coords m.Coords, barrierMap [][]bool) {
	curPos := m.Coords{
		X: int(e.Position.X),
		Y: int(e.Position.Y),
	}
	path := path_finding.FindPath(curPos, coords, barrierMap)
	if len(path) == 0 {
		fmt.Println("entity failed to find path to goal")
		return
	}
	// go along the path
	// TODO: add collision detection for other entities
	go e.startWalkingAnimation()
	for _, pos := range path {
		e.moveToCoords(pos)
	}
	e.IsMoving = false
}

// moves entity to a specific position.
//
// meant to represent a single movement to an adjacent position
func (e *Entity) moveToCoords(coords m.Coords) {
	// set facing direction
	if coords.X > int(e.X) {
		e.Facing = "R"
	} else if coords.X < int(e.X) {
		e.Facing = "L"
	} else if coords.Y < int(e.Y) {
		e.Facing = "U"
	} else {
		e.Facing = "D"
	}
	// move vertically and horizontally to the goal position
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		e.easeToPosY(float64(coords.Y))
	}()
	go func() {
		defer wg.Done()
		e.easeToPosX(float64(coords.X))
	}()
	wg.Wait()
}

func (e *Entity) easeToPosY(destY float64) {
	if destY > e.Y {
		// moving down
		for destY-e.Y >= e.MovementSpeed {
			e.Y += e.MovementSpeed
			time.Sleep(delay)
		}
	} else if e.Y > destY {
		// moving up
		for e.Y-destY >= e.MovementSpeed {
			e.Y -= e.MovementSpeed
			time.Sleep(delay)
		}
	}
	e.Position.Y = destY
}
func (e *Entity) easeToPosX(destX float64) {
	if destX > e.X {
		// moving right
		for destX-e.X >= e.MovementSpeed {
			e.X += e.MovementSpeed
			time.Sleep(delay)
		}
	} else if e.Position.X > destX {
		// moving left
		for e.X-destX >= e.MovementSpeed {
			e.X -= e.MovementSpeed
			time.Sleep(delay)
		}
	}
	e.X = destX
}

func (e *Entity) startWalkingAnimation() {
	e.IsMoving = true
	for e.IsMoving {
		// go to the next animation frame
		switch e.Facing {
		case "L":
			e.setNextAnimationFrame("left")
		case "R":
			e.setNextAnimationFrame("right")
		case "U":
			e.setNextAnimationFrame("up")
		case "D":
			e.setNextAnimationFrame("down")
		}
		time.Sleep(150 * time.Millisecond)
	}
}
