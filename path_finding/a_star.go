package path_finding

import (
	"ancient-rome/general_util"
	m "ancient-rome/model"
	"container/heap"
	"fmt"
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

func aStar(barrierMap [][]bool, start, goal m.Coords) []m.Coords {
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

	for open.Len() > 0 {
		// get the current best option to explore
		current := heap.Pop(&open).(Node).Coords

		if current == goal {
			return reconstructPath(parent, start, goal)
		}
		closed[current] = true

		// explore neighbors
		neighbors := getNeighbors(current, barrierMap)
		for _, neighbor := range neighbors {
			if closed[neighbor] {
				continue
			}

			tentativeG := gValues[current] + 1
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
	return nil
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

func getNeighbors(current m.Coords, barrierMap [][]bool) []m.Coords {
	neighbors := []m.Coords{
		{X: current.X, Y: current.Y - 1}, // UP
		{X: current.X, Y: current.Y + 1}, // DOWN
		{X: current.X - 1, Y: current.Y}, // LEFT
		{X: current.X + 1, Y: current.Y}, // RIGHT
	}
	validNeighbors := make([]m.Coords, 0)
	for _, neighbor := range neighbors {
		if isValidCoords(neighbor, barrierMap) {
			validNeighbors = append(validNeighbors, neighbor)
		}
	}
	return validNeighbors
}

func isValidCoords(p m.Coords, barrierMap [][]bool) bool {
	if len(barrierMap) == 0 || len(barrierMap[0]) == 0 {
		fmt.Println("isValidCoords: barrier map is empty!")
		return false
	}
	if p.X < 0 || p.X >= len(barrierMap[0]) {
		return false
	}
	if p.Y < 0 || p.Y >= len(barrierMap) {
		return false
	}
	return !barrierMap[p.Y][p.X]
}
