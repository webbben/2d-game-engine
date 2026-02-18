// Package definitions provides a centralized place for fetching definitions of items, shopkeepers, dialogs, etc.
package definitions

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/logz"
)

type DefinitionManager struct {
	ItemDefs map[defs.ItemID]defs.ItemDef

	ShopkeeperDefs      map[defs.ShopID]*defs.ShopkeeperDef
	ShopkeeperStates    map[defs.ShopID]*state.ShopkeeperState
	DialogProfiles      map[defs.DialogProfileID]*defs.DialogProfileDef
	DialogProfileStates map[defs.DialogProfileID]*state.DialogProfileState
	DialogTopics        map[defs.TopicID]*defs.DialogTopic

	NPCSchedules map[defs.ScheduleID]defs.ScheduleDef

	BodyPartDefs    map[defs.BodyPartID]defs.SelectedPartDef // only for body "skin" parts (not for equipment, since those are part of item defs)
	FootstepSFXDefs map[defs.FootstepSFXDefID]defs.FootstepSFXDef

	CharacterDefs   map[defs.CharacterDefID]defs.CharacterDef
	CharacterStates map[state.CharacterStateID]*state.CharacterState

	AttributeDefs map[defs.AttributeID]defs.AttributeDef
	SkillDefs     map[defs.SkillID]defs.SkillDef
	TraitDefs     map[defs.TraitID]defs.Trait
}

func NewDefinitionManager() *DefinitionManager {
	def := DefinitionManager{
		ItemDefs:            make(map[defs.ItemID]defs.ItemDef),
		ShopkeeperDefs:      make(map[defs.ShopID]*defs.ShopkeeperDef),
		ShopkeeperStates:    make(map[defs.ShopID]*state.ShopkeeperState),
		BodyPartDefs:        make(map[defs.BodyPartID]defs.SelectedPartDef),
		FootstepSFXDefs:     make(map[defs.FootstepSFXDefID]defs.FootstepSFXDef),
		AttributeDefs:       make(map[defs.AttributeID]defs.AttributeDef),
		SkillDefs:           make(map[defs.SkillID]defs.SkillDef),
		TraitDefs:           make(map[defs.TraitID]defs.Trait),
		DialogProfiles:      make(map[defs.DialogProfileID]*defs.DialogProfileDef),
		DialogTopics:        make(map[defs.TopicID]*defs.DialogTopic),
		DialogProfileStates: make(map[defs.DialogProfileID]*state.DialogProfileState),
		CharacterDefs:       make(map[defs.CharacterDefID]defs.CharacterDef),
		CharacterStates:     make(map[state.CharacterStateID]*state.CharacterState),
		NPCSchedules:        make(map[defs.ScheduleID]defs.ScheduleDef),
	}
	return &def
}

func (defMgr *DefinitionManager) LoadScheduleDef(sched defs.ScheduleDef) {
	if sched.ID == "" {
		panic("id was empty")
	}
	if _, exists := defMgr.NPCSchedules[sched.ID]; exists {
		logz.Panicln("DefinitionManager", "tried to load schedule def, but id already exists:", sched.ID)
	}
	sched.Validate()
	defMgr.NPCSchedules[sched.ID] = sched
}

func (defMgr *DefinitionManager) GetScheduleDef(id defs.ScheduleID) defs.ScheduleDef {
	if id == "" {
		panic("id was empty")
	}
	sched, exists := defMgr.NPCSchedules[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get schedule def, but id doesn't exist:", id)
	}
	return sched
}

func (defMgr *DefinitionManager) LoadCharacterDef(charDef defs.CharacterDef) {
	if charDef.ID == "" {
		panic("id was empty")
	}

	if _, exists := defMgr.CharacterDefs[charDef.ID]; exists {
		logz.Panicln("DefinitionManager", "tried to load character def, but ID already exists:", charDef.ID)
	}

	defMgr.CharacterDefs[charDef.ID] = charDef
}

func (defMgr *DefinitionManager) GetCharacterDef(id defs.CharacterDefID) defs.CharacterDef {
	if id == "" {
		panic("id was empty")
	}
	def, exists := defMgr.CharacterDefs[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get characterDef, but ID was not found", id)
	}
	return def
}

func (defMgr *DefinitionManager) LoadCharacterState(charState *state.CharacterState) {
	if charState == nil {
		panic("charstate was nil")
	}
	if charState.ID == "" {
		panic("id was empty")
	}
	if _, exists := defMgr.CharacterStates[charState.ID]; exists {
		logz.Panicln("DefinitionManager", "tried to load character state, but ID already exists:", charState.ID)
	}
	defMgr.CharacterStates[charState.ID] = charState
}

