package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/item"
)

// CharacterData contains all the info and data about a character, excluding things like mechanics and flags, etc.
//
// Things like the character's identity (name, ID, etc), the character's body state (visible appearance, hair, eyes, etc),
// the items the character has in its inventory, etc. Basically, all the data needed to actually save and load this character.
// (For example, the character builder only saves the data in this struct, and doesn't refer to anything else in an entity)
//
// The inner mechanisms used for things like movement, combat, etc are not included here.
type CharacterData struct {
	// Name, Identity

	DisplayName string // the actual name of the entity, as displayed in game to players
	ID          string // the unique identifier of this entity (not usually seen by players - only by developers)
	IsPlayer    bool   `json:"-"` // flag indicating if this entity is the player

	// Inventory and items

	InventoryItems []*item.InventoryItem

	EquipedHeadwear  *item.InventoryItem
	EquipedBodywear  *item.InventoryItem
	EquipedFootwear  *item.InventoryItem
	EquipedAmulet    *item.InventoryItem
	EquipedRing1     *item.InventoryItem
	EquipedRing2     *item.InventoryItem
	EquipedAmmo      *item.InventoryItem
	EquipedAuxiliary *item.InventoryItem

	// Body

	Body body.EntityBodySet

	// Attributes, Skills

	Vitals     Vitals
	Attributes Attributes

	WalkSpeed float64 `json:"walk_speed"` // value should be a TileSize / NumFrames calculation
	RunSpeed  float64 `json:"run_speed"`
}

