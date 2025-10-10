package inventory

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
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

	Item         *item.ItemInstance // the actual item in the slot
	ItemInfo     item.ItemDef       // the general item information (value, weight, etc)
	ItemQuantity int                // number of this item that is in the item slot (if item is group-able)

	Enabled    bool
	IsSelected bool
	IsEquiped  bool

	tooltip      string
	hoverTooltip ui.HoverTooltip
}

type ItemSlotParams struct {
	ItemSlotTiles ItemSlotTiles
	Enabled       bool
	Tooltip       string
}

func NewItemSlot(params ItemSlotParams, hoverWindowParams ui.TextWindowParams) ItemSlot {
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
	}

	if params.Tooltip != "" {
		tooltipTileset := config.DefaultTooltipBox.TilesetSrc
		tooltipOrigin := config.DefaultTooltipBox.OriginIndex
		itemSlot.hoverTooltip = ui.NewHoverTooltip(params.Tooltip, tooltipTileset, tooltipOrigin, 1000, -10, -10)
		itemSlot.tooltip = params.Tooltip
	}

	return itemSlot
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
	is.Item = itemInstance
	is.ItemInfo = itemInfo
	is.ItemQuantity = quantity

	// when an item is set, calculate the hover window
	// we have to do this on setting the item, since the text content may determine the actual size of the hover window.
	is.hoverWindow = ui.NewHoverWindow(itemInfo.GetName(), itemInfo.GetDescription(), is.hoverWindowParams)
}

func (is *ItemSlot) Clear() {
	is.Item = nil
	is.ItemInfo = nil
	is.ItemQuantity = 0
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

	slotSize, _ := is.Dimensions()

	drawImg := is.itemSlotTiles.EnabledTile
	if !is.Enabled {
		drawImg = is.itemSlotTiles.DisabledTile
	}
	ops := ebiten.DrawImageOptions{}
	if is.mouseBehavior.IsHovering {
		ops.ColorScale.Scale(1.1, 1.1, 1.1, 1)
	}
	rendering.DrawImageWithOps(screen, drawImg, x, y, config.UIScale, &ops)
	if is.ItemInfo == nil && is.itemSlotTiles.BgImage != nil {
		// only show bg image if item slot is empty
		rendering.DrawImage(screen, is.itemSlotTiles.BgImage, x, y, config.UIScale)
	}

	if is.ItemInfo != nil {
		if is.IsEquiped {
			rendering.DrawImage(screen, is.itemSlotTiles.EquipedTile, x, y, config.UIScale)
		}
		rendering.DrawImage(screen, is.ItemInfo.GetTileImg(), x, y, config.UIScale)
		// draw quantity if applicable
		if is.ItemQuantity > 1 {
			qS := fmt.Sprintf("%v", is.ItemQuantity)
			qDx, _, _ := text.GetStringSize(qS, config.DefaultFont)
			qX := is.x + slotSize - qDx - 3
			qY := is.y + (slotSize) - 5
			text.DrawOutlinedText(screen, fmt.Sprintf("%v", is.ItemQuantity), config.DefaultFont, qX, qY, color.Black, color.White, 0, 0)
		}
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
		if is.mouseBehavior.LeftClick.ClickReleased {
			if is.ItemInfo.IsEquipable() {
				is.IsEquiped = !is.IsEquiped
			}
		}
		if is.mouseBehavior.RightClick.ClickReleased {
			is.IsSelected = !is.IsSelected
		}
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
