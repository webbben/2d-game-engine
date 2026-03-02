// Package datamanager provides a centralized place for fetching definitions of items, shopkeepers, dialogs, etc.
package datamanager

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/general_util"
	"github.com/webbben/2d-game-engine/logz"
)

type DataManager struct {
	MapDefs   map[defs.MapID]defs.MapDef
	MapStates map[defs.MapID]*state.MapState

	ScenarioDef map[defs.ScenarioID]defs.ScenarioDef

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
	ClassDefs     map[defs.ClassDefID]defs.ClassDef
	CultureDefs   map[defs.CultureID]defs.CultureDef
}

func NewDataManager() *DataManager {
	dataman := DataManager{
		ItemDefs:            make(map[defs.ItemID]defs.ItemDef),
		MapDefs:             make(map[defs.MapID]defs.MapDef),
		MapStates:           make(map[defs.MapID]*state.MapState),
		ScenarioDef:         make(map[defs.ScenarioID]defs.ScenarioDef),
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
		CultureDefs:         make(map[defs.CultureID]defs.CultureDef),
		ClassDefs:           make(map[defs.ClassDefID]defs.ClassDef),
	}
	return &dataman
}

func (dataman *DataManager) LoadCultureDef(def defs.CultureDef) {
	if def.ID == "" {
		panic("id was empty")
	}
	if _, exists := dataman.CultureDefs[def.ID]; exists {
		logz.Panicln("DataManager", "tried to load culture def, but id already exists:", def.ID)
	}
	dataman.CultureDefs[def.ID] = def
}

func (dataman *DataManager) GetCultureDef(id defs.CultureID) defs.CultureDef {
	if id == "" {
		panic("id was empty")
	}
	def, exists := dataman.CultureDefs[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get culture def, but id doesn't exist:", id)
	}
	return def
}

func (dataman *DataManager) LoadClassDef(def defs.ClassDef) {
	if def.ID == "" {
		panic("id was empty")
	}
	if _, exists := dataman.ClassDefs[def.ID]; exists {
		logz.Panicln("DataManager", "tried to load class def, but id already exists:", def.ID)
	}
	dataman.ClassDefs[def.ID] = def
}

func (dataman *DataManager) GetClassDef(id defs.ClassDefID) defs.ClassDef {
	if id == "" {
		panic("id was empty")
	}
	def, exists := dataman.ClassDefs[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get class def, but id doesn't exist:", id)
	}
	return def
}

func (dataman *DataManager) LoadScenarioDef(def defs.ScenarioDef) {
	if def.ID == "" {
		panic("id was empty")
	}
	if _, exists := dataman.ScenarioDef[def.ID]; exists {
		logz.Panicln("DataManager", "tried to load scenario def, but id already exists:", def.ID)
	}
	dataman.ScenarioDef[def.ID] = def
}

func (dataman *DataManager) GetScenarioDef(id defs.ScenarioID) defs.ScenarioDef {
	if id == "" {
		panic("id was empty")
	}
	def, exists := dataman.ScenarioDef[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get scenario def, but id doesn't exist:", id)
	}
	return def
}

func (dataman *DataManager) LoadMapDef(def defs.MapDef) {
	if def.ID == "" {
		panic("id was empty")
	}
	if _, exists := dataman.MapDefs[def.ID]; exists {
		logz.Panicln("DataManager", "tried to load map def, but id already exists:", def.ID)
	}
	dataman.MapDefs[def.ID] = def
}

func (dataman *DataManager) GetMapDef(id defs.MapID) defs.MapDef {
	if id == "" {
		panic("id was empty")
	}
	def, exists := dataman.MapDefs[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get map def, but id doesn't exist:", id)
	}
	return def
}

func (dataman *DataManager) LoadMapState(mapState state.MapState) {
	if mapState.ID == "" {
		panic("id was empty")
	}
	if _, exists := dataman.MapStates[mapState.ID]; exists {
		logz.Panicln("DataManager", "tried to load map state, but id already exists:", mapState.ID)
	}
	dataman.MapStates[mapState.ID] = &mapState
}

