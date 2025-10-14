package trade

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui"
	"github.com/webbben/2d-game-engine/inventory"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/player"
)

type TradeScreen struct {
	Exit       bool // once set, trading will end
	defMgr     *definitions.DefinitionManager
	playerRef  *player.Player
	shopkeeper *definitions.Shopkeeper

	mainBox                     ui.BoxDef
	mainBoxImg                  *ebiten.Image
	mainBoxX, mainBoxY          int
	mainBoxWidth, mainBoxHeight int

	boxTitle ui.BoxTitle

	shopkeeperInventory            inventory.Inventory
	shopkeeperInvX, shopkeeperInvY int
	playerInventory                inventory.Inventory
	playerInvX, playerInvY         int

	boughtItems []tradeItem
	soldItems   []tradeItem

	playerGoldCount                      ui.TextBox
	playerGoldLabelX, playerGoldLabelY   int
	playerGoldCountX, playerGoldCountY   int
	transactionGoldCount                 ui.TextBox
	transactionLabelX, transactionLabelY int
	transactionX, transactionY           int

	acceptButton     *ui.Button
	acceptX, acceptY int
	cancelButton     *ui.Button
	cancelX, cancelY int

	itemTransfer inventory.ItemTransfer
}

type tradeItem struct {
	slot    *inventory.ItemSlot // slot the item is in
	invItem item.InventoryItem  // item type and amount traded
}

type TradeScreenParams struct {
	BoxTilesetSrc             string
	BoxTilesetOrigin          int
	BoxTitleOrigin            int
	ShopkeeperInventoryParams inventory.InventoryParams
	PlayerInventoryParams     inventory.InventoryParams
	TextBoxTilesetSrc         string
	TextBoxOrigin             int
}

func NewTradeScreen(params TradeScreenParams, defMgr *definitions.DefinitionManager, playerRef *player.Player) TradeScreen {
	if params.BoxTilesetSrc == "" {
		params.BoxTilesetSrc = config.DefaultUIBox.TilesetSrc
		params.BoxTilesetOrigin = config.DefaultUIBox.OriginIndex
	}

	tileSize := int(config.TileSize * config.UIScale)

	ts := TradeScreen{
		defMgr:    defMgr,
		playerRef: playerRef,
	}

	// main box
	ts.mainBox = ui.NewBox(params.BoxTilesetSrc, params.BoxTilesetOrigin)
	ts.mainBoxWidth = display.SCREEN_WIDTH * 3 / 4
	ts.mainBoxHeight = (display.SCREEN_HEIGHT * 2 / 3)
	ts.mainBoxWidth -= ts.mainBoxWidth % tileSize
	ts.mainBoxHeight -= ts.mainBoxHeight % tileSize
	ts.mainBoxImg = ts.mainBox.BuildBoxImage(ts.mainBoxWidth, ts.mainBoxHeight)

	// main box title
	ts.boxTitle = ui.NewBoxTitle(params.BoxTilesetSrc, params.BoxTitleOrigin, " Aurelius' Tradehouse ", config.DefaultTitleFont)

	// build inventories
	ts.shopkeeperInventory = inventory.NewInventory(defMgr, params.ShopkeeperInventoryParams)
	ts.playerInventory = inventory.NewInventory(defMgr, params.PlayerInventoryParams)

	ts.itemTransfer = inventory.NewItemTransfer(ts.playerInventory.GetItemSlots(), ts.shopkeeperInventory.GetItemSlots())

	// gold counters
	goldIcon := tiled.GetTileImage(params.PlayerInventoryParams.ItemSlotTilesetSource, 194)
	ts.playerGoldCount = ui.NewTextBox("0", params.TextBoxTilesetSrc, params.TextBoxOrigin, config.DefaultFont, goldIcon, &ui.TextBoxOptions{
		SetWidthPx: tileSize * 4,
	})
	ts.transactionGoldCount = ui.NewTextBox("0", params.TextBoxTilesetSrc, params.TextBoxOrigin, config.DefaultFont, goldIcon, &ui.TextBoxOptions{
		SetWidthPx:       tileSize * 4,
		HighlightOnHover: true,
	})

	// buttons
	ts.acceptButton = ui.NewButton("Accept", config.DefaultFont, tileSize*2, tileSize)
	ts.cancelButton = ui.NewButton("Cancel", config.DefaultFont, tileSize*2, tileSize)

	return ts
}

func (ts *TradeScreen) loadShopkeeper(sk *definitions.Shopkeeper) {
	ts.shopkeeper = sk
	ts.shopkeeperInventory.SetItemSlots(sk.ShopInventory)
}

