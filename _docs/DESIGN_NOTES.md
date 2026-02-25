# Design Notes

This document captures in-progress design ideas and decisions. Content here is preliminary and will be moved to authoritative documentation once features are implemented.

---

## Table of Contents

- [Locks System](#locks-system)
- [[Next Feature Name]](#next-feature)

---

## Locks System

### Overview

A system for locking doors in the game world. Locks prevent player access until unlocked via key or lockpicking. Locks can be static (always locked/unlocked) or time-based (e.g., shops locked at night).

### Motivation / Problem

- Doors in Tiled maps need a way to restrict player access
- Some doors should only be locked/unlocked at certain times (shop hours)
- Need to support two unlock methods: using a key, or lockpicking

### Proposed Design

#### Key Components

- **LockDef**: Definition of a lock's properties (difficulty, required key, etc.)
- **Door Object**: Tiled map object with attached lock properties
- **Lock State**: Runtime state tracking lock status (locked/unlocked, picked, broken)
- **Lockpick Skill**: Player skill that determines lockpicking success

#### Data Structures

```go
// In data/defs/ (or new file)

type LockID string     // unique identifier for a lock
type LockLevel int    // 1-100 scale, higher = harder to pick

type LockDef struct {
    ID              LockID
    Level           LockLevel
    BreakingChance float64        // chance lock breaks permanently when picking fails (0.0 - 1.0)
}

// Time range for time-based locks
type TimeRange struct {
    StartHour int // 0-23
    EndHour   int // 0-23
}

type LockType string

const (
    LockTypeStatic     LockType = "static"     // always locked or always unlocked
    LockTypeTimeBased LockType = "time_based" // changes based on time of day
)

// In data/state/ or entity/ (runtime state)
type DoorLockState struct {
    IsLocked   bool
    IsPicked   bool  // if true, lock was picked and cannot be locked again without reload
    IsBroken   bool  // if true, lock is broken and door can be opened freely
}
```

#### Tiled Integration

Doors in Tiled maps have these custom properties:

| Property | Type | Description |
|----------|------|-------------|
| `lock_ID` | string | Unique identifier for this lock (e.g., "shop_front_door", "dungeon_cell_1") - links to keys that can open it |
| `lock_level` | int | Difficulty level (1-100) - higher = harder to pick |
| `lock_type` | string | "static" or "time_based" |
| `locked` | bool | Whether door starts locked (for static type) |
| `breaking_chance` | float | Chance (0.0-1.0) lock breaks on failed pick attempt |
| `time_ranges` | JSON array | Time ranges when locked (for time_based type), e.g., `[{"start": 20, "end": 8}]` |

Example Tiled object properties:
```
lock_ID: "shop_front_door"
lock_level: 25
lock_type: "time_based"
locked: true
breaking_chance: 0.1
time_ranges: [{"start": 20, "end": 8}]
```

#### Behavior / Logic

1. **Door Interaction**
   - Player attempts to interact with door
   - Check if door has lock
   - If no lock or already unlocked → open door
   - If locked → prompt unlock options

2. **Unlock Options**
   - If player has required key → "Use Key" option
   - Always show "Pick Lock" option (if player has lockpicking tool)
   - Show difficulty indicator to player

3. **Lockpicking Mechanic**
   - Calculate success based on: Player Lockpick Skill vs Lock Difficulty
   - On success → door unlocks (state: picked)
   - On failure:
     - Random chance to break lock (per `breaking_chance`)
     - If broken → door opens freely
     - If not broken → lock remains, can try again
     - Optional: cooldown between attempts

4. **Time-Based Lock Updates**
   - On each in-game hour change, recalculate all time-based locks
   - Check current hour against all `time_ranges`
   - If hour falls within any range → locked
   - Update lock state accordingly
   - Notify player if lock state changes while they're nearby

### Edge Cases

- **Player inside building when lock engages**: Should not trap player. Lock only applies when approaching from outside?
- **Key required but not found**: Lock can still be picked (if not key-only)
- **Lock broken**: Door remains unlocked permanently for that save (or until game reload)
- **Multiple time ranges**: Locked if current time falls within ANY range (OR logic)
- **Wrap-around time ranges**: `start: 20, end: 8` means locked from 20:00 to 08:00 (next day)

### Open Questions

- [ ] Should locks have a "key only" option that disables lockpicking entirely?
- [ ] How do we handle locked doors that lead to interiors - should the player be able to exit if they got locked in?
- [ ] Do we need visual indicators for locked doors (padlock icon, etc.)?
- [ ] Should lockpicking require a tool item, or is it a bare-handed skill check?
- [ ] How do we handle NPCs with locked doors - they open can them regardless?

### Key Management

#### Problem

With small inventories (Minecraft-style), individual key items would quickly consume inventory slots. We need a streamlined approach.

#### Option 1: Key Ring

- New "Key Ring" inventory slot that holds any number of keys
- Keys are items that go into the key ring instead of main inventory
- Player can view/manage keys in a dedicated UI
- Keys persist in key ring across saves

```go
type Inventory struct {
    // ... existing fields
    KeyRing []*InventoryItem  // separate from main inventory
}
```

**Pros:**
- Player feels like they "have" keys, can see what they have
- Familiar mechanic (Morrowind, Skyrim)
- Keys can be dropped/given to NPCs if needed

**Cons:**
- Need UI for key ring management
- Still takes inventory space (even if separate slot)

#### Option 2: Automatic Key Collection

- Keys are given to player and automatically added to a "known keys" list
- No inventory representation - player cannot drop or manage keys
- Once obtained, always available for that save
- No UI needed for key management

```go
type PlayerKeyState struct {
    KnownKeyIDs map[ItemID]bool  // keys the player has obtained
}
```

**Pros:**
- No inventory management needed
- Simple to implement
- Player never loses keys

**Cons:**
- Less tactile - no item to hold/manage
- If player forgets they have a key, no way to check (need UI anyway?)
- Can't trade/give keys to NPCs

#### Recommendation

**Option 2 (Automatic Collection)** for simplicity, with a small "Keys" section in the player menu to show which keys the player has.

This could be expanded later if there's a design need to trade/drop keys.

#### Key Data Structure

Keys link to locks via `lock_ID` (not key ID):

```go
// In data/defs/
type KeyDef struct {
    ID          ItemID
    DisplayName string        // "Shop Front Key", "Dungeon Cell Key"
    Description string        // "Opens the front door of the general store"
    KeyRingIcon *ebiten.Image // optional icon for key ring UI
    
    // Lock IDs this key can open - allows one key to open multiple doors
    UnlocksLockIDs []LockID
}
```

---

## [[Next Feature Name]]