func (dataman *DataManager) GetMapState(id defs.MapID) *state.MapState {
	if id == "" {
		panic("id was empty")
	}
	mapState, exists := dataman.MapStates[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get map state, but id doesn't exist:", id)
	}
	return mapState
}

func (dataman *DataManager) MapStateExists(id defs.MapID) bool {
	_, exists := dataman.MapStates[id]
	return exists
}

func (dataman *DataManager) LoadScheduleDef(sched defs.ScheduleDef) {
	if sched.ID == "" {
		panic("id was empty")
	}
	if _, exists := dataman.NPCSchedules[sched.ID]; exists {
		logz.Panicln("DataManager", "tried to load schedule def, but id already exists:", sched.ID)
	}
	sched.Validate()
	dataman.NPCSchedules[sched.ID] = sched
}

func (dataman *DataManager) GetScheduleDef(id defs.ScheduleID) defs.ScheduleDef {
	if id == "" {
		panic("id was empty")
	}
	sched, exists := dataman.NPCSchedules[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get schedule def, but id doesn't exist:", id)
	}
	return sched
}

func (dataman *DataManager) LoadCharacterDef(charDef defs.CharacterDef) {
	if charDef.ID == "" {
		panic("id was empty")
	}

	if _, exists := dataman.CharacterDefs[charDef.ID]; exists {
		logz.Panicln("DataManager", "tried to load character def, but ID already exists:", charDef.ID)
	}

	dataman.CharacterDefs[charDef.ID] = charDef
}

func (dataman *DataManager) GetCharacterDef(id defs.CharacterDefID) defs.CharacterDef {
	if id == "" {
		panic("id was empty")
	}
	def, exists := dataman.CharacterDefs[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get characterDef, but ID was not found", id)
	}
	return def
}

func (dataman *DataManager) LoadCharacterState(charState *state.CharacterState) {
	if charState == nil {
		panic("charstate was nil")
	}
	if charState.ID == "" {
		panic("id was empty")
	}
	if _, exists := dataman.CharacterStates[charState.ID]; exists {
		logz.Panicln("DataManager", "tried to load character state, but ID already exists:", charState.ID)
	}
	dataman.CharacterStates[charState.ID] = charState
}

func (dataman *DataManager) GetCharacterState(id state.CharacterStateID) *state.CharacterState {
	if id == "" {
		panic("id was empty")
	}
	charState, exists := dataman.CharacterStates[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get character state, but ID was not found:", id)
	}
	return charState
}

// GetNewCharStateID generates a new and unique CharacterStateID that is guaranteed to not be defined in definitionMgr yet.
// Also uses the charDefID as its base, for convenience and search-ability
func (dataman DataManager) GetNewCharStateID(defID defs.CharacterDefID) state.CharacterStateID {
	charDef := dataman.GetCharacterDef(defID)
	var id state.CharacterStateID
	if charDef.Unique {
		// if unique, we just use the def ID
		id = state.CharacterStateID(defID)
	} else {
		id = state.CharacterStateID(fmt.Sprintf("%s_%s", defID, general_util.GenerateUUID()[:8]))
	}
	if _, exists := dataman.CharacterStates[id]; exists {
		if charDef.Unique {
			logz.Panicln("DataManager", "Tried to get charStateID for a unique character def, but the base characterDefID was already taken. ensure this characterDef is not instantiated more than once:", defID)
		}
		logz.Panicln("DataManager", "Generated Unique ID for Character State was already registered... is this possible?")
	}
	return id
}

func (dataman *DataManager) LoadDialogTopic(topic *defs.DialogTopic) {
	if topic.ID == "" {
		logz.Panicln("DataManager", "tried to load dialog topic, but ID was empty")
	}
	if _, exists := dataman.DialogTopics[topic.ID]; exists {
		logz.Panicln("DataManager", "tried to load dialog topic, but a topic with the same ID already exists:", topic.ID)
	}
	topic.Validate()
	dataman.DialogTopics[topic.ID] = topic
}