// does necessary preparation for a trade session (loads shopkeeper and items, syncs player items, etc)
// must be called before beginning trade
func (ts *TradeScreen) SetupTradeSession(shopkeeper *definitions.Shopkeeper) {
	if ts.playerRef == nil {
		panic("no player ref!")
	}

	// ensure any previous trade session state is reset
	for _, slot := range ts.playerInventory.GetItemSlots() {
		slot.IsSelected = false
	}
	for _, slot := range ts.shopkeeperInventory.GetItemSlots() {
		slot.IsSelected = false
	}
	ts.soldItems = make([]tradeItem, 0)
	ts.boughtItems = make([]tradeItem, 0)

	ts.transactionGoldCount.SetText("0")

	// setup inventories
	ts.loadShopkeeper(shopkeeper)
	ts.loadPlayerInventory()

	ts.Exit = false
}

func (ts *TradeScreen) Draw(screen *ebiten.Image, om *overlay.OverlayManager) {
	if ts.Exit {
		return
	}

	tileSize := int(config.TileSize * config.UIScale)
	// draw main box
	ts.mainBoxX = (display.SCREEN_WIDTH / 2) - (ts.mainBoxWidth / 2)
	ts.mainBoxY = (display.SCREEN_HEIGHT / 2) - (ts.mainBoxHeight / 2)
	rendering.DrawImage(screen, ts.mainBoxImg, float64(ts.mainBoxX), float64(ts.mainBoxY), 0)
	// draw title
	titleX := ts.mainBoxX + (ts.mainBoxWidth / 2) - (ts.boxTitle.Width() / 2)
	titleY := ts.mainBoxY - tileSize
	ts.boxTitle.Draw(screen, float64(titleX), float64(titleY))

	// draw shopkeeper inventory
	ts.shopkeeperInvX = ts.mainBoxX + (tileSize / 2)
	ts.shopkeeperInvY = ts.mainBoxY + (tileSize / 2) + 9
	ts.shopkeeperInventory.Draw(screen, float64(ts.shopkeeperInvX), float64(ts.shopkeeperInvY), om)

	// draw player inventory
	playerInvW, _ := ts.playerInventory.Dimensions()
	ts.playerInvX = (ts.mainBoxX + ts.mainBoxWidth) - playerInvW - (tileSize / 2)
	ts.playerInvY = ts.shopkeeperInvY
	ts.playerInventory.Draw(screen, float64(ts.playerInvX), float64(ts.playerInvY), om)

	// draw gold counters
	shopkeeperInvWidth, shopkeeperInvHeight := ts.shopkeeperInventory.Dimensions()
	ts.playerGoldLabelX = ts.shopkeeperInvX + shopkeeperInvWidth + tileSize
	ts.playerGoldLabelY = ts.shopkeeperInvY + (tileSize)
	text.DrawShadowText(screen, "Your Gold", config.DefaultFont, ts.playerGoldLabelX, ts.playerGoldLabelY, nil, nil, 0, 0)
	ts.playerGoldCountX = ts.playerGoldLabelX - tileSize
	ts.playerGoldCountY = ts.playerGoldLabelY
	ts.playerGoldCount.Draw(screen, float64(ts.playerGoldCountX), float64(ts.playerGoldCountY))

	ts.transactionLabelX = ts.playerGoldLabelX - (tileSize / 4)
	ts.transactionLabelY = ts.playerGoldCountY + (tileSize * 3)
	text.DrawShadowText(screen, "Transaction", config.DefaultFont, ts.transactionLabelX, ts.transactionLabelY, nil, nil, 0, 0)
	ts.transactionX = ts.playerGoldCountX
	ts.transactionY = ts.transactionLabelY
	ts.transactionGoldCount.Draw(screen, float64(ts.transactionX), float64(ts.transactionY))

	// accept and cancel buttons
	ts.cancelX = ts.shopkeeperInvX + shopkeeperInvWidth
	ts.cancelY = ts.shopkeeperInvY + shopkeeperInvHeight - (tileSize)
	ts.cancelButton.Draw(screen, ts.cancelX, ts.cancelY)

	ts.acceptX = ts.playerInvX - ts.acceptButton.Width
	ts.acceptY = ts.cancelY
	ts.acceptButton.Draw(screen, ts.acceptX, ts.acceptY)
}

func (ts *TradeScreen) Update() {
	if ts.Exit {
		return
	}

	ts.shopkeeperInventory.Update()
	ts.playerInventory.Update()

	transferResult := ts.itemTransfer.Update()
	if transferResult.TransferAttemptOccurred {
		ts.handleItemTrade(transferResult)
		// recalculate transaction price
		transactionAmount := ts.calculateTransaction()
		ts.transactionGoldCount.SetText(general_util.ConvertIntToCommaString(transactionAmount))
	}

	acceptResult := ts.acceptButton.Update()
	if acceptResult.Clicked {
		ts.acceptTransaction()
		return
	}
	cancelResult := ts.cancelButton.Update()
	if cancelResult.Clicked {
		ts.cancelTransaction()
		return
	}
}

