package datamanager

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
)

func (dataman *DataManager) NewItemState(itemID defs.ItemID, quantity int) *state.ItemState {
	if quantity <= 0 {
		panic("quantity <= 0")
	}
	itemDef := dataman.GetItemDef(itemID)
	is := state.ItemState{
		DefID:      itemID,
		Durability: itemDef.MaxDurability,
		Quantity:   quantity,
	}
	return &is
}
