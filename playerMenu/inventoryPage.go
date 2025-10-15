package playermenu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui"
	"github.com/webbben/2d-game-engine/internal/ui/box"
	"github.com/webbben/2d-game-engine/inventory"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/player"
	"golang.org/x/text/message"
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

	goldCountX, goldCountY          float64
	goldCountWidth, goldCountHeight int
	goldCount                       ui.TextBox // displays the amount gold in the inventory
	goldCountMouse                  mouse.MouseBehavior

	coinPurseX, coinPurseY float64
	coinPurse              inventory.Inventory // inventory section that holds the coins
	coinPurseBox           *ebiten.Image       // box that holds the coin purse inventory
	showCoinPurse          bool
	coinPurseMouse         mouse.MouseBehavior // for detecting if the user clicks outside the coin purse

	defMgr *definitions.DefinitionManager
}

// for first time loading
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

	tileSize := int(config.TileSize * config.UIScale)

	// gold counter and coin purse set up
	goldIcon := tiled.GetTileImage(inventoryParams.ItemSlotTilesetSource, 194)
	ip.goldCount = ui.NewTextBox("25", inventoryParams.HoverWindowParams.TilesetSource, 135, config.DefaultFont, goldIcon, &ui.TextBoxOptions{
		SetWidthPx:       tileSize * 4,
		HighlightOnHover: true,
	})

	coinPurseBox := box.NewBox(config.DefaultUIBox.TilesetSrc, config.DefaultUIBox.OriginIndex)
	ip.coinPurseBox = coinPurseBox.BuildBoxImage(tileSize*4, tileSize*3)
	coinPurseInvParams := inventoryParams
	coinPurseInvParams.RowCount = 2
	coinPurseInvParams.ColCount = 3
	coinPurseInvParams.EnabledSlotsCount = 6
	coinPurseInvParams.AllowedItemTypes = []string{item.TypeCurrency}
	ip.coinPurse = inventory.NewInventory(defMgr, coinPurseInvParams)

	// set up item mover
	itemSlots := []*inventory.ItemSlot{}
	itemSlots = append(itemSlots, ip.PlayerInventory.GetItemSlots()...)
	itemSlots = append(itemSlots, ip.coinPurse.GetItemSlots()...)
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

	ip.init = true
}

func (ip *InventoryPage) SyncPlayerItems() {
	// set equiped items
	setInventoryItem(ip.EquipedHead, ip.playerRef.EquipedHeadwear)
	setInventoryItem(ip.EquipedBody, ip.playerRef.EquipedBodywear)
	setInventoryItem(ip.EquipedFeet, ip.playerRef.EquipedFootwear)

	setInventoryItem(ip.EquipedAmulet, ip.playerRef.EquipedAmulet)
	setInventoryItem(ip.EquipedRing1, ip.playerRef.EquipedRing1)
	setInventoryItem(ip.EquipedRing2, ip.playerRef.EquipedRing2)

	setInventoryItem(ip.EquipedAmmo, ip.playerRef.EquipedAmmo)
	setInventoryItem(ip.EquipedAuxiliary, ip.playerRef.EquipedAuxiliary)

	// set inventory items
	ip.PlayerInventory.SetItemSlots(ip.playerRef.InventoryItems)

	// set coin purse items
	ip.coinPurse.SetItemSlots(ip.playerRef.CoinPurse)

	moneyCount := ip.CountMoney()
	p := message.NewPrinter(message.MatchLanguage("en"))
	ip.goldCount.SetText(p.Sprintf("%d", moneyCount))
}

func setInventoryItem(itemSlot *inventory.ItemSlot, invItem *item.InventoryItem) {
	if invItem == nil {
		itemSlot.Clear()
	} else {
		itemSlot.SetContent(
			&invItem.Instance,
			invItem.Def,
			invItem.Quantity,
		)
	}
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

	// gold counter and coin purse
	ip.goldCount.Update()
	ip.goldCountMouse.Update(int(ip.goldCountX), int(ip.goldCountY), ip.goldCountWidth, ip.goldCountHeight, false)
	if ip.goldCountMouse.LeftClick.ClickReleased {
		ip.showCoinPurse = !ip.showCoinPurse
		ip.coinPurseMouse.LeftClickOutside.Reset() // need this to prevent a dropped click
	}

	if ip.showCoinPurse {
		ip.coinPurse.Update()
		bounds := ip.coinPurseBox.Bounds()
		w, h := bounds.Dx(), bounds.Dy()
		ip.coinPurseMouse.Update(int(ip.coinPurseX), int(ip.coinPurseY), w, h, false)
		if ip.coinPurseMouse.LeftClickOutside.ClickReleased {
			ip.showCoinPurse = false
		}
	}
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

	ip.goldCountX = equipStartX + (tileSize * 4.5)
	ip.goldCountY = equipStartY
	ip.goldCountWidth, ip.goldCountHeight = ip.goldCount.Dimensions()
	ip.goldCount.Draw(screen, ip.goldCountX, ip.goldCountY)

	if ip.showCoinPurse {
		ip.coinPurseX = ip.goldCountX
		ip.coinPurseY = ip.goldCountY + (tileSize * 2)
		rendering.DrawImage(screen, ip.coinPurseBox, ip.coinPurseX, ip.coinPurseY, 0)
		ip.coinPurse.Draw(screen, ip.coinPurseX+(tileSize/2), ip.coinPurseY+(tileSize/2), om)
	}
}

func (ip InventoryPage) CountMoney() int {
	sum := 0
	for _, itemSlot := range ip.coinPurse.GetItemSlots() {
		if itemSlot.Item == nil {
			continue
		}
		if itemSlot.Item.Def.IsCurrencyItem() {
			sum += itemSlot.Item.Def.GetValue() * itemSlot.Item.Quantity
		}
	}

	// also check for coins not in coin purse
	for _, itemSlot := range ip.PlayerInventory.GetItemSlots() {
		if itemSlot.Item == nil {
			continue
		}
		if itemSlot.Item.Def.IsCurrencyItem() {
			sum += itemSlot.Item.Def.GetValue() * itemSlot.Item.Quantity
		}
	}

	return sum
}
