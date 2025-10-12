package playermenu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui"
	"github.com/webbben/2d-game-engine/inventory"
	"github.com/webbben/2d-game-engine/player"
)

type InventoryPage struct {
	init bool

	EquipedHead *inventory.ItemSlot // for hats, helmets, etc
	EquipedBody *inventory.ItemSlot // for shirts, cuirasses, robes,etc
	EquipedFeet *inventory.ItemSlot // for boots, shoes, etc

	EquipedAmulet *inventory.ItemSlot // can wear one amulet
	EquipedRing1  *inventory.ItemSlot // can wear two rings
	EquipedRing2  *inventory.ItemSlot

	EquipedAmmo      *inventory.ItemSlot // for arrows, sling bullets, etc
	EquipedAuxiliary *inventory.ItemSlot // for shields, torches, etc

	PlayerInventory inventory.Inventory
	playerAvatar    *ebiten.Image
	playerRef       *player.Player
	width, height   int

	itemMover inventory.ItemMover // for moving the items between slots

	goldCount ui.TextBox // displays the amount gold in the inventory
}

func (ip *InventoryPage) Load(pageWidth, pageHeight int, playerRef *player.Player, defMgr *definitions.DefinitionManager, inventoryParams inventory.InventoryParams) {
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
	ip.PlayerInventory = inventory.NewInventory(defMgr, inventoryParams)

	// load the item slot images for our equiped items slots
	itemSlotImages := inventory.LoadItemSlotTiles(
		inventoryParams.ItemSlotTilesetSource,
		inventoryParams.SlotEnabledTileID,
		inventoryParams.SlotDisabledTileID,
		inventoryParams.SlotEquipedBorderTileID,
		inventoryParams.SlotSelectedBorderTileID,
	)

	src := inventoryParams.ItemSlotTilesetSource

	ip.EquipedHead = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Headwear",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedHead.SetBGImage(tiled.GetTileImage(src, 5))
	ip.EquipedBody = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Bodywear",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedBody.SetBGImage(tiled.GetTileImage(src, 6))
	ip.EquipedFeet = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Footwear",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedFeet.SetBGImage(tiled.GetTileImage(src, 7))
	ip.EquipedAmmo = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Ammunition",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedAmmo.SetBGImage(tiled.GetTileImage(src, 8))
	ip.EquipedAuxiliary = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Auxiliary",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedAuxiliary.SetBGImage(tiled.GetTileImage(src, 9))

	ip.EquipedAmulet = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Amulet",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedAmulet.SetBGImage(tiled.GetTileImage(src, 10))
	ip.EquipedRing1 = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Ring",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedRing1.SetBGImage(tiled.GetTileImage(src, 11))
	ip.EquipedRing2 = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Ring",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedRing2.SetBGImage(tiled.GetTileImage(src, 11))

	ip.playerRef = playerRef
	ip.playerAvatar = ip.playerRef.Entity.DrawAvatarBox(100, 200)

	itemSlots := []*inventory.ItemSlot{}
	itemSlots = append(itemSlots, ip.PlayerInventory.GetItemSlots()...)
	itemSlots = append(
		itemSlots,
		ip.EquipedHead,
		ip.EquipedBody,
		ip.EquipedFeet,
		ip.EquipedAmulet,
		ip.EquipedRing1,
		ip.EquipedRing2,
		ip.EquipedAmmo,
		ip.EquipedAuxiliary,
	)

	ip.itemMover = inventory.NewItemMover(itemSlots)

	tileSize := int(config.TileSize * config.UIScale)

	goldIcon := tiled.GetTileImage(inventoryParams.ItemSlotTilesetSource, 194)
	ip.goldCount = ui.NewTextBox("25", inventoryParams.HoverWindowParams.TilesetSource, 135, config.DefaultFont, goldIcon, tileSize*4)

	ip.init = true
}

func (ip *InventoryPage) Update() {
	if !ip.init {
		panic("inventory page not initialized")
	}
	ip.PlayerInventory.Update()
	ip.EquipedHead.Update()
	ip.EquipedBody.Update()
	ip.EquipedFeet.Update()
	ip.EquipedAmmo.Update()
	ip.EquipedAuxiliary.Update()
	ip.EquipedAmulet.Update()
	ip.EquipedRing1.Update()
	ip.EquipedRing2.Update()

	ip.itemMover.Update()
}

func (ip *InventoryPage) Draw(screen *ebiten.Image, drawX, drawY float64, om *overlay.OverlayManager) {
	if !ip.init {
		panic("inventory page not initialized")
	}

	tileSize := config.TileSize * config.UIScale

	// draw player avatar
	rendering.DrawImage(screen, ip.playerAvatar, drawX, drawY, 0)

	// draw inventory item slots
	inventoryWidth, _ := ip.PlayerInventory.Dimensions()
	inventoryDrawX := int(drawX) + ip.width - inventoryWidth
	ip.PlayerInventory.Draw(screen, float64(inventoryDrawX), drawY, om)
	// player equipment item slots
	equipStartX := drawX + float64(ip.playerAvatar.Bounds().Dx()) + 10
	equipStartY := drawY + 10

	ip.EquipedHead.Draw(screen, equipStartX, equipStartY, om)
	ip.EquipedBody.Draw(screen, equipStartX, equipStartY+(1.5*tileSize), om)
	ip.EquipedFeet.Draw(screen, equipStartX, equipStartY+(3*tileSize), om)

	ip.EquipedAmulet.Draw(screen, equipStartX+(tileSize*1.5), equipStartY, om)
	ip.EquipedRing1.Draw(screen, equipStartX+(tileSize*1.5), equipStartY+(1.5*tileSize), om)
	ip.EquipedRing2.Draw(screen, equipStartX+(tileSize*1.5), equipStartY+(3*tileSize), om)

	ip.EquipedAmmo.Draw(screen, equipStartX+(tileSize*3), equipStartY+(0.75*tileSize), om)
	ip.EquipedAuxiliary.Draw(screen, equipStartX+(tileSize*3), equipStartY+(2.25*tileSize), om)

	ip.itemMover.Draw(om)

	ip.goldCount.Draw(screen, equipStartX+(tileSize*4.5), equipStartY)
}
