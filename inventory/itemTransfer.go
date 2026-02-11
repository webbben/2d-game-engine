package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
)

// ItemTransfer is for moving items from one inventory to another.
// doesn't involve drag and drop, but just automatic single-click transfers.
// mainly used for trade transactions or putting items into containers.
type ItemTransfer struct {
	playerInventory []*ItemSlot
	otherInventory  []*ItemSlot
}

func NewItemTransfer(playerInv []*ItemSlot, otherInv []*ItemSlot) ItemTransfer {
	itemTransfer := ItemTransfer{
		playerInventory: playerInv,
		otherInventory:  otherInv,
	}

	return itemTransfer
}

type ItemTransferResult struct {
	TransferedItem          defs.InventoryItem // item that was transfered
	ToPlayerInv             bool               // if true, was transfered to player inventory. otherwise, was transfered to other inventory
	TransferedTo            *ItemSlot          // slot item was transfered to
	Success                 bool               // if true, item was successfully transfered
	TransferAttemptOccurred bool               // if true, a transfer attempt occurred
}

func (it *ItemTransfer) Update() ItemTransferResult {
	// check if placer clicks an item in their own inventory
	// this means they are trying to move an item to the other inventory
	for i, itemSlot := range it.playerInventory {
		if itemSlot.Item == nil {
			continue
		}

		if itemSlot.mouseBehavior.LeftClick.ClickReleased {
			if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
				// shift + left click = transfer all
				result := transferItem(it.playerInventory, i, itemSlot.Item.Quantity, it.otherInventory)
				result.ToPlayerInv = false
				return result
			} else {
				// left click = transfer 1
				result := transferItem(it.playerInventory, i, 1, it.otherInventory)
				result.ToPlayerInv = false
				return result
			}
		}
		// right click = transfer half
		if itemSlot.mouseBehavior.RightClick.ClickReleased {
			q := max(itemSlot.Item.Quantity/2, 1)
			result := transferItem(it.playerInventory, i, q, it.otherInventory)
			result.ToPlayerInv = false
			return result
		}
	}

	// do the same check for the other inventory, to see if the player is trying to take items
	for i, itemSlot := range it.otherInventory {
		if itemSlot.Item == nil {
			continue
		}

		if itemSlot.mouseBehavior.LeftClick.ClickReleased {
			if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
				// shift + left click = transfer all
				result := transferItem(it.otherInventory, i, itemSlot.Item.Quantity, it.playerInventory)
				result.ToPlayerInv = true
				return result
			} else {
				// left click = transfer 1
				result := transferItem(it.otherInventory, i, 1, it.playerInventory)
				result.ToPlayerInv = true
				return result
			}
		}
		// right click = transfer half
		if itemSlot.mouseBehavior.RightClick.ClickReleased {
			q := max(itemSlot.Item.Quantity/2, 1)
			result := transferItem(it.otherInventory, i, q, it.playerInventory)
			result.ToPlayerInv = true
			return result
		}
	}

	return ItemTransferResult{
		TransferAttemptOccurred: false,
	}
}

func transferItem(from []*ItemSlot, fromIndex int, quantity int, to []*ItemSlot) ItemTransferResult {
	if quantity == 0 {
		panic("no quantity set for item transfer")
	}
	itemToMove := defs.InventoryItem{
		Instance: from[fromIndex].Item.Instance,
		Def:      from[fromIndex].Item.Def,
		Quantity: quantity,
	}

	// if groupable, find a matching slot
	if itemToMove.Def.IsGroupable() {
		for _, slot := range to {
			if slot.Item == nil {
				continue
			}
			if slot.Item.Instance.DefID == itemToMove.Instance.DefID {
				slot.Item.Quantity += quantity
				from[fromIndex].Item.Quantity -= quantity
				if from[fromIndex].Item.Quantity == 0 {
					from[fromIndex].Clear()
				}
				return ItemTransferResult{
					TransferedItem:          itemToMove,
					TransferedTo:            slot,
					Success:                 true,
					TransferAttemptOccurred: true,
				}
			}
		}
	}
	// otherwise, find an empty slot
	for _, slot := range to {
		if slot.Item == nil {
			slot.SetContent(&itemToMove.Instance, itemToMove.Def, itemToMove.Quantity)
			from[fromIndex].Item.Quantity -= quantity
			if from[fromIndex].Item.Quantity == 0 {
				from[fromIndex].Clear()
			}
			return ItemTransferResult{
				TransferedItem:          itemToMove,
				TransferedTo:            slot,
				Success:                 true,
				TransferAttemptOccurred: true,
			}
		}
	}

	// failed to add
	return ItemTransferResult{
		Success:                 false,
		TransferAttemptOccurred: true,
	}
}
