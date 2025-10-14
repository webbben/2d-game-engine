package definitions

import (
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/item"
)

type DefinitionManager struct {
	ItemDefs    map[string]item.ItemDef
	Shopkeepers map[string]*Shopkeeper
}

func NewDefinitionManager() *DefinitionManager {
	def := DefinitionManager{
		ItemDefs:    make(map[string]item.ItemDef),
		Shopkeepers: make(map[string]*Shopkeeper),
	}
	return &def
}

func (def *DefinitionManager) LoadItemDefs(itemDefs []item.ItemDef) {
	for _, itemDef := range itemDefs {
		itemDef.Validate()
		id := itemDef.GetID()
		itemDef.Load()
		def.ItemDefs[id] = itemDef
	}
}

func (def *DefinitionManager) GetItemDef(defID string) item.ItemDef {
	itemDef, exists := def.ItemDefs[defID]
	if !exists {
		panic("item def not found")
	}
	return itemDef
}

func (def *DefinitionManager) NewInventoryItem(defID string, quantity int) item.InventoryItem {
	if quantity <= 0 {
		panic("quantity must be a positive number")
	}
	itemDef := def.GetItemDef(defID)
	if itemDef == nil {
		panic("item def is nil")
	}

	invItem := item.InventoryItem{
		Instance: item.ItemInstance{
			DefID:      defID,
			Durability: itemDef.GetMaxDurability(),
		},
		Def:      itemDef,
		Quantity: quantity,
	}

	invItem.Validate()

	return invItem
}

func (def *DefinitionManager) LoadShopkeeper(shopkeeperID string, shopkeeper Shopkeeper) {
	def.Shopkeepers[shopkeeperID] = &shopkeeper
}

func (def DefinitionManager) GetShopkeeper(shopkeeperID string) *Shopkeeper {
	shopkeeper, exists := def.Shopkeepers[shopkeeperID]
	if !exists {
		logz.Panicf("shopkeeperID not found in defintionManager: %s", shopkeeperID)
	}
	return shopkeeper
}