func (cd CharacterData) WriteToJSON(outputFilePath string) error {
	if !filepath.IsAbs(outputFilePath) {
		return fmt.Errorf("given path is not abs (%s); please pass an absolute path", outputFilePath)
	}

	// Note: ItemDefs are not allowed to be written to JSONs since they can't be loaded from JSON properly.
	// That is handled by setting the Def field of inventoryItems to json:"-"

	data, err := json.MarshalIndent(cd, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(outputFilePath, data, 0o644)
}

func LoadCharacterDataJSON(src string, defMgr *definitions.DefinitionManager) (CharacterData, error) {
	if !config.FileExists(src) {
		return CharacterData{}, errors.New("no file found at path: " + src)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return CharacterData{}, fmt.Errorf("failed to read file data: %w", err)
	}

	var cd CharacterData
	err = json.Unmarshal(data, &cd)
	if err != nil {
		return CharacterData{}, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	// Load actual ItemDefs from DefinitionManager
	if cd.EquipedAmmo != nil {
		cd.EquipedAmmo.Def = defMgr.GetItemDef(cd.EquipedAmmo.Instance.DefID)
	}
	if cd.EquipedAmulet != nil {
		cd.EquipedAmulet.Def = defMgr.GetItemDef(cd.EquipedAmulet.Instance.DefID)
	}
	if cd.EquipedAuxiliary != nil {
		cd.EquipedAuxiliary.Def = defMgr.GetItemDef(cd.EquipedAuxiliary.Instance.DefID)
	}
	if cd.EquipedBodywear != nil {
		cd.EquipedBodywear.Def = defMgr.GetItemDef(cd.EquipedBodywear.Instance.DefID)
	}
	if cd.EquipedFootwear != nil {
		cd.EquipedFootwear.Def = defMgr.GetItemDef(cd.EquipedFootwear.Instance.DefID)
	}
	if cd.EquipedHeadwear != nil {
		cd.EquipedHeadwear.Def = defMgr.GetItemDef(cd.EquipedHeadwear.Instance.DefID)
	}
	if cd.EquipedRing1 != nil {
		cd.EquipedRing1.Def = defMgr.GetItemDef(cd.EquipedRing1.Instance.DefID)
	}
	if cd.EquipedRing2 != nil {
		cd.EquipedRing2.Def = defMgr.GetItemDef(cd.EquipedRing2.Instance.DefID)
	}

	for _, i := range cd.InventoryItems {
		if i == nil {
			continue
		}
		i.Def = defMgr.GetItemDef(i.Instance.DefID)
	}

	return cd, nil
}

// SetInventoryItems sets all the inventory items of an entity
func (cd *CharacterData) SetInventoryItems(invItems []*item.InventoryItem) {
	cd.InventoryItems = make([]*item.InventoryItem, 0)

	for _, newItem := range invItems {
		if newItem == nil {
			cd.InventoryItems = append(cd.InventoryItems, nil)
			continue
		}
		newItem.Validate()
		cd.InventoryItems = append(cd.InventoryItems, &item.InventoryItem{
			Instance: newItem.Instance,
			Def:      newItem.Def,
			Quantity: newItem.Quantity,
		})
	}
}

func (cd *CharacterData) AddItemToInventory(invItem item.InventoryItem) (bool, item.InventoryItem) {
	invItem.Validate()
	success, remaining := item.AddItemToInventory(invItem, cd.InventoryItems)
	cd.Validate()
	return success, remaining
}

func (cd *CharacterData) RemoveItemFromInventory(itemToRemove item.InventoryItem) (bool, item.InventoryItem) {
	itemToRemove.Validate()
	success, remaining := item.RemoveItemFromInventory(itemToRemove, cd.InventoryItems)
	cd.Validate()
	return success, remaining
}

// EquipItem equips a weapon, body armor, clothes, or other equipable items that go onto the entity's body or equipment slots
func (cd *CharacterData) EquipItem(i item.InventoryItem) (success bool) {
	i.Validate()
	if !i.Def.IsEquipable() {
		logz.Panicln(cd.DisplayName, "tried to equip an inequipable item:", i.Def.GetID())
	}

	switch i.Def.GetItemType() {
	case item.TypeHeadwear:
		if cd.EquipedHeadwear != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := cd.AddItemToInventory(*cd.EquipedHeadwear)
			if !succ {
				return false
			}
		}
		cd.EquipedHeadwear = &i
		part := i.Def.GetBodyPartDef()
		if part == nil {
			logz.Panicln(cd.DisplayName, "tried to equip an item with no part def:", i.Def.GetID())
		}
		cd.Body.SetEquipHead(*part)
	case item.TypeBodywear:
		if cd.EquipedBodywear != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := cd.AddItemToInventory(*cd.EquipedBodywear)
			if !succ {
				return false
			}
		}
		cd.EquipedBodywear = &i
		part := i.Def.GetBodyPartDef()
		if part == nil {
			logz.Panicln(cd.DisplayName, "tried to equip an item with no part def:", i.Def.GetID())
		}
		cd.Body.SetEquipBody(*part)
	case item.TypeWeapon:
		// weapons don't have an "equiped slot", so as of now there is no swapping to do here
		part, fxPart := item.GetWeaponParts(i.Def)

		cd.Body.SetWeapon(part, fxPart)
	case item.TypeAuxiliary:
		if cd.EquipedAuxiliary != nil {
			// already equiped; remove it and put it in a regular inventory slot
			succ, _ := cd.AddItemToInventory(*cd.EquipedAuxiliary)
			if !succ {
				return false
			}
		}
		cd.EquipedAuxiliary = &i
		part := i.Def.GetBodyPartDef()
		if part == nil {
			logz.Panicln(cd.DisplayName, "tried to equip an item with no part def:", i.Def.GetID())
		}
		cd.Body.SetAuxiliary(*part)
	default:
		logz.Panicln(cd.DisplayName, "tried to equip item, but it's type didn't match in the switch statement... (this probably should be caught by the IsEquipable check)")
	}

	cd.Validate()
	return true
}

func (cd *CharacterData) UnequipHeadwear() {
	if cd.EquipedHeadwear == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip headwear, but equiped headwear is nil")
	}

	cd.EquipedHeadwear = nil
	cd.Body.EquipHeadSet.Remove()
	// reload hair too, since it may have been cropped by the previously equiped headwear
	cd.Body.ReloadHair()
}

func (cd *CharacterData) UnequipBodywear() {
	if cd.EquipedBodywear == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip bodywear, but equiped bodywear is nil")
	}

	cd.EquipedBodywear = nil
	cd.Body.EquipBodySet.Remove()
	cd.Body.ReloadArms()
}

func (cd *CharacterData) UnequipAuxiliary() {
	if cd.EquipedAuxiliary == nil {
		logz.Panicln(cd.DisplayName, "tried to unequip auxiliary, but equiped auxiliary is nil")
	}

	cd.EquipedAuxiliary = nil
	cd.Body.RemoveAuxiliary()
}
