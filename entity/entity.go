// Package entity contains all the logic for an entity
package entity

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/model"
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

// An Entity makes a character exist in the world. The entity itself mainly handles logic for runtime stuff, like showing a body on a screen
// and doing animations. However, it links to the underlying state of a character too.
type Entity struct {
	// State and Definitions

	// Note: we don't keep a pointer to the CharacterDef here, since we don't need that data during runtime - just for instantiating or loading a character state.
	// The only thing the actual Entity needs from that is the BodyDef, which tells it which body parts to load.

	// records the last known equiped item ID for each spot; if a change is noticed, we should immediately update the body to match
	equipedBodywear, equipedHeadwear, equipedFootwear, equipedWeapon, equipedAuxiliary defs.ItemID

	CharacterStateRef *state.CharacterState

	// Runtime logic below

	Body body.EntityBodySet

	Loaded   bool     `json:"-"` // if the entity has been loaded into memory fully yet
	Movement Movement `json:"movement"`
	Position

	width float64 // used for collision rect

	footstepSFX audio.FootstepSFX // currently set footstep SFX. However, a default footstepSFX ID should be in entity def

	attackManager

	World WorldContext `json:"-"`

	stunTicks int
}

func (e Entity) DisplayName() string {
	if e.CharacterStateRef == nil {
		panic("character state ref was nil")
	}

	return e.CharacterStateRef.DisplayName
}

func (e Entity) ID() state.CharacterStateID {
	if e.CharacterStateRef == nil {
		panic("character state ref was nil")
	}

	return e.CharacterStateRef.ID
}

func (e Entity) IsPlayer() bool {
	if e.CharacterStateRef == nil {
		panic("character state ref was nil")
	}

	return e.CharacterStateRef.IsPlayer
}

type NewCharacterParams struct {
	IsPlayer bool
}

// LoadCharacterStateIntoEntity loads a character state into an Entity, for rendering and interacting in a game map.
// Grabs the character state from definition manager, outfits it in an entity, and prepares it for runtime use.
func LoadCharacterStateIntoEntity(charStateID state.CharacterStateID, defMgr *definitions.DefinitionManager, audioMgr *audio.AudioManager) *Entity {
	charState := defMgr.GetCharacterState(charStateID)

	// load body def from charDef
	charDef := defMgr.GetCharacterDef(charState.DefID)
	skin := LoadBodySkin(charDef.BodyDef, defMgr)

	// load footstep SFX
	sfxDef := defMgr.GetFootstepSFXDef(charDef.FootstepSFXDefID)

	ent := Entity{
		Body: skin,
		Movement: Movement{
			WalkAnimationTickInterval: defaultWalkAnimationTickInterval,
			RunAnimationTickInterval:  defaultRunAnimationTickInterval,
		},
		CharacterStateRef: charState,
		footstepSFX: audio.NewFootstepSFX(audio.FootstepSFXParams{
			Def:           sfxDef,
			TickDelay:     20,  // TODO: this should actually be calculated by the movement speed / speed of movement animation
			DefaultVolume: 0.2, // TODO: look into how this is being used, because I'm not really sure about it
		}, audioMgr),
	}

	// save char state to definition manager, since that's really where the character state will "live".
	// we just put a pointer to it in the entity so it can reference it too.

	if ent.CharacterStateRef.BaseAttributes == nil {
		panic("base attributes is nil. did character builder not save data?")
	}
	if ent.CharacterStateRef.BaseSkills == nil {
		panic("base skills are nil. did character builder not save data?")
	}

	if len(ent.CharacterStateRef.InventoryItems) == 0 {
		logz.Panicln(ent.CharacterStateRef.DisplayName, "inventory size is 0")
	}

	// prepare initial image frames
	ent.Movement.Direction = 'D'
	ent.Body.Load()
	ent.Body.SetAnimationTickCount(defaultIdleAnimationTickInterval)
	ent.Body.Name = charState.DisplayName
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

	// initialize body to match equiped items (this function is normally just done during update though)
	ent.SyncBodyToState()

	ent.Loaded = true

	return &ent
}

