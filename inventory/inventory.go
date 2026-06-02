// Package inventory contains logic for item inventories
package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/ui/textwindow"
)

type Inventory struct {
	init bool
	x, y int

	itemSlotTilesetSource    string // tileset where inventory tiles are loaded from
	slotEnabledTileID        int    // tile id from the tileset of the enabled slot image
	slotDisabledTileID       int    // tile id from the tileset of the disabled slot image
	slotEquipedBorderTileID  int    // tile id for the image of the border that signifies the slot is equiped
	slotSelectedBorderTileID int    // tile id for the image of the border that signifies the slot is selected

	hoverWindowParams textwindow.TextWindowParams

	RowCount          int // number of rows of item slots
	ColCount          int // number of columns of item slots
	EnabledSlotsCount int // number of item slots that are enabled

	itemSlots []*ItemSlot

	dataman *datamanager.DataManager
}

func (inv Inventory) GetItemSlots() []*ItemSlot {
	return inv.itemSlots
}

func (inv Inventory) GetInventoryItems() []*state.ItemState {
	invItems := []*state.ItemState{}
	for _, slot := range inv.itemSlots {
		slot.Validate()
		if slot.Item == nil {
			invItems = append(invItems, nil)
		} else {
			invItems = append(invItems, slot.Item)
		}
	}
	return invItems
}

func (inv *Inventory) ClearItemSlots() {
	for _, slot := range inv.itemSlots {
		slot.Clear()
	}
}

type InventoryParams struct {
	ItemSlotTilesetSource    string // tileset where inventory tiles are loaded from
	SlotEnabledTileID        int    // tile id from the tileset of the enabled slot image
	SlotDisabledTileID       int    // tile id from the tileset of the disabled slot image
	SlotEquipedBorderTileID  int    // tile id for the image of the border that signifies the slot is equiped
	SlotSelectedBorderTileID int    // tile id for the image of the border that signifies the slot is selected

	RowCount          int // number of rows of item slots
	ColCount          int // number of columns of item slots
	EnabledSlotsCount int // number of item slots that are enabled

	HoverWindowParams textwindow.TextWindowParams

	AllowedItemTypes []defs.ItemType // if set, all slots in this inventory will only allow items in this list of item IDs

	GroupID string // groupID to apply to the item slots in this inventory
}

func NewInventory(dataman *datamanager.DataManager, params InventoryParams) Inventory {
	inv := Inventory{
		dataman:                  dataman,
		itemSlotTilesetSource:    params.ItemSlotTilesetSource,
		slotEnabledTileID:        params.SlotEnabledTileID,
		slotDisabledTileID:       params.SlotDisabledTileID,
		slotEquipedBorderTileID:  params.SlotEquipedBorderTileID,
		slotSelectedBorderTileID: params.SlotSelectedBorderTileID,
		RowCount:                 params.RowCount,
		ColCount:                 params.ColCount,
		EnabledSlotsCount:        params.EnabledSlotsCount,
		hoverWindowParams:        params.HoverWindowParams,
	}

	if inv.itemSlotTilesetSource == "" {
		panic("no inventory tileset set")
	}
	if inv.RowCount == 0 || inv.ColCount == 0 {
		panic("row count or column count is 0")
	}
	if inv.hoverWindowParams.TilesetSource == "" {
		panic("hover window params: tileset source is empty")
	}

	itemSlotTiles := LoadItemSlotTiles(
		params.ItemSlotTilesetSource,
		params.SlotEnabledTileID,
		params.SlotDisabledTileID,
		params.SlotEquipedBorderTileID,
		params.SlotSelectedBorderTileID,
	)

	inv.itemSlots = make([]*ItemSlot, 0)

	for i := range inv.RowCount * inv.ColCount {
		itemSlot := NewItemSlot(ItemSlotParams{
			ItemSlotTiles:    itemSlotTiles,
			Enabled:          i < inv.EnabledSlotsCount,
			AllowedItemTypes: params.AllowedItemTypes,
			GroupID:          params.GroupID,
		}, inv.hoverWindowParams)

		inv.itemSlots = append(inv.itemSlots, itemSlot)
	}

	if len(inv.itemSlots) == 0 {
		panic("inventory has no item slots?")
	}

	inv.init = true

	return inv
}

