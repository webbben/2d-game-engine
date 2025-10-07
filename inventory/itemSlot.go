package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/item"
)

type ItemSlot struct {
	init          bool
	x, y          int
	mouseBehavior mouse.MouseBehavior

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

func NewItemSlot(enImg, disImg, equipImg, selectImg *ebiten.Image, enabled bool) ItemSlot {
	return ItemSlot{
		init:                true,
		enabledImg:          enImg,
		disabledImg:         disImg,
		equipedBorderImg:    equipImg,
		selectedBorderImg:   selectImg,
		Enabled:             enabled,
		selectedBorderFader: rendering.NewBounceFader(0.5, 0.5, 0.8, 0.1),
	}
}

func (is *ItemSlot) SetContent(itemInstance *item.ItemInstance, itemInfo item.ItemDef) {
	is.Item = itemInstance
	is.ItemInfo = itemInfo
}

func (is ItemSlot) Dimensions() (dx, dy int) {
	return is.enabledImg.Bounds().Dx() * int(config.UIScale), is.enabledImg.Bounds().Dy() * int(config.UIScale)
}

func (is *ItemSlot) Draw(screen *ebiten.Image, x, y float64) {
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

	if is.mouseBehavior.LeftClick.ClickReleased {
		if is.Item != nil {
			is.IsEquiped = !is.IsEquiped
		}
	}
	if is.mouseBehavior.RightClick.ClickReleased {
		if is.Item != nil {
			is.IsSelected = !is.IsSelected
		}
	}
	if is.IsSelected {
		is.selectedBorderFader.Update()
	}
}
