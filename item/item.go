// Package item defines the item concept, basic item types, and an interface for flexibly creating new item types.
package item

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

const (
	TypeWeapon     defs.ItemType = "WEAPON"
	TypeBodywear   defs.ItemType = "BODYWEAR"
	TypeHeadwear   defs.ItemType = "HEADWEAR"
	TypeFootwear   defs.ItemType = "FOOTWEAR"
	TypeAmulet     defs.ItemType = "AMULET"
	TypeRing       defs.ItemType = "RING"
	TypeAmmunition defs.ItemType = "AMMUNITION"
	TypeAuxiliary  defs.ItemType = "AUXILIARY"
	TypeConsumable defs.ItemType = "CONSUMABLE"
	TypeMisc       defs.ItemType = "MISC"
	TypeCurrency   defs.ItemType = "CURRENCY"
)

// ItemBase includes the basic functions required for an item to implement the ItemDef interface.
// embed into a struct to make it effectively an item type.
type ItemBase struct {
	init          bool // flag to indicate if item has been loaded yet
	ID            defs.ItemID
	Name          string
	Description   string
	Type          defs.ItemType
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

	BodyPartDef *defs.SelectedPartDef // made it a pointer so it can be nil-able
	LegsPartDef *defs.SelectedPartDef // bodywear has a legs component that moves separately from the body component
}

// panic is a helper for quickly throwing a panic based on this item, and giving helpful contextual info at the same time.
func (ib ItemBase) panic(s string) {
	logz.Panicf("[%s/%s] %s", ib.Name, ib.ID, s)
}

type ItemBaseParams struct {
	ID                   defs.ItemID
	Name, Description    string
	Type                 defs.ItemType
	Weight               float64
	Value, MaxDurability int
	TileImgTilesetSrc    string
	TileImgIndex         int
	Groupable            bool
	BodyPartDef          *defs.SelectedPartDef
	LegsPartDef          *defs.SelectedPartDef
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

func (ib ItemBase) GetID() defs.ItemID {
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

func (ib ItemBase) GetItemType() defs.ItemType {
	return ib.Type
}

// GetBodyPartDef gets the body part def for equiping this item visibly on the body.
func (ib ItemBase) GetBodyPartDef() *defs.SelectedPartDef {
	return ib.BodyPartDef
}

// GetLegsPartDef gets the legs part def for equiping this (bodywear) item. Should only exist for bodywear items.
func (ib ItemBase) GetLegsPartDef() *defs.SelectedPartDef {
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
	img, err := tileset.GetTileImage(ib.TileImgIndex, true)
	if err != nil {
		logz.Panicf("error while getting item tile image: %s", err)
	}
	ib.TileImg = img

	ib.init = true
}

func CompilerCheck() {
	_ = append([]defs.ItemDef{}, &WeaponDef{}, &PotionDef{}, &ArmorDef{})
}

type WeaponDef struct {
	ItemBase
	Damage        int     // damage per attack
	HitsPerSecond float64 // speed of attacks, in terms of number of attacks possible per second

	FxPartDef *defs.SelectedPartDef // only defined for weapon items
}

// GetWeaponParts gets the two bodyPartDefs for a weapon: the actual weapon, and the fx.
// Panics if given item is not a weaponDef, or if either part is not found.
func GetWeaponParts(i defs.ItemDef) (weaponPart defs.SelectedPartDef, fxPart defs.SelectedPartDef) {
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

	if part.None {
		logz.Panicln("GetWeaponParts", "weapon part is none")
	}
	if fx.None {
		logz.Panicln("GetWeaponParts", "fx part is none")
	}

	return *part, *fx
}

type ArmorDef struct {
	ItemBase
	Protection int // amount of protection this piece of armor gives
}

// IsArmor determines if the given ItemDef is an instance of ArmorDef (i.e. is assertable to ArmorDef)
func IsArmor(i defs.ItemDef) bool {
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
