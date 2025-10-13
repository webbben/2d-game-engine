package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui"
	"github.com/webbben/2d-game-engine/item"
)

type ItemSlot struct {
	init              bool
	x, y              int
	mouseBehavior     mouse.MouseBehavior
	hoverWindow       ui.HoverWindow
	hoverWindowParams ui.TextWindowParams // since we have to recalculate the hover window when text changes, save the params

	itemSlotTiles ItemSlotTiles

	selectedBorderFader rendering.BounceFader

	Item *item.InventoryItem

	Enabled    bool
	IsSelected bool
	IsEquiped  bool

	tooltip      string
	hoverTooltip ui.HoverTooltip

	allowedItemTypes []string // each item type in this array will be allowed; if nothing is set here, all items are allowed
}

type ItemSlotParams struct {
	ItemSlotTiles    ItemSlotTiles
	Enabled          bool
	Tooltip          string
	AllowedItemTypes []string // each item type in this array will be allowed; if nothing is set here, all items are allowed
}

func NewItemSlot(params ItemSlotParams, hoverWindowParams ui.TextWindowParams) *ItemSlot {
	if params.ItemSlotTiles.EnabledTile == nil {
		panic("EnabledImage is nil")
	}
	if params.ItemSlotTiles.DisabledTile == nil {
		panic("DisabledImage is nil")
	}
	if params.ItemSlotTiles.EquipedTile == nil {
		panic("EquipedBorder is nil")
	}
	if params.ItemSlotTiles.SelectedTile == nil {
		panic("SelectedBorder is nil")
	}
	if hoverWindowParams.TilesetSource == "" {
		panic("hover window tileset source is empty")
	}

	itemSlot := ItemSlot{
		init:                true,
		itemSlotTiles:       params.ItemSlotTiles,
		Enabled:             params.Enabled,
		selectedBorderFader: rendering.NewBounceFader(0.5, 0.5, 0.8, 0.1),
		hoverWindowParams:   hoverWindowParams,
		allowedItemTypes:    params.AllowedItemTypes,
	}

	if params.Tooltip != "" {
		tooltipTileset := config.DefaultTooltipBox.TilesetSrc
		tooltipOrigin := config.DefaultTooltipBox.OriginIndex
		itemSlot.hoverTooltip = ui.NewHoverTooltip(params.Tooltip, tooltipTileset, tooltipOrigin, 1000, -10, -10)
		itemSlot.tooltip = params.Tooltip
	}

	return &itemSlot
}

func (is ItemSlot) CanTakeItemType(itemType string) bool {
	if len(is.allowedItemTypes) == 0 {
		return true
	}
	for _, allowedType := range is.allowedItemTypes {
		if itemType == allowedType {
			return true
		}
	}
	return false
}

func (is *ItemSlot) SetContent(itemInstance *item.ItemInstance, itemInfo item.ItemDef, quantity int) {
	if itemInfo == nil {
		panic("item info is nil")
	}
	if itemInstance == nil {
		panic("item instance is nil")
	}
	if quantity <= 0 {
		panic("quantity is an invalid value. must be 1 or greater.")
	}
	if quantity > 1 && !itemInfo.IsGroupable() {
		panic("tried to add multiple of a non-groupable item to an item slot")
	}
	if !is.CanTakeItemType(itemInfo.GetItemType()) {
		panic("item slot can't take this item")
	}
	is.Item = &item.InventoryItem{
		Instance: *itemInstance,
		Def:      itemInfo,
		Quantity: quantity,
	}

	// when an item is set, calculate the hover window
	// we have to do this on setting the item, since the text content may determine the actual size of the hover window.
	is.hoverWindow = ui.NewHoverWindow(itemInfo.GetName(), itemInfo.GetDescription(), is.hoverWindowParams)
}

func (is *ItemSlot) Clear() {
	is.Item = nil
}