func (ts *TradeScreen) handleItemTrade(transferResult inventory.ItemTransferResult) {
	if transferResult.Success {
		if transferResult.ToPlayerInv {
			// bought an item
			// has the player sold any of this item already? if so, the player undoing the sale of the item
			for i, tradedItem := range ts.soldItems {
				if tradedItem.invItem.Instance.DefID == transferResult.TransferedItem.Instance.DefID {
					// negate from the already sold item
					ts.soldItems[i].invItem.Quantity -= transferResult.TransferedItem.Quantity
					if ts.soldItems[i].invItem.Quantity == 0 {
						// all previously sold items have been bought back, so remove the selection visual
						ts.soldItems[i].slot.IsSelected = false
						// remove this item from sold list
						ts.soldItems = append(ts.soldItems[:i], ts.soldItems[i+1:]...)
					}
					return
				}
			}
			// player has not sold this item yet - so they are buying it
			// show selection visual and add to buy list
			transferResult.TransferedTo.IsSelected = true
			added := false
			for i, tradedItem := range ts.boughtItems {
				if tradedItem.invItem.Instance.DefID == transferResult.TransferedItem.Instance.DefID {
					// we've already bought one of this item; increase the count
					ts.boughtItems[i].invItem.Quantity += transferResult.TransferedItem.Quantity
					added = true
					break
				}
			}
			if !added {
				ts.boughtItems = append(ts.boughtItems, tradeItem{
					slot:    transferResult.TransferedTo,
					invItem: transferResult.TransferedItem,
				})
			}
			return
		} else {
			// sold an item
			// has the player bought any of this item already? if so, the player undoing the purchase of the item
			for i, tradedItem := range ts.boughtItems {
				if tradedItem.invItem.Instance.DefID == transferResult.TransferedItem.Instance.DefID {
					// negate from the already bought item
					ts.boughtItems[i].invItem.Quantity -= transferResult.TransferedItem.Quantity
					if ts.boughtItems[i].invItem.Quantity == 0 {
						// all previously sold items have been bought back, so remove the selection visual
						ts.boughtItems[i].slot.IsSelected = false
						// remove this item from sold list
						ts.boughtItems = append(ts.boughtItems[:i], ts.boughtItems[i+1:]...)
					}
					return
				}
			}
			// player has not bought this item yet - so they are selling it
			// show selection visual and add to sell list
			transferResult.TransferedTo.IsSelected = true
			added := false
			for i, tradedItem := range ts.soldItems {
				if tradedItem.invItem.Instance.DefID == transferResult.TransferedItem.Instance.DefID {
					// we've already bought one of this item; increase the count
					ts.soldItems[i].invItem.Quantity += transferResult.TransferedItem.Quantity
					added = true
					break
				}
			}
			if !added {
				ts.soldItems = append(ts.soldItems, tradeItem{
					slot:    transferResult.TransferedTo,
					invItem: transferResult.TransferedItem,
				})
			}
			return
		}
	}
}

func (ts TradeScreen) calculateTransaction() int {
	totalBought := 0
	for _, tradedItem := range ts.boughtItems {
		totalBought += tradedItem.invItem.Def.GetValue() * tradedItem.invItem.Quantity
	}
	totalSold := 0
	for _, tradedItem := range ts.soldItems {
		totalSold += tradedItem.invItem.Def.GetValue() * tradedItem.invItem.Quantity
	}
	return totalSold - totalBought
}

func (ts *TradeScreen) loadPlayerInventory() {
	// set inventory items
	ts.playerInventory.SetItemSlots(ts.playerRef.InventoryItems)

	moneyCount := ts.playerRef.CountMoney()
	ts.playerGoldCount.SetText(general_util.ConvertIntToCommaString(moneyCount))
}

func (ts *TradeScreen) acceptTransaction() {
	transactionAmount := ts.calculateTransaction()

	if transactionAmount > 0 {
		// player earns money; add to coin purse
		ts.playerRef.EarnMoney(transactionAmount)
	} else if transactionAmount < 0 {
		// player spends money; take it out of their coin purse
		ts.playerRef.SpendMoney(-transactionAmount)
	}

	// // remove bought and sold items from their respective inventories
	// for _, boughtItem := range ts.boughtItems {
	// 	ts.playerRef.AddItemToInventory(boughtItem.invItem)
	// 	success, remaining := item.RemoveItemFromInventory(boughtItem.invItem, ts.shopKeeperItems)
	// 	if !success {
	// 		logz.Panicf("failed to remove items from shopkeeper inventory: %s", remaining.String())
	// 	}
	// }
	// for _, soldItem := range ts.soldItems {
	// 	success, remaining := ts.playerRef.RemoveItemFromInventory(soldItem.invItem)
	// 	if !success {
	// 		logz.Panicf("failed to remove items from player inventory: %s", remaining.String())
	// 	}
	// }

	// update player and shopkeeper inventories to match inventories in this trade screen
	ts.playerRef.SetInventoryItems(ts.playerInventory.GetInventoryItems())
	ts.shopkeeper.SetShopInventory(ts.shopkeeperInventory.GetInventoryItems())

	ts.Exit = true
}

func (ts *TradeScreen) cancelTransaction() {
	// we can just end the trade session, since the items should be automatically reset if the trade is started again later
	ts.Exit = true
}
