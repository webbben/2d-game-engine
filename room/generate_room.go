package room

import (
	"fmt"
	"math"
	"strings"

	m "github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/path_finding"
	"github.com/webbben/2d-game-engine/proc_gen"
	"github.com/webbben/2d-game-engine/tileset"
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

	// add slopes to cliffs
	jsonData.generateSlopes()

	// build the basic barrier layout based on cliffs and slopes
	jsonData.buildBarrierLayout()

	// setup a town center
	success := jsonData.SetTownCenter()
	if !success {
		fmt.Println("failed to find valid town center")
		return
	}
	fmt.Printf("town center: %v, %v (size=%v)\n", jsonData.TownCenter.X, jsonData.TownCenter.Y, jsonData.TownCenter.Size)

	// generate major roads through the map
	jsonData.generateMajorRoad(m.Coords{X: 0, Y: height / 2}, m.Coords{X: width - 1, Y: height / 2})
	jsonData.generateMajorRoad(m.Coords{X: width / 2, Y: 0}, m.Coords{X: width / 2, Y: height - 1})

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
		// paint all the tiles in this path
		for _, coords := range road.Path {
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
// it is expected that cliffs and barriers have already been mapped out beforehand, so the road can navigate around these things
func (jsonData *RoomData) generateMajorRoad(start, end m.Coords) {
	var road Road

	// find path from start to town center
	tc := jsonData.TownCenter
	goal := m.Coords{X: tc.X + (tc.Size / 2), Y: tc.Y + (tc.Size / 2)}
	path := path_finding.FindPath(start, goal, jsonData.BarrierLayout, nil)
	if len(path) == 0 {
		fmt.Printf("generateMajorRoad: failed to find path from %s to %s\n", start, goal)
		fmt.Println("aborting road generation")
		return
	}
	road.Path = append(road.Path, path...)

	// find path from town center to end
	path = path_finding.FindPath(goal, end, jsonData.BarrierLayout, nil)
	if len(path) == 0 {
		fmt.Printf("generateMajorRoad: failed to find path from %s to %s\n", start, goal)
		fmt.Println("aborting road generation")
		return
	}
	road.Path = append(road.Path, path...)

	jsonData.Roads = append(jsonData.Roads, road)
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

func (jsonData *RoomData) generateSlopes() {
	numSlopes := 4 // number of slopes to try to add to each cliff line
	jsonData.SlopeMap = make(map[string][]m.Coords)
	// find how many cliff lines there are
	cliffLines := findAllCliffLines(jsonData.CliffMap)
	if len(cliffLines) == 0 {
		fmt.Println("generateSlopes: no cliff lines found")
		return
	}

	// try to place some slopes on each cliff line
	// number of slopes should be based on length of the cliff
	threshold := jsonData.Width / 8
	for _, cliffLine := range cliffLines {
		// check horizontal range and add slopes if possible
		if cliffLine.xMax-cliffLine.xMin > threshold {
			for i := 0; i < numSlopes; i++ {
				success := jsonData.placeSlopeOnCliffLine(cliffLine, float64(threshold), true)
				if !success {
					break
				}
			}
		}
		// check vertical range and add slopes if possible
		if cliffLine.yMax-cliffLine.yMin > threshold {
			for i := 0; i < numSlopes; i++ {
				success := jsonData.placeSlopeOnCliffLine(cliffLine, float64(threshold), false)
				if !success {
					break
				}
			}
		}
	}
}

// attempts to place a slope along a cliff line. returns true if succeeded, or false if no viable slope position was found.
//
// minDistThresh: minimum distance new slope must be from old slopes - meant to help space them out a bit
//
// horizAxis: whether we are basing the slope placement on the horizontal axis or the vertical axis
func (jsonData *RoomData) placeSlopeOnCliffLine(cliffLine CliffLine, minDistThresh float64, horizAxis bool) bool {
	findNearestSlopeOption := func(coordVal int) (m.Coords, string) {
		minDist := float64(99999)   // minimum distance from the goal coords we've found so far
		closestOption := m.Coords{} // the current best option
		direction := ""             // direction of the current best

		for cliffDir, cliffCoords := range jsonData.CliffMap {
			// only search cliffs of the valid type
			if !(cliffDir == "U" || cliffDir == "D") {
				continue
			}
			for _, cliffPos := range cliffCoords {
				var compVal int
				if horizAxis {
					compVal = cliffPos.X
				} else {
					compVal = cliffPos.Y
				}
				axisDist := math.Abs(float64(coordVal) - float64(compVal))
				if axisDist < minDist {
					// make sure it's not too close to another slope
					tooClose := false
					for _, slopeList := range jsonData.SlopeMap {
						for _, slope := range slopeList {
							if euclideanDist(cliffPos, slope) < minDistThresh {
								tooClose = true
								break
							}
						}
						if tooClose {
							break
						}
					}
					if tooClose {
						continue
					}
					closestOption = m.Coords{X: cliffPos.X, Y: cliffPos.Y}
					direction = cliffDir
					minDist = axisDist
				}
			}
		}
		return closestOption, direction
	}

	var goalVal int
	// make the goal the center (average) of the range of x or y values in the cliff line
	if horizAxis {
		goalVal = (cliffLine.xMax + cliffLine.xMin) / 2
	} else {
		goalVal = (cliffLine.yMax + cliffLine.yMin) / 2
	}
	// find a slope candidate as close to centerX as possible, but also minDist from other slopes
	slopeOption, dir := findNearestSlopeOption(goalVal)
	if slopeOption == (m.Coords{}) {
		return false
	}
	if _, ok := jsonData.SlopeMap[dir]; !ok {
		jsonData.SlopeMap[dir] = []m.Coords{}
	}
	jsonData.SlopeMap[dir] = append(jsonData.SlopeMap[dir], slopeOption)
	return true
}

type CliffLine struct {
	coords                 []m.Coords
	xMin, xMax, yMin, yMax int
}

func (c *CliffLine) AddCliffCoords(coord m.Coords) {
	c.xMin = int(math.Min(float64(c.xMin), float64(coord.X)))
	c.xMax = int(math.Max(float64(c.xMax), float64(coord.X)))
	c.yMin = int(math.Min(float64(c.yMin), float64(coord.Y)))
	c.yMax = int(math.Max(float64(c.yMax), float64(coord.Y)))
	c.coords = append(c.coords, coord.Copy())
}

func findAllCliffLines(cliffMap map[string][]m.Coords) []CliffLine {
	cliffLines := []CliffLine{}
	// get a list of all the coords that have cliffs, regardless of direction
	cliffMasterList := make([]m.Coords, 0)
	for _, cliffCoords := range cliffMap {
		cliffMasterList = append(cliffMasterList, cliffCoords...)
	}

	removeElement := func(slice []m.Coords, index int) []m.Coords {
		if index < 0 || index >= len(slice) {
			return slice
		}
		return append(slice[:index], slice[index+1:]...)
	}
	findNearestCliff := func(cliffCoord m.Coords) (int, float64) {
		closestDist := 99999.0
		closestIndex := 0
		for index, compCliffCoord := range cliffMasterList {
			dist := euclideanDist(cliffCoord, compCliffCoord)
			if dist < float64(closestDist) {
				closestDist = dist
				closestIndex = index
			}
		}
		return closestIndex, closestDist
	}
	// now, group each cliff into a CliffLine
	for len(cliffMasterList) > 0 {
		cliffCoord := cliffMasterList[0]
		cliffMasterList = removeElement(cliffMasterList, 0)

		// a cliff line starting from an arbitrary point will have at most 2 directions.
		// start one direction, and when that direction ends, go back and try for another
		startPos := cliffCoord
		cliffLine := CliffLine{}
		cliffLine.AddCliffCoords(startPos)

		// search first direction
		curCoords := startPos.Copy()
		for {
			nearestIndex, nearestDist := findNearestCliff(curCoords)
			// if the nearest distance is too far, then we've exhausted this direction
			if nearestDist > 3 {
				break
			}
			curCoords = cliffMasterList[nearestIndex]
			cliffLine.AddCliffCoords(curCoords)
			cliffMasterList = removeElement(cliffMasterList, nearestIndex)
			if len(cliffMasterList) == 0 {
				break
			}
		}
		if len(cliffMasterList) == 0 {
			cliffLines = append(cliffLines, cliffLine)
			break
		}

		// search the other direction
		curCoords = startPos.Copy()
		for {
			nearestIndex, nearestDist := findNearestCliff(curCoords)
			// if the nearest distance is too far, then we've exhausted this direction
			if nearestDist > 3 {
				break
			}
			curCoords = cliffMasterList[nearestIndex]
			cliffLine.AddCliffCoords(curCoords)
			cliffMasterList = removeElement(cliffMasterList, nearestIndex)
			if len(cliffMasterList) == 0 {
				break
			}
		}
		cliffLines = append(cliffLines, cliffLine)
	}
	return cliffLines
}
