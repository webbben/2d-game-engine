package entity

import (
	"container/heap"
	"math"
)

type Pos struct {
	X float64
	Y float64
}

func FindPath(start, goal Pos, barrierMap [][]bool) []Point {
	startPoint := Point{X: int(start.X), Y: int(start.Y)}
	goalPoint := Point{X: int(goal.X), Y: int(goal.Y)}
	return aStar(barrierMap, startPoint, goalPoint)
}

type Point struct {
	X, Y int
}

type Node struct {
	Point
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

func (pq PriorityQueue) contains(p Point) bool {
	for _, node := range pq {
		if node.Point == p {
			return true
		}
	}
	return false
}

func aStar(barrierMap [][]bool, start, goal Point) []Point {
	open := make(PriorityQueue, 0)
	closed := make(map[Point]bool)

	heap.Init(&open)
	heap.Push(&open, &Node{Point: start})

	// map which point leads to the next for reconstructing the path
	parent := make(map[Point]Point)

	gValues := make(map[Point]int)
	hValues := make(map[Point]int)
	gValues[start] = 0
	hValues[start] = heuristic(start, goal)

	for open.Len() > 0 {
		// get the current best option to explore
		current := heap.Pop(&open).(Node).Point

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
					heap.Push(&open, &Node{Point: neighbor, G: tentativeG})
				}
			}
		}
	}
	return nil
}

func heuristic(start, goal Point) int {
	return int(math.Abs(float64(start.X)-float64(goal.X)) + math.Abs(float64(start.Y)-float64(goal.Y)))
}

func reconstructPath(parent map[Point]Point, start, goal Point) []Point {
	path := make([]Point, 0)
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

func getNeighbors(current Point, barrierMap [][]bool) []Point {
	neighbors := []Point{
		{X: current.X, Y: current.Y - 1}, // UP
		{X: current.X, Y: current.Y + 1}, // DOWN
		{X: current.X - 1, Y: current.Y}, // LEFT
		{X: current.X + 1, Y: current.Y}, // RIGHT
	}
	validNeighbors := make([]Point, 0)
	for _, neighbor := range neighbors {
		if isValidPoint(neighbor, barrierMap) {
			validNeighbors = append(validNeighbors, neighbor)
		}
	}
	return validNeighbors
}

func isValidPoint(p Point, barrierMap [][]bool) bool {
	if p.X < 0 || p.X >= len(barrierMap[0]) {
		return false
	}
	if p.Y < 0 || p.Y >= len(barrierMap) {
		return false
	}
	return !barrierMap[p.Y][p.X]
}