func (is ItemSlot) Dimensions() (dx, dy int) {
	dx = is.itemSlotTiles.EnabledTile.Bounds().Dx() * int(config.UIScale)
	dy = is.itemSlotTiles.EnabledTile.Bounds().Dy() * int(config.UIScale)
	if dx == 0 {
		panic("item slot has no width")
	}
	if dy == 0 {
		panic("item slot has no height")
	}
	return dx, dy
}

func (is *ItemSlot) Draw(screen *ebiten.Image, x, y float64, om *overlay.OverlayManager) {
	if !is.init {
		panic("item slot not initialized")
	}
	is.x = int(x)
	is.y = int(y)

	drawImg := is.itemSlotTiles.EnabledTile
	if !is.Enabled {
		drawImg = is.itemSlotTiles.DisabledTile
	}
	ops := ebiten.DrawImageOptions{}
	if is.mouseBehavior.IsHovering {
		ops.ColorScale.Scale(1.1, 1.1, 1.1, 1)
	}
	rendering.DrawImageWithOps(screen, drawImg, x, y, config.UIScale, &ops)
	if is.Item == nil && is.itemSlotTiles.BgImage != nil {
		// only show bg image if item slot is empty
		rendering.DrawImage(screen, is.itemSlotTiles.BgImage, x, y, config.UIScale)
	}

	if is.Item != nil {
		if is.IsEquiped {
			rendering.DrawImage(screen, is.itemSlotTiles.EquipedTile, x, y, config.UIScale)
		}

		is.Item.Draw(screen, x, y)

		if is.IsSelected {
			ops := ebiten.DrawImageOptions{}
			ops.ColorScale.Scale(1, 1, 1, is.selectedBorderFader.GetCurrentScale())
			rendering.DrawImageWithOps(screen, is.itemSlotTiles.SelectedTile, x, y, config.UIScale, &ops)
		}
		is.hoverWindow.Draw(om)
	} else {
		if is.tooltip != "" {
			is.hoverTooltip.Draw(om)
		}
	}
}

func (is *ItemSlot) Update() {
	if !is.init {
		panic("item slot not initialized")
	}
	if !is.Enabled {
		return
	}

	width, height := is.Dimensions()
	is.mouseBehavior.Update(is.x, is.y, width, height, false)

	if is.Item != nil {
		w, h := is.Dimensions()
		is.hoverWindow.Update(float64(is.x), float64(is.y), w, h)
	} else {
		if is.tooltip != "" {
			is.hoverTooltip.Update(float64(is.x), float64(is.y), width, height)
		}
	}

	if is.IsSelected {
		is.selectedBorderFader.Update()
	}

}

type ItemSlotTiles struct {
	EnabledTile  *ebiten.Image
	DisabledTile *ebiten.Image
	EquipedTile  *ebiten.Image
	SelectedTile *ebiten.Image
	BgImage      *ebiten.Image
}

func LoadItemSlotTiles(tilesetSrc string, enTileID, disTileID, eqTileID, selTileID int) ItemSlotTiles {
	ts, err := tiled.LoadTileset(tilesetSrc)
	if err != nil {
		logz.Panicf("failed to load tileset for inventory: %s", err)
	}
	enabledImg, err := ts.GetTileImage(enTileID)
	if err != nil {
		panic(err)
	}
	disabledImg, err := ts.GetTileImage(disTileID)
	if err != nil {
		panic(err)
	}
	selectedBorder, err := ts.GetTileImage(selTileID)
	if err != nil {
		panic(err)
	}
	equipedBorder, err := ts.GetTileImage(eqTileID)
	if err != nil {
		panic(err)
	}

	tiles := ItemSlotTiles{
		EnabledTile:  enabledImg,
		DisabledTile: disabledImg,
		SelectedTile: selectedBorder,
		EquipedTile:  equipedBorder,
	}

	return tiles
}

func (is *ItemSlot) SetBGImage(img *ebiten.Image) {
	is.itemSlotTiles.BgImage = img
}
