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
		return ObjectUpdateResult{AlreadyInUse: true}
	}

	obj.Bed.SleeperID = params.ActivatorID

	return ObjectUpdateResult{UpdateOccurred: true}
}
