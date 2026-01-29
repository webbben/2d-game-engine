package playermenu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui/box"
	"github.com/webbben/2d-game-engine/internal/ui/textbox"
	"github.com/webbben/2d-game-engine/inventory"
	"github.com/webbben/2d-game-engine/item"
	"golang.org/x/text/message"
)

type InventoryComponent struct {
	init bool

	EquipedHead *inventory.ItemSlot // for hats, helmets, etc
	EquipedBody *inventory.ItemSlot // for shirts, cuirasses, robes,etc
	EquipedFeet *inventory.ItemSlot // for boots, shoes, etc

	EquipedAmulet *inventory.ItemSlot // can wear one amulet
	EquipedRing1  *inventory.ItemSlot // can wear two rings
	EquipedRing2  *inventory.ItemSlot

	EquipedAmmo      *inventory.ItemSlot // for arrows, sling bullets, etc
	EquipedAuxiliary *inventory.ItemSlot // for shields, torches, etc

	EntityInventory inventory.Inventory
	entityRef       *entity.CharacterData
	width, height   int

	itemMover inventory.ItemMover // for moving the items between slots

	goldCountX, goldCountY          float64
	goldCountWidth, goldCountHeight int
	goldCount                       textbox.TextBox // displays the amount gold in the inventory
	goldCountMouse                  mouse.MouseBehavior

	coinPurseX, coinPurseY float64
	coinPurse              inventory.Inventory // inventory section that holds the coins
	coinPurseBox           *ebiten.Image       // box that holds the coin purse inventory
	showCoinPurse          bool
	coinPurseMouse         mouse.MouseBehavior // for detecting if the user clicks outside the coin purse

	defMgr *definitions.DefinitionManager
}

func (ic InventoryComponent) Dimensions() (dx, dy int) {
	return ic.width, ic.height
}

