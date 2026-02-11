package trade

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
)

func NewShopKeeperState(def defs.ShopkeeperDef) state.ShopkeeperState {
	def.Validate()

	sk := state.ShopkeeperState{
		ShopID: def.ID,
		Gold:   def.BaseGold,
	}

	for _, inv := range def.BaseInventory {
		sk.ShopInventory = append(sk.ShopInventory, &defs.InventoryItem{
			Instance: inv.Instance,
			Def:      inv.Def,
			Quantity: inv.Quantity,
		})
	}

	return sk
}
