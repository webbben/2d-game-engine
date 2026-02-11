package state

import "github.com/webbben/2d-game-engine/data/defs"

type ShopkeeperState struct {
	ShopID        defs.ShopID
	ShopInventory []*defs.InventoryItem
	Gold          int // current amount of gold this shopkeeper has
}

func (sk *ShopkeeperState) SetShopInventory(newItems []*defs.InventoryItem) {
	sk.ShopInventory = make([]*defs.InventoryItem, 0)
	for _, newItem := range newItems {
		if newItem == nil {
			sk.ShopInventory = append(sk.ShopInventory, nil)
			continue
		}
		sk.ShopInventory = append(sk.ShopInventory, &defs.InventoryItem{
			Instance: newItem.Instance,
			Def:      newItem.Def,
			Quantity: newItem.Quantity,
		})
	}
}
