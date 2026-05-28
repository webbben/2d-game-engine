package object

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/tiled"
)

const (
	PropDoorTo             string = "door_to"
	PropDoorSpawnIndex     string = "door_spawn_index"
	PropDoorActivate       string = "door_activate"
	PropDoorSFX            string = "SFX"
	PropDoorMapGeneratorID string = "map_generator_id"
)

type Door struct {
	TargetMapID      defs.MapID
	TargetSpawnIndex int
	openSoundID      defs.SoundID
	activateType     string // "click", "step"
}

func (obj *Object) loadDoorObject(props []tiled.Property) {
	doorTo := ""
	var toSpawn int

	for _, prop := range props {
		switch prop.Name {
		case PropDoorTo:
			doorTo = prop.GetStringValue()
		case PropDoorSpawnIndex:
			toSpawn = prop.GetIntValue()
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

	// if the door info is set in the props, then set it here. otherwise, it must be a generated map target,
	// in which case this data is added after LoadObject is done.
	if doorTo != "" {
		obj.SetDoorTarget(defs.MapID(doorTo), &toSpawn)
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

// SetDoorTarget handles setting the display name of the door too; also used for overriding door target after loading.
func (obj *Object) SetDoorTarget(targetMapID defs.MapID, targetSpawn *int) {
	if targetMapID == "" {
		panic("targetMapId was empty")
	}
	obj.Door.TargetMapID = targetMapID

	// Sometimes we only want to set the target Map ID, so making this nullable
	if targetSpawn != nil {
		obj.Door.TargetSpawnIndex = *targetSpawn
	}

	if obj.DisplayName == "" {
		mapInfo, _, _ := obj.dataman.GetAllMapData(obj.Door.TargetMapID)
		obj.DisplayName = fmt.Sprintf("Door to %s", mapInfo.DisplayName)
	}
}
