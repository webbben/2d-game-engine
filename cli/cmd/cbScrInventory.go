package cmd

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/ui/textwindow"
	"github.com/webbben/2d-game-engine/inventory"
	playermenu "github.com/webbben/2d-game-engine/playerMenu"
)

type inventoryScreen struct {
	inventoryComponent playermenu.InventoryComponent
}

func (bg *builderGame) setupInventoryPage() {
	tileSize := int(config.TileSize * config.UIScale)

	width := bg.windowWidth - (tileSize * 2)
	width -= width % tileSize // round it to the size of the box tile
	height := bg.windowHeight * 2 / 3
	height -= height % tileSize

	invParams := inventory.InventoryParams{
		ItemSlotTilesetSource:    "ui/ui-components.tsj",
		SlotEnabledTileID:        0,
		SlotDisabledTileID:       1,
		SlotEquipedBorderTileID:  3,
		SlotSelectedBorderTileID: 4,
		HoverWindowParams: textwindow.TextWindowParams{
			TilesetSource:   "boxes/boxes.tsj",
			OriginTileIndex: 20,
		},
		RowCount:          10,
		ColCount:          9,
		EnabledSlotsCount: 18,
	}

	bg.scrInventory.inventoryComponent.Load(width, height, &bg.characterData, bg.defMgr, invParams)
	bg.refreshInventory()
}

func (bg *builderGame) refreshInventory() {
	bg.scrInventory.inventoryComponent.SyncEntityItems()
}

func (bg *builderGame) saveInventory() {
	bg.scrInventory.inventoryComponent.SaveEntityInventory()
}

func (bg *builderGame) updateInventoryPage() {
	bg.scrInventory.inventoryComponent.Update()
}

func (bg *builderGame) drawInventoryPage(screen *ebiten.Image, om *overlay.OverlayManager) {
	bg.scrInventory.inventoryComponent.Draw(screen, 50, 150, om)
}