func (dataman DataManager) GetDialogTopic(id defs.TopicID) *defs.DialogTopic {
	topic, exists := dataman.DialogTopics[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get dialog topic, but no topic with the given id was found:", id)
	}
	return topic
}

func (dataman *DataManager) LoadDialogProfile(profile *defs.DialogProfileDef) {
	if profile.ProfileID == "" {
		logz.Panicln("DataManager", "tried to load dialog profile, but profile ID was empty")
	}
	if _, exists := dataman.DialogProfiles[profile.ProfileID]; exists {
		logz.Panicln("DataManager", "tried to load dialog profile, but a profile with the same ID already exists:", profile.ProfileID)
	}
	dataman.DialogProfiles[profile.ProfileID] = profile
}

func (dataman DataManager) GetDialogProfile(id defs.DialogProfileID) *defs.DialogProfileDef {
	profile, exists := dataman.DialogProfiles[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get dialog profile, but no profile with the given id was found:", id)
	}
	return profile
}

func (dataman *DataManager) LoadDialogProfileState(profileState *state.DialogProfileState) {
	if profileState.ProfileID == "" {
		logz.Panicln("DataManager", "tried to load dialog profile state, but profile ID was empty")
	}
	if _, exists := dataman.DialogProfileStates[profileState.ProfileID]; exists {
		logz.Panicln("DataManager", "tried to load dialog profile state, but a profile with the same ID already exists:", profileState.ProfileID)
	}
	dataman.DialogProfileStates[profileState.ProfileID] = profileState
}

func (dataman DataManager) GetDialogProfileState(id defs.DialogProfileID) *state.DialogProfileState {
	profileState, exists := dataman.DialogProfileStates[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get dialog profile state, but no profile with the given id was found:", id)
	}
	return profileState
}

func (dataman DataManager) DialogProfileStateExists(id defs.DialogProfileID) bool {
	if id == "" {
		panic("id was empty")
	}
	_, exists := dataman.DialogProfileStates[id]
	return exists
}

func (dataman *DataManager) LoadTraitDef(trait defs.Trait) {
	if trait.GetID() == "" {
		logz.Panicln("DataManager", "tried to load trait, but ID was empty")
	}
	if _, exists := dataman.TraitDefs[trait.GetID()]; exists {
		logz.Panicln("DataManager", "tried to load in a new trait, but an existing trait of the same ID already exists:", trait.GetID())
	}
	dataman.TraitDefs[trait.GetID()] = trait
}

func (dataman DataManager) GetTraitDef(id defs.TraitID) defs.Trait {
	trait, exists := dataman.TraitDefs[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get trait by ID that doesn't exist:", id)
	}
	return trait
}

func (dataman *DataManager) LoadAttributeDef(attr defs.AttributeDef) {
	if attr.ID == "" {
		logz.Panicln("DataManager", "tried to load attribute, but ID was empty")
	}
	if _, exists := dataman.AttributeDefs[attr.ID]; exists {
		logz.Panicln("DataManager", "tried to load in a new attribute, but an existing attribute of the same ID already exists:", attr.ID)
	}
	dataman.AttributeDefs[attr.ID] = attr
}

func (dataman DataManager) GetAttributeDef(id defs.AttributeID) defs.AttributeDef {
	attrDef, exists := dataman.AttributeDefs[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get attribute by ID that doesn't exist:", id)
	}
	return attrDef
}

func (dataman *DataManager) LoadSkillDef(sk defs.SkillDef) {
	if sk.ID == "" {
		logz.Panicln("DataManager", "tried to load skill, but ID was empty")
	}
	if _, exists := dataman.SkillDefs[sk.ID]; exists {
		logz.Panicln("DataManager", "tried to load in a new skill, but an existing skill of the same ID already exists:", sk.ID)
	}
	dataman.SkillDefs[sk.ID] = sk
}

