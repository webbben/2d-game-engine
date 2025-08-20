package model

import "fmt"

/*

Types related to world layouts and navigating rooms, etc.

*/

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
