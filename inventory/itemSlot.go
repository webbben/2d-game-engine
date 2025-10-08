package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/ui"
	"github.com/webbben/2d-game-engine/item"
)

type ItemSlot struct {
	init              bool
	x, y              int
	mouseBehavior     mouse.MouseBehavior
	hoverWindow       ui.HoverWindow
	hoverWindowParams ui.TextWindowParams // since we have to recalculate the hover window when text changes, save the params

	enabledImg        *ebiten.Image
	disabledImg       *ebiten.Image
	equipedBorderImg  *ebiten.Image
	selectedBorderImg *ebiten.Image

	selectedBorderFader rendering.BounceFader

	Item     *item.ItemInstance // the actual item in the slot
	ItemInfo item.ItemDef       // the general item information (value, weight, etc)

	Enabled    bool
	IsSelected bool
	IsEquiped  bool
}

type ItemSlotParams struct {
	EnabledImage, DisabledImage   *ebiten.Image
	EquipedBorder, SelectedBorder *ebiten.Image
	Enabled                       bool
}

func NewItemSlot(params ItemSlotParams, hoverWindowParams ui.TextWindowParams) ItemSlot {
	if params.EnabledImage == nil {
		panic("EnabledImage is nil")
	}
	if params.DisabledImage == nil {
		panic("DisabledImage is nil")
	}
	if params.EquipedBorder == nil {
		panic("EquipedBorder is nil")
	}
	if params.SelectedBorder == nil {
		panic("SelectedBorder is nil")
	}
	if hoverWindowParams.TilesetSource == "" {
		panic("tileset source is empty")
	}

	return ItemSlot{
		init:                true,
		enabledImg:          params.EnabledImage,
		disabledImg:         params.DisabledImage,
		equipedBorderImg:    params.EquipedBorder,
		selectedBorderImg:   params.SelectedBorder,
		Enabled:             params.Enabled,
		selectedBorderFader: rendering.NewBounceFader(0.5, 0.5, 0.8, 0.1),
		hoverWindowParams:   hoverWindowParams,
	}
}

func (is *ItemSlot) SetContent(itemInstance *item.ItemInstance, itemInfo item.ItemDef) {
	if itemInfo == nil {
		panic("item info is nil")
	}
	if itemInstance == nil {
		panic("item instance is nil")
	}
	is.Item = itemInstance
	is.ItemInfo = itemInfo

	// when an item is set, calculate the hover window
	// we have to do this on setting the item, since the text content may determine the actual size of the hover window.
	is.hoverWindow = ui.NewHoverWindow(itemInfo.GetName(), itemInfo.GetDescription(), is.hoverWindowParams)
}

func (is ItemSlot) Dimensions() (dx, dy int) {
	return is.enabledImg.Bounds().Dx() * int(config.UIScale), is.enabledImg.Bounds().Dy() * int(config.UIScale)
}

func (is *ItemSlot) Draw(screen *ebiten.Image, x, y float64, om *overlay.OverlayManager) {
	if !is.init {
		panic("item slot not initialized")
	}
	is.x = int(x)
	is.y = int(y)

	drawImg := is.enabledImg
	if !is.Enabled {
		drawImg = is.disabledImg
	}
	ops := ebiten.DrawImageOptions{}
	if is.mouseBehavior.IsHovering {
		ops.ColorScale.Scale(1.1, 1.1, 1.1, 1)
	}
	rendering.DrawImageWithOps(screen, drawImg, x, y, config.UIScale, &ops)

	if is.ItemInfo != nil {
		if is.IsEquiped {
			rendering.DrawImage(screen, is.equipedBorderImg, x, y, config.UIScale)
		}
		rendering.DrawImage(screen, is.ItemInfo.GetTileImg(), x, y, config.UIScale)
		if is.IsSelected {
			ops := ebiten.DrawImageOptions{}
			ops.ColorScale.Scale(1, 1, 1, is.selectedBorderFader.GetCurrentScale())
			rendering.DrawImageWithOps(screen, is.selectedBorderImg, x, y, config.UIScale, &ops)
		}
		is.hoverWindow.Draw(om)
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
			is.IsEquiped = !is.IsEquiped
		}
		if is.mouseBehavior.RightClick.ClickReleased {
			is.IsSelected = !is.IsSelected
		}
		w, h := is.Dimensions()
		is.hoverWindow.Update(float64(is.x), float64(is.y), w, h)
	}

	if is.IsSelected {
		is.selectedBorderFader.Update()
	}

}
