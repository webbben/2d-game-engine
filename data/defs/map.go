package defs

type MapID string

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
	DisplayName string
}
