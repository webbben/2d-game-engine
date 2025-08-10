// all things pertaining to the room the player is in
package room

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/tileset"

	"github.com/hajimehoshi/ebiten/v2"
)

type Room struct {
	RoomName        string
	Width           int
	Height          int
	MasterTileset   map[string]*ebiten.Image
	MasterObjectSet map[string]*ebiten.Image
	CliffTileset    map[string]*ebiten.Image
	TileLayout      [][]*ebiten.Image
	BarrierLayout   [][]bool
	ObjectLayout    [][]*ebiten.Image
	ObjectCoords    []model.Coords
	CliffMap        map[string][]model.Coords
	SlopeMap        map[string][]model.Coords
}

// creates a room object from the given room ID
func CreateRoom(roomID string) Room {
	fmt.Println("Creating room ", roomID)
	fmt.Println("loading json...")
	jsonData, err := loadRoomDataJson(roomID)
	if err != nil {
		panic(err)
	}
	room := Room{
		RoomName: jsonData.RoomName,
		Width:    jsonData.Width,
		Height:   jsonData.Height,
	}
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
	// load cliff tileset
	cliffTiles, err := tileset.LoadTileset(tileset.Tx_Cliff_01)
	if err != nil {
		panic(err)
	}
	r.CliffTileset = cliffTiles

	// remove cliff tiles if they are supposed to be slopes
	filteredCliffMap := make(map[string][]model.Coords)
	for dir, coords := range roomData.CliffMap {
		filteredCliffMap[dir] = []model.Coords{}
		// if there are slopes for this direction, replace regular cliffs with any corresponding slopes
		if slopeCoords, ok := roomData.SlopeMap[dir]; ok {
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
	r.CliffMap = filteredCliffMap
	r.SlopeMap = roomData.SlopeMap
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
	var objectCoords []model.Coords

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
			objectCoords = append(objectCoords, model.Coords{X: x, Y: y})
		}
		layout = append(layout, imgRow)
	}
	r.MasterObjectSet = objectSetMaster
	r.ObjectLayout = layout
	r.ObjectCoords = objectCoords
}

func (r *Room) buildBarrierLayout(rawBarrierLayout [][]bool) {
	barrierLayout := rawBarrierLayout

	// add barriers where there are objects
	for _, coords := range r.ObjectCoords {
		barrierLayout[coords.Y][coords.X] = true
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
	// draw regular cliffs
	for key, coordsList := range r.CliffMap {
		if len(coordsList) == 0 {
			continue
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
	// draw slopes
	for key, coordsList := range r.SlopeMap {
		if len(coordsList) == 0 {
			continue
		}
		imgKey := fmt.Sprintf("cliff_%s_slope", key)
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
