// all things pertaining to the room the player is in
package room

import (
	"ancient-rome/config"
	"ancient-rome/tileset"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type Room struct {
	JsonData      RoomData
	MasterTileset map[string]*ebiten.Image
	TileLayout    [][]*ebiten.Image
}

// data as loaded from the room_layout json
type RoomData struct {
	RoomName      string              `json:"room_name"`     // Display name of the room
	Width         int                 `json:"width"`         // width of the room
	Height        int                 `json:"height"`        // height of the room
	Tilesets      []string            `json:"tilesets"`      // names of the tilesets used in this room
	TileGroupKeys []string            `json:"tileGroupKeys"` // keys of the custom groups of tiles for this room
	TileGroups    map[string][]string `json:"tileGroups"`    // custom groups of tiles that can be used interchangeably, to "randomize" the layout a bit
	TileLayout    [][]string          `json:"tileLayout"`    // the keys for tiles at each position of the room. a key may point to a tile group, or be the key of an individual tile
	BarrierLayout [][]int             `json:"barrierLayout"` // a map of which tiles are accessible (0) and which are barriers/inaccessible (1)
}

func CreateRoom(roomID string) Room {
	fmt.Println("Creating room ", roomID)
	room := Room{}
	fmt.Println("loading json...")
	jsonData, err := loadRoomDataJson(roomID)
	fmt.Println("loading json complete!")
	if err != nil {
		panic(err)
	}
	room.JsonData = *jsonData
	fmt.Println("building the layout...")
	room.buildTileLayout()
	fmt.Println("build complete!")
	return room
}

// build the layout of tile images for this room
func (r *Room) buildTileLayout() {
	roomData := r.JsonData
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

func (r *Room) Draw(screen *ebiten.Image, op *ebiten.DrawImageOptions, offsetX float64, offsetY float64) {
	tileSize := config.TileSize
	for y, row := range r.TileLayout {
		for x, tileImg := range row {
			if tileImg == nil {
				fmt.Println("tileimg is nil!!")
			}
			drawX := float64(x*tileSize) - offsetX
			drawY := float64(y*tileSize) - offsetY
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(drawX, drawY)
			op.GeoM.Scale(config.GameScale, config.GameScale)
			screen.DrawImage(tileImg, op)
		}
	}
}

func loadRoomDataJson(roomKey string) (*RoomData, error) {
	path := fmt.Sprintf("room/room_layouts/%s.json", roomKey)

	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to load room data at %s", path))
	}
	var roomData RoomData
	if err = json.Unmarshal(jsonData, &roomData); err != nil {
		return nil, errors.New("json unmarshalling failed")
	}
	return &roomData, nil
}
