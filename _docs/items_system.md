# Items System

The items system defines all items in the game (weapons, armor, consumables, quest items) and manages inventory operations.

---

## Architecture Overview

```
ItemDef (interface)
├── ItemBase         // Common fields and methods
├── WeaponDef        // Weapons with damage/speed
├── ArmorDef         // Body armor with protection
├── PotionDef        // Consumables with effects
└── KeyDef           // Keys for locks

Inventory
├── StandardInventory    // Character backpack + equipment
│   ├── CoinPurse         // Currency items
│   ├── InventoryItems    // Backpack slots
│   └── Equipment         // Equipped items (slots)
└── InventoryItem        // Item instance with durability/quantity
```

---

## ItemDef Interface

All items implement this interface:

```go
type ItemDef interface {
    GetID() ItemID
    GetName() string
    GetDescription() string
    GetValue() int          // Sell price
    GetWeight() float64     // Inventory weight
    GetMaxDurability() int
    
    GetTileImg() *ebiten.Image           // Inventory icon
    GetEquipedTiles() []*ebiten.Image     // Equipped appearance
    
    GetBodyPartDef() *defs.SelectedPartDef    // Body appearance
    GetLegsPartDef() *defs.SelectedPartDef     // Legs (bodywear only)
    
    // Type checks
    IsGroupable() bool       // Stacks with same items
    IsWeapon() bool
    IsHeadwear() bool
    IsBodywear() bool
    IsFootwear() bool
    IsAmulet() bool
    IsRing() bool
    IsAmmunition() bool
    IsAuxiliary() bool       // Shields, torches
    IsConsumable() bool
    IsMiscItem() bool
    IsCurrencyItem() bool
    IsEquipable() bool
    
    GetItemType() ItemType
    Load()                    // Load images
    Validate()                // Validate definition
}
```

---

## Item Types

| Type | Equip Slot | Description |
|------|-----------|-------------|
| `TypeWeapon` | Right hand | Melee or ranged weapons |
| `TypeHeadwear` | Head | Helmets, hats |
| `TypeBodywear` | Body | Armor, clothing |
| `TypeFootwear` | Feet | Boots, shoes |
| `TypeAmulet` | Neck | Necklaces, amulets |
| `TypeRing` | Finger | Two rings allowed |
| `TypeAuxiliary` | Left hand | Shields, torches |
| `TypeAmmunition` | Quiver | Arrows, bolts |
| `TypeConsumable` | Inventory | Food, potions |
| `TypeCurrency` | Coin purse | Gold, coins |
| `TypeMisc` | Inventory | Quest items, general |
| `TypeKey` | Inventory | Opens locks |

---

## ItemBase

Embed `ItemBase` in your item struct to satisfy `ItemDef`:

```go
type ItemBase struct {
    ID            ItemID
    Name          string
    Description   string
    Type          ItemType
    Value         int
    Weight        float64
    MaxDurability int
    Groupable     bool
    
    TileImgTilesetSrc string  // Tileset for inventory icon
    TileImgIndex      int     // Tile index in tileset
    TileImg          *ebiten.Image  // Loaded image
    
    BodyPartDef   *defs.SelectedPartDef  // Visual for equipped
    LegsPartDef   *defs.SelectedPartDef  // Legs (bodywear only)
}
```

---

## Creating Items

### Using NewItemBase

For simple items without special logic:

```go
func NewRustySword() *item.ItemBase {
    return item.NewItemBase(item.ItemBaseParams{
        ID:                "rusty_sword",
        Name:              "Rusty Sword",
        Description:       "An old sword, pitted with rust.",
        Type:              item.TypeWeapon,
        Value:             15,
        Weight:            3.0,
        MaxDurability:     50,
        TileImgTilesetSrc: "items/weapons.png",
        TileImgIndex:      0,
        Groupable:         false,
        BodyPartDef:       getSwordBodyPart(),
    })
}
```

### WeaponDef

For weapons with damage and speed:

```go
type WeaponDef struct {
    ItemBase
    Damage        int     // Damage per attack
    HitsPerSecond float64 // Attack speed
    FxPartDef     *defs.SelectedPartDef // Visual effects
}

weapon := &item.WeaponDef{
    ItemBase: *item.NewItemBase(params),
    Damage:        12,
    HitsPerSecond: 1.5,
    FxPartDef:     getSlashFxPart(),
}
```

### ArmorDef

For body armor with protection:

