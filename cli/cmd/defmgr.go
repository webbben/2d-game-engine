package cmd

import (
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/item"
)

func LoadDefMgr(defMgr *definitions.DefinitionManager) {
	itemDefs := GetItemDefs()
	defMgr.LoadItemDefs(itemDefs)
	bodySkins, eyeSets, hairSets := GetAllEntityBodyPartSets()
	for _, skin := range bodySkins {
		defMgr.LoadBodyPartDef(skin.Body)
		defMgr.LoadBodyPartDef(skin.Arms)
		defMgr.LoadBodyPartDef(skin.Legs)
	}
	for _, eyes := range eyeSets {
		defMgr.LoadBodyPartDef(eyes)
	}
	for _, hair := range hairSets {
		defMgr.LoadBodyPartDef(hair)
	}

	shopKeeperInventory := []item.InventoryItem{}
	shopKeeperInventory = append(shopKeeperInventory, defMgr.NewInventoryItem("longsword_01", 1))
	shopkeeper := definitions.NewShopKeeper(1200, "Aurelius' Tradehouse", shopKeeperInventory)
	defMgr.LoadShopkeeper("aurelius_tradehouse", shopkeeper)

	defMgr.LoadDialog("dialog1", GetDialog())
}
