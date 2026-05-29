package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/ui/overlay"
)

type ItemMover struct {
	carryItem     *state.ItemState // the item that is currently being carried
	carryItemDef  defs.ItemDef
	dropableSlots []*ItemSlot   // item slots that an item can be dropped into
	itemImg       *ebiten.Image // image of the item as its being carried
}

func NewItemMover(itemSlots []*ItemSlot) ItemMover {
	return ItemMover{
		dropableSlots: itemSlots,
	}
}

func (im *ItemMover) pickupItem(itemToCarry *state.ItemState, itemDef defs.ItemDef, amount int) bool {
	if im.carryItem != nil {
		// already carrying an item
		return false
	}

	im.carryItem = itemToCarry
	im.carryItemDef = itemDef
	im.MakeItemImage()

	return true
}

func (im *ItemMover) MakeItemImage() {
	if im.carryItem == nil {
		panic("no carry item to generate image for")
	}

	tileSize := int(config.TileSize * config.UIScale)

	itemIcon := item.NewItemIcon(im.carryItemDef)

	im.itemImg = ebiten.NewImage(tileSize, tileSize)
	itemIcon.Draw(im.itemImg, 0, 0, im.carryItem.Quantity)
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
				if im.pickupItem(slot.Item, slot.ItemDef, slot.Item.Quantity) {
					// item was successfully picked up
					// clear item slot of its item
					slot.Clear()
					break
				}
			} else if slot.mouseBehavior.RightClick.ClickReleased {
				// right click = pick up half
				amount := slot.Item.Quantity
				if amount > 1 {
					amount /= 2
				}
				remaining := slot.Item.Quantity - amount
				if im.pickupItem(slot.Item, slot.ItemDef, amount) {
					if remaining > 0 {
						newItemState := *slot.Item
						newItemState.Quantity = remaining
						slot.SetContent(&newItemState, slot.ItemDef)
					} else {
						slot.Clear()
					}
					break
				}
			}
		}
	} else {
		// item is already being carried; check for item placement
		im.handleItemPlacement()
	}
}

func (im *ItemMover) handleItemPlacement() {
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
		if slot.mouseBehavior.LeftClick.ClickReleased {
			// first, detect if this is a double click; if so, we should gather all of this item, rather than placing
			if slot.Item == nil && slot.mouseBehavior.LeftClick.DoubleClicked() {
				// gather all of this item
				im.gatherAllOfItem(im.carryItemDef)
				return
			}

			// placing all of the item
			if slot.CanTakeItemType(im.carryItemDef.Type) {
				// check if the slot is empty
				if slot.Item == nil {
					// slot is empty, we can put this item here
					slot.SetContent(im.carryItem, im.carryItemDef)
					im.carryItem = nil
					return
				}

				// check if slot has the same item type (and is groupable)
				if slot.ItemDef.ID == im.carryItem.DefID {
					if im.carryItemDef.Groupable && slot.ItemDef.Groupable {
						// found a matching groupable item
						newItemState := *im.carryItem
						newItemState.Quantity += slot.Item.Quantity
						slot.SetContent(&newItemState, im.carryItemDef)
						im.carryItem = nil
						return
					}
				}
			}
		} else if slot.mouseBehavior.RightClick.ClickReleased {
			// placing single item
			if slot.CanTakeItemType(im.carryItemDef.Type) {
				if slot.Item == nil {
					newItemState := *im.carryItem
					newItemState.Quantity = 1
					slot.SetContent(&newItemState, im.carryItemDef)
					im.carryItem.Quantity--
					if im.carryItem.Quantity == 0 {
						im.carryItem = nil
					} else {
						im.MakeItemImage()
					}
					return
				}
				// check if slot has the same item type (and is groupable)
				if slot.ItemDef.ID == im.carryItem.DefID {
					if im.carryItemDef.Groupable && slot.ItemDef.Groupable {
						// found a matching groupable item
						newItemState := *slot.Item
						newItemState.Quantity++
						slot.SetContent(&newItemState, im.carryItemDef)
						im.carryItem.Quantity--
						if im.carryItem.Quantity == 0 {
							im.carryItem = nil
						}
						return
					}
				}
			}
		}
	}
}

func (im *ItemMover) gatherAllOfItem(def defs.ItemDef) {
	if !def.Groupable {
		return
	}

	for _, slot := range im.dropableSlots {
		if slot.Item == nil {
			continue
		}
		if slot.ItemDef.ID == def.ID {
			im.carryItem.Quantity += slot.Item.Quantity
			slot.Clear()
		}
	}

	im.MakeItemImage()
}
