package object

import (
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

const (
	TYPE_DOOR = "DOOR"
)

type Object struct {
	Name          string
	Type          string
	X, Y          float64 // logical position in the map
	DrawX, DrawY  float64 // the actual position on the screen where this was last drawn - for things like click detection
	Width, Height int

	MouseBehavior mouse.MouseBehavior

	OnLeftClick  func()
	OnRightClick func()

	Door
}

type Door struct {
	TargetMapID string
	openSound   *audio.Sound
}

func LoadObject(obj tiled.Object) *Object {
	o := Object{
		Name:   obj.Name,
		X:      obj.X,
		Y:      obj.Y,
		Width:  int(obj.Width),
		Height: int(obj.Height),
	}

	// get the type first - so we know what values to parse out
	for _, prop := range obj.Properties {
		if prop.Name == "TYPE" {
			o.Type = resolveObjectType(prop.GetStringValue())
			break
		}
	}
	if o.Type == "" {
		// no type found... this must be a malformed tileset?
		panic("no object type property found")
	}

	// load data for specific object type
	for _, prop := range obj.Properties {
		switch o.Type {
		case TYPE_DOOR:
			o.loadDoorProperty(prop)
		default:
			panic("object type invalid")
		}
	}

	return &o
}

func (obj *Object) loadDoorProperty(prop tiled.Property) {
	switch prop.Name {
	case "door_to":
		obj.Door.TargetMapID = prop.GetStringValue()
	case "door_sound":
		doorSound := prop.GetStringValue()
		switch doorSound {
		case "wood":
			// TODO work out a system for defining these default door sounds
			sound, err := audio.LoadSound("/Users/benwebb/dev/personal/ancient-rome/assets/audio/sfx/door/open_door_01.mp3", 0.5)
			if err != nil {
				panic("failed to load door sound:" + err.Error())
			}
			obj.Door.openSound = &sound
		default:
			panic("door_sound value not found:" + doorSound)
		}
	}
}

func resolveObjectType(objType string) string {
	switch objType {
	case TYPE_DOOR:
		return TYPE_DOOR
	default:
		panic("object type doesn't exist!")
	}
}
