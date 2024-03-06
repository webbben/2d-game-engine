package room

import (
	"ancient-rome/general_util"
	m "ancient-rome/model"
	"ancient-rome/path_finding"
	"ancient-rome/proc_gen"
	"ancient-rome/tileset"
	"fmt"
	"math"
	"strings"
)

type TownCenter struct {
	Size int `json:"size"`
	X    int `json:"x"`
	Y    int `json:"y"`
}

type Road struct {
	Path []m.Coords `json:"path"`
}

// data as loaded from the room_layout json
type RoomData struct {
	RoomName        string                `json:"roomName"`        // Display name of the room
	Width           int                   `json:"width"`           // width of the room
	Height          int                   `json:"height"`          // height of the room
	Tilesets        []string              `json:"tilesets"`        // names of the tilesets used in this room
	TileGroupKeys   []string              `json:"tileGroupKeys"`   // keys of the custom groups of tiles for this room
	TileGroups      map[string][]string   `json:"tileGroups"`      // custom groups of tiles that can be used interchangeably, to "randomize" the layout a bit
	TileLayout      [][]string            `json:"tileLayout"`      // the keys for tiles at each position of the room. a key may point to a tile group, or be the key of an individual tile
	BarrierLayout   [][]bool              `json:"barrierLayout"`   // a map of which tiles are accessible (0) and which are barriers/inaccessible (1)
	ObjectSets      []string              `json:"objectSets"`      // names of the tilesets used for objects in this room
	ObjectGroupKeys []string              `json:"objectGroupKeys"` // keys of the object groups for this room
	ObjectGroups    map[string][]string   `json:"objectGroups"`    // custom groups of objects that can be used interchangeably, to "randomize" the objects a bit
	ObjectLayout    [][]string            `json:"objectLayout"`    // the keys for objects at each position of the room. a key may point to a object group, or be the key of an individual object. '-' indicates no object in the given position.
	ElevationMap    [][]int               `json:"elevationMap"`
	CliffMap        map[string][]m.Coords `json:"cliffMap"`
	SlopeMap        map[string][]m.Coords `json:"slopeMap"`
	TownCenter      TownCenter            `json:"townCenter"`
	Roads           []Road                `json:"roads"`
}

// creates a room with randomized grass terrain and some hills
func GenerateRandomRoom(roomName string, width, height int) {
	jsonData := RoomData{
		RoomName: roomName,
		Width:    width,
		Height:   height,
	}

	// generate noise map for town elevation
	elevNoiseMap := proc_gen.GenerateTownElevation(width, height)
	jsonData.ElevationMap = elevNoiseMap

	// place cliffs based on elevation
	jsonData.generateCliffs()

	// setup a town center
	success := jsonData.SetTownCenter()
	if !success {
		fmt.Println("failed to find valid town center")
		return
	}
	fmt.Printf("town center: %v, %v (size=%v)\n", jsonData.TownCenter.X, jsonData.TownCenter.Y, jsonData.TownCenter.Size)

	// generate a major road through the map. may add slopes to cliffs if needed
	jsonData.generateMajorRoad(m.Coords{X: 0, Y: height / 2}, m.Coords{X: width - 1, Y: height / 2})
	fmt.Println("Road:", jsonData.Roads[0].Path)
	for _, slope := range jsonData.SlopeMap {
		fmt.Println("Slopes at:", slope)
	}

	// build the basic barrier layout based on cliffs and slopes
	jsonData.buildBarrierLayout()

	// set up the basic tile layout
	err := jsonData.GenerateRandomTileLayout(tileset.Tx_Grass_01, width, height)
	if err != nil {
		fmt.Println(err)
		return
	}
	// paint roads
	err = jsonData.setRoadLayout(tileset.Tx_Grass_01_road)
	if err != nil {
		fmt.Println(err)
		return
	}

	// save json file
	filename := fmt.Sprintf("room/room_layouts/%s.json", roomName)
	if err := writeToJson(filename, jsonData); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Generated room file at %s\n", filename)
}

