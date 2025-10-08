package playermenu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/inventory"
	"github.com/webbben/2d-game-engine/player"
)

type InventoryPage struct {
	init            bool
	PlayerInventory inventory.Inventory
	playerAvatar    *ebiten.Image
	playerRef       *player.Player
	width, height   int
}

func (ip *InventoryPage) Load(pageWidth, pageHeight int, playerRef *player.Player) {
	if playerRef == nil {
		panic("player ref is nil")
	}
	if playerRef.Entity == nil {
		panic("player entity is nil")
	}

	ip.width = pageWidth
	ip.height = pageHeight
	ip.PlayerInventory.RowCount = 9
	ip.PlayerInventory.ColCount = 9
	ip.PlayerInventory.EnabledSlotsCount = 18
	ip.PlayerInventory.Load()

	ip.playerRef = playerRef
	ip.playerAvatar = ip.playerRef.Entity.DrawAvatarBox(150, 300)

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

	// draw player avatar
	rendering.DrawImage(screen, ip.playerAvatar, drawX, drawY, 0)

	// draw inventory item slots
	inventoryWidth, _ := ip.PlayerInventory.Dimensions()
	inventoryDrawX := int(drawX) + ip.width - inventoryWidth
	ip.PlayerInventory.Draw(screen, float64(inventoryDrawX), drawY, om)
}
