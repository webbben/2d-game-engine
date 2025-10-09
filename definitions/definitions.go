package definitions

import "github.com/webbben/2d-game-engine/item"

type DefinitionManager struct {
	ItemDefs map[string]item.ItemDef
}

func NewDefinitionManager() *DefinitionManager {
	def := DefinitionManager{
		ItemDefs: make(map[string]item.ItemDef),
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
