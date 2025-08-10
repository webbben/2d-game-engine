package object

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

const (
	Latin_house_1 = "latin_house_1"
)

type Building struct {
	HasDoor bool `json:"has_door"`
	DoorX   int  `json:"door_x"`
}

type Object struct {
	Name       string   `json:"name"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	IsBuilding bool     `json:"is_building"`
	ImgPath    []string `json:"img_path"`
	X, Y       int
	Building
}

func loadObjectJson(objectID string) (*Object, error) {
	path := fmt.Sprintf("object/object_defs/%s.json", objectID)
	jsonData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load object def at %s", path)
	}
	var object Object
	if err = json.Unmarshal(jsonData, &object); err != nil {
		return nil, errors.New("json unmarshalling failed for object def")
	}
	return &object, nil
}

// instantiates an entity of the given definition
func CreateObject(objKey string, x, y int) (*Object, *ebiten.Image) {
	object, err := loadObjectJson(objKey)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	filePath := filepath.Join(object.ImgPath...)
	img, err := image.LoadImage(filePath)
	if err != nil {
		panic(err)
	}
	object.X = x
	object.Y = y
	return object, img
}

func (obj *Object) Draw(screen *ebiten.Image, offsetX float64, offsetY float64, imageMap map[string]*ebiten.Image) {
	img, ok := imageMap[obj.Name]
	if !ok {
		panic("failed to draw object - no corresponding image found in image map")
	}
	drawX, drawY := rendering.GetImageDrawPos(img, float64(obj.X), float64(obj.Y), offsetX, offsetY)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(img, op)
}
