package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/ui"
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

	hoverWindowParams ui.TextWindowParams

	RowCount          int // number of rows of item slots
	ColCount          int // number of columns of item slots
	EnabledSlotsCount int // number of item slots that are enabled

	itemSlots []ItemSlot
	Items     []InventoryItem // the items that are in this inventory

	defMgr *definitions.DefinitionManager
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

	HoverWindowParams ui.TextWindowParams
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

	inv.itemSlots = make([]ItemSlot, 0)

	for i := range inv.RowCount * inv.ColCount {
		var itemInstance item.ItemInstance
		var itemDef item.ItemDef
		if i < len(inv.Items) {
			itemInstance = inv.Items[i].Instance
			itemDef = inv.Items[i].Def
		}

		itemSlot := NewItemSlot(ItemSlotParams{
			ItemSlotTiles: itemSlotTiles,
			Enabled:       i < inv.EnabledSlotsCount,
		}, inv.hoverWindowParams)

		if itemDef != nil {
			itemSlot.SetContent(&itemInstance, itemDef, inv.Items[i].Quantity)
		}

		inv.itemSlots = append(inv.itemSlots, itemSlot)
	}

	if len(inv.itemSlots) == 0 {
		panic("inventory has no item slots?")
	}

	inv.init = true

	return inv
}

// returns the items that failed to be added (due to inventory being too full)
func (inv *Inventory) AddItems(items []item.ItemInstance) []item.ItemInstance {
	failedToAdd := []item.ItemInstance{}

	// find matching item that can merge
	for _, instance := range items {
		placed := false
		for i, invItem := range inv.Items {
			if invItem.Def.GetID() == instance.DefID {
				if invItem.Def.IsGroupable() {
					inv.Items[i].Quantity++
					placed = true
					break
				}
			}
		}
		if placed {
			continue
		}
		// if no matching groupable item was found, add it anew
		if len(inv.Items) == inv.EnabledSlotsCount {
			// no more space
			failedToAdd = append(failedToAdd, instance)
			continue
		}
		inventoryItem := InventoryItem{
			Instance: instance,
			Def:      inv.defMgr.GetItemDef(instance.DefID),
			Quantity: 1,
		}
		inv.Items = append(inv.Items, inventoryItem)
	}

	// put the items in the item slots
	for i := range inv.itemSlots {
		if i < len(inv.Items) {
			invItem := inv.Items[i]
			inv.itemSlots[i].SetContent(&invItem.Instance, invItem.Def, invItem.Quantity)
		} else {
			inv.itemSlots[i].Clear()
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

type InventoryItem struct {
	Instance item.ItemInstance
	Def      item.ItemDef
	Quantity int
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
	for _, invItem := range inv.Items {
		total += invItem.Def.GetWeight() * float64(invItem.Quantity)
	}
	return int(total)
}
