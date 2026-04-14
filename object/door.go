package object

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/tiled"
)

const (
	PropDoorTo         string = "door_to"
	PropDoorSpawnIndex string = "door_spawn_index"
	PropDoorActivate   string = "door_activate"
	PropDoorSFX        string = "SFX"
)

type Door struct {
	TargetMapID      defs.MapID
	TargetSpawnIndex int
	openSoundID      defs.SoundID
	activateType     string // "click", "step"
}

func (obj *Object) loadDoorObject(props []tiled.Property) {
	for _, prop := range props {
		switch prop.Name {
		case PropDoorTo:
			obj.Door.TargetMapID = defs.MapID(prop.GetStringValue())
		case PropDoorSpawnIndex:
			obj.Door.TargetSpawnIndex = prop.GetIntValue()
		case PropDoorActivate:
			obj.Door.activateType = prop.GetStringValue()
		case PropDoorSFX:
			doorSound := prop.GetStringValue()
			if doorSound == "" {
				logz.Panicln("Door", "no door sound found (prop SFX). object:", obj.Name, obj.ID)
			}
			obj.Door.openSoundID = defs.SoundID(doorSound)
		}
	}
}

func (obj Object) validateDoorObject() {
	// make sure required properties are defined
	if obj.Door.TargetMapID == "" {
		panic("door: no target map ID set. check Tiled object definition.")
	}
	if obj.Door.activateType == "" {
		panic("door: no activate type set. check Tiled object definition.")
	}
	switch obj.Door.activateType {
	case "click":
	case "step":
	default:
		logz.Panicf("door [%s]: invalid activation type set: %s. check Tiled object definition.", obj.Name, obj.Door.activateType)
	}

	if obj.Door.openSoundID == "" {
		logz.Panicf("door [%s]: no openSound defined. check Tiled object definition.", obj.Name)
	}
}

func (obj *Object) activateDoor(params ObjectActivationParams) ObjectUpdateResult {
	if obj.Type != TypeDoor {
		panic("tried to activate as door, but object is not a door")
	}
	if obj.Door.TargetMapID == "" {
		panic("activated door does not have a target map ID!")
	}
	if obj.Door.TargetSpawnIndex < 0 {
		panic("activated door's target spawn index is negative! it should be a positive integer, including 0.")
	}
	if obj.Door.openSoundID == "" {
		panic("activated door's open sound is empty!")
	}

	// TODO: should we define volume somewhere?
	obj.AudioMgr.PlaySFX(obj.Door.openSoundID, 0.5)

	// check if this is the player, or an NPC
	// It doesn't actually change the logic here, since the calling code will handle it, but good to know.
	if params.ActivatorID == id.CharacterStateID(defs.PlayerID) {
		// player has activated the door
		logz.Println("activateDoor", "Player going to map:", obj.Door.TargetMapID)
	} else {
		logz.Println("activateDoor", "NPC going to map:", obj.Door.TargetMapID)
	}

	return ObjectUpdateResult{
		UpdateOccurred:      true,
		ChangeMapID:         obj.Door.TargetMapID,
		ChangeMapSpawnIndex: obj.Door.TargetSpawnIndex,
	}
}

func (obj *Object) updateDoor() ObjectUpdateResult {
	switch obj.Door.activateType {
	case "click":
		// do nothing - object clicks are detected and handled within mapInfo handler function
	case "step":
		if obj.World.GetPlayerRect().Intersects(obj.rect) {
			return obj.activateDoor(ObjectActivationParams{ActivatorID: id.CharacterStateID(defs.PlayerID)})
		}
	default:
		panic("invalid activation type for door")
	}
	return ObjectUpdateResult{}
}
