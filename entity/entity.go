package entity

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/item"
)

var (
	defaultWalkSpeed float64 = float64(config.TileSize) / 18
	defaultRunSpeed  float64 = float64(config.TileSize) / 12
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
	EntityInfo
	Loaded   bool     `json:"-"` // if the entity has been loaded into memory fully yet
	Movement Movement `json:"movement"`
	Position
	MouseBehavior mouse.MouseBehavior

	width float64 // used for collision rect

	footstepSFX audio.FootstepSFX

	Body body.EntityBodySet

	World WorldContext `json:"-"`

	Vitals     Vitals
	Attributes Attributes

	// Inventory and Items

	InventoryItems []*item.InventoryItem

	EquipedHeadwear  *item.InventoryItem
	EquipedBodywear  *item.InventoryItem
	EquipedFootwear  *item.InventoryItem
	EquipedAmulet    *item.InventoryItem
	EquipedRing1     *item.InventoryItem
	EquipedRing2     *item.InventoryItem
	EquipedAmmo      *item.InventoryItem
	EquipedAuxiliary *item.InventoryItem
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

// TODO this probably doesn't work anymore
//
// create a duplicate entity from this one.
// both entities will share the same references to tiles and animations and such, but will be able to have different
// positions, movement targets, etc.
// Useful for when you need a bunch of NPC entities of the same kind
func (e Entity) Duplicate() Entity {
	copyEnt := Entity{
		EntityInfo: e.EntityInfo,
		World:      e.World,
	}

	copyEnt.IsPlayer = false // cannot have duplicate players

	copyEnt.Movement = e.Movement
	copyEnt.Movement.TargetPath = []model.Coords{}
	copyEnt.Movement.TargetTile = model.Coords{}
	copyEnt.Movement.IsMoving = false

	return copyEnt
}

type WorldContext interface {
	Collides(r model.Rect, excludeEntityId string, rectBased bool) model.CollisionResult
	FindPath(start, goal model.Coords) ([]model.Coords, bool)
	MapDimensions() (width int, height int)
	GetGroundMaterial(tileX, tileY int) string
	GetDistToPlayer(x, y float64) float64
}

type MovementProps struct {
	WalkSpeed float64
	RunSpeed  float64
}

type AudioProps struct {
	FootstepSFX audio.FootstepSFX
}

type GeneralProps struct {
	DisplayName   string
	IsPlayer      bool
	EntityBodySrc string // path to a JSON file containing the definition of the body
	InventorySize int
}

func NewEntity(general GeneralProps, mv MovementProps, ap AudioProps) Entity {
	if general.EntityBodySrc == "" {
		panic("no body source JSON specified")
	}
	if general.DisplayName == "" {
		panic("entity display name is empty")
	}
	if general.InventorySize <= 0 {
		panic("inventory size must be positive")
	}
	if mv.WalkSpeed == 0 {
		logz.Warnln("", "loaded entity does not have a walking speed; setting default value.")
		mv.WalkSpeed = GetDefaultWalkSpeed()
	}
	if mv.RunSpeed == 0 {
		logz.Warnln("", "loaded entity does not have a run speed; setting default value.")
		mv.RunSpeed = GetDefaultRunSpeed()
	}

	ent := Entity{
		EntityInfo: EntityInfo{
			IsPlayer:    general.IsPlayer,
			DisplayName: general.DisplayName,
		},
		Movement: Movement{
			WalkSpeed: mv.WalkSpeed,
			RunSpeed:  mv.RunSpeed,
		},
		InventoryItems: make([]*item.InventoryItem, general.InventorySize),
	}

	// load body
	entBody, err := body.ReadJSON(general.EntityBodySrc)
	if err != nil {
		panic(err)
	}
	ent.Body = entBody

	// load sounds
	ent.LoadFootstepSFX(ap.FootstepSFX)

	// prepare initial image frames
	ent.Movement.Direction = 'D'
	ent.Body.Load()
	ent.Body.Name = general.DisplayName
	ent.Movement.IsMoving = false

	// confirm that body image exists
	dx, dy := ent.Body.Dimensions()
	if dx == 0 || dy == 0 {
		panic("body image has no size?")
	}

	ent.width = float64(dx)

	ent.Loaded = true

	return ent
}

// Create an entity by opening an entity's definition JSON
// TODO delete? not being used currently
func OpenEntity(source string) (Entity, error) {
	data, err := os.ReadFile(source)
	if err != nil {
		return Entity{}, fmt.Errorf("error reading entity source file: %w", err)
	}
	var ent Entity
	err = json.Unmarshal(data, &ent)
	if err != nil {
		return Entity{}, fmt.Errorf("error while unmarshalling entity json data: %w", err)
	}

	if ent.Movement.WalkSpeed == 0 {
		logz.Warnln(ent.DisplayName, "loaded entity does not have a walking speed; setting default value.")
		ent.Movement.WalkSpeed = GetDefaultWalkSpeed()
	}

	return ent, nil
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

	IsMoving        bool    `json:"-"`
	movementStopped bool    // set when movement ends, so that animation knows when to prepare to go back to idle
	Interrupted     bool    `json:"-"`          // flag for if this entity's movement was stopped unexpectedly (e.g. by a collision)
	WalkSpeed       float64 `json:"walk_speed"` // value should be a TileSize / NumFrames calculation
	RunSpeed        float64
	Speed           float64 `json:"-"` // actual speed the entity is moving at

	TargetTile          model.Coords   `json:"-"` // next tile the entity is currently moving
	TargetPath          []model.Coords `json:"-"` // path the entity is currently trying to travel on
	SuggestedTargetPath []model.Coords `json:"-"` // a suggested path for this entity to consider merging into the target path
}

func (e *Entity) SetPosition(c model.Coords) {
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

func (e *Entity) AddItemToInventory(invItem item.InventoryItem) (bool, item.InventoryItem) {
	return item.AddItemToInventory(invItem, e.InventoryItems)
}

func (e *Entity) RemoveItemFromInventory(itemToRemove item.InventoryItem) (bool, item.InventoryItem) {
	return item.RemoveItemFromInventory(itemToRemove, e.InventoryItems)
}

func (e *Entity) UnequipWeaponFromBody() {
	e.Body.WeaponSet.None = true
	e.Body.WeaponFxSet.None = true
	e.Body.Load()
}

func (e *Entity) EquipWeapon(weaponDef body.SelectedPartDef, weaponFxDef body.SelectedPartDef) {
	e.Body.SetWeapon(weaponDef, weaponFxDef)
}

func (e *Entity) SwingWeapon() {
	if e.Body.WeaponSet.None {
		panic("tried to swing weapon, but no weapon is equiped")
	}
	e.Body.SetAnimationTickCount(8)
	e.Body.SetAnimation(body.ANIM_SLASH, body.SetAnimationOps{DoOnce: true, PreventSkip: true})
}
