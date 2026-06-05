package inventory

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/utils"
)

type ItemMover struct {
	carryItem     *state.ItemState // the item that is currently being carried
	carryItemDef  defs.ItemDef
	originalSlot  *ItemSlot     // slot the carried item was originally in, so it can be replaced if needed
	dropableSlots []*ItemSlot   // item slots that an item can be dropped into
	itemImg       *ebiten.Image // image of the item as its being carried

	possibleTransfers map[string][]string

	playerAvatarRect *model.Rect
}

func NewItemMover(itemSlots []*ItemSlot) ItemMover {
	return ItemMover{
		dropableSlots:     itemSlots,
		possibleTransfers: make(map[string][]string),
	}
}

// AddPlayerAvatorZone lets you set a zone on the screen where the player avatar is.
// If it is set, then the item mover can "drop" items onto the player, which can help with quickly using an item:
//
// - Consumables (potions, food, etc): consumes the item
//
// - Activatable items (books, etc): activates the item; reads the book, etc.
//
// - Equippable items (headwear, bodywear, footwear, etc): equips the item
func (im *ItemMover) AddPlayerAvatorZone(x, y float64, dx, dy int) {
	if dx <= 0 {
		panic("dx <= 0")
	}
	if dy <= 0 {
		panic("dy <= 0")
	}
	im.playerAvatarRect = &model.Rect{
		X: x,
		Y: y,
		W: float64(dx),
		H: float64(dy),
	}
}

func (im *ItemMover) AddPossibleGroupTransfer(fromGroupID string, toGroupID string) {
	transferDests := im.possibleTransfers[fromGroupID]

	if slices.Contains(transferDests, toGroupID) {
		return
	}

	transferDests = append(transferDests, toGroupID)
	im.possibleTransfers[fromGroupID] = transferDests
}

