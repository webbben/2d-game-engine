package pubsub

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
)

const (
	// NOTE: Only used by world to enact map occupancy changes from background threads; not meant for detecting map changes!
	SysEventChangeMapOccupancy defs.EventType = "SYS_CHANGE_MAP_OCCUPANCY"

	// fires when a character dies. Used by the world to do the cleanup that has to happen exactly once
	// per death -- dropping them out of any combat session, stopping their tasks, and scheduling the
	// corpse for removal later.
	//
	// Event Data:
	//
	// "charID": id.CharacterStateID
	SysEventEntityDied defs.EventType = "SYS_ENTITY_DIED"

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
	events := []defs.EventType{SysEventChangeMapOccupancy, SysScheduledWorldEffect, SysShowScreen, SysEventEntityDied}
	for _, eventType := range events {
		eb.Subscribe(fmt.Sprintf("%s_%s", subscriberID, eventType), eventType, fn)
	}
}

// SysEntityDied announces that a character has died, so the world can run its death cleanup.
func (eb *EventBus) SysEntityDied(charID id.CharacterStateID) {
	eb.Publish(defs.Event{
		Type: SysEventEntityDied,
		Data: map[string]any{
			"charID": charID,
		},
	})
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