// generates a random tile layout with tiles from the given tileset key
func (jsonData *RoomData) GenerateRandomTileLayout(tilesetKey string, width, height int) error {
	tileNames, err := tileset.GetTilesetNames(tilesetKey)
	if err != nil {
		return err
	}
	// group name will be the capitalized first letter
	groupName := strings.ToUpper(string(tileNames[0][0]))

	// put this group name into every cell
	tileLayout := make([][]string, height)
	for i := range tileLayout {
		tileLayout[i] = make([]string, width)
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			tileLayout[y][x] = groupName
		}
	}

	fmt.Printf("adding tileset %s (group %s)\n", tilesetKey, groupName)

	jsonData.Tilesets = append(jsonData.Tilesets, tilesetKey)
	if jsonData.TileGroups == nil {
		jsonData.TileGroups = make(map[string][]string)
	}
	jsonData.TileGroups[groupName] = tileNames
	jsonData.TileGroupKeys = append(jsonData.TileGroupKeys, groupName)
	jsonData.TileLayout = tileLayout
	return nil
}

func (jsonData *RoomData) setRoadLayout(tilesetKey string) error {
	// get the tileset for the road tiles
	tileNames, err := tileset.GetTilesetNames(tilesetKey)
	if err != nil {
		return err
	}
	// group name will be the capitalized first letter
	groupName := strings.ToUpper(string(tileNames[0][0]))

	for _, road := range jsonData.Roads {
		fullPath := []m.Coords{}
		from, to := m.Coords{}, m.Coords{}
		for _, coord := range road.Path {
			if from == (m.Coords{}) {
				from = m.Coords{X: coord.X, Y: coord.Y}
				continue
			}
			if to == (m.Coords{}) {
				to = m.Coords{X: coord.X, Y: coord.Y}
			}
			fmt.Printf("getting path from %s to %s\n", from, to)
			curPath := path_finding.FindPath(from, to, jsonData.BarrierLayout)
			if len(curPath) == 0 {
				fmt.Println("path is empty!")
			}
			fullPath = append(fullPath, curPath...)
			from = m.Coords{X: to.X, Y: to.Y}
			to = m.Coords{}
		}
		// paint all the tiles in this path
		for _, coords := range fullPath {
			jsonData.paintTiles(groupName, 0, coords)
		}
	}
	fmt.Printf("adding tileset %s (group %s)\n", tilesetKey, groupName)

	jsonData.Tilesets = append(jsonData.Tilesets, tilesetKey)
	if jsonData.TileGroups == nil {
		jsonData.TileGroups = make(map[string][]string)
	}
	jsonData.TileGroups[groupName] = tileNames
	jsonData.TileGroupKeys = append(jsonData.TileGroupKeys, groupName)
	return nil
}

// paints the given tile at a position with the given brush size
//
// tileName can either be the name of a tile, or the group name that a tile belongs to.
//
// brush size is an extra radius of tiles outside the center (coords) to paint too.
// 0 means the given coords alone are painted (1x1), 1 means a 3x3 area is painted, etc.
func (jsonData *RoomData) paintTiles(tileName string, brushSize int, coords m.Coords) {
	if brushSize == 0 {
		if posInRoomBounds(coords, jsonData.Width, jsonData.Height) {
			jsonData.TileLayout[coords.Y][coords.X] = tileName
		}
		return
	}

	// fill out all the tiles within the brushSize radius
	for y := 0 - brushSize; y <= brushSize; y++ {
		for x := 0 - brushSize; x <= brushSize; x++ {
			curPos := m.Coords{X: coords.X + x, Y: coords.Y + y}
			if posInRoomBounds(curPos, jsonData.Width, jsonData.Height) {
				jsonData.TileLayout[curPos.Y][curPos.X] = tileName
			}
		}
	}
}

