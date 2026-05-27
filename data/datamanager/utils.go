package datamanager

import "github.com/webbben/2d-game-engine/data/defs"

func (dm *DataManager) EnsureItemDefs(invItems []*defs.InventoryItem) {
	for _, invItem := range invItems {
		if invItem == nil {
			continue
		}
		if invItem.Def == nil {
			invItem.Def = dm.GetItemDef(invItem.Instance.DefID)
		}
	}
}
