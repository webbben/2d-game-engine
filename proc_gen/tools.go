package proc_gen

import "fmt"

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
				fmt.Println("exploring new island")
				islandDFS(x, y, visited, noiseMap, &island)
				fmt.Printf("found island of value %v of size %v\n", island.value, len(island.cells))
				islands = append(islands, island)
			}
		}
	}

	// fill in all the islands that are smaller than the threshold
	for _, island := range islands {
		if len(island.cells) <= threshold {
			fmt.Println("removing island")
			for _, cell := range island.cells {
				noiseMap[cell.Y][cell.X] = island.surroundedBy
			}
		}
	}
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
