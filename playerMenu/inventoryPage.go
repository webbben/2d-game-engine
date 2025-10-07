package playermenu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/inventory"
)

type InventoryPage struct {
	init bool
	inventory.Inventory
	width, height int
}

func (ip *InventoryPage) Load(pageWidth, pageHeight int) {
	ip.width = pageWidth
	ip.height = pageHeight
	ip.Inventory.RowCount = 9
	ip.Inventory.ColCount = 9
	ip.Inventory.EnabledSlotsCount = 18
	ip.Inventory.Load()

	ip.init = true
}

func (ip *InventoryPage) Update() {
	if !ip.init {
		panic("inventory page not initialized")
	}
	ip.Inventory.Update()
}

func (ip *InventoryPage) Draw(screen *ebiten.Image, drawX, drawY float64) {
	if !ip.init {
		panic("inventory page not initialized")
	}

	// draw inventory item slots
	inventoryWidth, _ := ip.Inventory.Dimensions()
	inventoryDrawX := int(drawX) + ip.width - inventoryWidth
	ip.Inventory.Draw(screen, float64(inventoryDrawX), drawY)
}
