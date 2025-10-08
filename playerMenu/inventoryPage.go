package playermenu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/inventory"
)

type InventoryPage struct {
	init            bool
	PlayerInventory inventory.Inventory
	width, height   int
}

func (ip *InventoryPage) Load(pageWidth, pageHeight int) {
	ip.width = pageWidth
	ip.height = pageHeight
	ip.PlayerInventory.RowCount = 9
	ip.PlayerInventory.ColCount = 9
	ip.PlayerInventory.EnabledSlotsCount = 18
	ip.PlayerInventory.Load()

	ip.init = true
}

func (ip *InventoryPage) Update() {
	if !ip.init {
		panic("inventory page not initialized")
	}
	ip.PlayerInventory.Update()
}

func (ip *InventoryPage) Draw(screen *ebiten.Image, drawX, drawY float64, om *overlay.OverlayManager) {
	if !ip.init {
		panic("inventory page not initialized")
	}

	// draw inventory item slots
	inventoryWidth, _ := ip.PlayerInventory.Dimensions()
	inventoryDrawX := int(drawX) + ip.width - inventoryWidth
	ip.PlayerInventory.Draw(screen, float64(inventoryDrawX), drawY, om)
}
