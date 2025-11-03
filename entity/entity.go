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
)

var (
	defaultWalkSpeed float64 = float64(config.TileSize) / 16
)

func GetDefaultWalkSpeed() float64 {
	if defaultWalkSpeed == 0 {
		panic("entity default walk speed is 0?")
	}
	return defaultWalkSpeed
}

type Entity struct {
	EntityInfo
	Loaded   bool     `json:"-"` // if the entity has been loaded into memory fully yet
	Movement Movement `json:"movement"`
	Position
	MouseBehavior mouse.MouseBehavior

	width float64 // used for collision rect

	footstepSFX audio.FootstepSFX

	FrameTilesetSources []string `json:"frame_tilesets"`
	Body                body.EntityBodySet

	World WorldContext `json:"-"`

	Vitals     Vitals
	Attributes Attributes
}

func (e Entity) CollisionRect() model.Rect {
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

// create a duplicate entity from this one.
// both entities will share the same references to tiles and animations and such, but will be able to have different
// positions, movement targets, etc.
// Useful for when you need a bunch of NPC entities of the same kind
func (e Entity) Duplicate() Entity {
	copyEnt := Entity{
		EntityInfo:          e.EntityInfo,
		FrameTilesetSources: e.FrameTilesetSources,
		World:               e.World,
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
}

type MovementProps struct {
	WalkSpeed float64
}

type AudioProps struct {
	FootstepSFX audio.FootstepSFX
}

type GeneralProps struct {
	DisplayName     string
	IsPlayer        bool
	FrameTilesetSrc string
}

func NewEntity(general GeneralProps, mv MovementProps, ap AudioProps) Entity {
	if general.FrameTilesetSrc == "" {
		panic("no frame tileset src specified")
	}
	if general.DisplayName == "" {
		panic("entity display name is empty")
	}
	if mv.WalkSpeed == 0 {
		logz.Warnln("", "loaded entity does not have a walking speed; setting default value.")
		mv.WalkSpeed = GetDefaultWalkSpeed()
	}

	ent := Entity{
		FrameTilesetSources: []string{general.FrameTilesetSrc},
		EntityInfo: EntityInfo{
			IsPlayer: general.IsPlayer,
		},
		Movement: Movement{
			WalkSpeed: mv.WalkSpeed,
		},
	}

	// load sounds
	ent.LoadFootstepSFX(ap.FootstepSFX)

	return ent
}

// Create an entity by opening an entity's definition JSON
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

// load fully entity data into memory for rendering in a map
func (e *Entity) Load() {
	e.Movement.Direction = 'D'
	e.Body.Load()
	e.Movement.IsMoving = false

	// confirm that body image exists
	dx, dy := e.Body.Dimensions()
	if dx == 0 || dy == 0 {
		panic("body image has no size?")
	}

	e.Loaded = true
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

	CanRun bool `json:"can_run"`

	IsMoving        bool    `json:"-"`
	movementStopped bool    // set when movement ends, so that animation knows when to prepare to go back to idle
	Interrupted     bool    `json:"-"`          // flag for if this entity's movement was stopped unexpectedly (e.g. by a collision)
	WalkSpeed       float64 `json:"walk_speed"` // value should be a TileSize / NumFrames calculation
	Speed           float64 `json:"-"`          // actual speed the entity is moving at

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
