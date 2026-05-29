package state

import "github.com/webbben/2d-game-engine/data/defs"

type ShopkeeperState struct {
	ShopID        defs.ShopID
	ShopInventory []*ItemState
	Gold          int // current amount of gold this shopkeeper has
}

func (sk *ShopkeeperState) SetShopInventory(newItems []*ItemState) {
	sk.ShopInventory = make([]*ItemState, 0)
	for _, newItem := range newItems {
		if newItem == nil {
			sk.ShopInventory = append(sk.ShopInventory, nil)
			continue
		}
		dereffed := *newItem
		sk.ShopInventory = append(sk.ShopInventory, &dereffed)
	}
}

func NewShopKeeperState(def defs.ShopkeeperDef) ShopkeeperState {
	def.Validate()

	sk := ShopkeeperState{
		ShopID: def.ID,
		Gold:   def.BaseGold,
	}

	for _, inv := range def.BaseInventory {
		itemState := ItemState{
			DefID:      inv.DefID,
			Durability: inv.Durability,
			Quantity:   inv.Quantity,
		}
		sk.ShopInventory = append(sk.ShopInventory, &itemState)
	}

	return sk
}
