package entity

import (
	"fmt"
	"math"
	"time"

	"github.com/webbben/2d-game-engine/internal/general_util"
	m "github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/player"
)

const (
	delay = time.Millisecond * 6
)

// Tells this entity to move to the given coordinates. Includes an optional "stopChannel" to tell this code to quit; mainly intended
// for redirecting the entity if its goal changes. If you don't want to use a stopChannel, just pass in nil for this parameter instead.
//
// coords: coordinates to travel to
//
// barrierMap: map of barriers the entity must avoid in the room
//
// stopChan: a channel to tell this function to stop, if it's running in its own go-routine
//
// nextTo: if you should travel next to the position, but not directly onto the position (e.g. traveling to another entity)
func (e *Entity) TravelToPosition(coords m.Coords, barrierMap [][]bool, stopChan <-chan struct{}, nextTo bool) {
	curPos := m.Coords{
		X: int(e.Position.X),
		Y: int(e.Position.Y),
	}
	path := path_finding.FindPath(curPos, coords, barrierMap, nil)
	if len(path) == 0 {
		fmt.Println("entity failed to find path to goal")
		return
	}
	if nextTo && len(path) > 1 {
		path = path[:len(path)-1]
	}
	// go along the path
	// TODO: add collision detection for other entities
	e.IsMoving = true
	stopAnimChan := make(chan struct{})
	defer func() {
		stopAnimChan <- struct{}{}
		e.IsMoving = false
	}()
	go e.startWalkingAnimation(stopAnimChan)

	for _, pos := range path {
		select {
		case <-stopChan:
			e.snapToGrid()
			return
		default:
			e.moveToCoords(pos)
		}
	}
	e.snapToGrid()
}

func (e *Entity) FollowPlayer(p *player.Player, barrierMap [][]bool) {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		fmt.Println("ticker stopped!")
	}()
	stopTravelChan := make(chan struct{})
	// defer close(stopTravelChan)
	targetLastPos := m.Coords{
		X: int(p.X),
		Y: int(p.Y),
	}
	targetPos := m.Coords{
		X: int(p.X),
		Y: int(p.Y),
	}

	go func() {
		for {
			select {
			case <-stopTravelChan:
				// keep listening on the channel even if the entity isn't travelling anywhere
			default:
				targetPos = m.Coords{
					X: int(p.X),
					Y: int(p.Y),
				}
				// continuously travel to the current target position. if stopTravelChan receives a stop signal,
				// then TravelToPosition returns early and we recalculate the route
				if dist := general_util.EuclideanDist(e.X, e.Y, float64(targetPos.X), float64(targetPos.Y)); dist > 1 {
					e.TravelToPosition(targetPos, barrierMap, stopTravelChan, true)
				}
			}

		}
	}()

	// every tick, check if the player has moved from the original position the entity is travelling to
	for range ticker.C {
		targetPos = m.Coords{
			X: int(p.X),
			Y: int(p.Y),
		}
		// send signal to interrupt the current travel path, if we need to reroute
		if general_util.EuclideanDistCoords(targetLastPos, targetPos) > 1 {
			stopTravelChan <- struct{}{}
		}
		targetLastPos = targetPos.Copy()
	}
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
	e.fadeToPosition(coords, e.MovementSpeed)
}

func (e *Entity) fadeToPosition(coords m.Coords, speed float64) {
	growX := e.X < float64(coords.X)
	growY := e.Y < float64(coords.Y)
	moveX, moveY := true, true

	for moveX || moveY {
		if moveX {
			if growX {
				e.X += speed
				if e.X >= float64(coords.X) {
					moveX = false
				}
			} else {
				e.X -= speed
				if e.X <= float64(coords.X) {
					moveX = false
				}
			}
		}
		if moveY {
			if growY {
				e.Y += speed
				if e.Y >= float64(coords.Y) {
					moveY = false
				}
			} else {
				e.Y -= speed
				if e.Y <= float64(coords.Y) {
					moveY = false
				}
			}
		}
		time.Sleep(delay)
	}
}

// starts a walking animation. the animation won't stop unless stopChan sends a signal.
func (e *Entity) startWalkingAnimation(stopChan chan struct{}) {
	ticker := time.NewTicker(time.Millisecond * 150)
	defer func() {
		ticker.Stop()
		e.switchToRestFrame()
	}()

	for {
		select {
		case <-ticker.C:
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
		case <-stopChan:
			return
		}
	}
}

func (e *Entity) snapToGrid() {
	e.Position.X = math.Round(e.Position.X)
	e.Position.Y = math.Round(e.Position.Y)
}
