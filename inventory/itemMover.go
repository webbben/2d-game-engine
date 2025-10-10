package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

type ItemMover struct {
	carryItem     *InventoryItem // the item that is currently being carried
	dropableSlots []*ItemSlot    // item slots that an item can be dropped into
	itemImg       *ebiten.Image  // image of the item as its being carried
}

func NewItemMover(itemSlots []*ItemSlot) ItemMover {
	return ItemMover{
		dropableSlots: itemSlots,
	}
}

func (im *ItemMover) pickupItem(itemToCarry *InventoryItem) bool {
	if im.carryItem != nil {
		// already carrying an item
		return false
	}

	im.carryItem = itemToCarry
	im.itemImg = rendering.ScaleImage(im.carryItem.Def.GetTileImg(), config.UIScale)
	return true
}

func (im *ItemMover) Draw(om *overlay.OverlayManager) {
	if im.carryItem == nil {
		return
	}
	if im.itemImg == nil {
		panic("item image hasn't been generated yet, but item seems to have been picked up")
	}
	mouseX, mouseY := ebiten.CursorPosition()

	om.AddOverlay(im.itemImg, float64(mouseX), float64(mouseY))
}

func (im *ItemMover) Update() {
	if im.carryItem == nil {
		// check for clicks on item slots - for picking up items
		for _, slot := range im.dropableSlots {
			if slot.Item == nil {
				continue
			}
			if slot.mouseBehavior.LeftClick.ClickReleased {
				// the slot has been clicked!
				if im.pickupItem(slot.Item) {
					logz.Println(slot.Item.Def.GetName(), "picked up item")
					// item was successfully picked up
					// clear item slot of its item
					slot.Clear()
					break
				}
			}
		}
	} else {
		// item is already being carried; check for item placement
		if im.checkItemPlacement() {
			logz.Println(im.carryItem.Def.GetName(), "placed item")
			// item was placed!
			im.carryItem = nil
		}
	}
}

func (im *ItemMover) checkItemPlacement() bool {
	if im.carryItem == nil {
		panic("tried to place item but no carryItem exists")
	}
	mouseX, mouseY := ebiten.CursorPosition()

	for _, slot := range im.dropableSlots {
		w, h := slot.Dimensions()
		slotRect := model.Rect{X: float64(slot.x), Y: float64(slot.y), W: float64(w), H: float64(h)}
		if !slotRect.Within(mouseX, mouseY) {
			continue
		}
		if !slot.mouseBehavior.LeftClick.ClickReleased {
			continue
		}

		// check if the slot is empty
		if slot.Item == nil {
			// slot is empty, we can put this item here
			slot.SetContent(&im.carryItem.Instance, im.carryItem.Def, im.carryItem.Quantity)
			return true
		}

		// check if slot has the same item type (and is groupable)
		if slot.Item.Def.GetID() == im.carryItem.Instance.DefID {
			if im.carryItem.Def.IsGroupable() && slot.Item.Def.IsGroupable() {
				// found a matching groupable item
				slot.SetContent(&im.carryItem.Instance, im.carryItem.Def, im.carryItem.Quantity+slot.Item.Quantity)
				return true
			}
		}
	}

	return false
}
