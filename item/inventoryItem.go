package item

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
)

type InventoryItem struct {
	Instance ItemInstance
	Def      ItemDef `json:"-"` // since this interface can't be properly loaded from JSON, lets exclude it from JSONs
	Quantity int
}

// NewInventoryItem puts the pieces of an inventoryItem together; use the one in definitionManager to actually get an inventory item from the game state.
func NewInventoryItem(def ItemDef, quantity int) InventoryItem {
	if def == nil {
		panic("def is nil")
	}
	invItem := InventoryItem{
		Instance: ItemInstance{
			DefID:      def.GetID(),
			Durability: def.GetMaxDurability(),
		},
		Def:      def,
		Quantity: quantity,
	}

	invItem.Validate()
	return invItem
}

func (invItem *InventoryItem) String() string {
	return fmt.Sprintf("{DefID: %s, Name: %s, Quant: %v}", invItem.Instance.DefID, invItem.Def.GetName(), invItem.Quantity)
}

func InventoryToString(inv []*InventoryItem) string {
	s := ""
	for _, invItem := range inv {
		if invItem == nil {
			continue
		}
		s += invItem.String() + ", "
	}
	return s
}

func (invItem InventoryItem) Validate() {
	if invItem.Def == nil {
		panic("item def is nil")
	}
	if invItem.Instance.DefID == "" {
		panic("item instance has no def ID")
	}
	if invItem.Quantity <= 0 {
		panic("item quantity is less than or equal to 0. all items must have a quantity of at least 1")
	}
	if invItem.Def.GetID() != invItem.Instance.DefID {
		panic("def.GetID() does not match instance.defID")
	}
}

func (i InventoryItem) Draw(screen *ebiten.Image, x, y float64) {
	tileSize := int(config.TileSize * config.UIScale)
	rendering.DrawImage(screen, i.Def.GetTileImg(), x, y, config.UIScale)

	if i.Quantity > 1 {
		qS := fmt.Sprintf("%v", i.Quantity)
		qDx, _, _ := text.GetStringSize(qS, config.DefaultFont)
		qX := int(x) + tileSize - qDx - 3
		qY := int(y) + tileSize - 5
		text.DrawOutlinedText(screen, fmt.Sprintf("%v", i.Quantity), config.DefaultFont, qX, qY, color.Black, color.White, 0, 0)
	}
}

func RemoveItemFromInventory(itemToRemove InventoryItem, removeFrom []*InventoryItem) (bool, InventoryItem) {
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
func AddItemToInventory(invItem InventoryItem, addTo []*InventoryItem) (bool, InventoryItem) {
	// if item is groupable, try to find a matching item already in the inventory
	if invItem.Def.IsGroupable() {
		for _, otherItem := range addTo {
			if otherItem == nil {
				continue
			}
			if otherItem.Instance.DefID == invItem.Instance.DefID {
				otherItem.Quantity += invItem.Quantity
				return true, InventoryItem{}
			}
		}
	}

	// not able to group together with existing item; place in empty slot if one is available
	for i, otherItem := range addTo {
		if otherItem == nil {
			addTo[i] = &invItem
			return true, InventoryItem{}
		}
	}

	return false, invItem
}