func (jsonData *RoomData) buildBarrierLayout() {
	// make barriers for all the cliffs, minus any slopes
	// remove cliff tiles if they are supposed to be slopes
	filteredCliffMap := make(map[string][]m.Coords)
	for dir, coords := range jsonData.CliffMap {
		filteredCliffMap[dir] = []m.Coords{}
		// if there are slopes for this direction, replace regular cliffs with any corresponding slopes
		if slopeCoords, ok := jsonData.SlopeMap[dir]; ok {
			for _, coord := range coords {
				skip := false
				for _, slopeCoord := range slopeCoords {
					if coord.Equals(slopeCoord) {
						skip = true
						break
					}
				}
				if !skip {
					filteredCliffMap[dir] = append(filteredCliffMap[dir], coord)
				}
			}
		} else {
			filteredCliffMap[dir] = coords
		}
	}
	jsonData.CliffMap = filteredCliffMap

	// add barriers where there are cliffs
	barrierLayout := make([][]bool, jsonData.Height)
	for i := range barrierLayout {
		barrierLayout[i] = make([]bool, jsonData.Width)
	}
	barriers_L := []m.Coords{{X: 1, Y: 0}, {X: 1, Y: 1}}
	barriers_R := []m.Coords{{X: 0, Y: 0}, {X: 0, Y: 1}}

	for key, coordsList := range jsonData.CliffMap {
		if len(coordsList) == 0 {
			continue
		}
		for _, coords := range coordsList {
			// with coords being the top left tile, mark the 2x2 tiles as barriers
			x, y := coords.X, coords.Y

			// handle L and R specially
			if key == "L" || key == "UL" {
				for _, mods := range barriers_L {
					barrierLayout[y+mods.Y][x+mods.X] = true
				}
				continue
			} else if key == "R" || key == "UR" {
				for _, mods := range barriers_R {
					barrierLayout[y+mods.Y][x+mods.X] = true
				}
				continue
			} else if key == "DL" {
				barrierLayout[y][x+1] = true
				continue
			} else if key == "DR" {
				barrierLayout[y][x] = true
				continue
			} else if key == "D" {
				barrierLayout[y][x] = true
				barrierLayout[y][x+1] = true
				continue
			} else if key == "DRC" {
				barrierLayout[y+1][x] = true
				barrierLayout[y][x] = true
				barrierLayout[y][x+1] = true
				continue
			}

			barrierLayout[y][x] = true
			if y+1 < jsonData.Height {
				barrierLayout[y+1][x] = true
				if x+1 < jsonData.Width {
					barrierLayout[y+1][x+1] = true
				}
			}
			if x+1 < jsonData.Width {
				barrierLayout[y][x+1] = true
			}
		}
	}
	jsonData.BarrierLayout = barrierLayout
}

