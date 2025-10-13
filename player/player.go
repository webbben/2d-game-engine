package player

import (
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/item"
)

type Player struct {
	Entity *entity.Entity // the entity that represents this player in a map

	InventoryItems []*item.InventoryItem // the regular items (not equiped) that are in the player's inventory
	CoinPurse      []*item.InventoryItem // items in the player's coin purse. note that coins can also be put in regular inventory slots.

	// equiped items
	EquipedHeadwear  *item.InventoryItem
	EquipedBodywear  *item.InventoryItem
	EquipedFootwear  *item.InventoryItem
	EquipedAmulet    *item.InventoryItem
	EquipedRing1     *item.InventoryItem
	EquipedRing2     *item.InventoryItem
	EquipedAmmo      *item.InventoryItem
	EquipedAuxiliary *item.InventoryItem

	defMgr *definitions.DefinitionManager
}

// needed for sorting renderables
func (p Player) Y() float64 {
	return p.Entity.Y
}

func NewPlayer(defMgr *definitions.DefinitionManager) Player {
	return Player{
		InventoryItems: make([]*item.InventoryItem, 18),
		CoinPurse:      make([]*item.InventoryItem, 6),
		defMgr:         defMgr,
	}
}

func (p *Player) AddItemToInventory(invItem item.InventoryItem) bool {
	// if currency, first try to place it in the coin purse
	if invItem.Def.IsCurrencyItem() {
		if invItem.Def.IsGroupable() {
			for _, coinItem := range p.CoinPurse {
				if coinItem != nil {
					if coinItem.Instance.DefID == invItem.Instance.DefID {
						coinItem.Quantity += invItem.Quantity
						return true
					}
				}
			}
		}
		// if no match, also check for an empty slot in the coin purse
		for i, coinItem := range p.CoinPurse {
			if coinItem == nil {
				p.CoinPurse[i] = &invItem
				return true
			}
		}
	}

	// if item is groupable, try to find a matching item already in the inventory
	if invItem.Def.IsGroupable() {
		if p.EquipedAmmo != nil && p.EquipedAmmo.Instance.DefID == invItem.Instance.DefID {
			p.EquipedAmmo.Quantity += invItem.Quantity
			return true
		}

		for _, otherItem := range p.InventoryItems {
			if otherItem == nil {
				continue
			}
			if otherItem.Instance.DefID == invItem.Instance.DefID {
				otherItem.Quantity += invItem.Quantity
				return true
			}
		}
	}

	// not able to group together with existing item; place in empty slot if one is available
	for i, otherItem := range p.InventoryItems {
		if otherItem == nil {
			p.InventoryItems[i] = &invItem
			return true
		}
	}

	return false
}
