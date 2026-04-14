package object

import (
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/tiled"
)

type Chair struct {
	InUse     bool
	SitterID  id.CharacterStateID
	OwnerID   id.CharacterStateID // only used to inform NPCs about which chairs to sit in
	Direction byte                // the direction this chair faces
}

const (
	PropChairDirection = "chair_direction"
	PropChairOwner     = "chair_owner"
)

func (obj *Object) loadChairObject(allProps []tiled.Property) {
	chairDirection, found := tiled.GetStringProperty(PropChairDirection, allProps)
	if !found {
		logz.Panicln("loadChairObject", "chair object didn't have chair_direction set:", obj.ID, obj.Name)
	}
	var dir byte
	switch chairDirection {
	case "L":
		dir = 'L'
	case "R":
		dir = 'R'
	case "U":
		dir = 'U'
	case "D":
		dir = 'D'
	default:
		logz.Panicln("loadChairObject", "invalid chair direction:", chairDirection)
	}

	obj.Chair.Direction = dir
}

func (c *Chair) LeaveChair(sitterID id.CharacterStateID) {
	if sitterID != c.SitterID {
		logz.Panicln("given sitterID didn't match the sitter ID in chair. passed in:", sitterID, "actual sitter ID:", c.SitterID)
	}
	c.InUse = false
	c.SitterID = ""
}

func (obj *Object) activateChair(params ObjectActivationParams) ObjectUpdateResult {
	if obj.Chair.InUse {
		if params.ActivatorID == obj.Chair.SitterID {
			// character is leaving chair
			obj.Chair.LeaveChair(params.ActivatorID)
			return ObjectUpdateResult{UpdateOccurred: true}
		}
		// a different character is already sitting here
		return ObjectUpdateResult{AlreadyInUse: true}
	}
	obj.Chair.InUse = true
	obj.Chair.SitterID = params.ActivatorID
	return ObjectUpdateResult{UpdateOccurred: true}
}
