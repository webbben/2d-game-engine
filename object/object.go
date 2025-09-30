package object

import (
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

const (
	TYPE_DOOR        = "DOOR"
	TYPE_SPAWN_POINT = "SPAWN_POINT"
)

type Object struct {
	Name          string
	Type          string
	X, Y          float64 // logical position in the map
	DrawX, DrawY  float64 // the actual position on the screen where this was last drawn - for things like click detection
	Width, Height int
	Rect          model.Rect

	MouseBehavior mouse.MouseBehavior

	OnLeftClick  func()
	OnRightClick func()

	Door

	SpawnPoint

	WorldContext
}

type WorldContext interface {
	GetPlayerRect() model.Rect
}

type Door struct {
	targetMapID      string
	targetSpawnIndex int
	openSound        *audio.Sound
	activateType     string // "click", "step"
}

type SpawnPoint struct {
	SpawnIndex int
}

func LoadObject(obj tiled.Object) *Object {
	o := Object{
		Name:   obj.Name,
		X:      obj.X,
		Y:      obj.Y,
		Width:  int(obj.Width),
		Height: int(obj.Height),
		Rect: model.Rect{
			X: obj.X,
			Y: obj.Y,
			W: obj.Width,
			H: obj.Height,
		},
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
		case TYPE_SPAWN_POINT:
			o.loadSpawnProperty(prop)
		default:
			panic("object type invalid")
		}
	}

	if o.Type == TYPE_DOOR {
		o.verifyDoorObject()
	}

	return &o
}

func (obj Object) verifyDoorObject() {
	// make sure required properties are defined
	if obj.Door.targetMapID == "" {
		panic("door: no target map ID set. check Tiled object definition.")
	}
	if obj.Door.activateType == "" {
		panic("door: no activate type set. check Tiled object definition.")
	}
	switch obj.Door.activateType {
	case "click":
	case "step":
	default:
		logz.Panicf("door: invalid activation type set: %s. check Tiled object definition.", obj.Door.activateType)
	}

	if obj.openSound == nil {
		panic("door: no openSound defined. check Tiled object definition.")
	}
}

func (obj *Object) loadDoorProperty(prop tiled.Property) {
	switch prop.Name {
	case "door_to":
		obj.Door.targetMapID = prop.GetStringValue()
	case "door_spawn_index":
		obj.Door.targetSpawnIndex = prop.GetIntValue()
	case "door_activate":
		obj.Door.activateType = prop.GetStringValue()
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

func (obj *Object) loadSpawnProperty(prop tiled.Property) {
	switch prop.Name {
	case "spawn_index":
		obj.SpawnIndex = prop.GetIntValue()
	}
}

func resolveObjectType(objType string) string {
	switch objType {
	case TYPE_DOOR:
		return TYPE_DOOR
	case TYPE_SPAWN_POINT:
		return TYPE_SPAWN_POINT
	default:
		panic("object type doesn't exist!")
	}
}
