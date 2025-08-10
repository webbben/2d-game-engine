package proc_gen

import (
	"fmt"
	"math/rand"
)

type Cell struct {
	X int
	Y int
}

type Island struct {
	value        int
	surroundedBy int
	cells        []Cell
}

// removes "islands" that are below the given size threshold.
//
// an "island" is a section on the noise map that is of one value and is surrounded by other values (including the edge of the map).
// technically, all bodies of the same value are islands, but this targets islands with an area smaller than the given threshold.
//
// noiseMap: noise map we are editing
//
// threshold: size (area) of islands that we want to remove;
// if an island's area is less than the threshold, it will be replaced by the value surrounding it.
func removeIslands(noiseMap [][]int, threshold int) {
	fmt.Printf("removing islands for threshold %v\n", threshold)
	islands := identifyIslands(noiseMap)
	// fill in all the islands that are smaller than the threshold
	for _, island := range islands {
		if len(island.cells) <= threshold {
			for _, cell := range island.cells {
				noiseMap[cell.Y][cell.X] = island.surroundedBy
			}
		}
	}
}

// identifies "islands" of same values in a noise map
//
// A noise map will entirely consist of islands, since all planes of values are then surrounded by other planes of a different (but consistent) value.
func identifyIslands(noiseMap [][]int) []Island {
	islands := []Island{}
	height := len(noiseMap)
	width := len(noiseMap[0])
	visited := make([][]bool, height)
	for r := range visited {
		visited[r] = make([]bool, width)
	}

	// get a list of all the islands (areas that are all of one value)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if !visited[y][x] {
				// dfs to find all cells and create an island
				island := Island{
					value:        noiseMap[y][x],
					surroundedBy: -1,
				}
				islandDFS(x, y, visited, noiseMap, &island)
				islands = append(islands, island)
			}
		}
	}
	return islands
}

func islandDFS(x, y int, visited [][]bool, noiseMap [][]int, island *Island) {
	// skip positions outside the map
	if x >= len(visited[0]) || y >= len(visited) || x < 0 || y < 0 {
		return
	}
	// skip positions of different values
	val := noiseMap[y][x]
	if val != island.value {
		if island.surroundedBy == -1 {
			island.surroundedBy = val
		}
		return
	}
	// skip positions already visited
	if visited[y][x] {
		return
	}
	visited[y][x] = true
	island.cells = append(island.cells, Cell{X: x, Y: y})
	dirs := [][]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, dir := range dirs {
		dy := dir[0]
		dx := dir[1]
		islandDFS(x+dx, y+dy, visited, noiseMap, island)
	}
}

// "thins out" the given noise map; randomly removes values from it to make it less dense.
//
// p - percentage of the map to remove, between 0 and 1. 1 is 100%, which means everything is removed, and 0 means nothing is removed.
func thinOut(noiseMap [][]int, p float64) {
	height := len(noiseMap)
	width := len(noiseMap[0])
	zero := noiseMap[0][0]

	// identify the "zero value"; the min value which will be used when thinning out
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if noiseMap[y][x] < zero {
				zero = noiseMap[y][x]
			}
		}
	}
	// replace values in noiseMap with zero value, based on p percentage
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if rand.Float64() <= p {
				noiseMap[y][x] = zero
			}
		}
	}
}
