// Package path_finding has logic and functions for finding paths in a map
package path_finding

import "github.com/webbben/2d-game-engine/model"

const (
	BlockThreshold int = 10 // if cost at a given tile reaches this level, then the tile is considered blocked and can't be used in a path.
)

// FindPath finds a path from the start position to the end position, navigating around barriers if possible.
// If the goal cannot be reached, a path to the nearest reachable point is returned.
// The boolean indicates if a path to the goal was successfully found.
//
// Notes:
//
//   - Q: why use a [][]int for costMap? would a map[Coords]int be faster?
//
//     A: No, [][]int is actually faster. If you know the index you are accessing, it's apparently faster than a map and less complex under the hood.
func FindPath(start, goal model.Coords, costMap [][]int) ([]model.Coords, bool) {
	foundPath, _, completePathFound := aStar(start, goal, costMap)
	return foundPath, completePathFound
}
