package room

import (
	"ancient-rome/proc_gen"
	"ancient-rome/tileset"
	"fmt"
	"strings"
)

// creates a room with randomized grass terrain and some hills
func GenerateRandomTerrain(roomName string, width, height int) {
	jsonData := RoomData{}

	// generate noise map for town elevation
	elevNoiseMap := proc_gen.GenerateTownElevation(width, height)
	jsonData.RoomName = roomName
	jsonData.ElevationMap = elevNoiseMap
	jsonData.Width = width
	jsonData.Height = height
	err := jsonData.GenerateRandomTileLayout(tileset.Tx_Grass_01, width, height)
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
	var tileLayout [][]string
	row := make([]string, width)
	for i := 0; i < width; i++ {
		row[i] = groupName
	}
	for i := 0; i < height; i++ {
		tileLayout = append(tileLayout, row)
	}

	jsonData.Tilesets = append(jsonData.Tilesets, tilesetKey)
	if jsonData.TileGroups == nil {
		jsonData.TileGroups = make(map[string][]string)
	}
	jsonData.TileGroups[groupName] = tileNames
	jsonData.TileGroupKeys = append(jsonData.TileGroupKeys, groupName)
	jsonData.TileLayout = tileLayout
	return nil
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
func (jsonData *RoomData) GenerateMajorRoad(start, end Coords) {

}
