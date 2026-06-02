package defs

import (
	"github.com/webbben/2d-game-engine/logz"
)

type (
	ItemType string
	ItemID   string
)

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
	TypeBook       ItemType = "BOOK"
)

type ItemDef struct {
	ID            ItemID
	Name          string
	Description   string
	Type          ItemType
	Value         int
	Weight        float64
	MaxDurability float64
	Groupable     bool

	TileImgTilesetSrc string // tileset where tile image is found
	TileImgIndex      int    // index of tile image in tileset

	// wearable item properties
	// Note: SelectedPartDef has ID, but purposely not using a central datamanager since only body parts (not items) are stored there.
	// I guess this works out since this is an item def, and so it serves as a central place to define something.

	BodyPartDef *SelectedPartDef // made it a pointer so it can be nil-able
	LegsPartDef *SelectedPartDef // bodywear has a legs component that moves separately from the body component

	// Item-type specific fields

	// Key

	LockIDs []string // the lockIDs that this key can unlock

	// Weapon

	Damage    int
	FxPartDef *SelectedPartDef

	// Armor (body/head/footwear, shield auxes, etc)

	Protection int // amount of protection this piece of armor gives

	// Book

	BookID BookID // if this is a book, set this to the book ID
}

func (id ItemDef) Validate() {
	if id.Name == "" {
		panic("item has no name")
	}
	if id.ID == "" {
		logz.Panicf("[%s] item has no ID", id.Name)
	}
	if id.Value < 0 {
		logz.Panic("value is less than 0" + "(" + string(id.ID) + ")")
	}
	if id.Description == "" {
		logz.Panic("item has no description" + "(" + string(id.ID) + ")")
	}
	if id.TileImgTilesetSrc == "" {
		logz.Panic("item has no tileset source for tile image" + "(" + string(id.ID) + ")")
	}
	if id.MaxDurability != 0 && id.Groupable {
		logz.Panic("items with durability cannot be groupable" + "(" + string(id.ID) + ")")
	}
	if id.Type == "" {
		logz.Panic("item has no type" + "(" + string(id.ID) + ")")
	}
	if id.Type == TypeBodywear || id.Type == TypeHeadwear || id.Type == TypeFootwear || id.Type == TypeAuxiliary || id.Type == TypeWeapon {
		if id.BodyPartDef == nil {
			logz.Panic("item is a visible equipable item, but no bodyPartDef is set" + "(" + string(id.ID) + ")")
		}
	} else if id.BodyPartDef != nil {
		logz.Panic("item is not a visible equipable item, but it has a defined bodyPartDef" + "(" + string(id.ID) + ")")
	}
	if id.Type == TypeBodywear {
		if id.BodyPartDef == nil {
			logz.Panic("bodywear must have a body part def" + "(" + string(id.ID) + ")")
		}
		if id.LegsPartDef == nil {
			logz.Panic("bodywear must have a legs part" + "(" + string(id.ID) + ")")
		}
	} else if id.Type == TypeHeadwear {
		if id.BodyPartDef == nil {
			logz.Panic("headwear must have a body part def" + "(" + string(id.ID) + ")")
		}
		if id.LegsPartDef != nil {
			logz.Panic("headwear must NOT have a legs component. that is only for bodywear." + "(" + string(id.ID) + ")")
		}
	}
}

func (def ItemDef) IsEquipable() bool {
	switch def.Type {
	case TypeBodywear, TypeHeadwear, TypeFootwear, TypeWeapon, TypeAmulet, TypeRing, TypeAmmunition, TypeAuxiliary:
		return true
	default:
		return false
	}
}

// ItemInitialStateDef is for defining an initial state for an item, for places such as chests or NPC inventories.
type ItemInitialStateDef struct {
	DefID      ItemID
	Quantity   int
	Durability float64
}

// InitialStandardInventoryDef is for defining the initial state for a character's inventory; what their inventory is when they first spawn
// into the game on starting a new game.
type InitialStandardInventoryDef struct {
	CoinPurse      []*ItemInitialStateDef
	InventoryItems []*ItemInitialStateDef

	EquipedHeadwear  *ItemInitialStateDef
	EquipedBodywear  *ItemInitialStateDef
	EquipedFootwear  *ItemInitialStateDef
	EquipedAmulet    *ItemInitialStateDef
	EquipedRing1     *ItemInitialStateDef
	EquipedRing2     *ItemInitialStateDef
	EquipedAmmo      *ItemInitialStateDef
	EquipedAuxiliary *ItemInitialStateDef
	EquipedWeapon    *ItemInitialStateDef
}
