package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/item"
)

type ItemMover struct {
	carryItem     *defs.InventoryItem // the item that is currently being carried
	dropableSlots []*ItemSlot         // item slots that an item can be dropped into
	itemImg       *ebiten.Image       // image of the item as its being carried
}

func NewItemMover(itemSlots []*ItemSlot) ItemMover {
	return ItemMover{
		dropableSlots: itemSlots,
	}
}

func (im *ItemMover) pickupItem(itemToCarry *defs.InventoryItem, amount int) bool {
	if im.carryItem != nil {
		// already carrying an item
		return false
	}

	im.carryItem = &defs.InventoryItem{
		Instance: itemToCarry.Instance,
		Def:      itemToCarry.Def,
		Quantity: amount,
	}

	im.MakeItemImage()

	return true
}

func (im *ItemMover) MakeItemImage() {
	if im.carryItem == nil {
		panic("no carry item to generate image for")
	}

	tileSize := int(config.TileSize * config.UIScale)

	im.itemImg = ebiten.NewImage(tileSize, tileSize)
	item.DrawInventoryItem(im.itemImg, *im.carryItem, 0, 0)
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
				if im.pickupItem(slot.Item, slot.Item.Quantity) {
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
				if im.pickupItem(slot.Item, amount) {
					if remaining > 0 {
						slot.SetContent(&slot.Item.Instance, slot.Item.Def, remaining)
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
				im.gatherAllOfItem(im.carryItem.Def)
				return
			}

			// placing all of the item
			if slot.CanTakeItemType(im.carryItem.Def.GetItemType()) {
				// check if the slot is empty
				if slot.Item == nil {
					// slot is empty, we can put this item here
					slot.SetContent(&im.carryItem.Instance, im.carryItem.Def, im.carryItem.Quantity)
					im.carryItem = nil
					return
				}

				// check if slot has the same item type (and is groupable)
				if slot.Item.Def.GetID() == im.carryItem.Instance.DefID {
					if im.carryItem.Def.IsGroupable() && slot.Item.Def.IsGroupable() {
						// found a matching groupable item
						slot.SetContent(&im.carryItem.Instance, im.carryItem.Def, im.carryItem.Quantity+slot.Item.Quantity)
						im.carryItem = nil
						return
					}
				}
			}
		} else if slot.mouseBehavior.RightClick.ClickReleased {
			// placing single item
			if slot.CanTakeItemType(im.carryItem.Def.GetItemType()) {
				if slot.Item == nil {
					slot.SetContent(&im.carryItem.Instance, im.carryItem.Def, 1)
					im.carryItem.Quantity--
					if im.carryItem.Quantity == 0 {
						im.carryItem = nil
					} else {
						im.MakeItemImage()
					}
					return
				}
				// check if slot has the same item type (and is groupable)
				if slot.Item.Def.GetID() == im.carryItem.Instance.DefID {
					if im.carryItem.Def.IsGroupable() && slot.Item.Def.IsGroupable() {
						// found a matching groupable item
						slot.SetContent(&im.carryItem.Instance, im.carryItem.Def, slot.Item.Quantity+1)
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
	if !def.IsGroupable() {
		return
	}

	for _, slot := range im.dropableSlots {
		if slot.Item == nil {
			continue
		}
		if slot.Item.Def.GetID() == def.GetID() {
			im.carryItem.Quantity += slot.Item.Quantity
			slot.Clear()
		}
	}

	im.MakeItemImage()
}