// The 'town center' is basically a central location in the map.
// It's meant to be a cross roads for paths, and if this is a town with population, it will have a higher concentration of buildings.
//
// It will be a square area of flat elevation, that aims to be as close to the center of the map as possible
func (jsonData *RoomData) SetTownCenter() bool {
	mapHeight := jsonData.Height
	mapWidth := jsonData.Width

	// returns true if the given position (top left of a rectangle of given size) is suitable for a town center, or false
	// and the coordinates at which there was a conflict
	isAreaSuitable := func(x, y int, size int) bool {
		expectedElevation := jsonData.ElevationMap[y][x]

		for y1 := 0; y1 < size; y1++ {
			if y+y1 >= mapHeight {
				return false
			}
			for x1 := 0; x1 < size; x1++ {
				if x+x1 >= mapWidth {
					return false
				}
				if jsonData.ElevationMap[y+y1][x+x1] != expectedElevation {
					return false
				}
			}
		}
		return true
	}

	getTownCenter := func(centerSize int) (bool, TownCenter) {
		// try the exact center first
		idealX := mapWidth/2 - (centerSize / 2)
		idealY := mapHeight/2 - (centerSize / 2)
		if isAreaSuitable(idealX, idealY, centerSize) {
			return true, TownCenter{Size: centerSize, X: idealX, Y: idealY}
		}
		// if there are hills in the way, then try to find a nearby suitable location
		// keep track of the best location (closest to the ideal position)
		idealSpot := m.Coords{X: idealX, Y: idealY}
		bestSpot := m.Coords{}
		bestDist := mapWidth * mapHeight // just an arbitrary value guaranteed to be replaced
		for y := mapHeight/2 - centerSize; y < mapHeight/2+centerSize; y++ {
			for x := mapWidth/2 - centerSize; x < mapWidth/2+centerSize; x++ {
				// if it's a suitable area, consider if it's better than the current winner
				if isAreaSuitable(x, y, centerSize) {
					newSpot := m.Coords{X: x, Y: y}
					newDist := int(euclideanDist(newSpot, idealSpot))
					if newDist < bestDist {
						bestSpot = newSpot.Copy()
						bestDist = newDist
					}
				}
			}
		}
		if bestSpot != (m.Coords{}) {
			return true, TownCenter{Size: centerSize, X: bestSpot.X, Y: bestSpot.Y}
		}
		return false, TownCenter{}
	}

	centerSize := int(math.Max(math.Min(float64(jsonData.Width)/4, 50), 20)) // 20 <= x <= 50
	// keep trying to find a suitable center, and decrease the center size each time if we fail to find it
	// if the center size gets too small, then give up - we probably need to regenerate a better map if so.
	attempts := 0
	for centerSize > 10 {
		found, townCenter := getTownCenter(centerSize)
		if found {
			fmt.Printf("town center setup took %v attempts\n", attempts)
			jsonData.TownCenter = townCenter
			return true
		}
		centerSize--
		attempts++
	}

	return false
}

