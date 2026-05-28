package defs

import "github.com/webbben/2d-game-engine/logz"

type (
	MapID    string
	RegionID string
	MapType  int
)

// MapDef defines a map in the game.
// Most of the details of a map and its "definition" come from the Tiled map (.tmj) file.
// We only put things here if it cannot be effectively "defined" in the map file, or if it's easier to manage here.
//
// Here are some things that are directly defined in the map file:
//
// - Initial items: we place items directly in the map as objects
//
// - Doors: these are placed in the map as objects
//   - initial door locks are set in the object properties
//
// TODO: should we add the following to this struct, instead of managing within the Tiled map?
// - the Daylight factor; how much daylight will show in this map.
//   - currently set in the tiled map, but I wonder if it's better set here. easier to discover, edit, etc. You can't actually see how much "daylight" there
//     is from the Tiled map editor anyway, so it can easily be forgotten there.
type MapDef struct {
	ID          MapID
	Region      RegionID // the overall location you are in, such as which city, or which forest, etc.
	DisplayName string

	// if true, this map def will be considered as just a template for map generators; so, it won't directly be built into the world as a unique map.
	// "templates" are just map defs that represent a generic map, not a specific, unique map.
	// To use this mapDef with a map generator, you MUST set this to true.
	IsMapGenTemplate bool
}

func (md MapDef) Validate() {
	if md.ID == "" {
		panic("id was empty")
	}
	if md.Region == "" {
		logz.Panicln(string(md.ID), "no region specified")
	}
	if md.DisplayName == "" {
		logz.Panicln(string(md.ID), "no display name set")
	}
}

// A MapLock is used for pairing a lock ID to the map that has the locked object in it.
// Mostly used in places like quest actions.
type MapLock struct {
	MapID  MapID
	LockID string
}

type ContainerDef struct {
	ID        string
	Inventory []*InventoryItem
}

type ContainerGenerator interface {
	ID() string
	GenerateItems(ctx GameContext) []*InventoryItem
}

// A MapGenerator is for generating an instance of a MapDef while customizing details, enabling maps to be reused.
// You'd use a MapGenerator for creating homes of generic NPCs, like an insignificant villager or something like that.
// It could be used for more complex maps too, but ideally it shouldn't be overused since these maps will ultimately just be
// "filler" in a game. But, better to use a MapGenerator than to make 20 individual identical maps for a bunch of generic villagers.
type MapGenerator struct {
	ID       string
	MapDefID MapID // the map def to use

	// Overrides: the following are for overriding fields from MapDef
	OverrideDisplayName string
	OverrideRegion      RegionID

	// The following defines how NPCs can be given a home/generated in this map
	// If more characters (in either slice) are defined than there are beds in the map, it panics.
	// both slices are not used at the same time; only one is used, with character defs getting priority
	// if both slices are defined, it panics

	// Character Defs are given priority; if any are defined, only these will be assigned to beds
	// Warning: CharacterDefs should represent unique NPCs; to reuse a character def, use/make a character generator for it.
	// (assignment is just based on the order in which bed objects are found)
	InhabitantCharacterDefs []CharacterDefID

	// If no character defs are defined, then any character gens defined here will get assigned to beds.
	InhabitantCharacterGens []string
}

func (mg MapGenerator) Validate() {
	if mg.ID == "" {
		logz.Panic("ID was empty")
	}
	if mg.MapDefID == "" {
		logz.Panicln("Validate", "map def ID was empty")
	}
	if len(mg.InhabitantCharacterDefs) > 0 && len(mg.InhabitantCharacterGens) > 0 {
		logz.Panicln("Validate", "both character defs and character gens are defined; that's not allowed, since only one or the other will get used (the others would be ignored)")
	}
}