// if quantity is 0, we pick up all
func (im *ItemMover) pickupItem(slot *ItemSlot, quantity int) bool {
	if im.carryItem != nil {
		// already carrying an item
		return false
	}
	if slot.Item == nil {
		logz.Panicln("ItemMover", "item slot was empty!")
	}

	if quantity < 0 {
		logz.Panic("quantity was < 0")
	}

	slot.Validate()

	if quantity == 0 {
		// pick up all of the item
		quantity = slot.Item.Quantity
	}

	utils.PanicAssert(quantity > 0, "quantity was <= 0")

	deref := *slot.Item
	itemToCarry := &deref
	itemToCarry.Quantity = quantity
	itemDef := slot.ItemDef

	slot.Item.Quantity -= quantity
	if slot.Item.Quantity < 0 {
		logz.Panicln("ItemMover", "picked up more than existed in an item slot!", quantity)
	}
	if slot.Item.Quantity == 0 {
		slot.Clear()
	}

	slot.Validate()

	im.originalSlot = slot

	im.carryItem = itemToCarry
	im.carryItemDef = itemDef
	im.MakeItemImage()

	im.carryItem.Validate()

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

type ItemMoverUpdateResult struct {
	OpenBook     bool
	BookID       defs.BookID
	LastHeldItem defs.ItemID
}

func (im *ItemMover) Update() ItemMoverUpdateResult {
	result := ItemMoverUpdateResult{}
	if im.carryItem != nil {
		result.LastHeldItem = im.carryItem.DefID
	}

	if im.carryItem == nil {
		// check for clicks on item slots - for picking up items
		for i, slot := range im.dropableSlots {
			if slot.Item == nil {
				continue
			}
			if slot.mouseBehavior.LeftClick.ClickReleased {
				// the slot has been clicked!
				if im.pickupItem(slot, 0) {
					// item was successfully picked up
					// (SHIFT+LEFT-CLICK) transfer the item to an appropriate slot
					if ebiten.IsKeyPressed(ebiten.KeyShift) {
						im.attemptGroupTransfer(i)
					}
					slot.Validate()
					return result
				}
			} else if slot.mouseBehavior.RightClick.ClickReleased {
				// right click = pick up half
				amount := slot.Item.Quantity
				if amount > 1 {
					amount /= 2
				}
				if im.pickupItem(slot, amount) {
					return result
				}
			}
		}
	} else {
		// item is already being carried; check for item placement
		return im.handleItemPlacement()
	}

	return result
}

// attemptGroupTransfer attempts to transfer the carried item to a valid new item slot.
//
// rules:
//   - item cannot move into a new slot of the same group ID
//   - item will first try to group up with another slot of the same item def (if item is groupable)
//   - item will then look for valid slots that list its type as an allowed type (prefers matching to item type match)
//   - if no item type match exists, it will go to the first valid slot found according to the group possible transfers map
func (im *ItemMover) attemptGroupTransfer(originIndex int) {
	if im.carryItem == nil {
		panic("no carry item")
	}

	originSlot := im.dropableSlots[originIndex]

	if im.carryItemDef.Groupable {
		// first, check if there is a slot that accepts only specifically this item type
		for i, slot := range im.dropableSlots {
			if i == originIndex {
				// can't place in the origin slot
				continue
			}

			if slot.Item == nil {
				continue
			}

			// if there is an item in this slot, see if its the same item def and we can group i t
			if slot.Item.DefID == im.carryItemDef.ID {
				// found the same item, and its groupable
				slot.Item.Quantity += im.carryItem.Quantity
				im.carryItem = nil
				return
			}
		}
	}

	// check if there is a slot that accepts only specifically this item type
	for i, slot := range im.dropableSlots {
		if i == originIndex {
			// can't place in the origin slot
			continue
		}

		if slot.Item != nil {
			continue
		}

		// check if the slot has an allow type match
		if slices.Contains(slot.allowedItemTypes, im.carryItemDef.Type) {
			slot.SetContent(im.carryItem, im.carryItemDef)
			im.carryItem = nil
			return
		}
	}

	// finally, if nothing else, then try to move according to the possibleTransfers map
	possibleTransfers := im.possibleTransfers[originSlot.groupID]

	for _, slot := range im.dropableSlots {
		if slot.Item != nil {
			continue
		}

		if slices.Contains(possibleTransfers, slot.groupID) {
			slot.SetContent(im.carryItem, im.carryItemDef)
			im.carryItem = nil
			return
		}
	}

	// if we got here, then the item failed to be transfered
	// put the carry item back in its origin slot
	originSlot.SetContent(im.carryItem, im.carryItemDef)
	im.carryItem = nil
}

func (im *ItemMover) putItemBack() {
	logz.Println("", "puttin it back...")
	if im.carryItem == nil {
		logz.Panic("not carrying any item!")
	}
	if im.originalSlot == nil {
		logz.Panic("original slot was nil!")
	}
	im.originalSlot.SetContent(im.carryItem, im.carryItemDef)
	im.carryItem = nil
}

func (im *ItemMover) handleItemPlacement() ItemMoverUpdateResult {
	if im.carryItem == nil {
		panic("tried to place item but no carryItem exists")
	}

	result := ItemMoverUpdateResult{
		LastHeldItem: im.carryItem.DefID,
	}

	mouseX, mouseY := ebiten.CursorPosition()

	// first, check if the mouse is over the player avatar (if it exists)
	if im.playerAvatarRect != nil {
		if im.playerAvatarRect.Within(mouseX, mouseY) {
			if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
				// attempting to place item onto the player avatar
				switch im.carryItemDef.Type {
				case defs.TypeBook:
					bookID := im.carryItemDef.BookID
					logz.Println("ItemMover", "you should read this book:", bookID)
					im.putItemBack()

					result.OpenBook = true
					result.BookID = bookID
					return result
				}
			}
		}
	}

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
				return result
			}

			// placing all of the item
			if slot.CanTakeItemType(im.carryItemDef.Type) {
				// check if the slot is empty
				if slot.Item == nil {
					// slot is empty, we can put this item here
					slot.SetContent(im.carryItem, im.carryItemDef)
					im.carryItem = nil
					return result
				}

				// check if slot has the same item type (and is groupable)
				if slot.ItemDef.ID == im.carryItem.DefID {
					if im.carryItemDef.Groupable && slot.ItemDef.Groupable {
						// found a matching groupable item
						newItemState := *im.carryItem
						newItemState.Quantity += slot.Item.Quantity
						slot.SetContent(&newItemState, im.carryItemDef)
						im.carryItem = nil
						return result
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
					return result
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
						return result
					}
				}
			}
		}
	}

	return result
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