func (defMgr *DefinitionManager) GetCharacterState(id state.CharacterStateID) *state.CharacterState {
	if id == "" {
		panic("id was empty")
	}
	charState, exists := defMgr.CharacterStates[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get character state, but ID was not found:", id)
	}
	return charState
}

// GetNewCharStateID generates a new and unique CharacterStateID that is guaranteed to not be defined in definitionMgr yet.
// Also uses the charDefID as its base, for convenience and search-ability
func (defMgr DefinitionManager) GetNewCharStateID(defID defs.CharacterDefID) state.CharacterStateID {
	charDef := defMgr.GetCharacterDef(defID)
	var id state.CharacterStateID
	if charDef.Unique {
		// if unique, we just use the def ID
		id = state.CharacterStateID(defID)
	} else {
		id = state.CharacterStateID(fmt.Sprintf("%s_%s", defID, general_util.GenerateUUID()))
	}
	if _, exists := defMgr.CharacterStates[id]; exists {
		if charDef.Unique {
			logz.Panicln("DefinitionManager", "Tried to get charStateID for a unique character def, but the base characterDefID was already taken. ensure this characterDef is not instantiated more than once:", defID)
		}
		logz.Panicln("DefinitionManager", "Generated Unique ID for Character State was already registered... is this possible?")
	}
	return id
}

func (defMgr *DefinitionManager) LoadDialogTopic(topic *defs.DialogTopic) {
	if topic.ID == "" {
		logz.Panicln("DefinitionManager", "tried to load dialog topic, but ID was empty")
	}
	if _, exists := defMgr.DialogTopics[topic.ID]; exists {
		logz.Panicln("DefinitionManager", "tried to load dialog topic, but a topic with the same ID already exists:", topic.ID)
	}
	topic.Validate()
	defMgr.DialogTopics[topic.ID] = topic
}

func (defMgr DefinitionManager) GetDialogTopic(id defs.TopicID) *defs.DialogTopic {
	topic, exists := defMgr.DialogTopics[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get dialog topic, but no topic with the given id was found:", id)
	}
	return topic
}

func (defMgr *DefinitionManager) LoadDialogProfile(profile *defs.DialogProfileDef) {
	if profile.ProfileID == "" {
		logz.Panicln("DefinitionManager", "tried to load dialog profile, but profile ID was empty")
	}
	if _, exists := defMgr.DialogProfiles[profile.ProfileID]; exists {
		logz.Panicln("DefinitionManager", "tried to load dialog profile, but a profile with the same ID already exists:", profile.ProfileID)
	}
	defMgr.DialogProfiles[profile.ProfileID] = profile
}

func (defMgr DefinitionManager) GetDialogProfile(id defs.DialogProfileID) *defs.DialogProfileDef {
	profile, exists := defMgr.DialogProfiles[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get dialog profile, but no profile with the given id was found:", id)
	}
	return profile
}

func (defMgr *DefinitionManager) LoadDialogProfileState(profileState *state.DialogProfileState) {
	if profileState.ProfileID == "" {
		logz.Panicln("DefinitionManager", "tried to load dialog profile state, but profile ID was empty")
	}
	if _, exists := defMgr.DialogProfileStates[profileState.ProfileID]; exists {
		logz.Panicln("DefinitionManager", "tried to load dialog profile state, but a profile with the same ID already exists:", profileState.ProfileID)
	}
	defMgr.DialogProfileStates[profileState.ProfileID] = profileState
}

func (defMgr DefinitionManager) GetDialogProfileState(id defs.DialogProfileID) *state.DialogProfileState {
	profileState, exists := defMgr.DialogProfileStates[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get dialog profile state, but no profile with the given id was found:", id)
	}
	return profileState
}

func (defMgr DefinitionManager) DialogProfileStateExists(id defs.DialogProfileID) bool {
	if id == "" {
		panic("id was empty")
	}
	_, exists := defMgr.DialogProfileStates[id]
	return exists
}

func (defMgr *DefinitionManager) LoadTraitDef(trait defs.Trait) {
	if trait.GetID() == "" {
		logz.Panicln("DefinitionManager", "tried to load trait, but ID was empty")
	}
	if _, exists := defMgr.TraitDefs[trait.GetID()]; exists {
		logz.Panicln("DefinitionManager", "tried to load in a new trait, but an existing trait of the same ID already exists:", trait.GetID())
	}
	defMgr.TraitDefs[trait.GetID()] = trait
}

func (defMgr DefinitionManager) GetTraitDef(id defs.TraitID) defs.Trait {
	trait, exists := defMgr.TraitDefs[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get trait by ID that doesn't exist:", id)
	}
	return trait
}

