package model

import (
	"fmt"
	"math"
)

/*

Types related to world layouts and navigating maps, etc.

*/

type directions struct {
	Left  byte
	Right byte
	Up    byte
	Down  byte
}

var Directions directions = directions{
	Left:  'L',
	Right: 'R',
	Up:    'U',
	Down:  'D',
}

func GetOppositeDirection(dir byte) byte {
	switch dir {
	case Directions.Left:
		return Directions.Right
	case Directions.Right:
		return Directions.Left
	case Directions.Up:
		return Directions.Down
	case Directions.Down:
		return Directions.Up
	default:
		panic("invalid direction given!")
	}
}

// gets the direction that b lies in relative to a
func GetRelativeDirection(a, b Coords) byte {
	dx := b.X - a.X
	dy := b.Y - a.Y
	if math.Abs(float64(dx)) > math.Abs(float64(dy)) {
		if dx < 0 {
			return Directions.Left
		} else {
			return Directions.Right
		}
	} else {
		if dy < 0 {
			return Directions.Up
		} else {
			return Directions.Down
		}
	}
}

// tile-based coordinate position in a room
type Coords struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (c Coords) String() string {
	return fmt.Sprintf("[X: %v, Y: %v]", c.X, c.Y)
}

func (c Coords) Equals(other Coords) bool {
	return c.X == other.X && c.Y == other.Y
}

// returns a copy of the coords struct, to avoid reference ties
func (c Coords) Copy() Coords {
	return Coords{X: c.X, Y: c.Y}
}

func (c Coords) GetAdj(direction byte) Coords {
	adj := c.Copy()
	switch direction {
	case 'L':
		adj.X--
	case 'R':
		adj.X++
	case 'U':
		adj.Y--
	case 'D':
		adj.Y++
	default:
		panic("GetAdj: invalid direction passed")
	}
	return adj
}