```go
type ArmorDef struct {
    ItemBase
    Protection int // Damage reduction
}

armor := &item.ArmorDef{
    ItemBase: *item.NewItemBase(params),
    Protection: 5,
}
```

### PotionDef

For consumables with effects:

```go
type PotionDef struct {
    ItemBase
    EffectDuration time.Duration
    HealAmount     int // HP per second
}

potion := &item.PotionDef{
    ItemBase: *item.NewItemBase(item.ItemBaseParams{
        ID:          "health_potion",
        Name:        "Health Potion",
        Description: "Restores health over time.",
        Type:        item.TypeConsumable,
        Value:       25,
        Weight:      0.2,
        Groupable:   true,
        // ... tile info
    }),
    EffectDuration: 5 * time.Second,
    HealAmount:     10,
}
```

### KeyDef

For keys that unlock locks:

```go
type KeyDef struct {
    ItemBase
    LockIDs []string // Locks this key can open
}

key := &item.KeyDef{
    ItemBase: *item.NewItemBase(params),
    LockIDs:  []string{"dungeon_door", "cell_key"},
}
```

---

## Equipment Slots

Characters have these equipment slots:

```go
type EquipedItems struct {
    EquipedHeadwear  *InventoryItem
    EquipedBodywear  *InventoryItem
    EquipedFootwear  *InventoryItem
    EquipedAmulet    *InventoryItem
    EquipedRing1     *InventoryItem
    EquipedRing2     *InventoryItem
    EquipedAmmo      *InventoryItem
    EquipedAuxiliary *InventoryItem  // Left hand (shields, torches)
    EquipedWeapon    *InventoryItem  // Right hand
}
```

---

## InventoryItem

An item instance in inventory:

```go
type ItemInstance struct {
    DefID      ItemID   // Reference to ItemDef
    Durability int     // Current condition (degrades with use)
}

type InventoryItem struct {
    Instance ItemInstance
    Def      ItemDef
    Quantity int       // Stack size
}
```

**Rules:**
- `Groupable` items stack (same `DefID`)
- Non-groupable items cannot stack
- Items with `MaxDurability > 0` cannot be groupable

---

## Inventory Operations

### Adding Items

```go
func (inv *Inventory) AddItems(items []defs.InventoryItem) []defs.InventoryItem {
    // Returns items that couldn't fit
}
```

**Behavior:**
1. For groupable items, finds existing stack and merges
2. Finds first empty slot for remaining items
3. Returns items that couldn't fit

### Removing Items

```go
slot.Clear()  // Remove item from slot
```

### Equipment Management

```go
// Equip item
slot.SetContent(&instance, itemDef, 1)

// Unequip item
slot.Clear()
```

---

## Visual Appearance

### TileImg

The image shown in inventory slots.

```go
TileImgTilesetSrc: "items/inventory.png"
TileImgIndex:      15  // Tile index in the tileset
```

### BodyPartDef

Visual appearance when equipped on character.

```go
BodyPartDef: &defs.SelectedPartDef{
    TilesLeft:  0,
    TilesRight: 1,
    TilesUp:    2,
    TilesDown:  3,
}
```

### Weapon FxPart

Additional visual effect for weapons:

```go
FxPartDef: &defs.SelectedPartDef{
    TilesLeft:  4,
    TilesRight: 5,
    // ...
}
```

---

## Item Validation

`ItemBase.Validate()` enforces:

1. `Name` and `ID` are set
2. `Value >= 0`
3. `Description` is set
4. `TileImgTilesetSrc` is set
5. Non-groupable if `MaxDurability > 0`
6. Valid type-specific requirements:
   - Visible equipables must have `BodyPartDef`
   - Bodywear must have both `BodyPartDef` and `LegsPartDef`
   - Headwear must NOT have `LegsPartDef`

---

## Best Practices

1. **Use consistent tilesets** - Group items by tileset for efficient loading
2. **Define fallback items** - For dynamically generated items
3. **Track durability** - Items with durability can't stack
4. **Weight limits** - Consider carrying capacity
5. **Value balance** - Prices should feel fair
6. **Descriptive names** - Players should understand items from names
7. **Load items at startup** - Don't load during gameplay

---

## Loading Items

Items are loaded via DataManager:

```go
datamanager.LoadItemDef(itemDef defs.ItemDef)
```

Call `itemDef.Load()` before loading to initialize images:

```go
mySword.Load()
dataman.LoadItemDef(mySword)
```

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
