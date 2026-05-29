package item

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/ui/text"
)

type ItemIcon struct {
	tileImage *ebiten.Image
}

func NewItemIcon(itemDef defs.ItemDef) ItemIcon {
	tileImage := tiled.GetTileImage(itemDef.TileImgTilesetSrc, itemDef.TileImgIndex, true)

	return ItemIcon{
		tileImage: tileImage,
	}
}

func (ic ItemIcon) Draw(screen *ebiten.Image, x, y float64, quantity int) {
	rendering.DrawImage(screen, ic.tileImage, x, y, config.UIScale)
	tileSize := int(config.TileSize * config.UIScale)

	if quantity > 1 {
		qS := fmt.Sprintf("%v", quantity)
		qDx, _, _ := text.GetStringSize(qS, config.DefaultFont)
		qX := int(x) + tileSize - qDx - 3
		qY := int(y) + tileSize - 5
		text.DrawOutlinedText(screen, qS, config.DefaultFont, qX, qY, color.Black, color.White, 0, 0)
	}
}

func RemoveItemFromStandardInventory(inv *state.StandardInventory, itemToRemove state.ItemState, dataman *datamanager.DataManager) (bool, state.ItemState) {
	itemToRemove.Validate()

	itemDef := dataman.GetItemDef(itemToRemove.DefID)

	if (itemDef.Type == defs.TypeCurrency) && len(inv.CoinPurse) > 0 {
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

func AddItemToStandardInventory(inv *state.StandardInventory, invItem state.ItemState, dataman *datamanager.DataManager) (bool, state.ItemState) {
	invItem.Validate()

	itemDef := dataman.GetItemDef(invItem.DefID)

	// if currency, first try to place it in the coin purse
	if len(inv.CoinPurse) > 0 {
		if itemDef.Type == defs.TypeCurrency {
			success, remaining := AddItemToInventory(invItem, inv.CoinPurse, dataman)
			if success {
				return true, state.ItemState{}
			}
			invItem = remaining
		}
	}

	success, remaining := AddItemToInventory(invItem, inv.InventoryItems, dataman)

	return success, remaining
}

func RemoveItemFromInventory(itemToRemove state.ItemState, removeFrom []*state.ItemState) (bool, state.ItemState) {
	for i, invItem := range removeFrom {
		if invItem == nil {
			continue
		}
		if invItem.DefID == itemToRemove.DefID {
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
func AddItemToInventory(invItem state.ItemState, addTo []*state.ItemState, dataman *datamanager.DataManager) (bool, state.ItemState) {
	itemDef := dataman.GetItemDef(invItem.DefID)

	// if item is groupable, try to find a matching item already in the inventory
	if itemDef.Groupable {
		for _, otherItem := range addTo {
			if otherItem == nil {
				continue
			}
			if otherItem.DefID == invItem.DefID {
				otherItem.Quantity += invItem.Quantity
				return true, state.ItemState{}
			}
		}
	}

	// not able to group together with existing item; place in empty slot if one is available
	for i, otherItem := range addTo {
		if otherItem == nil {
			addTo[i] = &invItem
			return true, state.ItemState{}
		}
	}

	return false, invItem
}