// generates a main thoroughfare from the start position to the end position
//
// start and end are expected to be opposite sides of the map, and this algorithm will try to go to these points via the "town center" point
//
// Overall idea of this algorithm:
//
//  1. from start, begin tracing directly towards the town center
//
//  2. if obstructions are hit, start searching for a way around them
//
//     - if there's no way around the obstructions (e.g. the search is unable to go further right than a certain point) then we start looking for cliffs to turn into stairs
//
//     - once a point is found that can be turned into stairs, this point is marked as part of the road's path
//
//     - repeat this obstruction/path finding process as needed
//
//  3. once we reach the town center, then we aim to go towards end
//
//  4. repeat the same strategy from 2 to get from the town center to the end point
func (jsonData *RoomData) generateMajorRoad(start, end m.Coords) {
	// points along the path from start and end to draw the road via
	var waypoints Road

	// checks if there is any obstruction in the path going directly from start to end. also returns coordinates of the obstruction
	isPathBlocked := func(start, end m.Coords) (bool, m.Coords) {
		curPos := m.Coords{X: start.X, Y: start.Y}
		startElev := jsonData.ElevationMap[start.Y][start.X]
		dirs := []m.Coords{
			{X: -1, Y: 0}, // Left
			{X: 1, Y: 0},  // Right
			{X: 0, Y: -1}, // Up
			{X: 0, Y: 1},  // Down
		}
		// continuously move to the adjacent tile that is the closest to the goal, until an obstacle (i.e. cliff) is encountered
		for euclideanDist(curPos, end) > 1 {
			// check if the tile we are on is blocking the path (i.e. changed elevation)
			curElev := jsonData.ElevationMap[curPos.Y][curPos.X]
			if curElev != startElev {
				fmt.Printf("elevation changed from %v to %v at %s. need to find way around\n", startElev, curElev, curPos)
				return true, curPos
			}
			// find the best tile to move to
			minDist := float64(jsonData.Width + jsonData.Height)
			nextPos := m.Coords{}
			for _, dir := range dirs {
				movePos := m.Coords{X: curPos.X + dir.X, Y: curPos.Y + dir.Y}
				if !posInRoomBounds(movePos, jsonData.Width, jsonData.Height) {
					continue
				}
				dist := euclideanDist(movePos, end)
				if dist < minDist {
					minDist = dist
					nextPos = movePos
				}
			}
			curPos = m.Coords{X: nextPos.X, Y: nextPos.Y}
		}

		return false, m.Coords{}
	}

	findNearestSlopeOption := func(pos m.Coords) (m.Coords, string) {
		minDist := float64(99999)
		closestOption := m.Coords{}
		direction := ""

		for cliffDir, cliffCoords := range jsonData.CliffMap {
			// only search cliffs of the valid type
			if !(cliffDir == "U" || cliffDir == "D") {
				continue
			}
			for _, cliffPos := range cliffCoords {
				dist := euclideanDist(pos, cliffPos)
				if dist < minDist {
					closestOption = m.Coords{X: cliffPos.X, Y: cliffPos.Y}
					direction = cliffDir
					minDist = dist
				}
			}
		}
		return closestOption, direction
	}

	// first, aim to get to the town center
	waypoints.Path = append(waypoints.Path, start)
	curPos := start
	goal := m.Coords{X: jsonData.TownCenter.X + jsonData.TownCenter.Size/2, Y: jsonData.TownCenter.Y + jsonData.TownCenter.Size/2}
	dirs := []string{"U", "D"}
	slopeMap := make(map[string][]m.Coords)
	for _, dir := range dirs {
		slopeMap[dir] = []m.Coords{}
	}

	fmt.Println("starting from:", start)
	fmt.Println("finding path to town center:", goal)
	for general_util.EuclideanDist(float64(curPos.X), float64(curPos.Y), float64(goal.X), float64(goal.Y)) > 2 {
		// check if there is an obstruction to the goal
		blocked, blockedPos := isPathBlocked(curPos, goal)
		if blocked {
			// find a way around the obstacle - in this case, we need to make a slope down/up a cliff
			newSlopePos, newSlopeDir := findNearestSlopeOption(blockedPos)
			if newSlopePos == (m.Coords{}) {
				fmt.Println("failed to find a slope option!")
				return
			}
			fmt.Println("New slope at:", newSlopePos, " - dir:", newSlopeDir)
			slopeMap[newSlopeDir] = append(slopeMap[newSlopeDir], newSlopePos)
			waypoints.Path = append(waypoints.Path, newSlopePos)
			curPos = m.Coords{X: newSlopePos.X, Y: newSlopePos.Y}
			if newSlopeDir == "U" {
				curPos.Y -= 2
			} else if newSlopeDir == "D" {
				curPos.Y += 2
			}
			fmt.Println("starting from pos:", curPos)
		} else {
			curPos = m.Coords{X: goal.X, Y: goal.Y}
		}
	}
	waypoints.Path = append(waypoints.Path, goal)

	// next, aim to get to the end
	fmt.Println("finding path to end:", end)
	goal = end
	for general_util.EuclideanDist(float64(curPos.X), float64(curPos.Y), float64(goal.X), float64(goal.Y)) > 2 {
		// check if there is an obstruction to the goal
		blocked, blockedPos := isPathBlocked(curPos, goal)
		if blocked {
			// find a way around the obstacle - in this case, we need to make a slope down/up a cliff
			newSlopePos, newSlopeDir := findNearestSlopeOption(blockedPos)
			if newSlopePos == (m.Coords{}) {
				fmt.Println("failed to find a slope option!")
				return
			}
			fmt.Println("New slope at:", newSlopePos, " - dir:", newSlopeDir)
			slopeMap[newSlopeDir] = append(slopeMap[newSlopeDir], newSlopePos)
			waypoints.Path = append(waypoints.Path, newSlopePos)
			curPos = m.Coords{X: newSlopePos.X, Y: newSlopePos.Y}
		} else {
			curPos = m.Coords{X: goal.X, Y: goal.Y}
		}

	}
	waypoints.Path = append(waypoints.Path, end)
	jsonData.Roads = append(jsonData.Roads, waypoints)
	jsonData.SlopeMap = slopeMap
}