func (defMgr *DefinitionManager) LoadAttributeDef(attr defs.AttributeDef) {
	if attr.ID == "" {
		logz.Panicln("DefinitionManager", "tried to load attribute, but ID was empty")
	}
	if _, exists := defMgr.AttributeDefs[attr.ID]; exists {
		logz.Panicln("DefinitionManager", "tried to load in a new attribute, but an existing attribute of the same ID already exists:", attr.ID)
	}
	defMgr.AttributeDefs[attr.ID] = attr
}

func (defMgr DefinitionManager) GetAttributeDef(id defs.AttributeID) defs.AttributeDef {
	attrDef, exists := defMgr.AttributeDefs[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get attribute by ID that doesn't exist:", id)
	}
	return attrDef
}

func (defMgr *DefinitionManager) LoadSkillDef(sk defs.SkillDef) {
	if sk.ID == "" {
		logz.Panicln("DefinitionManager", "tried to load skill, but ID was empty")
	}
	if _, exists := defMgr.SkillDefs[sk.ID]; exists {
		logz.Panicln("DefinitionManager", "tried to load in a new skill, but an existing skill of the same ID already exists:", sk.ID)
	}
	defMgr.SkillDefs[sk.ID] = sk
}

func (defMgr DefinitionManager) GetSkillDef(id defs.SkillID) defs.SkillDef {
	skillDef, exists := defMgr.SkillDefs[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to get skill by ID that doesn't exist:", id)
	}
	return skillDef
}

func (def *DefinitionManager) LoadBodyPartDef(partDef defs.SelectedPartDef) {
	if partDef.ID == "" {
		logz.Panicln("DefinitionManager", "tried to load body part def, but ID is empty")
	}
	if _, exists := def.BodyPartDefs[partDef.ID]; exists {
		logz.Panicln("DefinitionManager", "tried to load in a new entity body part def, but the id already exists:", partDef.ID)
	}
	def.BodyPartDefs[partDef.ID] = partDef
}

func (def DefinitionManager) GetBodyPartDef(id defs.BodyPartID) defs.SelectedPartDef {
	partDef, exists := def.BodyPartDefs[id]
	if !exists {
		logz.Panicln("DefinitionManager", "entity body part def not found:", id)
	}
	return partDef
}

func (def *DefinitionManager) LoadItemDefs(itemDefs []defs.ItemDef) {
	for _, itemDef := range itemDefs {
		itemDef.Validate()
		id := itemDef.GetID()
		itemDef.Load()
		def.ItemDefs[id] = itemDef
	}
}

func (def *DefinitionManager) GetItemDef(defID defs.ItemID) defs.ItemDef {
	itemDef, exists := def.ItemDefs[defID]
	if !exists {
		logz.Panicln("DefinitionManager", "item def not found:", defID)
	}
	return itemDef
}

func (def *DefinitionManager) NewInventoryItem(defID defs.ItemID, quantity int) defs.InventoryItem {
	if quantity <= 0 {
		panic("quantity must be a positive number")
	}
	itemDef := def.GetItemDef(defID)
	if itemDef == nil {
		panic("item def is nil")
	}

	return defs.NewInventoryItem(itemDef, quantity)
}

func (def *DefinitionManager) LoadShopkeeperDef(sk defs.ShopkeeperDef) {
	def.ShopkeeperDefs[sk.ID] = &sk
}

func (def DefinitionManager) GetShopkeeperDef(id defs.ShopID) *defs.ShopkeeperDef {
	shopkeeper, exists := def.ShopkeeperDefs[id]
	if !exists {
		logz.Panicf("shopkeeperID not found in defintionManager: %s", id)
	}
	return shopkeeper
}

func (def *DefinitionManager) LoadShopkeeperState(sk state.ShopkeeperState) {
	def.ShopkeeperStates[sk.ShopID] = &sk
}

func (def DefinitionManager) GetShopkeeperState(id defs.ShopID) *state.ShopkeeperState {
	shopkeeper, exists := def.ShopkeeperStates[id]
	if !exists {
		logz.Panicf("shopkeeperID not found in defintionManager: %s", id)
	}
	return shopkeeper
}

func (def *DefinitionManager) LoadFootstepSFXDef(sfxDef defs.FootstepSFXDef) {
	id := sfxDef.ID
	if id == "" {
		panic("id was empty")
	}

	if _, exists := def.FootstepSFXDefs[id]; exists {
		logz.Panicln("DefinitionManager", "tried to load footstep sfx def that already exists:", id)
	}
	def.FootstepSFXDefs[id] = sfxDef
}

func (def DefinitionManager) GetFootstepSFXDef(id defs.FootstepSFXDefID) defs.FootstepSFXDef {
	if id == "" {
		panic("id was empty")
	}
	sfxDef, exists := def.FootstepSFXDefs[id]
	if !exists {
		logz.Panicln("DefinitionManager", "tried to load footstepSfxDef that doesn't exist yet:", id)
	}
	return sfxDef
}
