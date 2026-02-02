// Package definitions provides a centralized place for fetching definitions of items, shopkeepers, dialogs, etc.
package definitions

import (
	"github.com/webbben/2d-game-engine/dialog"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/skills"
)

type DefinitionManager struct {
	ItemDefs    map[string]item.ItemDef
	Shopkeepers map[string]*Shopkeeper
	Dialogs     map[string]dialog.Dialog

	BodyPartDefs map[string]body.SelectedPartDef // only for body "skin" parts (not for equipment, since those are part of item defs)

	AttributeDefs map[skills.AttributeID]skills.AttributeDef
	SkillDefs     map[skills.SkillID]skills.SkillDef
	TraitDefs     map[skills.TraitID]skills.Trait
}

func NewDefinitionManager() *DefinitionManager {
	def := DefinitionManager{
		ItemDefs:      make(map[string]item.ItemDef),
		Shopkeepers:   make(map[string]*Shopkeeper),
		Dialogs:       make(map[string]dialog.Dialog),
		BodyPartDefs:  make(map[string]body.SelectedPartDef),
		AttributeDefs: make(map[skills.AttributeID]skills.AttributeDef),
		SkillDefs:     make(map[skills.SkillID]skills.SkillDef),
		TraitDefs:     make(map[skills.TraitID]skills.Trait),
	}
	return &def
}

func (defMgr *DefinitionManager) LoadTraitDef(trait skills.Trait) {
	if trait.GetID() == "" {
		logz.Panicln("DefinitionManager", "tried to load trait, but ID was empty")
	}
	if _, exists := defMgr.TraitDefs[trait.GetID()]; exists {
		logz.Panicln("DefinitionManager", "tried to load in a new trait, but an existing trait of the same ID already exists:", trait.GetID())
	}
	defMgr.TraitDefs[trait.GetID()] = trait
}

func (defMgr DefinitionManager) GetTraitDef(id skills.TraitID) skills.Trait {
	trait, exists := defMgr.TraitDefs[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get trait by ID that doesn't exist:", id)
	}
	return trait
}

func (defMgr *DefinitionManager) LoadAttributeDef(attr skills.AttributeDef) {
	if attr.ID == "" {
		logz.Panicln("DefinitionManager", "tried to load attribute, but ID was empty")
	}
	if _, exists := defMgr.AttributeDefs[attr.ID]; exists {
		logz.Panicln("DefinitionManager", "tried to load in a new attribute, but an existing attribute of the same ID already exists:", attr.ID)
	}
	defMgr.AttributeDefs[attr.ID] = attr
}

func (defMgr DefinitionManager) GetAttributeDef(id skills.AttributeID) skills.AttributeDef {
	attrDef, exists := defMgr.AttributeDefs[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get attribute by ID that doesn't exist:", id)
	}
	return attrDef
}

func (defMgr *DefinitionManager) LoadSkillDef(sk skills.SkillDef) {
	if sk.ID == "" {
		logz.Panicln("DefinitionManager", "tried to load skill, but ID was empty")
	}
	if _, exists := defMgr.SkillDefs[sk.ID]; exists {
		logz.Panicln("DefinitionManager", "tried to load in a new skill, but an existing skill of the same ID already exists:", sk.ID)
	}
	defMgr.SkillDefs[sk.ID] = sk
}

func (defMgr DefinitionManager) GetSkillDef(id skills.SkillID) skills.SkillDef {
	skillDef, exists := defMgr.SkillDefs[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get skill by ID that doesn't exist:", id)
	}
	return skillDef
}

func (def *DefinitionManager) LoadBodyPartDef(partDef body.SelectedPartDef) {
	if partDef.ID == "" {
		logz.Panicln("DefinitionManager", "tried to load body part def, but ID is empty")
	}
	if _, exists := def.BodyPartDefs[partDef.ID]; exists {
		logz.Panicln("DefinitionManager", "tried to load in a new entity body part def, but the id already exists:", partDef.ID)
	}
	def.BodyPartDefs[partDef.ID] = partDef
}

func (def DefinitionManager) GetBodyPartDef(id string) body.SelectedPartDef {
	partDef, exists := def.BodyPartDefs[id]
	if !exists {
		logz.Panicln("DefinitionManager", "entity body part def not found:", id)
	}
	return partDef
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
		logz.Panicln("DefinitionManager", "item def not found:", defID)
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

	return item.NewInventoryItem(itemDef, quantity)
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

func (def *DefinitionManager) LoadDialog(dialogID string, d dialog.Dialog) {
	def.Dialogs[dialogID] = d
}

func (def DefinitionManager) GetDialog(dialogID string) dialog.Dialog {
	d, exists := def.Dialogs[dialogID]
	if !exists {
		logz.Panicf("dialogID not found in defMgr: %s", dialogID)
	}
	return d
}
