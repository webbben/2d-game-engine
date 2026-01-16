// Package item defines the item concept, basic item types, and an interface for flexibly creating new item types.
package item

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/entity/body"
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

	GetTileImg() *ebiten.Image        // gets the "tile image", i.e. the image used in places like the inventory slots
	GetEquipedTiles() []*ebiten.Image // gets the tiles for the equiped view of this item (if equipable)

	GetBodyPartDef() *body.SelectedPartDef // gets the body part set for wearing when this item is equiped, if it exists. Only exists for visible equipable items.
	GetLegsPartDef() *body.SelectedPartDef // gets a legs body part set, if it exists (should only exist for bodywear items)

	// if true, item can be grouped together with other same items in inventories
	// this tends to be true for items that don't have durability, like potions or arrows, but not for weapons and armor.
	IsGroupable() bool
	// if true, this item is treated as an equipable weapon that is held
	IsWeapon() bool
	// if true, this item is treated as an equipable piece of armor that is worn
	IsHeadwear() bool
	// if true, this item is equipable on the body
	IsBodywear() bool
	// if true, this item is equipable on the feet
	IsFootwear() bool
	// if true, this item is an amulet
	IsAmulet() bool
	// if true, this item is a ring
	IsRing() bool
	// if true, this item is treated as an equipable ammunition type (arrows, etc)
	IsAmmunition() bool
	// if true, this item is an "auxiliary" item which is held in the left hand (torches, shields, etc)
	IsAuxiliary() bool
	// if true, this item is treated as an item that can be consumed (food, potions, etc)
	IsConsumable() bool
	// if true, this item has no specific use or utility; it just exists in your inventory and may have value or weight
	IsMiscItem() bool
	// if true, this item is a piece of currency (gold, coins, etc) used in transactions
	IsCurrencyItem() bool
	// determines if this item can be equiped
	IsEquipable() bool

	GetItemType() ItemType // returns the item type; to simply confirm a single item type, use the specific Is<itemType> functions instead.

	Load() // load things like images

	Validate() // checks if item def is properly defined
}

type ItemType string

const (
	TypeWeapon     ItemType = "WEAPON"
	TypeBodywear   ItemType = "BODYWEAR"
	TypeHeadwear   ItemType = "HEADWEAR"
	TypeFootwear   ItemType = "FOOTWEAR"
	TypeAmulet     ItemType = "AMULET"
	TypeRing       ItemType = "RING"
	TypeAmmunition ItemType = "AMMUNITION"
	TypeAuxiliary  ItemType = "AUXILIARY" // TODO
	TypeConsumable ItemType = "CONSUMABLE"
	TypeMisc       ItemType = "MISC"
	TypeCurrency   ItemType = "CURRENCY"
)

// ItemBase includes the basic functions required for an item to implement the ItemDef interface.
// embed into a struct to make it effectively an item type.
type ItemBase struct {
	init          bool // flag to indicate if item has been loaded yet
	ID            string
	Name          string
	Description   string
	Type          ItemType
	Value         int
	Weight        float64
	MaxDurability int
	Groupable     bool

	TileImgTilesetSrc       string // tileset where tile image is found
	TileImgIndex            int    // index of tile image in tileset
	TileImg                 *ebiten.Image
	OriginIndexEquipedTiles int // index of origin where equiped tiles are in tileset
	EquipedTiles            []*ebiten.Image

	// wearable item properties

	BodyPartDef *body.SelectedPartDef // made it a pointer so it can be nil-able
	LegsPartDef *body.SelectedPartDef // bodywear has a legs component that moves separately from the body component
}

// panic is a helper for quickly throwing a panic based on this item, and giving helpful contextual info at the same time.
func (ib ItemBase) panic(s string) {
	logz.Panicf("[%s/%s] %s", ib.Name, ib.ID, s)
}

type ItemBaseParams struct {
	ID, Name, Description string
	Type                  ItemType
	Weight                float64
	Value, MaxDurability  int
	TileImgTilesetSrc     string
	TileImgIndex          int
	Groupable             bool
	BodyPartDef           *body.SelectedPartDef
	LegsPartDef           *body.SelectedPartDef
}

func NewItemBase(params ItemBaseParams) *ItemBase {
	ib := ItemBase{
		ID:                params.ID,
		Name:              params.Name,
		Description:       params.Description,
		Type:              params.Type,
		Value:             params.Value,
		Weight:            params.Weight,
		MaxDurability:     params.MaxDurability,
		TileImgTilesetSrc: params.TileImgTilesetSrc,
		TileImgIndex:      params.TileImgIndex,
		Groupable:         params.Groupable,
		BodyPartDef:       params.BodyPartDef,
		LegsPartDef:       params.LegsPartDef,
	}

	ib.Validate()
	return &ib
}

