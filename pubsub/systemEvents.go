package pubsub

import (
	"fmt"

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

	// fires when a screen should be shown while in an active map (e.g. when a container is opened).
	// Not used for things like screens that show during dialog (e.g. trade screens); those use the dialog action instead.
	//
	// Event Data:
	// 	- "screen_id" (string) ScreenID
	// 	- "params" (any) the params to pass to the screen
	SysShowScreen defs.EventType = "SYS_SHOW_SCREEN"
)

func (eb *EventBus) SubscribeToWorldEvents(subscriberID string, fn func(defs.Event)) {
	events := []defs.EventType{SysEventChangeMapOccupancy, SysScheduledWorldEffect, SysShowScreen}
	for _, eventType := range events {
		eb.Subscribe(fmt.Sprintf("%s_%s", subscriberID, eventType), eventType, fn)
	}
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
