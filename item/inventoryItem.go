package item

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
)

type InventoryItem struct {
	Instance ItemInstance
	Def      ItemDef
	Quantity int
}

func (invItem *InventoryItem) String() string {
	return fmt.Sprintf("{DefID: %s, Name: %s, Quant: %v}", invItem.Instance.DefID, invItem.Def.GetName(), invItem.Quantity)
}

func (i InventoryItem) Draw(screen *ebiten.Image, x, y float64) {
	tileSize := int(config.TileSize * config.UIScale)
	rendering.DrawImage(screen, i.Def.GetTileImg(), x, y, config.UIScale)

	if i.Quantity > 1 {
		qS := fmt.Sprintf("%v", i.Quantity)
		qDx, _, _ := text.GetStringSize(qS, config.DefaultFont)
		qX := int(x) + tileSize - qDx - 3
		qY := int(y) + tileSize - 5
		text.DrawOutlinedText(screen, fmt.Sprintf("%v", i.Quantity), config.DefaultFont, qX, qY, color.Black, color.White, 0, 0)
	}
}
