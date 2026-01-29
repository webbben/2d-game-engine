package cmd

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/ui/button"
	"github.com/webbben/2d-game-engine/internal/ui/dropdown"
	"github.com/webbben/2d-game-engine/internal/ui/textfield"
	"github.com/webbben/2d-game-engine/internal/ui/textwindow"
	"github.com/webbben/2d-game-engine/inventory"
	playermenu "github.com/webbben/2d-game-engine/playerMenu"
)

type inventoryScreen struct {
	inventoryComponent playermenu.InventoryComponent

	moneyInput      textfield.TextField
	moneyConfirmBtn *button.Button

	itemDropdown dropdown.OptionSelect
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

	// money setters
	dx, _, _ := text.GetStringSize("1000000000", config.DefaultFont)
	bg.scrInventory.moneyInput = *textfield.NewTextField(textfield.TextFieldParams{
		WidthPx:            dx,
		NumericOnly:        true,
		TextColor:          color.White,
		BorderColor:        color.White,
		BgColor:            color.Black,
		MaxCharacterLength: 9,
		FontFace:           config.DefaultFont,
	})
	bg.scrInventory.moneyConfirmBtn = button.NewLinearBoxButton("Set Gold", "ui/ui-components.tsj", 352, config.DefaultTitleFont)

	// item setters
	itemDefs := GetItemDefs()
	itemIDs := []string{}
	for _, itemDef := range itemDefs {
		itemIDs = append(itemIDs, itemDef.GetID())
	}
	bg.scrInventory.itemDropdown = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               itemIDs,
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
		InputEnabled:          true,
	}, &bg.popupMgr)

	bg.refreshInventory()
}

func (bg *builderGame) refreshInventory() {
	bg.scrInventory.inventoryComponent.SyncEntityItems()
	bg.scrInventory.moneyInput.SetNumber(bg.characterData.CountMoney()) // initialize to current money
}

func (bg *builderGame) saveInventory() {
	bg.scrInventory.inventoryComponent.SaveEntityInventory()
}

func (bg *builderGame) updateInventoryPage() {
	bg.scrInventory.inventoryComponent.Update()

	// money input
	bg.scrInventory.moneyInput.Update()
	if bg.scrInventory.moneyConfirmBtn.Update().Clicked {
		logz.Println("", "money confirm clicked")
		inputMoney := bg.scrInventory.moneyInput.GetNumber()
		currentMoney := bg.characterData.CountMoney()
		if inputMoney != currentMoney {
			// reset character money to this value
			bg.saveInventory() // save any changes first
			if currentMoney > 0 {
				bg.characterData.SpendMoney(currentMoney, bg.defMgr)
			}
			if inputMoney > 0 {
				bg.characterData.EarnMoney(inputMoney, bg.defMgr)
			}
			afterUpdate := bg.characterData.CountMoney()
			if afterUpdate != inputMoney {
				logz.Panicln("CB Inventory Page", "tried to set money, but looks like it failed. input:", inputMoney, "after update:", afterUpdate)
			}
			bg.refreshInventory() // refresh so we see the changes
		}
	}

	// item input
	bg.scrInventory.itemDropdown.Update()
}

func (bg *builderGame) drawInventoryPage(screen *ebiten.Image, om *overlay.OverlayManager) {
	bg.scrInventory.inventoryComponent.Draw(screen, 50, 150, om)

	_, dy := bg.scrInventory.inventoryComponent.Dimensions()
	moneyInputX := 50
	moneyInputY := 150 + dy
	bg.scrInventory.moneyInput.Draw(screen, float64(moneyInputX), float64(moneyInputY))
	moneyInputDx, moneyInputDy := bg.scrInventory.moneyInput.Dimensions()
	bg.scrInventory.moneyConfirmBtn.Draw(screen, moneyInputX, moneyInputY+moneyInputDy+10)

	itemInputX := moneyInputX + moneyInputDx + 150
	itemInputY := moneyInputY
	bg.scrInventory.itemDropdown.Draw(screen, float64(itemInputX), float64(itemInputY), om)
}
