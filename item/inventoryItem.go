package item

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
)

func DrawInventoryItem(screen *ebiten.Image, invItem defs.InventoryItem, x, y float64) {
	tileSize := int(config.TileSize * config.UIScale)
	rendering.DrawImage(screen, invItem.Def.GetTileImg(), x, y, config.UIScale)

	if invItem.Quantity > 1 {
		qS := fmt.Sprintf("%v", invItem.Quantity)
		qDx, _, _ := text.GetStringSize(qS, config.DefaultFont)
		qX := int(x) + tileSize - qDx - 3
		qY := int(y) + tileSize - 5
		text.DrawOutlinedText(screen, fmt.Sprintf("%v", invItem.Quantity), config.DefaultFont, qX, qY, color.Black, color.White, 0, 0)
	}
}

func RemoveItemFromStandardInventory(inv *defs.StandardInventory, itemToRemove defs.InventoryItem) (bool, defs.InventoryItem) {
	itemToRemove.Validate()

	if itemToRemove.Def.IsCurrencyItem() && len(inv.CoinPurse) > 0 {
		// first try the coin purse
		success, remaining := RemoveItemFromInventory(itemToRemove, inv.CoinPurse)
		if success {
			if remaining.Quantity != 0 {
				logz.Panicln("RemoveItemFromStandardInventory", "item removal was supposedly successful, but remaining is not 0:", remaining.Quantity)
			}
			return true, remaining
		}
		if remaining.Quantity == 0 {
			panic("why is Quantity 0 if no success?")
		}
		// if there is still some left to remove, use the `remaining` value
		// e.g. if some coins are not in the coin purse for some reason
		itemToRemove.Quantity = remaining.Quantity
	}

	success, remaining := RemoveItemFromInventory(itemToRemove, inv.InventoryItems)
	if success {
		if remaining.Quantity != 0 {
			logz.Panicln("RemoveItemFromInventory", "item removal was supposedly successful, but remaining is not 0:", remaining.Quantity)
		}
	}
	// cd.Validate()
	return success, remaining
}

// LoadStandardInventoryItemDefs confirms all itemDefs are loaded for every inventory item in a standard inventory.
// The reason we need to do this is, if we are loading an inventory from a JSON file, the item defs won't be defined.
// The reason for that is, itemdefs are interfaces, which don't really work for writing to JSON files.
// So, whenever loading an inventory, ensure that all the defs are loaded in.
func LoadStandardInventoryItemDefs(inv *defs.StandardInventory, defMgr *definitions.DefinitionManager) {
	reloadItemDef := func(invItem *defs.InventoryItem) {
		if invItem == nil {
			return
		}
		if invItem.Def == nil {
			if invItem.Instance.DefID == "" {
				panic("tried to reload invItem def, but instance didn't have defID set")
			}
			invItem.Def = defMgr.GetItemDef(invItem.Instance.DefID)
		}
		invItem.Validate()
	}

	for _, coinItem := range inv.CoinPurse {
		reloadItemDef(coinItem)
	}
	for _, invItem := range inv.InventoryItems {
		reloadItemDef(invItem)
	}
	reloadItemDef(inv.Equipment.EquipedBodywear)
	reloadItemDef(inv.Equipment.EquipedHeadwear)
	reloadItemDef(inv.Equipment.EquipedFootwear)
	reloadItemDef(inv.Equipment.EquipedAuxiliary)
	reloadItemDef(inv.Equipment.EquipedWeapon)
	reloadItemDef(inv.Equipment.EquipedAmulet)
	reloadItemDef(inv.Equipment.EquipedRing1)
	reloadItemDef(inv.Equipment.EquipedRing2)
	reloadItemDef(inv.Equipment.EquipedAmmo)
}

func AddItemToStandardInventory(inv *defs.StandardInventory, invItem defs.InventoryItem) (bool, defs.InventoryItem) {
	invItem.Validate()

	// if currency, first try to place it in the coin purse
	if len(inv.CoinPurse) > 0 {
		if invItem.Def.IsCurrencyItem() {
			success, remaining := AddItemToInventory(invItem, inv.CoinPurse)
			if success {
				return true, defs.InventoryItem{}
			}
			invItem = remaining
		}
	}

	success, remaining := AddItemToInventory(invItem, inv.InventoryItems)

	return success, remaining
}

func RemoveItemFromInventory(itemToRemove defs.InventoryItem, removeFrom []*defs.InventoryItem) (bool, defs.InventoryItem) {
	for i, invItem := range removeFrom {
		if invItem == nil {
			continue
		}
		if invItem.Instance.DefID == itemToRemove.Instance.DefID {
			sub := min(itemToRemove.Quantity, invItem.Quantity)
			invItem.Quantity -= sub
			if invItem.Quantity == 0 {
				removeFrom[i] = nil
			}
			itemToRemove.Quantity -= sub
			if itemToRemove.Quantity == 0 {
				return true, itemToRemove
			}
		}
	}
	// item wasn't found, or not enough of it at least
	return false, itemToRemove
}

// AddItemToInventory attempts to add the given inventory item to an inventory.
// Returns true if successfully in placing the item; otherwise, returns the inventory item back (however much failed to be added)
func AddItemToInventory(invItem defs.InventoryItem, addTo []*defs.InventoryItem) (bool, defs.InventoryItem) {
	// if item is groupable, try to find a matching item already in the inventory
	if invItem.Def.IsGroupable() {
		for _, otherItem := range addTo {
			if otherItem == nil {
				continue
			}
			if otherItem.Instance.DefID == invItem.Instance.DefID {
				otherItem.Quantity += invItem.Quantity
				return true, defs.InventoryItem{}
			}
		}
	}

	// not able to group together with existing item; place in empty slot if one is available
	for i, otherItem := range addTo {
		if otherItem == nil {
			addTo[i] = &invItem
			return true, defs.InventoryItem{}
		}
	}

	return false, invItem
}