func (dataman DataManager) GetSkillDef(id defs.SkillID) defs.SkillDef {
	if id == "" {
		panic("id was empty")
	}
	skillDef, exists := dataman.SkillDefs[id]
	if !exists {
		logz.Panicln("DataManager", "tried to get skill by ID that doesn't exist:", id)
	}
	return skillDef
}

func (dataman *DataManager) LoadBodyPartDef(partDef defs.SelectedPartDef) {
	if partDef.ID == "" {
		logz.Panicln("DataManager", "tried to load body part def, but ID is empty")
	}
	if _, exists := dataman.BodyPartDefs[partDef.ID]; exists {
		logz.Panicln("DataManager", "tried to load in a new entity body part def, but the id already exists:", partDef.ID)
	}
	dataman.BodyPartDefs[partDef.ID] = partDef
}

func (dataman DataManager) GetBodyPartDef(id defs.BodyPartID) defs.SelectedPartDef {
	partDef, exists := dataman.BodyPartDefs[id]
	if !exists {
		logz.Panicln("DataManager", "entity body part def not found:", id)
	}
	return partDef
}

func (dataman *DataManager) LoadItemDefs(itemDefs []defs.ItemDef) {
	for _, itemDef := range itemDefs {
		itemDef.Validate()
		id := itemDef.GetID()
		itemDef.Load()
		dataman.ItemDefs[id] = itemDef
	}
}

func (dataman *DataManager) GetItemDef(defID defs.ItemID) defs.ItemDef {
	itemDef, exists := dataman.ItemDefs[defID]
	if !exists {
		logz.Panicln("DataManager", "item def not found:", defID)
	}
	return itemDef
}

func (dataman *DataManager) NewInventoryItem(defID defs.ItemID, quantity int) defs.InventoryItem {
	if quantity <= 0 {
		panic("quantity must be a positive number")
	}
	itemDef := dataman.GetItemDef(defID)
	if itemDef == nil {
		panic("item def is nil")
	}

	return defs.NewInventoryItem(itemDef, quantity)
}

func (dataman *DataManager) LoadShopkeeperDef(sk defs.ShopkeeperDef) {
	dataman.ShopkeeperDefs[sk.ID] = &sk
}

func (dataman *DataManager) GetShopkeeperDef(id defs.ShopID) *defs.ShopkeeperDef {
	shopkeeper, exists := dataman.ShopkeeperDefs[id]
	if !exists {
		logz.Panicf("shopkeeperID not found in defintionManager: %s", id)
	}
	return shopkeeper
}

func (dataman *DataManager) LoadShopkeeperState(sk state.ShopkeeperState) {
	dataman.ShopkeeperStates[sk.ShopID] = &sk
}

func (dataman *DataManager) GetShopkeeperState(id defs.ShopID) *state.ShopkeeperState {
	shopkeeper, exists := dataman.ShopkeeperStates[id]
	if !exists {
		logz.Panicf("shopkeeperID not found in defintionManager: %s", id)
	}
	return shopkeeper
}

func (dataman *DataManager) LoadFootstepSFXDef(sfxDef defs.FootstepSFXDef) {
	id := sfxDef.ID
	if id == "" {
		panic("id was empty")
	}

	if _, exists := dataman.FootstepSFXDefs[id]; exists {
		logz.Panicln("DataManager", "tried to load footstep sfx def that already exists:", id)
	}
	dataman.FootstepSFXDefs[id] = sfxDef
}

func (dataman *DataManager) GetFootstepSFXDef(id defs.FootstepSFXDefID) defs.FootstepSFXDef {
	if id == "" {
		panic("id was empty")
	}
	sfxDef, exists := dataman.FootstepSFXDefs[id]
	if !exists {
		logz.Panicln("DataManager", "tried to load footstepSfxDef that doesn't exist yet:", id)
	}
	return sfxDef
}
