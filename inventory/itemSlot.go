package inventory

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/item"
)

type ItemSlot struct {
	x, y          int
	mouseBehavior mouse.MouseBehavior

	enabledImg        *ebiten.Image
	disabledImg       *ebiten.Image
	equipedBorderImg  *ebiten.Image
	selectedBorderImg *ebiten.Image

	Item     *item.ItemInstance // the actual item in the slot
	ItemInfo item.ItemDef       // the general item information (value, weight, etc)

	Enabled    bool
	IsSelected bool
	IsEquiped  bool
}

func (is *ItemSlot) SetContent(itemInstance *item.ItemInstance, itemInfo item.ItemDef) {
	is.Item = itemInstance
	is.ItemInfo = itemInfo
}

func (is ItemSlot) Dimensions() (dx, dy int) {
	return is.enabledImg.Bounds().Dx() * int(config.UIScale), is.enabledImg.Bounds().Dy() * int(config.UIScale)
}

func (is *ItemSlot) Draw(screen *ebiten.Image, x, y float64) {
	is.x = int(x)
	is.y = int(y)

	drawImg := is.enabledImg
	if !is.Enabled {
		drawImg = is.disabledImg
	}
	rendering.DrawImage(screen, drawImg, x, y, config.UIScale)

	if is.ItemInfo != nil {
		if is.IsEquiped {
			rendering.DrawImage(screen, is.equipedBorderImg, x, y, config.UIScale)
		}
		rendering.DrawImage(screen, is.ItemInfo.GetTileImg(), x, y, config.UIScale)
		if is.IsSelected {
			rendering.DrawImage(screen, is.selectedBorderImg, x, y, config.UIScale)
		}
	}
}

func (is *ItemSlot) Update() {
	if !is.Enabled {
		return
	}

	width, height := is.Dimensions()
	is.mouseBehavior.Update(is.x, is.y, width, height, false)

	if is.mouseBehavior.LeftClick.ClickReleased {
		if is.Item != nil {
			is.IsEquiped = !is.IsEquiped
		}
		logz.Println("itemslot", "enabled:", is.IsEquiped)
	}
}
