// Package state defines stately state in a stateful estate
package state

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/internal/logz"
)

type (
	NPCID            string
	CharacterStateID string
)

type CharacterState struct {
	// Name, Identity

	DisplayName string              // the actual name of the entity, as displayed in game to players
	ID          CharacterStateID    // the unique identifier of this entity (not usually seen by players - only by developers)
	DefID       defs.CharacterDefID // the ID of the characterDef used for creating this character state originally. Use this to look things up like which dialogProfileID to use.
	IsPlayer    bool                `json:"-"` // flag indicating if this entity is the player RUNTIME

	// Inventory and items
	// Each InventoryItem slot represents an actual individual item slot in the inventory.
	// Equiped armor, weapon, rings, amulets, arrows, etc all have their own specific slots.
	// Other inventory items just go in the array.
	// Note: as of now, we are not doing the "top level scrollable items row" concept that you see in Minecraft or SDV.
	// Could implement that later, but for now there are just specific slots for equipped items, and then an array of items for the rest of your inventory.

	defs.StandardInventory

	// InventoryItems []*defs.InventoryItem

	// items in the entity's coin purse. note that coins can also be put in regular inventory slots.
	// the coin purse is effectively disabled if its size is 0.
	// CoinPurse []*defs.InventoryItem

	// defs.EquipedItems

	// Attributes, Skills

	Vitals         defs.Vitals
	BaseAttributes map[defs.AttributeID]int // Base attribute levels (not including modifiers from traits, etc)
	BaseSkills     map[defs.SkillID]int     // Base skill levels (not including modifiers from traits, etc)
	Traits         []defs.TraitID
}

// WalkSpeed returns a walking speed, calculated by character stats (chiefly Agility)
// value should be a TileSize / NumFrames calculation (probably? this was originally suggested by ChatGPT a while back)
func (cs CharacterState) WalkSpeed() float64 {
	// TODO: calculate walk speed
	return 1.0
}

func (cs CharacterState) RunSpeed() float64 {
	// TODO: calculate run speed
	return 2.0
}

func (cd CharacterState) Validate() {
	if cd.ID == "" {
		logz.Panicln(cd.DisplayName, "entity ID must not be empty!")
	}
	if cd.DisplayName == "" {
		logz.Panicln(cd.DisplayName, "entity displayName must not be empty!")
	}
	if cd.WalkSpeed() == 0 {
		logz.Panicln(cd.DisplayName, "walk speed is 0")
	}
	if cd.RunSpeed() == 0 {
		logz.Panicln(cd.DisplayName, "run speed is 0")
	}
	if cd.WalkSpeed() == cd.RunSpeed() {
		logz.Panicln(cd.DisplayName, "run speed and walk speed are the same:", cd.RunSpeed())
	}
	if len(cd.InventoryItems) == 0 {
		logz.Panicln(cd.DisplayName, "inventory size is 0")
	}

	cd.StandardInventory.Validate()
}

// SetInventoryItems sets all the inventory items of an entity
func (cd *CharacterState) SetInventoryItems(invItems []*defs.InventoryItem) {
	cd.StandardInventory.SetInventoryItems(invItems)
}

func (cd *CharacterState) SetCoinPurseItems(invItems []*defs.InventoryItem) {
	cd.StandardInventory.SetCoinPurseItems(invItems)
}

func (cd CharacterState) CountMoney() int {
	return cd.StandardInventory.CountMoney()
}

func (cd *CharacterState) UnequipHeadwear() {
	if cd.Equipment.EquipedHeadwear == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip headwear, but equiped headwear is nil")
	}

	cd.Equipment.EquipedHeadwear = nil
}

func (cd *CharacterState) UnequipFootwear() {
	if cd.Equipment.EquipedFootwear == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip footwear, but equiped footwear is nil")
	}

	cd.Equipment.EquipedFootwear = nil
}

func (cd *CharacterState) UnequipBodywear() {
	if cd.Equipment.EquipedBodywear == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip bodywear, but equiped bodywear is nil")
	}

	cd.Equipment.EquipedBodywear = nil
}

func (cd *CharacterState) UnequipAuxiliary() {
	if cd.Equipment.EquipedAuxiliary == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip auxiliary, but equiped auxiliary is nil")
	}

	cd.Equipment.EquipedAuxiliary = nil
}

func (cd *CharacterState) UnequipWeapon() {
	if cd.Equipment.EquipedWeapon == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip weapon, but equiped weapon is nil")
	}

	cd.Equipment.EquipedWeapon = nil
	// cd.Body.RemoveWeapon()
}

/*
*
* We need to split NPC into the following concepts:
*
* NPC Definition:
* - Name, Culture, Factions (CharacterData JSON)
* - Base Stats (CharacterData JSON)
*   - Q: should we just store a CharacterDataID that loads the JSON file?
* - Inventory template (starting items) (CharacterData)
*   - this is loaded from JSON, but we wont' persist this; as the NPC gets new items, we will just overwrite.
*   - Q: would it be better to define specific item sets? could make character creation faster.
* - DialogProfileID
* - anything immutable at runtime that we define when we create the character
*
* future:
* - Schedule def
* - Class
*
*
* NPC State:
* - persistent, save-game data
* - health (CharacterData)
* - inventory (actual items) (CharacterData)
*
* future:
* - schedule progress
* - current location
* - alive / dead
* - disposition to player
*
* NPC Runtime / Controller data should not be in either of these. that can go in the plain NPC struct, which will hold the above two structs too.
*
 */

// type NPCDef struct {
// 	ID              NPCID
// 	CharacterID     CharacterID
// 	DialogProfileID DialogProfileID
// }
//
// type NPCState struct {
// 	// CharacterData entity.CharacterData
// }
