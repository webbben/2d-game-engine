// Package entity contains all the logic for an entity
package entity

import (
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
	Collides(r model.Rect, excludeEntityID string) model.CollisionResult
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
	CoinPurseSize    int    // if greater than 0, the entity will have a "coin purse" section in their inventory which stores money.
	CharacterDataSrc string // path to a JSON file containing the definition of the character data (entity body, items, etc)
}

// NewEntity Create a new entity from a character data JSON file
func NewEntity(general GeneralProps, ap AudioProps, defMgr *definitions.DefinitionManager) Entity {
	if general.CharacterDataSrc == "" {
		panic("no character data source JSON specified")
	}
	if general.CoinPurseSize < 0 {
		panic("coin purse size cannot be negative")
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

	ent.IsPlayer = general.IsPlayer

	if len(ent.InventoryItems) == 0 {
		logz.Panicln(ent.DisplayName, "inventory size is 0")
	}
	if ent.WalkSpeed == 0 {
		logz.Warnln("", "loaded entity does not have a walking speed; setting default value.")
		ent.WalkSpeed = GetDefaultWalkSpeed()
	}
	if ent.RunSpeed == 0 {
		logz.Warnln("", "loaded entity does not have a run speed; setting default value.")
		ent.RunSpeed = GetDefaultRunSpeed()
	}

	ent.Validate()

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
	if cd.WalkSpeed == cd.RunSpeed {
		logz.Panicln(cd.DisplayName, "run speed and walk speed are the same:", cd.RunSpeed)
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
			if bodyPart.PartSrc.None {
				logz.Panicln(cd.DisplayName, "equipment is not nil, but body part is set to none")
			}
			if equipedItem.Def == nil {
				logz.Panicln(cd.DisplayName, "equipment item def was found to be nil")
			}
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
	// the validation function checks the bodyPartDef, but for legs we want to check the LegsPartDef... so just handle it here separately
	if cd.EquipedBodywear == nil {
		if !cd.Body.EquipLegsSet.PartSrc.None {
			logz.Panicln(cd.DisplayName, "equiped bodywear is nil, but equiped legs part is not none")
		}
	} else {
		equipedLegsPart := cd.EquipedBodywear.Def.GetLegsPartDef()
		if equipedLegsPart == nil {
			logz.Panicln(cd.DisplayName, "bodywear set, but equiped legs part seems to be nil")
		}
		if !equipedLegsPart.IsEqual(cd.Body.EquipLegsSet.PartSrc) {
			logz.Panicln(cd.DisplayName, "equiped legs dont appear to match actual item legs equipment")
		}
	}
	validateEquipment(cd.EquipedAuxiliary, cd.Body.AuxItemSet)
	validateEquipment(cd.EquipedWeapon, cd.Body.WeaponSet)
	// handle weapon fx separately since we get that part with a specific function
	if cd.EquipedWeapon == nil {
		if !cd.Body.WeaponFxSet.PartSrc.None {
			logz.Panicln(cd.DisplayName, "equiped weapon is nil, but weapon fx part is not none")
		}
	} else {
		_, fxPart := item.GetWeaponParts(cd.EquipedWeapon.Def)
		if !fxPart.IsEqual(cd.Body.WeaponFxSet.PartSrc) {
			logz.Panicln(cd.DisplayName, "equiped weapon fx doesn't appear to match actual fx part")
		}
	}
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

// SetPosition sets a tile position
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