type NewCharacterStateParams struct {
	IsPlayer             bool
	OverwriteDisplayName string // if set, the display name used in charDef will be ignored in favor of this one
}

// CreateNewCharacterState instantiates a new Character State from a CharacterDef. This should only be done when:
//
// - a player starts a new game, and creates his character for the first time.
//
// - NPCs are undergoing first-time creation/instantiation in the game world.
//
// - Or, perhaps a "generic" characterDef is being used to dynamically generate certain types of characters. In which case, this might go on later in the game.
//
// ... Basically, DON'T use this to "load an existing character back into the world". Each character only has this done to them once in their existence.
func CreateNewCharacterState(charDefID defs.CharacterDefID, params NewCharacterStateParams, defMgr *definitions.DefinitionManager) state.CharacterStateID {
	charDef := defMgr.GetCharacterDef(charDefID)

	// find unique ID based on this characterDefID
	id := defMgr.GetNewCharStateID(charDefID)
	if charDef.Unique {
		if string(id) != string(charDef.ID) {
			logz.Panicln("CreateNewCharacterState", "new charStateID should match the defID since the charDef is unique, but it doesn't:", charDef.ID, id)
		}
	}

	// check if this characters' dialog profile has a state created yet. if not, create it now.
	// Note: the player as a character doesn't have a dialog profile. so skip the player.
	if !params.IsPlayer {
		dialogProfileID := charDef.DialogProfileID
		if !defMgr.DialogProfileStateExists(dialogProfileID) {
			dialogProfileState := state.DialogProfileState{
				ProfileID: dialogProfileID,
				Memory:    make(map[string]bool),
			}
			defMgr.LoadDialogProfileState(&dialogProfileState)
		}
	}

	charState := &state.CharacterState{
		ID:          id,
		DefID:       charDefID,
		DisplayName: charDef.DisplayName,
		IsPlayer:    params.IsPlayer,

		StandardInventory: charDef.InitialInventory,

		Vitals:         charDef.BaseVitals,
		BaseAttributes: charDef.BaseAttributes,
		BaseSkills:     charDef.BaseSkills,
		Traits:         charDef.InitialTraits,
	}

	item.LoadStandardInventoryItemDefs(&charState.StandardInventory, defMgr)

	charState.Validate()

	defMgr.LoadCharacterState(charState)

	return id
}

// LoadBodySkin loads up a bodyDef into an EntityBodySet. Does not handle loading any equipment.
func LoadBodySkin(bodyDef defs.BodyDef, defMgr *definitions.DefinitionManager) body.EntityBodySet {
	if defMgr == nil {
		panic("defMgr was nil")
	}

	skin := body.NewHumanBodyFramework()

	// Load body "skin" parts
	if bodyDef.BodyID == "" {
		logz.Panicln("LoadBodySkin", "failed to load body set; id is empty")
	}
	skin.BodySet.PartSrc = defMgr.GetBodyPartDef(bodyDef.BodyID)
	if bodyDef.ArmsID == "" {
		logz.Panicln("LoadBodySkin", "failed to load arms set; id is empty")
	}
	skin.ArmsSet.PartSrc = defMgr.GetBodyPartDef(bodyDef.ArmsID)
	if bodyDef.LegsID == "" {
		logz.Panicln("LoadBodySkin", "failed to load legs set; id is empty")
	}
	skin.LegsSet.PartSrc = defMgr.GetBodyPartDef(bodyDef.LegsID)
	if bodyDef.EyesID == "" {
		logz.Panicln("LoadBodySkin", "failed to load eyes set; id is empty")
	}
	skin.EyesSet.PartSrc = defMgr.GetBodyPartDef(bodyDef.EyesID)
	if bodyDef.HairID == "" {
		logz.Panicln("LoadBodySkin", "failed to load hair set; id is empty")
	}
	skin.HairSet.PartSrc = defMgr.GetBodyPartDef(bodyDef.HairID)

	skin.BodyHSV = bodyDef.BodyHSV
	skin.HairHSV = bodyDef.HairHSV
	skin.EyesHSV = bodyDef.EyesHSV

	return skin
}