func (ib ItemBase) Validate() {
	if ib.Name == "" {
		panic("item has no name")
	}
	if ib.ID == "" {
		logz.Panicf("[%s] item has no ID", ib.Name)
	}
	if ib.Value < 0 {
		ib.panic("value is less than 0")
	}
	if ib.Description == "" {
		ib.panic("item has no description")
	}
	if ib.TileImgTilesetSrc == "" {
		ib.panic("item has no tileset source for tile image")
	}
	if ib.MaxDurability != 0 && ib.Groupable {
		ib.panic("items with durability cannot be groupable")
	}
	if ib.Type == "" {
		ib.panic("item has no type")
	}
	if ib.Type == TypeBodywear || ib.Type == TypeHeadwear || ib.Type == TypeFootwear || ib.Type == TypeAuxiliary || ib.Type == TypeWeapon {
		if ib.BodyPartDef == nil {
			ib.panic("item is a visible equipable item, but no bodyPartDef is set")
		}
	} else if ib.BodyPartDef != nil {
		ib.panic("item is not a visible equipable item, but it has a defined bodyPartDef")
	}
	if ib.IsBodywear() {
		if ib.BodyPartDef == nil {
			ib.panic("bodywear must have a body part def")
		}
		if ib.LegsPartDef == nil {
			ib.panic("bodywear must have a legs part")
		}
	} else if ib.IsHeadwear() {
		if ib.BodyPartDef == nil {
			ib.panic("headwear must have a body part def")
		}
		if ib.LegsPartDef != nil {
			ib.panic("headwear must NOT have a legs component. that is only for bodywear.")
		}
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

func (ib ItemBase) GetEquipedTiles() []*ebiten.Image {
	return ib.EquipedTiles
}

func (ib ItemBase) IsGroupable() bool {
	return ib.Groupable
}

func (ib ItemBase) IsWeapon() bool {
	return ib.Type == TypeWeapon
}

func (ib ItemBase) IsBodywear() bool {
	return ib.Type == TypeBodywear
}

func (ib ItemBase) IsHeadwear() bool {
	return ib.Type == TypeHeadwear
}

func (ib ItemBase) IsFootwear() bool {
	return ib.Type == TypeFootwear
}

func (ib ItemBase) IsAmulet() bool {
	return ib.Type == TypeAmulet
}

func (ib ItemBase) IsRing() bool {
	return ib.Type == TypeRing
}

func (ib ItemBase) IsAmmunition() bool {
	return ib.Type == TypeAmmunition
}

func (ib ItemBase) IsAuxiliary() bool {
	return ib.Type == TypeAuxiliary
}

func (ib ItemBase) IsConsumable() bool {
	return ib.Type == TypeConsumable
}

func (ib ItemBase) IsMiscItem() bool {
	return ib.Type == TypeMisc
}

func (ib ItemBase) IsCurrencyItem() bool {
	return ib.Type == TypeCurrency
}

func (ib ItemBase) IsEquipable() bool {
	switch ib.Type {
	case TypeBodywear, TypeHeadwear, TypeFootwear, TypeWeapon, TypeAmulet, TypeRing, TypeAmmunition, TypeAuxiliary:
		return true
	default:
		return false
	}
}

func (ib ItemBase) GetItemType() ItemType {
	return ib.Type
}

// GetBodyPartDef gets the body part def for equiping this item visibly on the body.
func (ib ItemBase) GetBodyPartDef() *body.SelectedPartDef {
	return ib.BodyPartDef
}

// GetLegsPartDef gets the legs part def for equiping this (bodywear) item. Should only exist for bodywear items.
func (ib ItemBase) GetLegsPartDef() *body.SelectedPartDef {
	return ib.LegsPartDef
}

func (ib *ItemBase) Load() {
	if ib.TileImgTilesetSrc == "" {
		panic("no tileset source defined for item tile image")
	}

	// load tile image
	tileset, err := tiled.LoadTileset(ib.TileImgTilesetSrc)
	if err != nil {
		logz.Panicf("error while loading tileset for item tile image: %s", err)
	}
	img, err := tileset.GetTileImage(ib.TileImgIndex)
	if err != nil {
		logz.Panicf("error while getting item tile image: %s", err)
	}
	ib.TileImg = img

	ib.init = true
}

func CompilerCheck() {
	_ = append([]ItemDef{}, &WeaponDef{}, &PotionDef{}, &ArmorDef{})
}

type ItemInstance struct {
	DefID      string // ID of the ItemDef that defines this item
	Durability int    // the current condition of this item
}

type WeaponDef struct {
	ItemBase
	Damage        int     // damage per attack
	HitsPerSecond float64 // speed of attacks, in terms of number of attacks possible per second

	FxPartDef *body.SelectedPartDef // only defined for weapon items
}

// GetWeaponParts gets the two bodyPartDefs for a weapon: the actual weapon, and the fx.
// Panics if given item is not a weaponDef, or if either part is not found.
func GetWeaponParts(i ItemDef) (weaponPart body.SelectedPartDef, fxPart body.SelectedPartDef) {
	part := i.GetBodyPartDef()
	if part == nil {
		logz.Panicln("GetWeaponParts", "weapon part is nil:", i.GetID())
	}
	weaponDef, ok := i.(*WeaponDef)
	if !ok {
		logz.Panicln("GetWeaponParts", "failed to assert as weapon:", i.GetID())
	}
	fx := weaponDef.FxPartDef
	if fx == nil {
		logz.Panicln("GetWeaponParts", "fx part is nil:", i.GetID())
	}

	return *part, *fx
}

type ArmorDef struct {
	ItemBase
	Protection int // amount of protection this piece of armor gives
}

// IsArmor determines if the given ItemDef is an instance of ArmorDef (i.e. is assertable to ArmorDef)
func IsArmor(i ItemDef) bool {
	_, ok := i.(*ArmorDef)
	return ok
}

type PotionDef struct {
	ItemBase
	EffectDuration time.Duration
	// TODO: add an Effect concept, which will encompass potion effects and enchantments on weapons or items
	// for now, just going to make a "heal amount" value.
	HealAmount int // how much health will be healed per second
}
