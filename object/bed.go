package object

import (
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/logz"
)

type Bed struct {
	InUse     bool
	SleeperID id.CharacterStateID // ID of character sleeping in the bed right now
}

func (b *Bed) LeaveBed(sleeperID id.CharacterStateID) {
	if sleeperID != b.SleeperID {
		logz.Panicln("LeaveBed", "id passed in did not match sleeper ID. passed in:", sleeperID, "sleeperID:", b.SleeperID)
	}
	b.SleeperID = ""
	b.InUse = false
}

func (obj *Object) activateBed(params ObjectActivationParams) ObjectUpdateResult {
	if params.ActivatorID == "" {
		panic("activator ID was empty")
	}

	if obj.Bed.InUse {
		if obj.Bed.SleeperID == params.ActivatorID {
			// the character sleeping in the bed is leaving it now
			// TODO: right now, the player and NPCs don't activate a bed to leave it. only for sleeping in it.
			obj.Bed.LeaveBed(params.ActivatorID)
			return ObjectUpdateResult{UpdateOccurred: true}
		}
		// a different character is already sleeping in this bed
		return ObjectUpdateResult{AlreadyInUse: true}
	}

	obj.Bed.SleeperID = params.ActivatorID
	obj.Bed.InUse = true

	return ObjectUpdateResult{UpdateOccurred: true}
}
