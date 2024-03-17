package object

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Building struct {
	HasDoor bool `json:"has_door"`
	DoorX   int  `json:"door_x"`
}

type Object struct {
	Name       string `json:"name"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	IsBuilding bool   `json:"is_building"`
	Building
}

func LoadObject(objectID string) (*Object, error) {
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
