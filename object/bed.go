package object

import "github.com/webbben/2d-game-engine/data/state"

type Bed struct {
	InUse     bool
	SleeperID state.CharacterStateID // ID of character sleeping in the bed right now
	OwnerID   state.CharacterStateID // ID of character who owns this bed.
}

func (obj *Object) activateBed(params ObjectActivationParams) ObjectUpdateResult {
	if params.ActivatorID == "" {
		panic("activator ID was empty")
	}

	if obj.Bed.InUse {
		if obj.Bed.SleeperID == params.ActivatorID {
			// the character sleeping in the bed is leaving it now
			obj.Bed.SleeperID = ""
			obj.Bed.InUse = false
			return ObjectUpdateResult{UpdateOccurred: true}
		}
		// a different character is already sleeping in this bed
		return ObjectUpdateResult{AlreadyInUse: true}
	}

	obj.Bed.SleeperID = params.ActivatorID
	obj.Bed.InUse = true

	return ObjectUpdateResult{UpdateOccurred: true}
}
