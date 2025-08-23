package path_finding

import (
	"container/heap"
	"fmt"

	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/logz"
	m "github.com/webbben/2d-game-engine/internal/model"
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
	return pq[i].G+pq[i].H < pq[j].G+pq[j].H
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
	return *node
}

func (pq PriorityQueue) contains(p m.Coords) bool {
	for _, node := range pq {
		if node.Coords == p {
			return true
		}
	}
	return false
}

// performs A* search and returns a path to the goal, or to the closest reachable node to the goal.
// returns true if successfully reached the goal, or false if not.
func aStar(start, goal m.Coords, costMap [][]int) ([]m.Coords, bool) {
	if start.Equals(goal) {
		logz.Warnln("aStar", "start and goal are the same position")
		return []m.Coords{}, false
	}
	open := make(PriorityQueue, 0)
	closed := make(map[m.Coords]bool)

	heap.Init(&open)
	heap.Push(&open, &Node{Coords: start})

	// map which m.Coords leads to the next for reconstructing the path
	parent := make(map[m.Coords]m.Coords)

	gValues := make(map[m.Coords]int)
	hValues := make(map[m.Coords]int)
	gValues[start] = 0
	hValues[start] = heuristic(start, goal)

	// track the closest position found to the goal we've seen so far
	closest := start.Copy()
	closestDist := hValues[start]

	for open.Len() > 0 {
		// get the current best option to explore
		current := heap.Pop(&open).(Node).Coords

		// update closest found position
		if hValues[current] < closestDist {
			closest = current.Copy()
			closestDist = hValues[current]
		}

		if current == goal {
			return reconstructPath(parent, start, goal), true
		}
		closed[current] = true

		// explore neighbors
		neighbors := getNeighbors(current, costMap)
		for _, neighbor := range neighbors {
			if closed[neighbor] {
				continue
			}
			// add other costs, such as terrain, which can influence the path
			tentativeG := gValues[current] + 1
			if costMap != nil {
				tentativeG += costMap[neighbor.Y][neighbor.X]
			}
			// check if tentative g value is better than current one
			if !open.contains(neighbor) || tentativeG < gValues[neighbor] {
				gValues[neighbor] = tentativeG
				hValues[neighbor] = heuristic(neighbor, goal)
				parent[neighbor] = current

				if !open.contains(neighbor) {
					heap.Push(&open, &Node{Coords: neighbor, G: tentativeG})
				}
			}
		}
	}

	// if we didn't reach the goal, return the path to the closest found position
	if !closest.Equals(start) {
		return reconstructPath(parent, start, closest), false
	}

	return nil, false
}

// use euclidean distance or manhattan distance?
func heuristic(start, goal m.Coords) int {
	// return int(math.Abs(float64(start.X)-float64(goal.X)) + math.Abs(float64(start.Y)-float64(goal.Y)))
	return int(general_util.EuclideanDist(float64(start.X), float64(start.Y), float64(goal.X), float64(goal.Y)))
}

func reconstructPath(parent map[m.Coords]m.Coords, start, goal m.Coords) []m.Coords {
	path := make([]m.Coords, 0)
	current := goal

	for current != start {
		path = append(path, current)
		current = parent[current]
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
		fmt.Println("isValidCoords: barrier map is empty!")
		return false
	}
	if p.X < 0 || p.X >= len(costMap[0]) {
		return false
	}
	if p.Y < 0 || p.Y >= len(costMap) {
		return false
	}
	return costMap[p.Y][p.X] < 10
}
