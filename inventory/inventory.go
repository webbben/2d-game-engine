package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui"
	"github.com/webbben/2d-game-engine/item"
)

type Inventory struct {
	init bool
	x, y int

	ItemSlotTilesetSource    string // tileset where inventory tiles are loaded from
	SlotEnabledTileID        int    // tile id from the tileset of the enabled slot image
	SlotDisabledTileID       int    // tile id from the tileset of the disabled slot image
	SlotEquipedBorderTileID  int    // tile id for the image of the border that signifies the slot is equiped
	SlotSelectedBorderTileID int    // tile id for the image of the border that signifies the slot is selected

	HoverWindowParams ui.TextWindowParams

	RowCount          int // number of rows of item slots
	ColCount          int // number of columns of item slots
	EnabledSlotsCount int // number of item slots that are enabled

	itemSlots []ItemSlot
	Items     []InventoryItem // the items that are in this inventory
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
}

func (inv *Inventory) Load() {
	if inv.ItemSlotTilesetSource == "" {
		panic("no inventory tileset set")
	}
	if inv.RowCount == 0 || inv.ColCount == 0 {
		panic("row count or column count is 0")
	}
	if inv.HoverWindowParams.TilesetSource == "" {
		panic("hover window params: tileset source is empty")
	}

	ts, err := tiled.LoadTileset(inv.ItemSlotTilesetSource)
	if err != nil {
		logz.Panicf("failed to load tileset for inventory: %s", err)
	}
	enabledImg, err := ts.GetTileImage(inv.SlotEnabledTileID)
	if err != nil {
		panic(err)
	}
	disabledImg, err := ts.GetTileImage(inv.SlotDisabledTileID)
	if err != nil {
		panic(err)
	}
	selectedBorder, err := ts.GetTileImage(inv.SlotSelectedBorderTileID)
	if err != nil {
		panic(err)
	}
	equipedBorder, err := ts.GetTileImage(inv.SlotEquipedBorderTileID)
	if err != nil {
		panic(err)
	}

	inv.itemSlots = make([]ItemSlot, 0)

	for i := range inv.RowCount * inv.ColCount {
		var itemInstance item.ItemInstance
		var itemDef item.ItemDef
		if i < len(inv.Items) {
			itemInstance = inv.Items[i].Instance
			itemDef = inv.Items[i].Def
			logz.Println("inventory", "item:", itemDef.GetID())
		}

		itemSlot := NewItemSlot(ItemSlotParams{
			EnabledImage:   enabledImg,
			DisabledImage:  disabledImg,
			EquipedBorder:  equipedBorder,
			SelectedBorder: selectedBorder,
			Enabled:        i < inv.EnabledSlotsCount,
		}, inv.HoverWindowParams)

		if itemDef != nil {
			itemSlot.SetContent(&itemInstance, itemDef)
		}

		inv.itemSlots = append(inv.itemSlots, itemSlot)
	}

	if len(inv.itemSlots) == 0 {
		panic("inventory has no item slots?")
	}

	inv.init = true
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
