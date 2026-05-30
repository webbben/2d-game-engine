# Items System

The items system defines all items in the game (weapons, armor, consumables, quest items) and manages inventory operations.

---

## Architecture Overview

Item definitions and item state are now separate concepts:

- **ItemDef (struct)** — the immutable definition of an item (name, type, stats, visual references)
- **ItemState** — the mutable runtime state (quantity, durability, which item def it refers to)
- **StandardInventory** — a character's inventory composed of `ItemState` objects

```
ItemDef (struct)
├── ID, Name, Description
├── Type                    // ItemType enum
├── Value, Weight, MaxDurability, Groupable
├── TileImgTilesetSrc, TileImgIndex   // Visual references
├── BodyPartDef, LegsPartDef          // Equipped appearance
├── LockIDs                           // Key-specific
├── Damage, FxPartDef                 // Weapon-specific
└── Protection                        // Armor-specific

ItemState (struct)
├── DefID          // References ItemDef by ID
├── Durability     // Current condition
└── Quantity       // Stack size

StandardInventory (struct)
├── CoinPurse          []*ItemState
├── InventoryItems     []*ItemState
├── EquipedHeadwear    *ItemState
├── EquipedBodywear    *ItemState
├── EquipedFootwear    *ItemState
├── EquipedAmulet      *ItemState
├── EquipedRing1       *ItemState
├── EquipedRing2       *ItemState
├── EquipedAmmo        *ItemState
├── EquipedAuxiliary   *ItemState
└── EquipedWeapon      *ItemState
```

---

## ItemDef Struct

Item definitions live in `data/defs/item.go`. `ItemDef` is a plain struct (not an interface):

```go
type ItemDef struct {
    ID            ItemID
    Name          string
    Description   string
    Type          ItemType
    Value         int
    Weight        float64
    MaxDurability float64
    Groupable     bool

    TileImgTilesetSrc string   // tileset for inventory icon
    TileImgIndex      int      // tile index in tileset

    // Equipped visual appearance (nil-able pointers)
    BodyPartDef *SelectedPartDef
    LegsPartDef *SelectedPartDef   // bodywear only

    // Key-specific
    LockIDs []string

    // Weapon-specific
    Damage    int
    FxPartDef *SelectedPartDef

    // Armor (body/head/footwear/auxiliary)
    Protection int
}
```

All item types use the same `ItemDef` struct — there are no separate `WeaponDef`, `ArmorDef`, `PotionDef`, or `KeyDef` structs. Type-specific fields like `Damage`, `Protection`, and `LockIDs` are present on all `ItemDef` values but only meaningful for the relevant types.

### Key methods on ItemDef

```go
func (id ItemDef) Validate()        // Panics if definition is invalid
func (def ItemDef) IsEquipable() bool // Returns true for wearable/weapon types
```

---

## Item Types

| Type | Equip Slot | Description |
|------|-----------|-------------|
| `TypeWeapon` | Weapon slot | Melee or ranged weapons |
| `TypeHeadwear` | Head | Helmets, hats |
| `TypeBodywear` | Body | Armor, clothing |
| `TypeFootwear` | Feet | Boots, shoes |
| `TypeAmulet` | Neck | Necklaces, amulets |
| `TypeRing` | Finger | Two rings allowed |
| `TypeAuxiliary` | Auxiliary slot | Shields, torches |
| `TypeAmmunition` | Ammo slot | Arrows, bolts |
| `TypeConsumable` | Inventory | Food, potions |
| `TypeCurrency` | Coin purse | Gold, coins |
| `TypeMisc` | Inventory | Quest items, general |
| `TypeKey` | Inventory | Opens locks |

---

## ItemState — Runtime State

Item state lives in `data/state/item.go`. It tracks the mutable part of items:

```go
type ItemState struct {
    DefID      defs.ItemID
    Durability float64
    Quantity   int
}
```

**Rules:**
- `Groupable` items stack (same `DefID`). Quantity tracks the stack size.
- Non-groupable items have `Quantity == 1`.
- Items with `MaxDurability > 0` cannot be groupable (enforced by `ItemDef.Validate()`).

---

## StandardInventory — Character Inventory

`StandardInventory` is embedded in `state.CharacterState`:

```go
type StandardInventory struct {
    CoinPurse      []*ItemState
    InventoryItems []*ItemState

    EquipedHeadwear  *ItemState
    EquipedBodywear  *ItemState
    EquipedFootwear  *ItemState
    EquipedAmulet    *ItemState
    EquipedRing1     *ItemState
    EquipedRing2     *ItemState
    EquipedAmmo      *ItemState
    EquipedAuxiliary *ItemState
    EquipedWeapon    *ItemState
}
```

Equipment slots are direct fields on the struct (no separate `EquipedItems` wrapper).

---

## Creating Items

### Defining Item Definitions

Item definitions are created as literal `defs.ItemDef` values and loaded via the `DataManager`:

```go
// define a weapon
rustySword := defs.ItemDef{
    ID:                "rusty_sword",
    Name:              "Rusty Sword",
    Description:       "An old sword, pitted with rust.",
    Type:              defs.TypeWeapon,
    Value:             15,
    Weight:            3.0,
    MaxDurability:     50,
    TileImgTilesetSrc: "items/weapons.png",
    TileImgIndex:      0,
    BodyPartDef:       getSwordBodyPart(),
    FxPartDef:         getSlashFxPart(),
    Damage:            12,
}

// define a key
dungeonKey := defs.ItemDef{
    ID:                "dungeon_key",
    Name:              "Dungeon Key",
    Description:       "Opens the old dungeon door.",
    Type:              defs.TypeKey,
    Value:             0,
    Weight:            0.1,
    TileImgTilesetSrc: "items/misc.png",
    TileImgIndex:      3,
    LockIDs:           []string{"dungeon_door"},
}
```