// Load loads the inventory page for first time loading
func (ip *InventoryComponent) Load(pageWidth, pageHeight int, entRef *entity.CharacterData, defMgr *definitions.DefinitionManager, inventoryParams inventory.InventoryParams) {
	if entRef == nil {
		panic("player ref is nil")
	}
	if pageWidth == 0 {
		panic("width is 0")
	}
	if pageHeight == 0 {
		panic("height is 0")
	}

	ip.defMgr = defMgr

	ip.width = pageWidth
	ip.height = pageHeight
	ip.EntityInventory.RowCount = 9
	ip.EntityInventory.ColCount = 9
	ip.EntityInventory.EnabledSlotsCount = 18
	ip.EntityInventory = inventory.NewInventory(defMgr, inventoryParams)

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
	ip.EquipedHead.SetBGImage(tiled.GetTileImage(src, 5, true))
	ip.EquipedBody = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Bodywear",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedBody.SetBGImage(tiled.GetTileImage(src, 6, true))
	ip.EquipedFeet = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Footwear",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedFeet.SetBGImage(tiled.GetTileImage(src, 7, true))
	ip.EquipedAmmo = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Ammunition",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedAmmo.SetBGImage(tiled.GetTileImage(src, 8, true))
	ip.EquipedAuxiliary = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Auxiliary",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedAuxiliary.SetBGImage(tiled.GetTileImage(src, 9, true))

	ip.EquipedAmulet = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Amulet",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedAmulet.SetBGImage(tiled.GetTileImage(src, 10, true))
	ip.EquipedRing1 = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Ring",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedRing1.SetBGImage(tiled.GetTileImage(src, 11, true))
	ip.EquipedRing2 = inventory.NewItemSlot(inventory.ItemSlotParams{
		ItemSlotTiles: itemSlotImages,
		Enabled:       true,
		Tooltip:       "Ring",
	}, inventoryParams.HoverWindowParams)
	ip.EquipedRing2.SetBGImage(tiled.GetTileImage(src, 11, true))

	ip.entityRef = entRef

	tileSize := int(config.TileSize * config.UIScale)

	// gold counter and coin purse set up
	goldIcon := tiled.GetTileImage(inventoryParams.ItemSlotTilesetSource, 194, true)
	ip.goldCount = textbox.NewTextBox("No Data!", inventoryParams.HoverWindowParams.TilesetSource, 135, config.DefaultFont, goldIcon, &textbox.TextBoxOptions{
		SetWidthPx:       tileSize * 4,
		HighlightOnHover: true,
	})

	coinPurseBox := box.NewBox(config.DefaultUIBox.TilesetSrc, config.DefaultUIBox.OriginIndex)
	ip.coinPurseBox = coinPurseBox.BuildBoxImage(tileSize*4, tileSize*3)
	coinPurseInvParams := inventoryParams
	coinPurseInvParams.RowCount = 2
	coinPurseInvParams.ColCount = 3
	coinPurseInvParams.EnabledSlotsCount = 6
	coinPurseInvParams.AllowedItemTypes = []item.ItemType{item.TypeCurrency}
	ip.coinPurse = inventory.NewInventory(defMgr, coinPurseInvParams)

	// set up item mover
	itemSlots := []*inventory.ItemSlot{}
	itemSlots = append(itemSlots, ip.EntityInventory.GetItemSlots()...)
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

// SyncEntityItems syncs the entity's items into the inventory item slots
func (ip *InventoryComponent) SyncEntityItems() {
	if ip.entityRef == nil {
		panic("no entity ref found")
	}
	// set equiped items
	setInventoryItem(ip.EquipedHead, ip.entityRef.EquipedHeadwear)
	setInventoryItem(ip.EquipedBody, ip.entityRef.EquipedBodywear)
	setInventoryItem(ip.EquipedFeet, ip.entityRef.EquipedFootwear)

	setInventoryItem(ip.EquipedAmulet, ip.entityRef.EquipedAmulet)
	setInventoryItem(ip.EquipedRing1, ip.entityRef.EquipedRing1)
	setInventoryItem(ip.EquipedRing2, ip.entityRef.EquipedRing2)

	setInventoryItem(ip.EquipedAmmo, ip.entityRef.EquipedAmmo)
	setInventoryItem(ip.EquipedAuxiliary, ip.entityRef.EquipedAuxiliary)

	// set inventory items
	ip.EntityInventory.SetItemSlots(ip.entityRef.InventoryItems)

	// set coin purse items
	ip.coinPurse.SetItemSlots(ip.entityRef.CoinPurse)

	moneyCount := ip.entityRef.CountMoney()
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

func (ip *InventoryComponent) Update() {
	if !ip.init {
		panic("inventory page not initialized")
	}
	ip.EntityInventory.Update()

	// update equiped item slots
	ip.EquipedHead.Update()
	ip.EquipedBody.Update()
	ip.EquipedFeet.Update()
	ip.EquipedAmmo.Update()
	ip.EquipedAuxiliary.Update()
	ip.EquipedAmulet.Update()
	ip.EquipedRing1.Update()
	ip.EquipedRing2.Update()

	ip.itemMover.Update()

	// check if any differences exist between equiped item slots and actual equiped items
	// if so, update the playerRef version to match the equiped item slots
	if ip.EquipedHead.Item == nil {
		if ip.entityRef.EquipedHeadwear != nil {
			ip.entityRef.UnequipHeadwear()
		}
	} else {
		if ip.entityRef.EquipedHeadwear == nil {
			logz.Println(ip.entityRef.DisplayName, "equiping headwear:", ip.EquipedHead.Item.Def.GetID())
			succ := ip.entityRef.EquipItem(*ip.EquipedHead.Item)
			if !succ {
				logz.Panicln(ip.entityRef.DisplayName, "somehow failed to equip headwear")
			}
		} else if ip.entityRef.EquipedHeadwear.Def.GetID() != ip.EquipedHead.Item.Def.GetID() {
			logz.Panicln(ip.entityRef.DisplayName, "somehow, equiped headwear slot in inventory does not match equiped headwear on body")
		}
	}
	if ip.EquipedBody.Item == nil {
		if ip.entityRef.EquipedBodywear != nil {
			ip.entityRef.UnequipBodywear()
		}
	} else {
		if ip.entityRef.EquipedBodywear == nil {
			ip.entityRef.EquipItem(*ip.EquipedBody.Item)
		} else if ip.entityRef.EquipedBodywear.Def.GetID() != ip.EquipedBody.Item.Def.GetID() {
			logz.Panicln(ip.entityRef.DisplayName, "somehow, equiped bodywear slot in inventory does not match equiped bodywear on body")
		}
	}
	if ip.EquipedAuxiliary.Item == nil {
		if ip.entityRef.EquipedAuxiliary != nil {
			ip.entityRef.UnequipAuxiliary()
		}
	} else {
		if ip.entityRef.EquipedAuxiliary == nil {
			ip.entityRef.EquipItem(*ip.EquipedAuxiliary.Item)
		} else if ip.entityRef.EquipedAuxiliary.Def.GetID() != ip.EquipedAuxiliary.Item.Def.GetID() {
			logz.Panicln(ip.entityRef.DisplayName, "somehow, equiped auxiliary slot in inventory does not match equiped auxiliary on body")
		}
	}

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

func (ip *InventoryComponent) Draw(screen *ebiten.Image, drawX, drawY float64, om *overlay.OverlayManager) {
	if !ip.init {
		panic("inventory page not initialized")
	}

	tileSize := config.TileSize * config.UIScale

	// draw player avatar
	ip.entityRef.Body.Draw(screen, drawX, drawY, config.UIScale)

	// draw inventory item slots
	inventoryWidth, _ := ip.EntityInventory.Dimensions()
	inventoryDrawX := int(drawX) + ip.width - inventoryWidth
	ip.EntityInventory.Draw(screen, float64(inventoryDrawX), drawY, om)
	// player equipment item slots
	playerAvatarDx, _ := ip.entityRef.Body.Dimensions()
	playerAvatarDx = int(float64(playerAvatarDx) * config.UIScale)

	equipStartX := drawX + float64(playerAvatarDx) + 10
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

func (ip InventoryComponent) CountMoney() int {
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
	for _, itemSlot := range ip.EntityInventory.GetItemSlots() {
		if itemSlot.Item == nil {
			continue
		}
		if itemSlot.Item.Def.IsCurrencyItem() {
			sum += itemSlot.Item.Def.GetValue() * itemSlot.Item.Quantity
		}
	}

	return sum
}

// SaveEntityInventory saves the items in the inventory page into the entity's actual inventory
func (ip *InventoryComponent) SaveEntityInventory() {
	if len(ip.entityRef.CoinPurse) > 0 {
		ip.entityRef.SetCoinPurseItems(ip.coinPurse.GetInventoryItems())
	}
	ip.entityRef.SetInventoryItems(ip.EntityInventory.GetInventoryItems())

	ip.entityRef.EquipedHeadwear = ip.EquipedHead.Item
	ip.entityRef.EquipedBodywear = ip.EquipedBody.Item
	ip.entityRef.EquipedFootwear = ip.EquipedFeet.Item

	ip.entityRef.EquipedAmulet = ip.EquipedAmulet.Item
	ip.entityRef.EquipedRing1 = ip.EquipedRing1.Item
	ip.entityRef.EquipedRing2 = ip.EquipedRing2.Item

	ip.entityRef.EquipedAmmo = ip.EquipedAmmo.Item
	ip.entityRef.EquipedAuxiliary = ip.EquipedAuxiliary.Item
}
