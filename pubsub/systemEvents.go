package pubsub

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
)

const (
	// NOTE: Only used by world to enact map occupancy changes from background threads; not meant for detecting map changes!
	SysEventChangeMapOccupancy defs.EventType = "SYS_CHANGE_MAP_OCCUPANCY"

	// data:
	// 	- "type" (string) the name of the WorldEffect
	// 	- "effect" (any) the actual struct data for the WorldEffect
	SysScheduledWorldEffect defs.EventType = "SYS_SCHED_WORLD_EFFECT"

	// fires when a time lapse has occurred. should never be fired by anything other than the game engine; game projects can listen for this event
	// if they want to know when a time lapse has finished.
	//
	// Event Data:
	//
	// "time": clock.GameTime
	SysTimeLapse defs.EventType = "SYS_TIME_LAPSE"
)

func (eb *EventBus) SubscribeToWorldEvents(subscriberID string, fn func(defs.Event)) {
	eb.Subscribe(subscriberID, SysEventChangeMapOccupancy, fn)
}

type SysEventChangeMapOccupancyParams struct {
	CharacterStateID id.CharacterStateID
	From, To         defs.MapID
	ToSpawn          int
}

// SysChangeMapOccupancy changes map occupancy for an NPC.
// WARNING: Do not use outside of the game engine! Game projects should not call system events.
func (eb *EventBus) SysChangeMapOccupancy(charStateID id.CharacterStateID, from, to defs.MapID, toSpawn int) {
	params := SysEventChangeMapOccupancyParams{
		CharacterStateID: charStateID,
		From:             from,
		To:               to,
		ToSpawn:          toSpawn,
	}
	e := defs.Event{
		Type: SysEventChangeMapOccupancy,
		Data: map[string]any{
			"params": params,
		},
	}

	eb.Publish(e)
}
