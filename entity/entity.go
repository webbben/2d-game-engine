package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/item"
)

var (
	defaultWalkSpeed                 float64 = float64(config.TileSize) / 18
	defaultRunSpeed                  float64 = float64(config.TileSize) / 12
	defaultWalkAnimationTickInterval         = 10
	defaultRunAnimationTickInterval          = 6
	defaultIdleAnimationTickInterval         = 10
)

func GetDefaultWalkSpeed() float64 {
	if defaultWalkSpeed == 0 {
		panic("entity default walk speed is 0?")
	}
	return defaultWalkSpeed
}
func GetDefaultRunSpeed() float64 {
	if defaultRunSpeed == 0 {
		panic("default run speed is 0?")
	}
	return defaultRunSpeed
}

type Entity struct {
	Loaded   bool     `json:"-"` // if the entity has been loaded into memory fully yet
	Movement Movement `json:"movement"`
	Position

	width float64 // used for collision rect

	footstepSFX audio.FootstepSFX

	attackManager

	World WorldContext `json:"-"`

	stunTicks int

	CharacterData
}

// contains all the info and data about a character, excluding things like mechanics and flags, etc.
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

	// ItemDefs (interfaces in general) can't be loaded from JSON, so lets nullify each of the ItemDefs.
	// Then, during loading, we can use the definitionManager to load ItemDefs by ID.
	if cd.EquipedAmmo != nil {
		cd.EquipedAmmo.Def = nil
	}
	if cd.EquipedAmulet != nil {
		cd.EquipedAmulet.Def = nil
	}
	if cd.EquipedAuxiliary != nil {
		cd.EquipedAuxiliary.Def = nil
	}
	if cd.EquipedBodywear != nil {
		cd.EquipedBodywear.Def = nil
	}
	if cd.EquipedFootwear != nil {
		cd.EquipedFootwear.Def = nil
	}
	if cd.EquipedHeadwear != nil {
		cd.EquipedHeadwear.Def = nil
	}
	if cd.EquipedRing1 != nil {
		cd.EquipedRing1.Def = nil
	}
	if cd.EquipedRing2 != nil {
		cd.EquipedRing2.Def = nil
	}

	for _, i := range cd.InventoryItems {
		if i == nil {
			continue
		}
		i.Def = nil
	}

	data, err := json.MarshalIndent(cd, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(outputFilePath, data, 0644)
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

func (e Entity) CollisionRect() model.Rect {
	if e.width == 0 {
		panic("entity width is unset or unexpectedly 0!")
	}
	offsetX := (config.TileSize - e.width) / 2
	return model.Rect{
		X: e.X + offsetX,
		Y: e.Y,
		W: e.width,
		H: e.width,
	}
}

func (e *Entity) LoadFootstepSFX(source audio.FootstepSFX) {
	e.footstepSFX = source
	e.footstepSFX.Load()
}

type WorldContext interface {
	Collides(r model.Rect, excludeEntityId string) model.CollisionResult
	FindPath(start, goal model.Coords) ([]model.Coords, bool)
	MapDimensions() (width int, height int)
	GetGroundMaterial(tileX, tileY int) string
	GetDistToPlayer(x, y float64) float64
	AttackArea(attackInfo AttackInfo)
}

type AudioProps struct {
	FootstepSFX audio.FootstepSFX
}

type GeneralProps struct {
	IsPlayer         bool
	CharacterDataSrc string // path to a JSON file containing the definition of the character data (entity body, items, etc)
	InventorySize    int
}

// Create a new entity from a character data JSON file
func NewEntity(general GeneralProps, ap AudioProps, defMgr *definitions.DefinitionManager) Entity {
	if general.CharacterDataSrc == "" {
		panic("no character data source JSON specified")
	}
	if general.InventorySize <= 0 {
		panic("inventory size must be positive")
	}

	ent := Entity{
		Movement: Movement{
			WalkAnimationTickInterval: defaultWalkAnimationTickInterval,
			RunAnimationTickInterval:  defaultRunAnimationTickInterval,
		},
	}

	// load character data, entity body
	characterData, err := LoadCharacterDataJSON(general.CharacterDataSrc, defMgr)
	if err != nil {
		panic(err)
	}
	ent.CharacterData = characterData

	ent.CharacterData.IsPlayer = general.IsPlayer

	if len(ent.CharacterData.InventoryItems) == 0 {
		ent.CharacterData.InventoryItems = make([]*item.InventoryItem, general.InventorySize)
	}
	if ent.CharacterData.WalkSpeed == 0 {
		logz.Warnln("", "loaded entity does not have a walking speed; setting default value.")
		ent.CharacterData.WalkSpeed = GetDefaultWalkSpeed()
	}
	if ent.CharacterData.RunSpeed == 0 {
		logz.Warnln("", "loaded entity does not have a run speed; setting default value.")
		ent.CharacterData.RunSpeed = GetDefaultRunSpeed()
	}

	ent.CharacterData.Validate()

	// load sounds
	ent.LoadFootstepSFX(ap.FootstepSFX)

	// prepare initial image frames
	ent.Movement.Direction = 'D'
	ent.Body.Load()
	ent.Body.SetAnimationTickCount(defaultIdleAnimationTickInterval)
	ent.Body.Name = ent.DisplayName
	ent.Movement.IsMoving = false

	// confirm that body image exists
	dx, dy := ent.Body.Dimensions()
	if dx == 0 || dy == 0 {
		panic("body image has no size?")
	}

	ent.width = float64(dx)
	if ent.width == 0 {
		panic("ent width is 0?")
	}

	ent.Loaded = true

	return ent
}

func (cd CharacterData) Validate() {
	if cd.ID == "" {
		logz.Panicln(cd.DisplayName, "entity ID must not be empty!")
	}
	if cd.DisplayName == "" {
		logz.Panicln(cd.DisplayName, "entity displayName must not be empty!")
	}
	if cd.WalkSpeed == 0 {
		logz.Panicln(cd.DisplayName, "walk speed is 0")
	}
	if cd.RunSpeed == 0 {
		logz.Panicln(cd.DisplayName, "run speed is 0")
	}
	if len(cd.InventoryItems) == 0 {
		logz.Panicln(cd.DisplayName, "inventory size is 0")
	}
	// ensure equiped items match the set body part def
	validateEquipment := func(equipedItem *item.InventoryItem, bodyPart body.BodyPartSet) {
		if equipedItem == nil {
			if !bodyPart.PartSrc.None {
				logz.Panicln(cd.DisplayName, "equipment is nil, but body part is not set to None")
			}
		} else {
			equiped := equipedItem.Def.GetBodyPartDef()
			if equiped == nil {
				logz.Panicln(cd.DisplayName, "equipment body part def was found to be nil")
			}
			if !equiped.IsEqual(bodyPart.PartSrc) {
				logz.Panicln(cd.DisplayName, "equipment item def does not appear to match body's equiped part")
			}
		}
	}
	validateEquipment(cd.EquipedHeadwear, cd.Body.EquipHeadSet)
	validateEquipment(cd.EquipedBodywear, cd.Body.EquipBodySet)
	validateEquipment(cd.EquipedAuxiliary, cd.Body.AuxItemSet)
}

type EntityInfo struct {
	DisplayName string `json:"display_name"`
	ID          string `json:"id"`
	Source      string `json:"-"` // JSON source file for this entity
	IsPlayer    bool   `json:"-"` // flag indicating if this entity is the player
}

type Position struct {
	X, Y             float64      `json:"-"` // the exact position the entity is at on the map
	drawX, drawY     float64      `json:"-"` // the actual position on the screen where the entity would be drawn
	TargetX, TargetY float64      `json:"-"` // the target position the entity is moving to
	TilePos          model.Coords `json:"-"` // the tile the entity is technically inside of
}

type Movement struct {
	Direction byte `json:"-"` // L R U D

	IsMoving                  bool    `json:"-"`
	Interrupted               bool    `json:"-"` // flag for if this entity's movement was stopped unexpectedly (e.g. by a collision)
	Speed                     float64 `json:"-"` // actual speed the entity is moving at
	WalkAnimationTickInterval int
	RunAnimationTickInterval  int

	TargetTile          model.Coords   `json:"-"` // next tile the entity is currently moving
	TargetPath          []model.Coords `json:"-"` // path the entity is currently trying to travel on
	SuggestedTargetPath []model.Coords `json:"-"` // a suggested path for this entity to consider merging into the target path
}

// for setting a tile position
func (e *Entity) SetPosition(c model.Coords) {
	mapWidth, mapHeight := e.World.MapDimensions()
	if c.X > mapWidth {
		panic("c.X is outside of map bounds")
	}
	if c.Y > mapHeight {
		panic("c.Y is outside of map bounds")
	}
	e.TilePos = c
	e.X = float64(c.X) * float64(config.TileSize)
	e.Y = float64(c.Y) * float64(config.TileSize)
	e.TargetX = e.X
	e.TargetY = e.Y
}

func (e *Entity) SetPositionPx(x, y float64) {
	e.TilePos = model.ConvertPxToTilePos(int(x), int(y))
	e.X = x
	e.Y = y
	e.TargetX = e.X
	e.TargetY = e.Y
}

// for setting the entire inventory
func (e *Entity) SetInventoryItems(invItems []*item.InventoryItem) {
	e.InventoryItems = make([]*item.InventoryItem, 0)

	for _, newItem := range invItems {
		if newItem == nil {
			e.InventoryItems = append(e.InventoryItems, nil)
			continue
		}
		e.InventoryItems = append(e.InventoryItems, &item.InventoryItem{
			Instance: newItem.Instance,
			Def:      newItem.Def,
			Quantity: newItem.Quantity,
		})
	}
}

func (cd *CharacterData) AddItemToInventory(invItem item.InventoryItem) (bool, item.InventoryItem) {
	return item.AddItemToInventory(invItem, cd.InventoryItems)
}

func (cd *CharacterData) RemoveItemFromInventory(itemToRemove item.InventoryItem) (bool, item.InventoryItem) {
	return item.RemoveItemFromInventory(itemToRemove, cd.InventoryItems)
}

// Equips a weapon, body armor, clothes, or other equipable items that go onto the entity's body or equipment slots
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
		return true
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
		return true
	case item.TypeWeapon:
		// weapons don't have an "equiped slot", so as of now there is no swapping to do here
		part, fxPart := item.GetWeaponParts(i.Def)

		cd.Body.SetWeapon(part, fxPart)
		return true
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
		return true
	default:
		logz.Panicln(cd.DisplayName, "tried to equip item, but it's type didn't match in the switch statement... (this probably should be caught by the IsEquipable check)")
	}
	return false
}
