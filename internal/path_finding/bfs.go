package path_finding

import (
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/utils"
)

// GetAllReachablePositions finds all tile positions that are reachable starting from the given position.
// Note: this will not include start in the returned slice of reachable positions, because we don't consider that to be "reachable".
func GetAllReachablePositions(start model.Coords, costMap [][]int) []model.Coords {
	// the positions we haven't explored yet
	open := []model.Coords{start}
	// use this set to track what's in open, so we don't have to do an O(n) search each loop
	openSet := map[model.Coords]bool{start: true}
	// the positions that we've explored already
	closed := make(map[model.Coords]bool)

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

	visited := []model.Coords{}
	for c := range closed {
		visited = append(visited, c)
	}
	return visited
}

// FindNearestOpenPosition does a BFS to find the nearest open position to the given position.
// If the given position is open, then this will just return that position.
func FindNearestOpenPosition(c model.Coords, distLimit int, costMap [][]int) (model.Coords, bool) {
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
	open := []model.Coords{c}
	// use this set to track what's in open, so we don't have to do an O(n) search each loop
	openSet := map[model.Coords]bool{c: true}
	// the positions that we've explored already
	closed := make(map[model.Coords]bool)

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
	return model.Coords{}, false
}

// BuildDistanceMap builds a map of how distant every position is from the `from` position.
// You can optionally supply a center and search radius to limit the area of the distance map, as an optimization.
//
// params:
//   - from: the position we are charting distances from
//   - center: (OPT) if set, the returned distance map will only be populated in a certain radius based on this position
//   - searchRadius: (OPT) if center is set, this must be set too (> 0). the search radius around center.
//   - costMap: the costMap that represents the map we are building the distance map for.
//
// returns the distMap and the index of the most distant position.
func BuildDistanceMap(from []model.Coords, center *model.Coords, searchRadius int, costMap [][]int) (distMap []int, farthestIdx int) {
	utils.PanicAssert(len(from) > 0, "no coordinates passed!")
	height := len(costMap)
	utils.PanicAssert(height > 0, "costmap was empty!")
	width := len(costMap[0])
	utils.PanicAssert(width > 0, "costmap has no width!")

	if center != nil {
		// we don't check isValidPosition here since it's possible for an NPC to slide into a step door object - which, while not technically a collision,
		// is a pathfinding block, and so isValidPosition returns false. instead, we will do the same as what we do for `from` positions and just
		// ensure there are positions to travel to from here
		if len(getNeighbors(*center, costMap)) == 0 {
			logz.PanicCtx("BuildDistanceMap", "center position was invalid! no way to travel from it", center.String())
		}
		utils.PanicAssert(searchRadius > 0, "search radius must be > 0 if center is defined")

	}

	for _, c := range from {
		// ensure there is a possible path from the start position
		// note that we don't care if the position itself is invalid/a collision, because the search won't directly check that position.
		// we only care that it's possible to travel from that position somewhere else.
		if len(getNeighbors(c, costMap)) == 0 {
			logz.PanicCtx("BuildDistanceMap", "from position was invalid! no way to travel from it", c.String())
		}
		if center != nil {
			// ensure that `from` is within the search radius. otherwise no distMap can be created
			if utils.ManhattanDistCoords(c, *center) > searchRadius {
				logz.PanicCtx("BuildDistanceMap", "from is outside search radius range of center!", c, center, searchRadius)
			}
		}
	}

	size := width * height

	distances := make([]int, size)
	for i := range distances {
		distances[i] = -1
	}

	open := make([]model.Coords, 0, len(from))

	for _, threat := range from {
		threatIdx := idx(threat, width)
		distances[threatIdx] = 0
		open = append(open, threat)
	}

	best := from[0]
	bestDistance := 0

	for head := 0; head < len(open); head++ {
		current := open[head]
		currentIdx := idx(current, width)
		currentDist := distances[currentIdx]

		// track the furthest position we've found within the search radius
		if currentDist > bestDistance {
			best = current
			bestDistance = currentDist
		}

		neighbors := getNeighbors(current, costMap)
		for _, neighbor := range neighbors {
			neighborIdx := idx(neighbor, width)

			// already visited
			if distances[neighborIdx] != -1 {
				continue
			}

			if center != nil {
				// only search the area within the radius
				if utils.ManhattanDistCoords(*center, neighbor) > searchRadius {
					continue
				}
			}

			distances[neighborIdx] = currentDist + 1
			open = append(open, neighbor)
		}
	}

	return distances, idx(best, width)
}

func FleeFromPositions(from []model.Coords, start model.Coords, searchRadius int, costMap [][]int) (fleePath []model.Coords, reachable bool, cannotFlee bool) {
	// first, get the distance map
	distMap, _ := BuildDistanceMap(from, &start, searchRadius, costMap)

	height := len(costMap)
	utils.PanicAssert(height > 0, "costmap is empty!")
	width := len(costMap[0])
	utils.PanicAssert(width > 0, "costmap has no width!")

	startIdx := idx(start, width)
	startDist := distMap[startIdx]
	if startDist == -1 {
		// the distance map never reached the NPC
		// this implies the position we are fleeing from is not accessible anyway, so fleeing probably isn't necessary
		return nil, false, false
	}

	// Do BFS from start, using distMap to guide which path to explore
	size := width * height
	parent := make([]int, size)
	for i := range parent {
		parent[i] = -1
	}
	parent[startIdx] = startIdx
	open := []model.Coords{start}

	best := start
	bestDist := startDist

	// explore outward from start, but only along tiles whose distance from the threat is strictly increasing
	for head := 0; head < len(open); head++ {
		current := open[head]
		currentIdx := idx(current, width)
		currentDist := distMap[currentIdx]

		if currentDist > bestDist {
			// new furthest tile we've found so far
			best = current
			bestDist = currentDist
		}

		for _, neighbor := range getNeighbors(current, costMap) {
			neighborIdx := idx(neighbor, width)

			// the neighbor wasn't reached by build distance map
			if distMap[neighborIdx] == -1 {
				continue
			}
			// only move in directions that are increasing in distance
			if distMap[neighborIdx] != currentDist+1 {
				continue
			}
			// Already visited
			if parent[neighborIdx] != -1 {
				continue
			}

			parent[neighborIdx] = currentIdx
			open = append(open, neighbor)
		}
	}

	if best == start {
		// nowhere to move while getting further away from flee target
		return nil, true, true
	}

	path := reconstructPath(parent, width, start, best)
	return path, true, false
}