// SetItemSlots sets all item slots; a nil spot represents an empty item slot.
// items can be less than the actual total slots number, since some slots may be disabled.
func (inv *Inventory) SetItemSlots(items []*state.ItemState) {
	if len(items) > len(inv.itemSlots) {
		logz.Panicf("trying to set more items than there are item slots. slots: %v items: %v", len(inv.itemSlots), len(items))
	}

	inv.ClearItemSlots()

	for i, invItem := range items {
		if invItem == nil {
			inv.itemSlots[i].Clear()
		} else {
			invItem.Validate()
			itemDef := inv.dataman.GetItemDef(invItem.DefID)
			inv.itemSlots[i].SetContent(invItem, itemDef)
		}
	}
}

// AddItems adds items to an inventory and returns the items that failed to be added (due to inventory being too full)
func (inv *Inventory) AddItems(items []state.ItemState) []state.ItemState {
	failedToAdd := []state.ItemState{}

	// find matching item that can merge
	for _, newItem := range items {
		placed := false
		for _, itemSlot := range inv.itemSlots {
			if itemSlot.Item == nil {
				continue
			}
			if newItem.DefID == itemSlot.Item.DefID {
				if itemSlot.ItemDef.Groupable {
					itemSlot.Item.Quantity += newItem.Quantity
					placed = true
					break
				}
			}
		}
		if placed {
			continue
		}

		// if no matching groupable item was found, add it anew
		// find an empty slot if one exists
		for _, itemSlot := range inv.itemSlots {
			if itemSlot.Item == nil {
				itemDef := inv.dataman.GetItemDef(newItem.DefID)
				itemSlot.SetContent(&newItem, itemDef)
				placed = true
				break
			}
		}

		if !placed {
			// still haven't placed it - this is a failed item then
			failedToAdd = append(failedToAdd, newItem)
		}
	}

	return failedToAdd
}

func (inv Inventory) Dimensions() (dx, dy int) {
	if !inv.init {
		panic("getting dimensions of inventory before initializing")
	}
	if len(inv.itemSlots) == 0 {
		panic("no item slots in inventory when getting inventory dimensions. did something load out or order?")
	}
	slotWidth, slotHeight := inv.itemSlots[0].Dimensions()
	return slotWidth * inv.ColCount, slotHeight * inv.RowCount
}

func (inv *Inventory) Draw(screen *ebiten.Image, drawX, drawY float64, om *overlay.OverlayManager) {
	if !inv.init {
		panic("inventory not initialized before drawing")
	}
	if len(inv.itemSlots) == 0 {
		panic("inventory has no item slots")
	}

	inv.x = int(drawX)
	inv.y = int(drawY)

	slotWidth, slotHeight := inv.itemSlots[0].Dimensions()

	for i := range inv.itemSlots {
		row := i / inv.ColCount
		col := i % inv.ColCount
		x := inv.x + (slotWidth * col)
		y := inv.y + (slotHeight * row)
		inv.itemSlots[i].Draw(screen, float64(x), float64(y), om)
	}
}

func (inv *Inventory) Update() {
	if !inv.init {
		panic("inventory not initialized before update called")
	}
	for i := range inv.itemSlots {
		inv.itemSlots[i].Update()
	}
}

func (inv *Inventory) TotalWeight() int {
	var total float64
	for _, itemSlot := range inv.itemSlots {
		if itemSlot.Item == nil {
			continue
		}
		total += itemSlot.ItemDef.Weight * float64(itemSlot.Item.Quantity)
	}
	return int(total)
}