// SyncBodyToState detects changes to equiped items in character state, and applies those changes to the body
func (e *Entity) SyncBodyToState() {
	if e.CharacterStateRef == nil {
		panic("character state ref was nil")
	}

	var actualBodywearID, actualHeadwearID, actualFootwearID, actualAuxID, actualWeaponID defs.ItemID
	change := false

	equipment := e.CharacterStateRef.Equipment
	if equipment.EquipedBodywear != nil {
		actualBodywearID = equipment.EquipedBodywear.Instance.DefID
	}
	if equipment.EquipedHeadwear != nil {
		actualHeadwearID = equipment.EquipedHeadwear.Instance.DefID
	}
	if equipment.EquipedFootwear != nil {
		actualFootwearID = equipment.EquipedFootwear.Instance.DefID
	}
	if equipment.EquipedAuxiliary != nil {
		actualAuxID = equipment.EquipedAuxiliary.Instance.DefID
	}
	if equipment.EquipedWeapon != nil {
		actualWeaponID = equipment.EquipedWeapon.Instance.DefID
	}

	if actualBodywearID != e.equipedBodywear {
		change = true
		if actualBodywearID == "" {
			e.Body.RemoveBodywear()
		} else {
			e.Body.EquipBodyItem(equipment.EquipedBodywear.Def)
		}
		e.equipedBodywear = actualBodywearID
	}
	if actualHeadwearID != e.equipedHeadwear {
		change = true
		if actualHeadwearID == "" {
			e.Body.RemoveHeadwear()
		} else {
			e.Body.EquipHeadItem(equipment.EquipedHeadwear.Def)
		}
		e.equipedHeadwear = actualHeadwearID
	}
	if actualFootwearID != e.equipedFootwear {
		change = true
		if actualFootwearID == "" {
			e.Body.RemoveFootwear()
		} else {
			e.Body.EquipFootItem(equipment.EquipedFootwear.Def)
		}
		e.equipedFootwear = actualFootwearID
	}
	if actualAuxID != e.equipedAuxiliary {
		change = true
		if actualAuxID == "" {
			e.Body.RemoveAuxiliary()
		} else {
			e.Body.EquipAuxItem(equipment.EquipedAuxiliary.Def)
		}
		e.equipedAuxiliary = actualAuxID
	}
	if actualWeaponID != e.equipedWeapon {
		change = true
		if actualWeaponID == "" {
			e.Body.RemoveWeapon()
		} else {
			e.Body.EquipWeaponItem(equipment.EquipedWeapon.Def)
		}
		e.equipedWeapon = actualWeaponID
	}

	// if we changed a body part, let's do a full validation to confirm that all body parts match their equiped items
	if change {
		logz.Println(e.DisplayName(), "equipment change detected")
		e.validateEquipment()
	}
}