### Creating ItemState

Use the DataManager to create `ItemState` instances from a def ID:

```go
itemState := dataman.NewItemState("rusty_sword", 1)
// Returns: &state.ItemState{DefID: "rusty_sword", Quantity: 1, Durability: <max durability from def>}
```

### Initial Standard Inventory Definition

Character definitions specify starting inventories using `InitialStandardInventoryDef`:

```go
type InitialStandardInventoryDef struct {
    CoinPurse      []*ItemInitialStateDef
    InventoryItems []*ItemInitialStateDef

    EquipedHeadwear  *ItemInitialStateDef
    EquipedBodywear  *ItemInitialStateDef
    // ... other equipment slots
}
```

Where `ItemInitialStateDef` specifies the def ID, quantity, and initial durability.

---

## Inventory Operations

### Adding Items

```go
func AddItemToStandardInventory(inv *state.StandardInventory, invItem state.ItemState, dataman *datamanager.DataManager) (bool, state.ItemState)
```

**Behavior:**
1. If currency, tries coin purse first
2. For groupable items, finds existing stack and merges
3. Finds first empty slot for remaining items
4. Returns remaining items that couldn't fit

### Removing Items

```go
func RemoveItemFromStandardInventory(inv *state.StandardInventory, itemToRemove state.ItemState, dataman *datamanager.DataManager) (bool, state.ItemState)
```

### Equipment Management

Equipment is managed via `characterstate.EquipItem()`:

```go
func EquipItem(cs *state.CharacterState, i *state.ItemState, dataman *datamanager.DataManager) (success bool)
```

- Unequips current item in that slot (returns it to inventory) before equipping the new one
- Panics if item is not equippable

Direct slot access is also possible:

```go
charState.EquipedWeapon = nil       // Unequip weapon
charState.EquipedWeapon = itemState // Equip weapon directly
```

### Money Management

```go
item.CountMoney(inv state.StandardInventory, dataman)  // Count total money
characterstate.SpendMoney(inv, value, dataman)          // Spend money
characterstate.EarnMoney(inv, value, dataman)           // Earn money
```

---

## Visual Appearance

### Tile Image (Inventory Icon)

The inventory icon is loaded using the tileset source and tile index from the ItemDef:

```go
tileImage := tiled.GetTileImage(itemDef.TileImgTilesetSrc, itemDef.TileImgIndex, true)
```

### BodyPartDef (Equipped Appearance)

Visual appearance when equipped on a character, defined by a `SelectedPartDef`:

```go
BodyPartDef: &defs.SelectedPartDef{
    TilesLeft:  0,
    TilesRight: 1,
    TilesUp:    2,
    TilesDown:  3,
}
```

### Weapon FxPart

Weapons have an additional visual effects part:

```go
FxPartDef: &defs.SelectedPartDef{
    TilesLeft:  4,
    TilesRight: 5,
    // ...
}
```

The weapon and fx parts are retrieved together:

```go
weaponPart, fxPart := item.GetWeaponParts(itemDef)
```

---

## Item Validation

`ItemDef.Validate()` enforces:

1. `Name` and `ID` are set
2. `Value >= 0`
3. `Description` is set
4. `TileImgTilesetSrc` is set
5. Items with `MaxDurability > 0` cannot be groupable
6. `Type` is set to a valid value
7. Valid type-specific requirements:
   - Visible equipables (weapon, bodywear, headwear, footwear, auxiliary) must have `BodyPartDef`
   - Non-visible equipables must NOT have `BodyPartDef`
   - Bodywear must have both `BodyPartDef` and `LegsPartDef`
   - Headwear must NOT have `LegsPartDef`

`ItemState.Validate()` enforces:
1. `DefID` is set
2. `Quantity > 0`
3. `Durability >= 0`

---

## Loading Items

Item definitions are loaded via DataManager:

```go
func (dataman *DataManager) LoadItemDefs(itemDefs []defs.ItemDef)
```

This calls `Validate()` on each definition and registers it in the DataManager's map.

Retrieve item definitions at runtime:

```go
itemDef := dataman.GetItemDef(defID)
```

Tile images are loaded on demand via `tiled.GetTileImage()` — there is no separate `Load()` method on `ItemDef`.

---

## Best Practices

1. **Use consistent tilesets** — Group items by tileset for efficient loading
2. **Define fallback items** — For dynamically generated items
3. **Track durability** — Items with durability can't stack
4. **Weight limits** — Consider carrying capacity
5. **Value balance** — Prices should feel fair
6. **Descriptive names** — Players should understand items from names
7. **Load items at startup** — Don't load during gameplay
8. **One struct, all types** — `ItemDef` handles all item types via its fields; type-specific logic reads the relevant fields based on `Type`

---

## Item Constants

```go
const (
    TypeWeapon     ItemType = "WEAPON"
    TypeBodywear   ItemType = "BODYWEAR"
    TypeHeadwear   ItemType = "HEADWEAR"
    TypeFootwear   ItemType = "FOOTWEAR"
    TypeAmulet     ItemType = "AMULET"
    TypeRing       ItemType = "RING"
    TypeAmmunition ItemType = "AMMUNITION"
    TypeAuxiliary  ItemType = "AUXILIARY"
    TypeConsumable ItemType = "CONSUMABLE"
    TypeMisc       ItemType = "MISC"
    TypeCurrency   ItemType = "CURRENCY"
    TypeKey        ItemType = "KEY"
)
```
