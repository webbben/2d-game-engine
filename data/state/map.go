package state

import (
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
)

type MapState struct {
	ID defs.MapID

	// scenarios that have been queued to run in this map. when this map loads, it will run the scenario at the top of this slice,
	// if any scenario exists. otherwise, it would run it's default behavior.
	QueuedScenarios []defs.ScenarioID

	// all items that exist in the map. there are two types of items that might be here:
	//
	// 1) an item that is "part of the map" and was put there during map creation (but can be picked up by the player).
	// These will always be there, unless the player or an NPC picks it up.
	//
	// 2) an item that is dropped by the player or an NPC. these will expire and disappear eventually.
	// - TODO: once this expiry stuff is implemented, we need to decide how a player can put an item down that shouldn't expire,
	//   vs putting an item down with the intention of "throwing it away".
	//
	// On MapState creation, we should populate the items from category 1 into this map state.
	MapItems []MapItemState

	// all objects that have a lock in the map have their locks tracked here.
	// a door or container object can have a lock, which prevents its opening unless the player has a key or can unlock it.
	MapLocks map[string]LockState
}

type MapItemState struct {
	ItemInstance defs.ItemInstance
	Quantity     int
	X, Y         float64
	ExpiresAt    *clock.GameTime // if set, this item will expire and disappear from the map at the set game time
}

type LockState struct {
	LockID    string
	LockLevel int
	Unlocked  bool
}
