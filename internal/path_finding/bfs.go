package path_finding

import (
	"github.com/webbben/2d-game-engine/logz"
	m "github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/utils"
)

// GetAllReachablePositions finds all tile positions that are reachable starting from the given position.
// Note: this will not include start in the returned slice of reachable positions, because we don't consider that to be "reachable".
func GetAllReachablePositions(start m.Coords, costMap [][]int) []m.Coords {
	// the positions we haven't explored yet
	open := []m.Coords{start}
	// use this set to track what's in open, so we don't have to do an O(n) search each loop
	openSet := map[m.Coords]bool{start: true}
	// the positions that we've explored already
	closed := make(map[m.Coords]bool)

	for len(open) > 0 {
		// get the current best option to explore
		current := open[0]
		open = open[1:]
		delete(openSet, current)

		closed[current] = true

		// explore neighbors
		neighbors := getNeighbors(current, costMap)
		for _, neighbor := range neighbors {
			if closed[neighbor] {
				continue
			}
			if openSet[neighbor] {
				continue
			}
			open = append(open, neighbor)
			openSet[neighbor] = true
		}
	}

	// delete start from closed, since the start position shouldn't be included as a "reachable position".
	delete(closed, start)

	visited := []m.Coords{}
	for c := range closed {
		visited = append(visited, c)
	}
	return visited
}

// FindNearestOpenPosition does a BFS to find the nearest open position to the given position.
// If the given position is open, then this will just return that position.
func FindNearestOpenPosition(c m.Coords, distLimit int, costMap [][]int) (m.Coords, bool) {
	rows := len(costMap)
	if rows == 0 {
		panic("costmap had no rows")
	}
	cols := len(costMap[0])
	if cols == 0 {
		panic("costmap had no columns")
	}
	if c.X < 0 || c.Y < 0 || c.X >= cols || c.Y >= rows {
		panic("given position was outside the bounds of the cost map")
	}
	if distLimit <= 0 {
		panic("distLimit must be greater than 0")
	}

	// the positions we haven't explored yet
	open := []m.Coords{c}
	// use this set to track what's in open, so we don't have to do an O(n) search each loop
	openSet := map[m.Coords]bool{c: true}
	// the positions that we've explored already
	closed := make(map[m.Coords]bool)

	for len(open) > 0 {
		// get the current best option to explore
		current := open[0]
		open = open[1:]
		delete(openSet, current)

		closed[current] = true

		logz.Println("BFS", current)

		if costMap[current.Y][current.X] < BlockThreshold {
			// found an open spot
			return current, true
		}

		// explore neighbors
		neighbors := getNeighbors(current, costMap)
		for _, neighbor := range neighbors {
			if closed[neighbor] {
				continue
			}
			if openSet[neighbor] {
				continue
			}
			// make sure new neighbor option isn't too far away
			if utils.EuclideanDistCoords(c, neighbor) > float64(distLimit) {

				logz.Println("BFS", "too far away:", neighbor)
				continue
			}
			logz.Println("BFS", "neighbor:", neighbor)
			open = append(open, neighbor)
			openSet[neighbor] = true
		}
	}

	// no open spot found...
	logz.Println("BFS", "no spot found")
	return m.Coords{}, false
}
