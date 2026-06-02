package inventory

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/mouse"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/ui/textwindow"
	"github.com/webbben/2d-game-engine/utils"
)

type ItemSlot struct {
	init              bool
	x, y              int
	mouseBehavior     mouse.MouseBehavior
	hoverWindow       textwindow.HoverWindow
	hoverWindowParams textwindow.TextWindowParams // since we have to recalculate the hover window when text changes, save the params

	itemSlotTiles ItemSlotTiles

	selectedBorderFader rendering.BounceFader

	Item     *state.ItemState
	ItemDef  defs.ItemDef
	ItemIcon item.ItemIcon

	Enabled    bool
	IsSelected bool
	IsEquiped  bool

	tooltip      string
	hoverTooltip textwindow.HoverTooltip

	allowedItemTypes []defs.ItemType // each item type in this array will be allowed; if nothing is set here, all items are allowed

	groupID string // ID of the "group" this item slot belongs to; groups are used for auto-moving items to specific slots
}

// InventoryItem is the UI component/runtime of an item in an inventory created when opening an inventory UI.
// All it really gives you is the tile's image, and any other inventory runtime/UI stuff you need.
// also def data, for ease of access.
type InventoryItem struct {
	TileImage *ebiten.Image

	ItemDef defs.ItemDef
}

type ItemSlotParams struct {
	ItemSlotTiles    ItemSlotTiles
	Enabled          bool
	Tooltip          string
	AllowedItemTypes []defs.ItemType // each item type in this array will be allowed; if nothing is set here, all items are allowed
	GroupID          string
}

func NewItemSlot(params ItemSlotParams, hoverWindowParams textwindow.TextWindowParams) *ItemSlot {
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
		groupID:             params.GroupID,
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
		itemSlot.hoverTooltip = textwindow.NewHoverTooltip(params.Tooltip, tooltipTileset, tooltipOrigin, 1000, -10, -10)
		itemSlot.tooltip = params.Tooltip
	}

	return &itemSlot
}

func (is ItemSlot) CanTakeItemType(itemType defs.ItemType) bool {
	if len(is.allowedItemTypes) == 0 {
		return true
	}

	return slices.Contains(is.allowedItemTypes, itemType)
}

func (is *ItemSlot) SetContent(itemState *state.ItemState, itemDef defs.ItemDef) {
	if itemState == nil {
		panic("item state is nil")
	}
	if itemState.Quantity <= 0 {
		panic("quantity is an invalid value. must be 1 or greater.")
	}
	if itemState.Quantity > 1 && !itemDef.Groupable {
		panic("tried to add multiple of a non-groupable item to an item slot")
	}
	if !is.CanTakeItemType(itemDef.Type) {
		panic("item slot can't take this item")
	}
	deref := *itemState
	is.Item = &deref
	is.ItemDef = itemDef
	is.ItemIcon = item.NewItemIcon(itemDef)

	// when an item is set, calculate the hover window
	// we have to do this on setting the item, since the text content may determine the actual size of the hover window.
	is.hoverWindow = textwindow.NewHoverWindow(itemDef.Name, itemDef.Description, is.hoverWindowParams)

	is.Validate()
}

func (is *ItemSlot) Clear() {
	is.Item = nil
}

func (is ItemSlot) Validate() {
	if is.Item != nil {
		utils.PanicAssert(is.Item.DefID == is.ItemDef.ID, "itemID doesn't match def ID")
		utils.PanicAssert(is.Item.Quantity > 0, "item quantity was <= 0")
		is.Item.Validate()
	}
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

		is.ItemIcon.Draw(screen, x, y, is.Item.Quantity)

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
	enabledImg, err := ts.GetTileImage(enTileID, true)
	if err != nil {
		panic(err)
	}
	disabledImg, err := ts.GetTileImage(disTileID, true)
	if err != nil {
		panic(err)
	}
	selectedBorder, err := ts.GetTileImage(selTileID, true)
	if err != nil {
		panic(err)
	}
	equipedBorder, err := ts.GetTileImage(eqTileID, true)
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
