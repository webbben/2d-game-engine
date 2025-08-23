package path_finding

import "github.com/webbben/2d-game-engine/internal/model"

type Pos struct {
	X float64
	Y float64
}

// Finds a path from the start position to the end position, navigating around barriers if possible.
// If the goal cannot be reached, a path to the nearest reachable point is returned.
// The boolean indicates if a path to the goal was successfully found.
func FindPath(start, goal model.Coords, costMap [][]int) ([]model.Coords, bool) {
	return aStar(start, goal, costMap)
}
