package item

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

type ItemDef interface {
	GetID() string          // the internal ID of this item.
	GetName() string        // the display name of this item.
	GetDescription() string // the description of the item
	GetValue() int          // the full value of this item, if it were sold at maximum price.
	GetWeight() float64     // the weight of this item, which factors into the player's inventory weight.
	GetMaxDurability() int  // the full durability of this item. a higher value means it takes longer to break.

	GetTileImg() *ebiten.Image // gets the "tile image", i.e. the image used in places like the inventory slots

	// if true, item can be grouped together with other same items in inventories
	// this tends to be true for items that don't have durability, like potions or arrows, but not for weapons and armor.
	IsGroupable() bool
	// if true, this item is treated as an equipable weapon that is held
	IsWeapon() bool
	// if true, this item is treated as an equipable piece of armor that is worn
	IsArmor() bool
	// if true, this item is treated as an equipable accessory (amulet, ring, etc) that is worn
	IsAccessory() bool
	// if true, this item is treated as an equipable ammunition type (arrows, etc)
	IsAmmunition() bool
	// if true, this item is treated as an item that can be consumed (food, potions, etc)
	IsConsumable() bool
	// if true, this item has no specific use or utility; it just exists in your inventory and may have value or weight
	IsMiscItem() bool
	// determines if this item can be equiped
	IsEquipable() bool

	Load() // load things like images

	Validate() // checks if item def is properly defined
}

// includes the basic functions required for an item to implement the ItemDef interface.
// embed into a struct to make it effectively an item type.
type ItemBase struct {
	init          bool // flag to indicate if item has been loaded yet
	ID            string
	Name          string
	Description   string
	Value         int
	Weight        float64
	MaxDurability int
	Groupable     bool

	Weapon     bool
	Armor      bool
	Accessory  bool
	Ammunition bool
	Consumable bool
	MiscItem   bool

	TilesetSource string
	TileID        int
	TileImg       *ebiten.Image
}

func (ib ItemBase) Validate() {
	if ib.Value < 0 {
		panic("value < 0")
	}
	if ib.Name == "" {
		panic("item has no name")
	}
	if ib.ID == "" {
		panic("item has no ID")
	}
	if ib.Description == "" {
		panic("item has no description")
	}
	i := 0
	if ib.Weapon {
		i++
	}
	if ib.Armor {
		i++
	}
	if ib.Accessory {
		i++
	}
	if ib.Ammunition {
		i++
	}
	if ib.Consumable {
		i++
	}
	if ib.MiscItem {
		i++
	}
	if i == 0 {
		panic("item has no designated type")
	}
	if i > 1 {
		panic("item has more than one designated type")
	}
	if ib.TilesetSource == "" {
		panic("item has no tileset source")
	}
	if ib.MaxDurability != 0 && ib.Groupable {
		panic("items with durability cannot be groupable")
	}
}

func (ib ItemBase) GetID() string {
	return ib.ID
}
func (ib ItemBase) GetName() string {
	return ib.Name
}
func (ib ItemBase) GetDescription() string {
	return ib.Description
}
func (ib ItemBase) GetValue() int {
	return ib.Value
}
func (ib ItemBase) GetWeight() float64 {
	return ib.Weight
}
func (ib ItemBase) GetMaxDurability() int {
	return ib.MaxDurability
}
func (ib ItemBase) GetTileImg() *ebiten.Image {
	return ib.TileImg
}
func (ib ItemBase) IsGroupable() bool {
	return ib.Groupable
}
func (ib ItemBase) IsWeapon() bool {
	return ib.Weapon
}
func (ib ItemBase) IsArmor() bool {
	return ib.Armor
}
func (ib ItemBase) IsAccessory() bool {
	return ib.Accessory
}
func (ib ItemBase) IsAmmunition() bool {
	return ib.Ammunition
}
func (ib ItemBase) IsConsumable() bool {
	return ib.Consumable
}
func (ib ItemBase) IsMiscItem() bool {
	return ib.MiscItem
}
func (ib ItemBase) IsEquipable() bool {
	return ib.Armor || ib.Weapon || ib.Accessory || ib.Ammunition
}

func (ib *ItemBase) Load() {
	if ib.TilesetSource == "" {
		panic("no tileset source defined for item")
	}
	tileset, err := tiled.LoadTileset(ib.TilesetSource)
	if err != nil {
		logz.Panicf("error while loading tileset for item: %s", err)
	}
	img, err := tileset.GetTileImage(ib.TileID)
	if err != nil {
		logz.Panicf("error while getting item tile image: %s", err)
	}

	ib.TileImg = img

	ib.init = true
}

// just for the compiler to automatically confirm each item type implements ItemDef interface
func CompilerCheck() {
	_ = append([]ItemDef{}, &WeaponDef{}, &PotionDef{}, &ArmorDef{})
}

type ItemInstance struct {
	DefID      string // ID of the ItemDef that defines this item
	Durability int    // the current condition of this item
}

// ItemDef for weapons
type WeaponDef struct {
	ItemBase
	Damage        int     // damage per attack
	HitsPerSecond float64 // speed of attacks, in terms of number of attacks possible per second
}

// ItemDef for armor
type ArmorDef struct {
	ItemBase
	Protection int // amount of protection this piece of armor gives
}

// ItemDef for potions
type PotionDef struct {
	ItemBase
	EffectDuration time.Duration
	// TODO: add an Effect concept, which will encompass potion effects and enchantments on weapons or items
	// for now, just going to make a "heal amount" value.
	HealAmount int // how much health will be healed per second
}
