package state

import (
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
)

type MapState struct {
	ID defs.MapID

	DisplayName string        // override from map def
	RegionID    defs.RegionID // override from map def

	// if set, this map state was generated from a map generator.
	// This means the ID of this map state does NOT match the map def (likely has a random UUID appended), since it's potentially not unique.
	// Generated maps use the GeneratedMapDefID field to point to their map def.
	IsGenerated       bool
	GeneratedMapDefID defs.MapID // (generated maps only) ID of the MapDef used to make this map state.

	// scenarios that have been queued to run in this map. when this map loads, it will run the scenario at the top of this slice,
	// if any scenario exists. otherwise, it would run it's default behavior.
	QueuedScenarios []defs.ScenarioID

	// all items that exist in the map. there are two types of items that might be here:
	//
	// 1) an item that is "part of the map" and was put there during map creation (i.e. part of the tmj file).
	// its ID matches the Tiled object ID of the object that represents it, and it stays here until it is taken by the player.
	//
	// 2) an item that was dropped by the player or an NPC at runtime. these will expire and disappear eventually.
	//
	// On MapState creation, we should populate the items from category 1 into this map state. Category 2 items are populated at runtime.
	// ID is the join key in both directions: an Object finds its state with ID, and loadDroppedItems rebuilds an Object from the state.
	// Dropped records which category an item belongs to, so nothing has to be inferred from ID.
	MapItems []MapItemState

	// all objects that have a lock in the map have their locks tracked here.
	// a door or container object can have a lock, which prevents its opening unless the player has a key or can unlock it.
	MapLocks map[string]LockState

	// bed objects are tracked here, and are mapped by the integer object ID as found in the Tiled map.
	// beds are linked to a single character, and the character will go to this bed when sleeping.
	MapBeds map[int]BedState

	// container objects have their state tracked here. key is the integer object ID as found in the Tiled map.
	MapContainers map[int]*ContainerState

	// if a door has an override (e.g. goes to a generated map), it'll be defined here
	DoorOverrides map[int]DoorState
}

type MapItemState struct {
	// ID of the object from the Tiled map that represents this item.
	// 	- authored items: ID is directly derived from an object that is set in the tmj file.
	// 	- runtime drops: ID is incremented starting from tiled.Map.NextObjectID - the next object ID that *would've* been used for authoring.
	//
	// If you edit a map to add a new object and then load an existing save, any existing runtime drop items in that map could have their IDs overlap with
	// the newly added map objects. So if you edit a map in this way, existing saves should be considered invalid.
	ID int

	ItemState ItemState

	// Position where the item exists in the map.
	X, Y float64

	// if set, this item will expire and disappear from the map at the set game time
	ExpiresAt *clock.GameTime

	// If true, this item was created at runtime by World.DropItemOnGround; if false, authored in a tmj.
	Dropped bool
}

type LockState struct {
	LockID            string
	OriginalLockLevel int // the level this was originally defined as
	LockLevel         int
	Unlocked          bool
}

type BedState struct {
	MapObjID int                 // ID (in Tiled) of the object that represents this bed
	OwnerID  id.CharacterStateID // ID of the character (state) that owns this bed
}

type ContainerState struct {
	Inventory []*ItemState // slots / items in this container
}

type DoorState struct {
	// If this door goes to a generated map, then this value will be set (since generated maps don't know their true map (state) IDs until runtime)
	OverrideDestinationMap   defs.MapID
	OverrideDestinationSpawn *int
}
