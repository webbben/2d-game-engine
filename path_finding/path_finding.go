package path_finding

import "ancient-rome/model"

type Pos struct {
	X float64
	Y float64
}

// Finds a path from the start position to the end position, navigating around barriers if possible.
func FindPath(start, goal model.Coords, barrierMap [][]bool) []model.Coords {
	return aStar(barrierMap, start, goal)
}
