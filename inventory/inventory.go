package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/ui/textwindow"
	"github.com/webbben/2d-game-engine/item"
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

	defMgr *definitions.DefinitionManager
}

func (inv Inventory) GetItemSlots() []*ItemSlot {
	return inv.itemSlots
}

func (inv Inventory) GetInventoryItems() []*item.InventoryItem {
	invItems := []*item.InventoryItem{}
	for _, slot := range inv.itemSlots {
		if slot.Item == nil {
			invItems = append(invItems, nil)
		} else {
			invItems = append(invItems, &item.InventoryItem{
				Instance: slot.Item.Instance,
				Def:      slot.Item.Def,
				Quantity: slot.Item.Quantity,
			})
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

	AllowedItemTypes []item.ItemType // if set, all slots in this inventory will only allow items in this list of item IDs
}

func NewInventory(defMgr *definitions.DefinitionManager, params InventoryParams) Inventory {
	inv := Inventory{
		defMgr:                   defMgr,
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
		}, inv.hoverWindowParams)

		inv.itemSlots = append(inv.itemSlots, itemSlot)
	}

	if len(inv.itemSlots) == 0 {
		panic("inventory has no item slots?")
	}

	inv.init = true

	return inv
}

// sets all item slots; a nil spot represents an empty item slot.
// items can be less than the actual total slots number, since some slots may be disabled.
func (inv *Inventory) SetItemSlots(items []*item.InventoryItem) {
	if len(items) > len(inv.itemSlots) {
		logz.Panicf("trying to set more items than there are item slots. slots: %v items: %v", len(inv.itemSlots), len(items))
	}

	inv.ClearItemSlots()

	for i, invItem := range items {
		if invItem == nil {
			inv.itemSlots[i].Clear()
		} else {
			invItem.Validate()
			if invItem.Quantity == 0 {
				logz.Println(invItem.Instance.DefID)
				panic("trying to set an item that has 0 quantity")
			}
			inv.itemSlots[i].SetContent(&invItem.Instance, invItem.Def, invItem.Quantity)
		}
	}
}

// returns the items that failed to be added (due to inventory being too full)
func (inv *Inventory) AddItems(items []item.InventoryItem) []item.InventoryItem {
	failedToAdd := []item.InventoryItem{}

	// find matching item that can merge
	for _, newItem := range items {
		placed := false
		for _, itemSlot := range inv.itemSlots {
			if itemSlot.Item == nil {
				continue
			}
			if newItem.Instance.DefID == itemSlot.Item.Instance.DefID {
				if newItem.Def.IsGroupable() {
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
				itemSlot.SetContent(&newItem.Instance, newItem.Def, newItem.Quantity)
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
		total += itemSlot.Item.Def.GetWeight() * float64(itemSlot.Item.Quantity)
	}
	return int(total)
}
