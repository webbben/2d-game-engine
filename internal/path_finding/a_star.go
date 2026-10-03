package path_finding

import (
	"container/heap"

	"github.com/webbben/2d-game-engine/logz"
	m "github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/utils"
)

type Node struct {
	m.Coords
	// known cost of reaching a node
	//
	// sum of the edge costs from the start node to this node (i.e. cost of the path traveled so far)
	G int
	// heuristic estimate of cost to travel from this node to the goal
	H int
}

type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	fi := pq[i].G + pq[i].H
	fj := pq[j].G + pq[j].H
	if fi != fj {
		return fi < fj
	}
	// tie-breaker: prefer the one nearer to the goal based on heuristic alone
	return pq[i].H < pq[j].H
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(n interface{}) {
	*pq = append(*pq, n.(*Node))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := old.Len()
	node := old[n-1]
	*pq = old[0 : n-1]
	return node
}

func (pq PriorityQueue) Contains(p m.Coords) bool {
	for _, node := range pq {
		if node.Coords == p {
			return true
		}
	}
	return false
}

// performs A* search and returns a path to the goal, or to the closest reachable node to the goal.
// returns true if successfully reached the goal, or false if not.
//
// Revisions:
//   - 2026-10-03: start tracking H values in Nodes, ignore stale G values, use slices instead of maps to optimize performance
func aStar(start, goal m.Coords, costMap [][]int) (foundPath []m.Coords, visited []bool, completePathFound bool) {
	utils.PanicAssert(costMap != nil, "costmap was nil!")
	if start.Equals(goal) {
		logz.Warnln("aStar", "start and goal are the same position")
		return []m.Coords{}, nil, false
	}

	height := len(costMap)
	if height == 0 {
		logz.Panicln("PATH_FINDING", "costmap is empty!")
	}
	width := len(costMap[0])
	if width == 0 {
		logz.Panicln("PATH_FINDING", "costmap has no width!")
	}

	if start.X < 0 || start.Y < 0 || start.X >= width || start.Y >= height {
		logz.PanicCtx("PATH_FINDING", "start position is outside the cost map!", start.String(), width, height)
	}

	size := width * height

	// the positions we haven't explored yet
	open := make(PriorityQueue, 0)
	heap.Init(&open)

	// the positions that we've explored already
	// we use slices instead of maps to improve performance by reducing allocations
	closed := make([]bool, size)
	gValues := make([]int, size)
	hValues := make([]int, size)

	// map which tile position leads to the next for reconstructing the path
	// holds parent tile's index within the slice
	parent := make([]int, size)

	for i := range size {
		gValues[i] = -1 // not discovered yet
		parent[i] = -1  // no parent
	}

	startIdx := start.Y*width + start.X
	gValues[startIdx] = 0
	hValues[startIdx] = heuristic(start, goal)
	heap.Push(&open, &Node{Coords: start})

	// track the closest position found to the goal we've seen so far
	closest := start.Copy()
	closestDist := hValues[startIdx]

	for open.Len() > 0 {
		// get the current best option to explore
		node := heap.Pop(&open).(*Node)
		current := node.Coords
		currentIdx := current.Y*width + current.X

		// a node may have been pushed more than once (once per improvement)
		// an entry whose G no longer matches the best known cost is stale, so skip it
		// (it will reappear later with the correct G)
		if node.G != gValues[currentIdx] {
			continue
		}

		// update closest found position
		if hValues[currentIdx] < closestDist {
			closest = current.Copy()
			closestDist = hValues[currentIdx]
		}

		if current == goal {
			return reconstructPath(parent, width, start, goal), closed, true
		}
		closed[currentIdx] = true

		// explore neighbors
		neighbors := getNeighbors(current, costMap)
		for _, neighbor := range neighbors {
			neighborIdx := neighbor.Y*width + neighbor.X
			if closed[neighborIdx] {
				continue
			}
			// add other costs, such as terrain, which can influence the path
			tentativeG := gValues[currentIdx] + 1
			tentativeG += costMap[neighbor.Y][neighbor.X]

			// check if tentative g value is better than current one
			// note: gValues holds no entry for a node we haven't discovered yet, and a missing map key reads as 0.
			// so the "not seen" case must be tested explicitly.
			prevG := gValues[neighborIdx]
			if prevG == -1 || tentativeG < prevG {
				gValues[neighborIdx] = tentativeG
				hValues[neighborIdx] = heuristic(neighbor, goal)
				parent[neighborIdx] = currentIdx

				heap.Push(&open, &Node{Coords: neighbor, G: tentativeG, H: hValues[neighborIdx]})
			}
		}
	}

	// if we didn't reach the goal, return the path to the closest found position
	// (but, of course, if we didn't get past the start position, then there's no point)
	if !closest.Equals(start) {
		p := reconstructPath(parent, width, start, closest)
		return p, closed, false
	}

	return nil, nil, false
}

func heuristic(start, goal m.Coords) int {
	return utils.ManhattanDist(start.X, start.Y, goal.X, goal.Y)
}

func reconstructPath(parent []int, width int, start, goal m.Coords) []m.Coords {
	utils.PanicAssert(width >= 1, "width was <= 0")
	path := make([]m.Coords, 0)
	current := goal

	for current != start {
		path = append(path, current)
		next := parent[current.Y*width+current.X]
		if next < 0 {
			// shouldn't be reachable; every tile we return was discovered first, and discovery always sets a parent.
			logz.PanicCtx("PATH_FINDING", "reconstructPath: parent chain was broken.", current)
		}
		current = m.Coords{X: next % width, Y: next / width}
	}

	// reverse
	for i := 0; i < len(path)/2; i++ {
		j := len(path) - 1 - i
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func getNeighbors(current m.Coords, costMap [][]int) []m.Coords {
	neighbors := []m.Coords{
		{X: current.X, Y: current.Y - 1}, // UP
		{X: current.X, Y: current.Y + 1}, // DOWN
		{X: current.X - 1, Y: current.Y}, // LEFT
		{X: current.X + 1, Y: current.Y}, // RIGHT
	}
	validNeighbors := make([]m.Coords, 0)
	for _, neighbor := range neighbors {
		if isValidCoords(neighbor, costMap) {
			validNeighbors = append(validNeighbors, neighbor)
		}
	}
	return validNeighbors
}

func isValidCoords(p m.Coords, costMap [][]int) bool {
	if len(costMap) == 0 || len(costMap[0]) == 0 {
		logz.Panicln("PATH_FINDING", "costmap was empty!")
	}
	if p.X < 0 || p.X >= len(costMap[0]) {
		return false
	}
	if p.Y < 0 || p.Y >= len(costMap) {
		return false
	}
	return costMap[p.Y][p.X] < BlockThreshold
}
