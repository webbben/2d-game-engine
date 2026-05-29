// Package item defines the item concept, basic item types, and an interface for flexibly creating new item types.
package item

import (
	"time"

	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/logz"
)

type WeaponDef struct {
	// ItemBase
	Damage int // damage per attack

	FxPartDef *defs.SelectedPartDef // only defined for weapon items
}

// GetWeaponParts gets the two bodyPartDefs for a weapon: the actual weapon, and the fx.
// Panics if given item is not a weaponDef, or if either part is not found.
func GetWeaponParts(i defs.ItemDef) (weaponPart defs.SelectedPartDef, fxPart defs.SelectedPartDef) {
	part := i.BodyPartDef
	if part == nil {
		logz.Panicln("GetWeaponParts", "weapon part is nil:", i.ID)
	}
	fx := i.FxPartDef
	if fx == nil {
		logz.Panicln("GetWeaponParts", "fx part is nil:", i.ID)
	}

	if part.None {
		logz.Panicln("GetWeaponParts", "weapon part is none")
	}
	if fx.None {
		logz.Panicln("GetWeaponParts", "fx part is none")
	}

	return *part, *fx
}

type PotionDef struct {
	EffectDuration time.Duration
	// TODO: add an Effect concept, which will encompass potion effects and enchantments on weapons or items
	// for now, just going to make a "heal amount" value.
	HealAmount int // how much health will be healed per second
}

func CountMoney(inv state.StandardInventory, dataman *datamanager.DataManager) int {
	sum := 0
	for _, coinItem := range inv.CoinPurse {
		if coinItem == nil {
			continue
		}
		itemDef := dataman.GetItemDef(coinItem.DefID)
		if itemDef.Type == defs.TypeCurrency {
			sum += itemDef.Value * coinItem.Quantity
		}
	}

	// also check for coins not in coin purse
	for _, coinItem := range inv.InventoryItems {
		if coinItem == nil {
			continue
		}
		itemDef := dataman.GetItemDef(coinItem.DefID)
		if itemDef.Type == defs.TypeCurrency {
			sum += itemDef.Value * coinItem.Quantity
		}
	}

	return sum
}

func ConvertInitialItemStateDefs(initialItemDefs []*defs.ItemInitialStateDef) []*state.ItemState {
	inv := []*state.ItemState{}
	for _, invItem := range initialItemDefs {
		if invItem == nil {
			inv = append(inv, nil)
			continue
		}
		inv = append(inv, &state.ItemState{
			DefID:      invItem.DefID,
			Durability: invItem.Durability,
			Quantity:   invItem.Quantity,
		})
	}
	return inv
}