func (e Entity) validateEquipment() {
	// ensure equiped items match the set body part def
	validateEquipment := func(equipedItem *defs.InventoryItem, bodyPart body.BodyPartSet) {
		if equipedItem == nil {
			if !bodyPart.PartSrc.None {
				logz.Panicln(e.DisplayName(), "equipment is nil, but body part is not set to None")
			}
		} else {
			if bodyPart.PartSrc.None {
				logz.Panicln(e.DisplayName(), "equipment is not nil, but body part is set to none")
			}
			if equipedItem.Def == nil {
				logz.Panicln(e.DisplayName(), "equipment item def was found to be nil")
			}
			equiped := equipedItem.Def.GetBodyPartDef()
			if equiped == nil {
				logz.Panicln(e.DisplayName(), "equipment body part def was found to be nil")
			}
			if !equiped.IsEqual(bodyPart.PartSrc) {
				logz.Panicln(e.DisplayName(), "equipment item def does not appear to match body's equiped part")
			}
		}
	}

	equipment := e.CharacterStateRef.Equipment

	validateEquipment(equipment.EquipedHeadwear, e.Body.EquipHeadSet)
	validateEquipment(equipment.EquipedBodywear, e.Body.EquipBodySet)
	// the validation function checks the bodyPartDef, but for legs we want to check the LegsPartDef... so just handle it here separately
	if equipment.EquipedBodywear == nil {
		if !e.Body.EquipLegsSet.PartSrc.None {
			logz.Panicln(e.DisplayName(), "equiped bodywear is nil, but equiped legs part is not none")
		}
	} else {
		equipedLegsPart := equipment.EquipedBodywear.Def.GetLegsPartDef()
		if equipedLegsPart == nil {
			logz.Panicln(e.DisplayName(), "bodywear set, but equiped legs part seems to be nil")
		}
		if !equipedLegsPart.IsEqual(e.Body.EquipLegsSet.PartSrc) {
			logz.Panicln(e.DisplayName(), "equiped legs dont appear to match actual item legs equipment")
		}
	}
	validateEquipment(equipment.EquipedAuxiliary, e.Body.AuxItemSet)
	validateEquipment(equipment.EquipedWeapon, e.Body.WeaponSet)
	// handle weapon fx separately since we get that part with a specific function
	if equipment.EquipedWeapon == nil {
		if !e.Body.WeaponFxSet.PartSrc.None {
			logz.Panicln(e.DisplayName(), "equiped weapon is nil, but weapon fx part is not none")
		}
	} else {
		_, fxPart := item.GetWeaponParts(equipment.EquipedWeapon.Def)
		if !fxPart.IsEqual(e.Body.WeaponFxSet.PartSrc) {
			logz.Panicln(e.DisplayName(), "equiped weapon fx doesn't appear to match actual fx part")
		}
	}
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

type WorldContext interface {
	Collides(r model.Rect, excludeEntityID string) model.CollisionResult
	FindPath(start, goal model.Coords) ([]model.Coords, bool)
	MapDimensions() (width int, height int)
	GetGroundMaterial(tileX, tileY int) string
	GetDistToPlayer(x, y float64) float64
	AttackArea(attackInfo AttackInfo)
}

type GeneralProps struct {
	IsPlayer bool
	EntityID string // entity ID links this entity to the character data JSON file of the same name
}

type Position struct {
	X, Y             float64 `json:"-"` // the exact position the entity is at on the map
	drawX, drawY     float64 `json:"-"` // the actual position on the screen where the entity would be drawn
	TargetX, TargetY float64 `json:"-"` // the target position the entity is moving to
}

// TilePos converts the current X/Y position to an absolute tile position
func (e Entity) TilePos() model.Coords {
	return model.ConvertPxToTilePos(int(e.X), int(e.Y))
}

// TargetTilePos converts the current target X/Y position to an absolute tile position
func (e Entity) TargetTilePos() model.Coords {
	return model.ConvertPxToTilePos(int(e.TargetX), int(e.TargetY))
}

// Movement is Runtime logic for movement
type Movement struct {
	Direction byte `json:"-"` // L R U D

	IsMoving                  bool    `json:"-"`
	Interrupted               bool    `json:"-"` // flag for if this entity's movement was stopped unexpectedly (e.g. by a collision)
	Speed                     float64 `json:"-"` // actual speed the entity is moving at
	WalkAnimationTickInterval int
	RunAnimationTickInterval  int

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
	e.X = float64(c.X) * float64(config.TileSize)
	e.Y = float64(c.Y) * float64(config.TileSize)
	e.TargetX = e.X
	e.TargetY = e.Y
}

func (e *Entity) SetPositionPx(x, y float64) {
	e.X = x
	e.Y = y
	e.TargetX = e.X
	e.TargetY = e.Y
}

func ValidateCharacterDef(cd defs.CharacterDef) {
	if cd.DisplayName == "" {
		panic("display name empty")
	}
	if cd.ID == "" {
		panic("ID empty")
	}
	if cd.ClassName == "" {
		panic("class name empty")
	}
	if cd.ClassDefID == "" {
		panic("class def ID is empty")
	}
	if cd.DialogProfileID == "" {
		panic("dialog profile ID is empty")
	}
	if cd.FootstepSFXDefID == "" {
		panic("footstep sfx is empty")
	}
}
