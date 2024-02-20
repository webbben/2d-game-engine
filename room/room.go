// all things pertaining to the room the player is in
package room

import (
	"ancient-rome/config"
	"ancient-rome/proc_gen"
	"ancient-rome/rendering"
	"ancient-rome/tileset"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

type Coords struct {
	X int
	Y int
}

type Room struct {
	Width           int
	Height          int
	MasterTileset   map[string]*ebiten.Image
	MasterObjectSet map[string]*ebiten.Image
	CliffTileset    map[string]*ebiten.Image
	TileLayout      [][]*ebiten.Image
	BarrierLayout   [][]bool
	ObjectLayout    [][]*ebiten.Image
	ObjectCoords    []Coords
	CliffMap        map[string][]Coords
}

// data as loaded from the room_layout json
type RoomData struct {
	RoomName        string              `json:"room_name"`       // Display name of the room
	Width           int                 `json:"width"`           // width of the room
	Height          int                 `json:"height"`          // height of the room
	Tilesets        []string            `json:"tilesets"`        // names of the tilesets used in this room
	TileGroupKeys   []string            `json:"tileGroupKeys"`   // keys of the custom groups of tiles for this room
	TileGroups      map[string][]string `json:"tileGroups"`      // custom groups of tiles that can be used interchangeably, to "randomize" the layout a bit
	TileLayout      [][]string          `json:"tileLayout"`      // the keys for tiles at each position of the room. a key may point to a tile group, or be the key of an individual tile
	BarrierLayout   [][]int             `json:"barrierLayout"`   // a map of which tiles are accessible (0) and which are barriers/inaccessible (1)
	ObjectSets      []string            `json:"objectSets"`      // names of the tilesets used for objects in this room
	ObjectGroupKeys []string            `json:"objectGroupKeys"` // keys of the object groups for this room
	ObjectGroups    map[string][]string `json:"objectGroups"`    // custom groups of objects that can be used interchangeably, to "randomize" the objects a bit
	ObjectLayout    [][]string          `json:"objectLayout"`    // the keys for objects at each position of the room. a key may point to a object group, or be the key of an individual object. '-' indicates no object in the given position.
	ElevationMap    [][]int             `json:"elevationMap"`
}

func CreateRoom(roomID string) Room {
	fmt.Println("Creating room ", roomID)
	room := Room{}
	fmt.Println("loading json...")
	jsonData, err := loadRoomDataJson(roomID)
	if err != nil {
		panic(err)
	}
	room.Width = jsonData.Width
	room.Height = jsonData.Height
	fmt.Println("building the tile layout...")
	room.buildTileLayout(*jsonData)
	fmt.Println("building cliff layout...")
	room.buildElevation(*jsonData)
	fmt.Println("building the object layout...")
	room.buildObjectLayout(*jsonData)
	fmt.Println("building the barrier layout...")
	room.buildBarrierLayout(jsonData.BarrierLayout)
	fmt.Println("room creation complete!")
	return room
}

// build the layout of tile images for this room
func (r *Room) buildTileLayout(roomData RoomData) {
	tileGroups := roomData.TileGroups
	tileLayout := roomData.TileLayout
	tilesets := roomData.Tilesets

	// first, get the map of tile images from all tilesets used
	tilesetMaster := make(map[string]*ebiten.Image)

	fmt.Println("loading tilesets and their images")
	for _, setKey := range tilesets {
		tileMap, err := tileset.LoadTileset(setKey)
		if err != nil {
			panic(err)
		}
		// if there are duplicate keys, the previous image at that key will be overwritten
		for key, image := range tileMap {
			tilesetMaster[key] = image
		}
	}

	var layout [][]*ebiten.Image

	fmt.Println("building the layout")
	for _, row := range tileLayout {
		var imgRow []*ebiten.Image
		for _, tileKey := range row {
			key := tileKey
			// if this is part of a tileGroup, get a random tile key from it
			tileGroup, ok := tileGroups[tileKey]
			if ok {
				key = tileGroup[rand.Intn(len(tileGroup))]
			}
			img, ok := tilesetMaster[key]
			if !ok {
				panic("image file not found?")
			}
			imgRow = append(imgRow, img)
		}
		layout = append(layout, imgRow)
	}
	r.MasterTileset = tilesetMaster
	r.TileLayout = layout
}

func (r *Room) buildElevation(roomData RoomData) {
	elevationMap := roomData.ElevationMap
	dirs := []string{"U", "D", "L", "R", "UL", "UR", "DL", "DR", "ULC", "URC", "DLC", "DRC"}
	cliffMap := make(map[string][]Coords)
	for _, dir := range dirs {
		cliffMap[dir] = []Coords{}
	}

	// for every 2x2 square of tiles, detect if there are is raised elevation anywhere surrounding it
	// elevation is always mapped in blocks of 2x2
	width := roomData.Width
	height := roomData.Height

	for y := 0; y < height; y += 2 {
		if y >= height {
			break
		}
		for x := 0; x < width; x += 2 {
			if x >= width {
				break
			}
			coords := Coords{X: x, Y: y}
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

	// load cliff tileset
	cliffTiles, err := tileset.LoadTileset(tileset.Tx_Cliff_01)
	if err != nil {
		panic(err)
	}
	r.CliffTileset = cliffTiles
	r.CliffMap = cliffMap
}

func isCliff(x, y int, direction string, elevationMap [][]int, width, height int) bool {
	curElevation := elevationMap[y][x]
	switch direction {
	case "U":
		if y-1 < 0 {
			return false
		}
		return elevationMap[y-1][x] > curElevation
	case "D":
		if y+2 >= height {
			return false
		}
		return elevationMap[y+2][x] > curElevation
	case "L":
		if x-1 < 0 {
			return false
		}
		return elevationMap[y][x-1] > curElevation
	case "R":
		if x+2 >= width {
			return false
		}
		return elevationMap[y][x+2] > curElevation
	case "UL":
		if y-1 < 0 || x-1 < 0 {
			return false
		}
		return elevationMap[y-1][x-1] > curElevation
	case "UR":
		if y-1 < 0 || x+2 >= width {
			return false
		}
		return elevationMap[y-1][x+2] > curElevation
	case "DL":
		if y+2 >= height || x-1 < 0 {
			return false
		}
		return elevationMap[y+2][x-1] > curElevation
	case "DR":
		if y+2 >= height || x+2 >= width {
			return false
		}
		return elevationMap[y+2][x+2] > curElevation
	}
	return false
}

func (r *Room) buildObjectLayout(roomData RoomData) {
	objectGroups := roomData.ObjectGroups
	objectLayout := roomData.ObjectLayout
	objectSets := roomData.ObjectSets

	if objectLayout == nil || objectSets == nil {
		return
	}

	// first, get the map of tile images from all tilesets used
	objectSetMaster := make(map[string]*ebiten.Image)

	fmt.Println("loading object sets and their images")
	for _, setKey := range objectSets {
		imageMap, err := tileset.LoadTileset(setKey)
		if err != nil {
			panic(err)
		}
		// if there are duplicate keys, the previous image at that key will be overwritten
		for key, image := range imageMap {
			objectSetMaster[key] = image
		}
	}

	var layout [][]*ebiten.Image
	var objectCoords []Coords

	fmt.Println("building the object layout")
	for y, row := range objectLayout {
		var imgRow []*ebiten.Image
		for x, objectKey := range row {
			if objectKey == "-" {
				imgRow = append(imgRow, nil)
				continue
			}
			key := objectKey
			// if this is part of a tileGroup, get a random tile key from it
			objectGroup, ok := objectGroups[objectKey]
			if ok {
				key = objectGroup[rand.Intn(len(objectGroup))]
			}
			img, ok := objectSetMaster[key]
			if !ok {
				panic("image file not found?")
			}
			imgRow = append(imgRow, img)
			// store the coords so we can be a little quicker at finding objects
			objectCoords = append(objectCoords, Coords{X: x, Y: y})
		}
		layout = append(layout, imgRow)
	}
	r.MasterObjectSet = objectSetMaster
	r.ObjectLayout = layout
	r.ObjectCoords = objectCoords
}

func (r *Room) buildBarrierLayout(rawBarrierLayout [][]int) {
	// we start with a free movement room, and apply barriers as needed (for structures, objects, cliffs, etc)
	barrierLayout := make([][]bool, r.Height)
	for i := range barrierLayout {
		barrierLayout[i] = make([]bool, r.Width)
	}

	// add barriers where there are objects
	for _, coords := range r.ObjectCoords {
		barrierLayout[coords.Y][coords.X] = true
	}

	// add barriers where there are cliffs
	for _, coordsList := range r.CliffMap {
		if len(coordsList) == 0 {
			continue
		}
		for _, coords := range coordsList {
			// with coords being the top left tile, mark the 2x2 tiles as barriers
			x, y := coords.X, coords.Y
			barrierLayout[y][x] = true
			if y+1 < r.Height {
				barrierLayout[y+1][x] = true
				if x+1 < r.Width {
					barrierLayout[y+1][x+1] = true
				}
			}
			if x+1 < r.Width {
				barrierLayout[y][x+1] = true
			}
		}
	}

	r.BarrierLayout = barrierLayout
}

func (r *Room) DrawFloor(screen *ebiten.Image, offsetX float64, offsetY float64) {
	for y, row := range r.TileLayout {
		// skip this row if it's above the camera
		if rendering.RowAboveCameraView(float64(y), offsetY) {
			continue
		}
		// skip all remaining rows if it's below the camera
		if rendering.RowBelowCameraView(float64(y), offsetY) {
			break
		}
		drawY := float64(y*config.TileSize) - offsetY

		for x, tileImg := range row {
			if rendering.ColBeforeCameraView(float64(x), offsetX) {
				continue
			}
			// skip the rest of the columns if it's past the screen
			if rendering.ColAfterCameraView(float64(x), offsetX) {
				break
			}
			drawX := float64(x*config.TileSize) - offsetX
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(drawX, drawY)
			op.GeoM.Scale(config.GameScale, config.GameScale)
			screen.DrawImage(tileImg, op)
		}
	}
}

func (r *Room) DrawCliffs(screen *ebiten.Image, offsetX float64, offsetY float64) {
	for key, coordsList := range r.CliffMap {
		if len(coordsList) == 0 {
			continue
		}
		if key == "UL" || key == "UR" || key == "DL" || key == "DR" {
			key = "D"
		}
		imgKey := fmt.Sprintf("cliff_%s", key)
		img := r.CliffTileset[imgKey]
		for _, coords := range coordsList {
			if !rendering.ObjectInsideCameraView(float64(coords.X), float64(coords.Y), 32, 32, offsetX, offsetY) {
				continue
			}
			drawY := float64(coords.Y*config.TileSize) - offsetY
			drawX := float64(coords.X*config.TileSize) - offsetX
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(drawX, drawY)
			op.GeoM.Scale(config.GameScale, config.GameScale)
			screen.DrawImage(img, op)
		}
	}
}

func (r *Room) DrawObjects(screen *ebiten.Image, offsetX float64, offsetY float64) {
	if r.ObjectCoords == nil {
		return
	}

	for _, coords := range r.ObjectCoords {
		x := coords.X
		y := coords.Y
		if rendering.RowAboveCameraView(float64(y), offsetY) {
			continue
		}
		objectImg := r.ObjectLayout[y][x]
		drawX, drawY := rendering.GetImageDrawPos(objectImg, float64(x), float64(y), offsetX, offsetY)
		//drawX := float64(x*config.TileSize) - offsetX
		//drawY := float64(y*config.TileSize) - offsetY
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(drawX, drawY)
		op.GeoM.Scale(config.GameScale, config.GameScale)
		screen.DrawImage(objectImg, op)
	}
}

func loadRoomDataJson(roomKey string) (*RoomData, error) {
	path := fmt.Sprintf("room/room_layouts/%s.json", roomKey)

	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load room data at %s", path)
	}
	var roomData RoomData
	if err = json.Unmarshal(jsonData, &roomData); err != nil {
		return nil, errors.New("json unmarshalling failed")
	}
	return &roomData, nil
}

// creates a room with randomized grass terrain and some hills
func GenerateRandomTerrain(roomName string, width, height int) {

	jsonData := RoomData{}

	noiseMap := proc_gen.GenerateTownElevation(width, height)
	jsonData.RoomName = roomName
	jsonData.ElevationMap = noiseMap
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

func writeToJson(filepath string, jsonData RoomData) error {
	if !strings.HasSuffix(filepath, ".json") {
		return errors.New("failed to write to json: given filepath doesn't end in '.json'")
	}
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(jsonData)
	if err != nil {
		return err
	}
	return nil
}
