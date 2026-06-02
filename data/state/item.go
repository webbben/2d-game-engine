package state

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

// ItemState represents the state of an individual item, or group of same item ID that have been grouped.
// The state simply tracks the things that can change overtime, like durability for example.
type ItemState struct {
	DefID      defs.ItemID
	Durability float64
	Quantity   int // if groupable, then this represents how many of the item have been clustered up (e.g. in an inventory slot)
}

func (is ItemState) String() string {
	return fmt.Sprintf("{DefID: %s, Quant: %v, Dur: %v}", is.DefID, is.Quantity, is.Durability)
}

func InventoryToString(inv []*ItemState) string {
	s := ""
	for _, invItem := range inv {
		if invItem == nil {
			continue
		}
		s += invItem.String() + ", "
	}
	return s
}

// StandardInventory is the state of the entire inventory of a character, including coin purse
// and equipped items
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

func (inv StandardInventory) Validate() {
	for _, i := range inv.CoinPurse {
		if i != nil {
			i.Validate()
		}
	}
	for _, i := range inv.InventoryItems {
		if i != nil {
			i.Validate()
		}
	}

	allEquipment := []*ItemState{}
	allEquipment = append(allEquipment, inv.EquipedBodywear)
	allEquipment = append(allEquipment, inv.EquipedHeadwear)
	allEquipment = append(allEquipment, inv.EquipedFootwear)
	allEquipment = append(allEquipment, inv.EquipedAuxiliary)
	allEquipment = append(allEquipment, inv.EquipedWeapon)
	allEquipment = append(allEquipment, inv.EquipedRing1)
	allEquipment = append(allEquipment, inv.EquipedRing2)
	allEquipment = append(allEquipment, inv.EquipedAmulet)
	allEquipment = append(allEquipment, inv.EquipedAmmo)

	for _, i := range allEquipment {
		if i != nil {
			i.Validate()
		}
	}
}

func (is ItemState) Validate() {
	if is.DefID == "" {
		panic("item state has no def ID")
	}
	if is.Quantity <= 0 {
		logz.Panicln(string(is.DefID), "item state quantity is less than or equal to 0. all items must have a quantity of at least 1.", is.Quantity)
	}
	if is.Durability < 0 {
		logz.Panicln(string(is.DefID), "item state has durability < 0")
	}
}

func (inv *StandardInventory) SetInventoryItems(invItems []*ItemState) {
	inv.InventoryItems = make([]*ItemState, 0)

	for _, newItem := range invItems {
		if newItem == nil {
			inv.InventoryItems = append(inv.InventoryItems, nil)
			continue
		}
		newItem.Validate()

		dereffed := *newItem
		inv.InventoryItems = append(inv.InventoryItems, &dereffed)
	}

	inv.Validate()
}

func (inv *StandardInventory) SetCoinPurseItems(invItems []*ItemState) {
	inv.CoinPurse = make([]*ItemState, 0)

	for _, newItem := range invItems {
		if newItem == nil {
			inv.CoinPurse = append(inv.CoinPurse, nil)
			continue
		}

		newItem.Validate()
		dereffed := *newItem
		inv.CoinPurse = append(inv.CoinPurse, &dereffed)
	}
}
