# Tiled Integration

This document lists all custom properties used in Tiled maps and tilesets for the game engine.

---

## Overview

The game engine uses Tiled for map creation. Custom properties are defined on:
- **Map objects** (doors, gates, spawn points, etc.)
- **Tiles** (lights, terrain types, costs)
- **Layers** (terrain properties)

---

## Object Properties by Type

### TYPE (Required)

**Location:** Objects, Tiles  
**Type:** String  
**Description:** The primary type identifier for an object. Determines which struct to load.

| Type Value | Description |
|------------|-------------|
| `DOOR` | A portal to another map |
| `GATE` | An openable/closable barrier (door, gate) |
| `SPAWN_POINT` | Player spawn location |
| `LIGHT` | A light source |
| `CONTAINER` | A container (chest, crate) - TODO |
| `MISC` | General purpose, just collision |

---

## Door Properties

Used when `TYPE = "DOOR"`

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `door_to` | string | Yes | MapID to teleport to |
| `door_spawn_index` | int | No | Spawn point index on target map |
| `door_activate` | string | Yes | Activation type: `"click"` or `"step"` |
| `SFX` | string | Yes | Sound effect ID for door opening |

---

## Gate Properties

Used when `TYPE = "GATE"`

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `SFX` | string | Yes | Sound effect ID for gate opening |

---

## Spawn Point Properties

Used when `TYPE = "SPAWN_POINT"`

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `spawn_index` | int | Yes | Index identifier for this spawn point |

---

## Light Properties

Used when `TYPE = "LIGHT"` (on tiles or objects)

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `light_color_preset` | string | No | Color preset name (e.g., "warm", "cool") |
| `light_glow_factor` | float | No | Glow intensity (0.0-1.0) |
| `light_offset_y` | int | No | Vertical offset for light position |
| `light_radius` | int | No | Light radius in pixels |
| `light_inner_radius_factor` | float | No | Inner radius as fraction of radius |
| `light_flicker_interval` | int | No | Flicker interval in milliseconds |
| `light_max_brightness` | float | No | Maximum brightness (0.0-1.0) |
| `light_core_radius` | float | No | Core radius as fraction of radius |

---

## Tile Properties

### TYPE (Tile Classification)

**Location:** Tiles (in tilesets)  
**Type:** String  
**Description:** Classifies a tile for special handling.

---

## Layer Properties

### cost

**Location:** Tile layers  
**Type:** Int  
**Description:** Additional pathfinding cost for tiles with this property. Values are added to the tile's base cost.

---

## Future / Planned Properties

### Lock Properties (Design Draft)

These properties are planned for the locks system:

| Property | Type | Description |
|----------|------|-------------|
| `lock_ID` | string | Unique identifier for this lock |
| `lock_level` | int | Difficulty level (1-100) for lockpicking |
| `lock_type` | string | "static" or "time_based" |
| `locked` | bool | Whether door starts locked (for static type) |
| `breaking_chance` | float | Chance (0.0-1.0) lock breaks on failed pick |
| `time_ranges | Time ranges when locked,` | JSON array e.g., `[{"start": 20, "end": 8}]` |

---

## Property Access Functions

The engine provides helper functions in `internal/tiled/`:

```go
// Get a string property
tiled.GetStringProperty("property_name", props)

// Get an int property
tiled.GetIntProperty("property_name", props)

// Get a float property
tiled.GetFloatProperty("property_name", props)

// Get a bool property
tiled.GetBoolProperty("property_name", props)

// Get tile type (TYPE property)
tiled.GetTileType(tile)

// Get light properties from tile/object
tiled.GetLightProps(props)
```

---

## How Properties Are Loaded

1. **Object loading** (`object/object.go`):
   - Properties are collected from both the object level AND the embedded tile level
   - The `TYPE` property is checked first to determine object type
   - Type-specific properties are then parsed

2. **Tile loading** (`internal/tiled/tileset.go`):
   - Each tile can have properties defined in the tileset
   - Light properties are extracted using `GetLightProps()`

3. **Layer loading** (`internal/tiled/map.go`):
   - Layer properties can define terrain costs
   - Tile-level properties (like `cost`) are applied to the pathfinding cost map

---

## Adding New Properties

To add a new property:

1. **Define the property in Tiled**:
   - Open the map/tileset in Tiled
   - Add a new custom property to the object/tile/layer
   - Set the type (string, int, float, bool)

2. **Add property parsing in code**:
   - Find the appropriate loading function (e.g., `loadDoorObject`)
   - Add a case to the switch statement for the property name
   - Parse the value using the appropriate getter

3. **Update this documentation**:
   - Add the property to the appropriate table above

---

## Notes

- Property names are case-sensitive
- Tiled's `class` field is also available but currently unused
- Some properties accept JSON arrays (like `time_ranges` for locks)
- Properties can be defined at multiple levels (object + tile) and are merged
