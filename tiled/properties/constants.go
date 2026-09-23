// Package properties centralizes the Tiled custom property names used by the engine.
//
// All Tiled property names (for layers, tiles, and objects) should be referenced
// through the consts in this package, never as raw string literals. This is the
// single place to look when you want to know every property name the engine reads.
// It also ensures names are spelled consistently across the codebase.
package properties

// Layer properties
const (
	// PropTop marks a layer to be drawn on top of everything else.
	PropTop = "TOP"
)

// Tile/object properties shared by tiles and objects
const (
	// PropType is the object/tile type identifier (e.g. "DOOR", "GATE", "LIGHT").
	PropType = "TYPE"
	// PropSeeThrough, when true, means the tile/object never blocks line of sight,
	// even if it is collidable.
	PropSeeThrough = "see_through"
)

// Tile properties
const (
	// PropCollision is the collision shape of a tile (e.g. "WHOLE", "HALF_B", ...).
	PropCollision = "COLLISION"
	// PropWater marks a tile as water (treated as a collision).
	PropWater = "WATER"
	// PropMaterial is the ground material of a tile (used for footstep sounds, etc.).
	PropMaterial = "MATERIAL"
	// PropCost is a tile's pathfinding cost. (Currently unused.)
	PropCost = "cost"
	// PropNextTile is the next tile in a state-change animation chain.
	PropNextTile = "nextTile"
	// PropCoverHair marks an equipment head tile that covers the character's hair.
	PropCoverHair = "COVER_HAIR"
	// PropTodo marks an object as an authoring to-do (surfaces a warning in map validation). Not to be used in actual finalized game code or files.
	PropTodo = "todo"
)

// Object properties (general)
const (
	PropDisplayName      = "displayName"
	PropNoCollision      = "no_collision"     // if true, object has no collision
	PropCollisionHeight  = "collision_height" // height of the object's collision rect, in tiles
	PropOnObjID          = "on_obj_id"        // if set, render this object on top of another object
	PropOwnerCharID      = "owner_id"         // a unique character of the given ID owns this object
	PropRoleID           = "role_id"          // characters with this role can use this object
	PropContainerDefID   = "container_def_id"
	PropContainerGenID   = "container_gen_id"
	PropLockID           = "lock_id"
	PropLockLevel        = "lock_level"
	PropEmitterID        = "emitter_id"
	PropEmitterOffsetY   = "emitter_offset_y"
	PropItemID           = "item_id"
	PropZOffset          = "z_offset"
	PropSpawnIndex       = "spawn_index"
	PropFaceDirection    = "face_direction"
	PropReturnSpawnIndex = "return_spawn_index"
)

// Door/gate properties
const (
	PropDoorTo             = "door_to"
	PropDoorSpawnIndex     = "door_spawn_index"
	PropDoorActivate       = "door_activate"
	PropDoorMapGeneratorID = "map_generator_id"
	PropSFX                = "SFX" // open/close sound effect (doors and gates)
)

// Chair properties
const (
	PropChairDirection = "chair_direction"
	PropChairOwner     = "chair_owner" // if set, the chair is reserved for a specific character
)

// Sign properties
const (
	PropBookID = "book_id"
)

// Container properties
const (
	PropSFXOpen  = "sfx_open"
	PropSFXClose = "sfx_close"
)

// Task area properties
const (
	PropTaskID      = "task_id"
	PropTaskDir     = "task_dir"
	PropPatrolOrder = "patrol_order"
)

// Bed properties
const (
	PropCharacterGeneratorID = "characterGeneratorID"
	PropCharacterDefID       = "characterDefID"
)

// Light properties (on tiles and objects)
const (
	PropLightColorR            = "light_color_r"
	PropLightColorG            = "light_color_g"
	PropLightColorB            = "light_color_b"
	PropLightGlowFactor        = "light_glow_factor"
	PropLightOffsetY           = "light_offset_y"
	PropLightRadius            = "light_radius"
	PropLightInnerRadiusFactor = "light_inner_radius_factor"
	PropLightFlickerInterval   = "light_flicker_interval"
	PropLightMaxBrightness     = "light_max_brightness"
	PropLightCoreRadius        = "light_core_radius"
	PropLightPreset            = "light_preset" // deprecated; use PropLightColorR/G/B instead
)

// Window properties
const (
	PropWindowWidth        = "window_width"
	PropWindowLength       = "window_length"
	PropWindowMaxIntensity = "window_max_intensity"
	PropWindowDirX         = "window_dir_x"
	PropWindowDirY         = "window_dir_y"
)
