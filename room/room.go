// all things pertaining to the room the player is in
package room

import (
	"ancient-rome/config"
	"ancient-rome/rendering"
	"ancient-rome/tileset"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type Coords struct {
	X int
	Y int
}

type Room struct {
	MasterTileset   map[string]*ebiten.Image
	MasterObjectSet map[string]*ebiten.Image
	TileLayout      [][]*ebiten.Image
	BarrierLayout   [][]bool
	ObjectLayout    [][]*ebiten.Image
	ObjectCoords    []Coords
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
}

func CreateRoom(roomID string) Room {
	fmt.Println("Creating room ", roomID)
	room := Room{}
	fmt.Println("loading json...")
	jsonData, err := loadRoomDataJson(roomID)
	if err != nil {
		panic(err)
	}
	fmt.Println("building the tile layout...")
	room.buildTileLayout(*jsonData)
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

func (r *Room) buildObjectLayout(roomData RoomData) {
	objectGroups := roomData.ObjectGroups
	objectLayout := roomData.ObjectLayout
	objectSets := roomData.ObjectSets

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
	// convert to bools since it's less memory than ints
	// TODO just use bools in the first place?
	var barrierLayout [][]bool

	for _, row := range rawBarrierLayout {
		var rowVals []bool
		for _, val := range row {
			if val == 1 {
				rowVals = append(rowVals, true)
			} else {
				rowVals = append(rowVals, false)
			}
		}
		barrierLayout = append(barrierLayout, rowVals)
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
			if tileImg == nil {
				fmt.Println("tileimg is nil!!")
			}
			drawX := float64(x*config.TileSize) - offsetX
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(drawX, drawY)
			op.GeoM.Scale(config.GameScale, config.GameScale)
			screen.DrawImage(tileImg, op)
		}
	}
}

func (r *Room) DrawObjects(screen *ebiten.Image, offsetX float64, offsetY float64) {
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
