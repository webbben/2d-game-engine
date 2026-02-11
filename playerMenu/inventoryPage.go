package playermenu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/inventory"
)

type InventoryPage struct {
	init bool

	inventoryComponent InventoryComponent

	playerRef     *player.Player
	width, height int

	defMgr *definitions.DefinitionManager
}

// Load loads the inventory page for first time loading
func (ip *InventoryPage) Load(pageWidth, pageHeight int, playerRef *player.Player, defMgr *definitions.DefinitionManager, inventoryParams inventory.InventoryParams) {
	if playerRef == nil {
		panic("player ref is nil")
	}
	if playerRef.Entity == nil {
		panic("player entity is nil")
	}

	ip.defMgr = defMgr

	ip.width = pageWidth
	ip.height = pageHeight
	ip.playerRef = playerRef

	ip.inventoryComponent.Load(pageWidth, pageHeight, &playerRef.Entity.CharacterStateRef.StandardInventory, defMgr, inventoryParams)

	ip.init = true
}

func (ip *InventoryPage) Update() {
	ip.inventoryComponent.Update()
}

func (ip *InventoryPage) Draw(screen *ebiten.Image, drawX, drawY float64, om *overlay.OverlayManager) {
	ip.inventoryComponent.Draw(screen, drawX, drawY, om)
}

func (ip *InventoryPage) LoadPlayerItemsIn() {
	ip.inventoryComponent.SyncCharacterItems()
}

func (ip *InventoryPage) SaveAndClose() {
	ip.inventoryComponent.SaveCharacterInventory()
}