// searches the elevation map and builds cliffs on the map according to elevation changes
func (jsonData *RoomData) generateCliffs() {
	elevationMap := jsonData.ElevationMap
	dirs := []string{"U", "D", "L", "R", "UL", "UR", "DL", "DR", "ULC", "URC", "DLC", "DRC"}
	cliffMap := make(map[string][]m.Coords)
	for _, dir := range dirs {
		cliffMap[dir] = []m.Coords{}
	}

	// for every 2x2 square of tiles, detect if there are is raised elevation anywhere surrounding it
	// elevation is always mapped in blocks of 2x2
	width := jsonData.Width
	height := jsonData.Height

	for y := 0; y < height; y += 2 {
		if y >= height {
			break
		}
		for x := 0; x < width; x += 2 {
			if x >= width {
				break
			}
			coords := m.Coords{X: x, Y: y}
			// check sides first
			up := isCliff(x, y, "U", elevationMap, width, height)
			down := isCliff(x, y, "D", elevationMap, width, height)
			left := isCliff(x, y, "L", elevationMap, width, height)
			right := isCliff(x, y, "R", elevationMap, width, height)
			// ignore cliffs if they are too smashed together
			if up && down {
				continue
			}
			if left && right {
				continue
			}
			// detect if there are inner corners
			if up && left {
				cliffMap["UL"] = append(cliffMap["UL"], coords)
				continue
			}
			if up && right {
				cliffMap["UR"] = append(cliffMap["UR"], coords)
				continue
			}
			if down && left {
				cliffMap["DL"] = append(cliffMap["DL"], coords)
				continue
			}
			if down && right {
				cliffMap["DR"] = append(cliffMap["DR"], coords)
				continue
			}
			// otherwise mark it as a flat side
			if up {
				cliffMap["U"] = append(cliffMap["U"], coords)
				continue
			}
			if down {
				cliffMap["D"] = append(cliffMap["D"], coords)
				continue
			}
			if left {
				cliffMap["L"] = append(cliffMap["L"], coords)
				continue
			}
			if right {
				cliffMap["R"] = append(cliffMap["R"], coords)
				continue
			}
			// if no cliffs have been found, check for outer corners
			if isCliff(x, y, "UL", elevationMap, width, height) {
				cliffMap["ULC"] = append(cliffMap["ULC"], coords)
			}
			if isCliff(x, y, "UR", elevationMap, width, height) {
				cliffMap["URC"] = append(cliffMap["URC"], coords)
			}
			if isCliff(x, y, "DL", elevationMap, width, height) {
				cliffMap["DLC"] = append(cliffMap["DLC"], coords)
			}
			if isCliff(x, y, "DR", elevationMap, width, height) {
				cliffMap["DRC"] = append(cliffMap["DRC"], coords)
			}
		}
	}
	jsonData.CliffMap = cliffMap
}

func isCliff(x, y int, direction string, elevationMap [][]int, width, height int) bool {
	curElevation := elevationMap[y][x]
	if strings.Contains(direction, "U") {
		if y-1 < 0 {
			return false
		}
	}
	if strings.Contains(direction, "D") {
		if y+2 >= height {
			return false
		}
	}
	if strings.Contains(direction, "L") {
		if x-1 < 0 {
			return false
		}
	}
	if strings.Contains(direction, "R") {
		if x+2 >= width {
			return false
		}
	}
	switch direction {
	case "U":
		return elevationMap[y-1][x] > curElevation
	case "D":
		return elevationMap[y+2][x] > curElevation
	case "L":
		return elevationMap[y][x-1] > curElevation
	case "R":
		return elevationMap[y][x+2] > curElevation
	case "UL":
		return elevationMap[y-1][x-1] > curElevation
	case "UR":
		return elevationMap[y-1][x+2] > curElevation
	case "DL":
		return elevationMap[y+2][x-1] > curElevation
	case "DR":
		return elevationMap[y+2][x+2] > curElevation
	}
	return false
}
