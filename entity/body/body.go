// Package body contains all drawing and update logic for entity bodies for moving in worlds
package body

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

type HSV struct {
	H, S, V float64
}

var Default HSV = HSV{0.5, 0.5, 0.5}

type EntityBodySet struct {
	Name string

	animation                 string `json:"-"`
	nextAnimation             string `json:"-"`
	stopAnimationOnCompletion bool   `json:"-"`
	animationTickCount        int    `json:"-"` // the "duration" of ticks until the next animation frame should trigger
	ticks                     int    `json:"-"` // number of ticks elapsed
	currentDirection          byte   `json:"-"` // L R U D

	dmgFlicker damageFlickerFX `json:"-"`

	stretchX, stretchY int `json:"-"` // amount to stretch certain body parts - set by body set

	// actual body definition - not including equiped items

	stagingImg *ebiten.Image `json:"-"` // just for putting everything together before drawing to screen (for adding flicker fx)

	BodyHSV HSV
	BodySet BodyPartSet
	EyesHSV HSV
	EyesSet BodyPartSet
	HairHSV HSV
	HairSet BodyPartSet
	ArmsSet BodyPartSet
	LegsSet BodyPartSet

	// currently equiped items

	EquipBodySet BodyPartSet // An equiped piece of body armor or shirt, on the entity's torso and arms.
	EquipLegsSet BodyPartSet // The corresponding leg equipment for the body set
	EquipFeetSet BodyPartSet // Equiped boots, shoes, or other footwear
	EquipHeadSet BodyPartSet // An equiped helmet or hat, on the entity's head
	WeaponSet    BodyPartSet // The weapon as shown in the entity's hands
	WeaponFxSet  BodyPartSet // Fx from using a weapon or tool. For showing things like sword slash Fx
	AuxItemSet   BodyPartSet // Item held in the left hand, such as a torch or shield.

	globalOffsetY  float64 `json:"-"` // amount to offset placement of (non-body) parts by, when body is taller or shorter
	nonBodyYOffset int     `json:"-"` // amount to offset placement of (non-body) parts by, simply dictated by the body's movements
}

func (eb EntityBodySet) shouldCropHair() bool {
	if eb.EquipHeadSet.PartSrc.None {
		return false
	}
	return eb.EquipHeadSet.PartSrc.CropHairToHead
}

func (eb EntityBodySet) GetDebugString() string {
	s := fmt.Sprintf("ANIM: %s DIR: %s (next: %s, stopOnComp: %v)\n", eb.animation, string(eb.currentDirection), eb.nextAnimation, eb.stopAnimationOnCompletion)
	s += fmt.Sprintf("ticks: %v tickCount: %v globalOffY: %v nonBodyOffY: %v cropHair: %v\n", eb.ticks, eb.animationTickCount, eb.globalOffsetY, eb.nonBodyYOffset, eb.shouldCropHair())
	// get a single line status for each bodypart
	s += eb.BodySet.animationDebugString() + "\n"
	s += eb.ArmsSet.animationDebugString() + "\n"
	s += eb.LegsSet.animationDebugString() + "\n"
	s += eb.EyesSet.animationDebugString() + "\n"
	s += eb.HairSet.animationDebugString() + "\n"
	s += eb.EquipBodySet.animationDebugString() + "\n"
	s += eb.EquipLegsSet.animationDebugString() + "\n"
	s += eb.EquipHeadSet.animationDebugString() + "\n"
	s += eb.EquipFeetSet.animationDebugString() + "\n"
	s += eb.WeaponSet.animationDebugString() + "\n"
	return s
}

// Load is for loading all body parts, assuming that they all already have PartSrc set. E.g. for after loading from JSON.
func (eb *EntityBodySet) Load() {
	// load body first, since it dictates stretchX and stretchY (which impact several sets)
	eb.SetBody(eb.BodySet.PartSrc, eb.ArmsSet.PartSrc, eb.LegsSet.PartSrc)
	// load head second, since it impacts the hair set
	eb.SetEquipHead(eb.EquipHeadSet.PartSrc)
	eb.SetEquipFeet(eb.EquipFeetSet.PartSrc)
	eb.SetHair(eb.HairSet.PartSrc)
	eb.SetEyes(eb.EyesSet.PartSrc)
	eb.SetEquipBody(eb.EquipBodySet.PartSrc, eb.EquipLegsSet.PartSrc)
	eb.SetWeapon(eb.WeaponSet.PartSrc, eb.WeaponFxSet.PartSrc)
	eb.SetAuxiliary(eb.AuxItemSet.PartSrc)

	// set an initial direction and ensure img is set
	eb.animation = AnimIdle
	eb._initializeDirection(model.Directions.Down)
	if eb.BodySet.img == nil {
		panic("body image is nil!")
	}

	// make sure everything looks correct
	eb.validate()

	tilesize := config.TileSize
	eb.stagingImg = ebiten.NewImage(tilesize*5, tilesize*5)
}

func (eb EntityBodySet) validate() {
	if eb.BodySet.PartSrc.None {
		panic("body cannot be None")
	}
	if eb.ArmsSet.PartSrc.None {
		panic("arms cannot be None")
	}
	if eb.LegsSet.PartSrc.None {
		panic("legs cannot be None")
	}
	if eb.EyesSet.PartSrc.None {
		panic("eyes cannot be None")
	}
	if eb.HairSet.PartSrc.None {
		// TODO should we allow no hair to be set?
		panic("hair cannot be None")
	}

	eb.HairSet.validate()
	eb.EyesSet.validate()
	eb.EquipBodySet.validate()
	eb.EquipLegsSet.validate()
	eb.EquipHeadSet.validate()
	eb.EquipFeetSet.validate()
	eb.BodySet.validate()
	eb.WeaponSet.validate()
	eb.WeaponFxSet.validate()
	eb.AuxItemSet.validate()
}

func ReadJSON(jsonFilePath string) (EntityBodySet, error) {
	if !config.FileExists(jsonFilePath) {
		return EntityBodySet{}, errors.New("no file found at path: " + jsonFilePath)
	}

	data, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return EntityBodySet{}, fmt.Errorf("failed to read file data: %w", err)
	}

	var eb EntityBodySet
	err = json.Unmarshal(data, &eb)
	if err != nil {
		return EntityBodySet{}, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return eb, nil
}

func (eb *EntityBodySet) SetBodyHSV(h, s, v float64) {
	eb.BodyHSV = HSV{h, s, v}
}

func (eb EntityBodySet) GetBodyHSV() (h, s, v float64) {
	return eb.BodyHSV.H, eb.BodyHSV.S, eb.BodyHSV.V
}

func (eb *EntityBodySet) SetEyesHSV(h, s, v float64) {
	eb.EyesHSV = HSV{h, s, v}
}

func (eb EntityBodySet) GetEyesHSV() (h, s, v float64) {
	return eb.EyesHSV.H, eb.EyesHSV.S, eb.EyesHSV.V
}

func (eb *EntityBodySet) SetHairHSV(h, s, v float64) {
	eb.HairHSV = HSV{h, s, v}
}

func (eb EntityBodySet) GetHairHSV() (h, s, v float64) {
	return eb.HairHSV.H, eb.HairHSV.S, eb.HairHSV.V
}

// NewEntityBodySet creates a base body set, without anything equiped
func NewEntityBodySet(bodySet, armsSet, legsSet, hairSet, eyesSet, equipHeadSet, equipFeetSet, equipBodySet, equipLegsSet, weaponSet, weaponFxSet, auxSet BodyPartSet, bodyHSV, eyesHSV, hairHSV *HSV) EntityBodySet {
	if bodyHSV == nil {
		bodyHSV = &Default
	}
	if eyesHSV == nil {
		eyesHSV = &Default
	}
	if hairHSV == nil {
		hairHSV = &Default
	}

	eb := EntityBodySet{
		animation:          AnimIdle,
		animationTickCount: 15,
		currentDirection:   'D',
		BodySet:            bodySet,
		BodyHSV:            *bodyHSV,
		ArmsSet:            armsSet,
		LegsSet:            legsSet,
		HairSet:            hairSet,
		HairHSV:            *hairHSV,
		EyesSet:            eyesSet,
		EyesHSV:            *eyesHSV,
		EquipBodySet:       equipBodySet,
		EquipLegsSet:       equipLegsSet,
		EquipHeadSet:       equipHeadSet,
		EquipFeetSet:       equipFeetSet,
		WeaponSet:          weaponSet,
		WeaponFxSet:        weaponFxSet,
		AuxItemSet:         auxSet,
		stagingImg:         ebiten.NewImage(config.TileSize*5, config.TileSize*5),
	}

	return eb
}

func (eb *EntityBodySet) Dimensions() (dx, dy int) {
	if eb.BodySet.img == nil {
		panic("body image is nil")
	}
	bounds := eb.BodySet.img.Bounds()
	return bounds.Dx(), bounds.Dy()
}

func (eb *EntityBodySet) SetBody(bodyDef, armDef, legDef SelectedPartDef) {
	if bodyDef.None {
		panic("body must be defined")
	}
	if armDef.None {
		panic("arms must be defined")
	}
	if legDef.None {
		panic("legs must be defined")
	}

	eb.BodySet.setImageSource(bodyDef, 0, 0, eb.IsAuxEquipped())

	// reload any body parts that are influenced by stretch properties
	// ensure these stretch values are set before calling subtract arms, since it uses equipBodyStretchY
	eb.stretchX = bodyDef.StretchX
	eb.stretchY = bodyDef.StretchY
	if eb.HairSet.HasLoaded() {
		eb.HairSet.load(eb.stretchX, 0, eb.IsAuxEquipped())
	}
	if eb.EquipHeadSet.HasLoaded() {
		eb.EquipHeadSet.load(eb.stretchX, 0, eb.IsAuxEquipped())
	}
	if eb.EquipBodySet.HasLoaded() {
		eb.EquipBodySet.load(eb.stretchX, eb.stretchY, eb.IsAuxEquipped())
	}
	// FYI: this hasn't been tested yet, since we've stopped using body stretching (for now)
	if eb.EquipLegsSet.HasLoaded() {
		eb.EquipLegsSet.load(eb.stretchX, eb.stretchY, eb.IsAuxEquipped())
	}
	if eb.EquipFeetSet.HasLoaded() {
		eb.EquipFeetSet.load(eb.stretchX, 0, eb.IsAuxEquipped())
	}

	// ensure this is set before calling subtractArms, since it uses this value
	eb.globalOffsetY = float64(bodyDef.OffsetY)

	// arms are directly set with body
	eb.ArmsSet.setImageSource(armDef, 0, 0, eb.IsAuxEquipped())
	if eb.EquipBodySet.sourceSet && !eb.EquipBodySet.PartSrc.None {
		// subtract arms by equip body image (remove parts hidden by it)
		eb.subtractArms()
	}

	// legs are also set directly with body
	eb.LegsSet.setImageSource(legDef, 0, 0, eb.IsAuxEquipped())
}

func (eb *EntityBodySet) SetEyes(def SelectedPartDef) {
	if def.None {
		panic("eyes must be defined")
	}
	eb.EyesSet.setImageSource(def, 0, 0, eb.IsAuxEquipped())
}

func (eb *EntityBodySet) SetHair(def SelectedPartDef) {
	eb.HairSet.setImageSource(def, eb.stretchX, 0, eb.IsAuxEquipped())
	if eb.shouldCropHair() {
		eb.cropHair()
	}
}

func (eb *EntityBodySet) ReloadHair() {
	if !eb.HairSet.HasLoaded() {
		logz.Panicln(eb.Name, "tried to reload hair, but hair hasn't been loaded yet")
	}

	eb.HairSet.load(eb.stretchX, 0, eb.IsAuxEquipped())
	if eb.shouldCropHair() {
		eb.cropHair()
	}

	eb.HairSet.setCurrentFrame(eb.currentDirection, eb.animation)
}

func (eb *EntityBodySet) SetEquipHead(def SelectedPartDef) {
	eb.EquipHeadSet.setImageSource(def, eb.stretchX, 0, eb.IsAuxEquipped())

	// always reload hair when equiping head, since it could either need to crop or un-crop the hair
	if eb.HairSet.HasLoaded() {
		eb.ReloadHair()
	}
	// if we are already in game (animation has been defined) then ensure first frame is set.
	// We do this here for a couple reasons: firstly, so that in the inventory screen, the change is visible immediately.
	// But also, for sets like Hair, if it's nil in draw we panic. So, this ensures that it's not ever nil when the draw function is called.
	if eb.animation != "" {
		eb.EquipHeadSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}
}

func (eb *EntityBodySet) SetEquipFeet(def SelectedPartDef) {
	eb.EquipFeetSet.setImageSource(def, eb.stretchX, 0, eb.IsAuxEquipped())

	if eb.animation != "" {
		eb.EquipFeetSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}
}

func (eb *EntityBodySet) ReloadArms() {
	if !eb.ArmsSet.HasLoaded() {
		logz.Panicln(eb.Name, "trying to reload arms, but they haven't been loaded yet")
	}

	eb.ArmsSet.load(0, 0, eb.IsAuxEquipped())
	if !eb.EquipBodySet.PartSrc.None {
		eb.subtractArms()
	}

	if eb.animation != "" {
		eb.ArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}
}

func (eb *EntityBodySet) SetEquipBody(bodyDef, legsDef SelectedPartDef) {
	eb.EquipBodySet.setImageSource(bodyDef, eb.stretchX, eb.stretchY, eb.IsAuxEquipped())

	// redo the arms subtraction
	if eb.ArmsSet.HasLoaded() {
		eb.ReloadArms()
	}

	eb.EquipLegsSet.setImageSource(legsDef, eb.stretchX, eb.stretchY, eb.IsAuxEquipped())

	if eb.animation != "" {
		eb.EquipBodySet.setCurrentFrame(eb.currentDirection, eb.animation)
		eb.EquipLegsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}
}

func (eb *EntityBodySet) SetAuxiliary(def SelectedPartDef) {
	eb.AuxItemSet.setImageSource(def, 0, 0, eb.IsAuxEquipped())

	if eb.animation != "" {
		eb.AuxItemSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}

	eb.reloadAuxAffectedParts()
}

func (eb *EntityBodySet) reloadAuxAffectedParts() {
	// body
	eb.BodySet.load(0, 0, eb.IsAuxEquipped())
	// equip body
	eb.EquipBodySet.load(eb.stretchX, eb.stretchY, eb.IsAuxEquipped())
	// arms
	eb.ReloadArms()
}

func (eb *EntityBodySet) RemoveAuxiliary() {
	eb.AuxItemSet.Remove()

	if eb.animation != "" {
		eb.EquipBodySet.setCurrentFrame(eb.currentDirection, eb.animation)
		eb.ArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}
}

// IsAuxEquipped determines if an aux item is currently equiped.
// An "Aux" item is an item that is held in the left hand (e.g. a torch.).
func (eb EntityBodySet) IsAuxEquipped() bool {
	return !eb.AuxItemSet.PartSrc.None
}

func (eb *EntityBodySet) SetWeapon(weaponDef, weaponFxDef SelectedPartDef) {
	if weaponDef.None != weaponFxDef.None {
		logz.Panicln("SetWeapon", "weapon and weaponFx should have the same None value (so they always equip or unequip together)", "weapon:", weaponDef.None, "weaponFx:", weaponFxDef.None)
	}

	// as of now, we are assuming that weaponFx will never have an idle animation, so setting it to skip here.
	// this is to prevent the weaponFx frames from showing while idle is active.
	weaponFxDef.IdleAnimation.Skip = true

	eb.WeaponSet.setImageSource(weaponDef, 0, 0, eb.IsAuxEquipped())
	eb.WeaponFxSet.setImageSource(weaponFxDef, 0, 0, eb.IsAuxEquipped())
}

func (eb EntityBodySet) GetCurrentAnimation() string {
	return eb.animation
}

func (eb *EntityBodySet) SetAnimationTickCount(tickCount int) {
	if tickCount == 0 {
		logz.Panic("tick count cannot be 0")
	}
	eb.animationTickCount = tickCount
}

// SelectedPartDef represents the currently selected body part and it's individual definition
type SelectedPartDef struct {
	None      bool // if true, this part will not be shown
	FlipRForL bool // if true, instead of using an L source, we just flip the frames for right

	// Idle animation def
	IdleAnimation      AnimationParams // this is defined separately from other animations, since it behaves uniquely (see body.md)
	WalkAnimation      AnimationParams
	RunAnimation       AnimationParams
	SlashAnimation     AnimationParams
	BackslashAnimation AnimationParams
	ShieldAnimation    AnimationParams

	// body-specific props

	// FYI: these are not currently being used anymore (but remain functional) since we don't have other body options (tall, short, fat, etc) anymore.

	StretchX int // amount to stretch hair and equip body on X axis. Defined here, this represents the value that is applied to ALL (applicable) parts - not to this one.
	StretchY int // amount to stretch equip body on the Y axis. Defined here, this represents the value that is applied to ALL (applicable) parts - not to this one.
	OffsetY  int // amount to offset positions of hair, eyes, equip body, etc on the Y axis

	// headwear-specific props

	CropHairToHead bool // set to have hair not go outside the head image. used for helmets or certain hats.
}

type PartDefParams struct {
	None      bool
	FlipRForL bool // if true, frames for Right directions will be flipped horizontally and reused for the Left direction.

	Idle, Walk, Run, Slash, Backslash, Shield *AnimationParams

	StretchX, StretchY int
	OffsetY            int

	CropHairToHead bool
}

// NewPartDef creates a new SelectedPartDef, which essentially defines a specific body part's animations, visuals, etc.
// Use this function to create a SelectedPartDef, rather than directly making the struct, since this will handle some important validation.
func NewPartDef(params PartDefParams) SelectedPartDef {
	if params.None {
		return SelectedPartDef{None: true}
	}
	def := SelectedPartDef{
		FlipRForL:          params.FlipRForL,
		StretchX:           params.StretchX,
		StretchY:           params.StretchY,
		OffsetY:            params.OffsetY,
		CropHairToHead:     params.CropHairToHead,
		IdleAnimation:      AnimationParams{Skip: true},
		WalkAnimation:      AnimationParams{Skip: true},
		RunAnimation:       AnimationParams{Skip: true},
		SlashAnimation:     AnimationParams{Skip: true},
		BackslashAnimation: AnimationParams{Skip: true},
		ShieldAnimation:    AnimationParams{Skip: true},
	}
	validateAnimParams := func(animParams AnimationParams) {
		if animParams.Skip {
			return
		}
		if animParams.TilesetSrc == "" {
			panic("tilesetSrc must not be empty")
		}
	}

	if params.Idle != nil {
		def.IdleAnimation = *params.Idle
		def.IdleAnimation.Name = "idle"
	}
	if params.Walk != nil {
		def.WalkAnimation = *params.Walk
		def.WalkAnimation.Name = "walk"
	}
	if params.Run != nil {
		def.RunAnimation = *params.Run
		def.RunAnimation.Name = "run"
	}
	if params.Slash != nil {
		def.SlashAnimation = *params.Slash
		def.SlashAnimation.Name = "slash"
	}
	if params.Backslash != nil {
		def.BackslashAnimation = *params.Backslash
		def.BackslashAnimation.Name = "backslash"
	}
	if params.Shield != nil {
		def.ShieldAnimation = *params.Shield
		def.ShieldAnimation.Name = "shield"
	}

	validateAnimParams(def.IdleAnimation)
	validateAnimParams(def.WalkAnimation)
	validateAnimParams(def.RunAnimation)
	validateAnimParams(def.SlashAnimation)
	validateAnimParams(def.BackslashAnimation)
	validateAnimParams(def.ShieldAnimation)

	return def
}

// IsEqual checks if the two are equal. mainly used for validation.
func (def SelectedPartDef) IsEqual(other SelectedPartDef) bool {
	if def.None != other.None {
		return false
	}
	if def.FlipRForL != other.FlipRForL {
		return false
	}
	if def.StretchX != other.StretchX || def.StretchY != other.StretchY || def.OffsetY != other.OffsetY {
		return false
	}
	if def.CropHairToHead != other.CropHairToHead {
		return false
	}
	if !def.IdleAnimation.IsEqual(other.IdleAnimation) {
		return false
	}
	if !def.WalkAnimation.IsEqual(other.WalkAnimation) {
		return false
	}
	if !def.RunAnimation.IsEqual(other.RunAnimation) {
		return false
	}
	if !def.SlashAnimation.IsEqual(other.SlashAnimation) {
		return false
	}
	if !def.BackslashAnimation.IsEqual(other.BackslashAnimation) {
		return false
	}
	if !def.ShieldAnimation.IsEqual(other.ShieldAnimation) {
		return false
	}
	return true
}

// Requires BodySet and HairSet to be loaded already
func (eb *EntityBodySet) cropHair() {
	eb.BodySet.validate()
	eb.HairSet.validate()
	leftHead := ebiten.NewImage(config.TileSize, config.TileSize)
	leftHead.DrawImage(eb.BodySet.WalkAnimation.L[0], nil)
	rightHead := ebiten.NewImage(config.TileSize, config.TileSize)
	rightHead.DrawImage(eb.BodySet.WalkAnimation.R[0], nil)
	upHead := ebiten.NewImage(config.TileSize, config.TileSize)
	upHead.DrawImage(eb.BodySet.WalkAnimation.U[0], nil)
	downHead := ebiten.NewImage(config.TileSize, config.TileSize)
	downHead.DrawImage(eb.BodySet.WalkAnimation.D[0], nil)

	cropper := func(a *Animation) {
		for i, img := range a.L {
			a.L[i] = rendering.CropImageByOtherImage(img, leftHead)
		}
		for i, img := range a.R {
			a.R[i] = rendering.CropImageByOtherImage(img, rightHead)
		}
		for i, img := range a.U {
			a.U[i] = rendering.CropImageByOtherImage(img, upHead)
		}
		for i, img := range a.D {
			a.D[i] = rendering.CropImageByOtherImage(img, downHead)
		}
	}

	cropper(&eb.HairSet.WalkAnimation)
	cropper(&eb.HairSet.RunAnimation)
	cropper(&eb.HairSet.SlashAnimation)
	cropper(&eb.HairSet.BackslashAnimation)
	cropper(&eb.HairSet.ShieldAnimation)
	cropper(&eb.HairSet.IdleAnimation)
}

// subtractArms "subtracts" the arms from the equip body set. i.e. the arms are cut off anywhere the the body equipment overlaps.
// this produces an arms set that is just the hands or lower arms, and allows for more careful placement of different body parts.
// for example, when doing a sword slash while facing up, we need the body equipment to show "on top" of the arms, and the hair to show on top of the body equipment,
// but we still need the arms (hands) to be visible as they are behind the head. So, this subtraction allows for that more fine grained control.
//
// Specific cases:
// - facing up, doing a sword slash: the arm is cocked back behind the head. body equipment is "on top" of the arms, hair is on top of body equipment, but hands can still show
// over the hair.
func (eb *EntityBodySet) subtractArms() {
	logz.Println(eb.Name, "subtracting arms")
	if eb.EquipBodySet.PartSrc.None {
		logz.Panicln(eb.Name, "trying to subtract arms, but no bodywear is set")
	}
	cropper := func(a *Animation, subtractorA Animation) {
		equipBodyOffsetY := int(eb.globalOffsetY + eb.getEquipBodyOffsetY())

		// LEFT
		if len(a.L) == 0 {
			logz.Panicln(eb.Name, "subtract arms: no left arms frames?")
		}
		if len(a.L) != len(subtractorA.L) {
			logz.Panicln(eb.Name, "subtract arms: subtractor and subtractee not same size (L)")
		}
		for i, img := range a.L {
			equipBodyImg := subtractorA.L[i]
			a.L[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg, 0, equipBodyOffsetY)
		}

		// RIGHT
		if len(a.R) == 0 {
			logz.Panicln(eb.Name, "subtract arms: no right arms frames?")
		}
		if len(a.R) != len(subtractorA.R) {
			logz.Panicln(eb.Name, "subtract arms: subtractor and subtractee not same size (R)")
		}
		for i, img := range a.R {
			equipBodyImg := subtractorA.R[i]
			a.R[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg, 0, equipBodyOffsetY)
		}

		// UP
		if len(a.U) == 0 {
			logz.Panicln(eb.Name, "subtract arms: no up arms frames?")
		}
		if len(a.U) != len(subtractorA.U) {
			logz.Panicln(eb.Name, "subtract arms: subtractor and subtractee not same size (U)")
		}
		for i, img := range a.U {
			equipBodyImg := subtractorA.U[i]
			a.U[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg, 0, equipBodyOffsetY)
		}

		// DOWN
		if len(a.D) == 0 {
			logz.Panicln(eb.Name, "subtract arms: no down arms frames?")
		}
		if len(a.D) != len(subtractorA.D) {
			logz.Panicln(eb.Name, "subtract arms: subtractor and subtractee not same size (D)")
		}
		for i, img := range a.D {
			equipBodyImg := subtractorA.D[i]
			a.D[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg, 0, equipBodyOffsetY)
		}
	}

	// Note: had issues when a tileset moved down a row without me knowing, and suddenly we were subtracting empty equip body frames from arms.
	// If something similar happens again, you can try using something like the below debug string to confirm the frame indices are wrong.
	// logz.Println(eb.Name, eb.EquipBodySet.PartSrc.WalkAnimation.DebugString())

	cropper(&eb.ArmsSet.WalkAnimation, eb.EquipBodySet.WalkAnimation)
	cropper(&eb.ArmsSet.RunAnimation, eb.EquipBodySet.RunAnimation)
	cropper(&eb.ArmsSet.SlashAnimation, eb.EquipBodySet.SlashAnimation)
	cropper(&eb.ArmsSet.BackslashAnimation, eb.EquipBodySet.BackslashAnimation)
	cropper(&eb.ArmsSet.ShieldAnimation, eb.EquipBodySet.ShieldAnimation)
	cropper(&eb.ArmsSet.IdleAnimation, eb.EquipBodySet.IdleAnimation)
}

func (eb *EntityBodySet) Draw(screen *ebiten.Image, x, y, characterScale float64) {
	// Warning: Do not use characterScale anywhere except the bottom - where we draw stagingImg onto screen!
	// we first make a "staging image" which is drawn without scale, and then we draw that image into screen using characterScale.
	eb.stagingImg.Clear()
	// eb.stagingImg.Fill(color.RGBA{100, 0, 0, 50})  // for testing

	// render order decisions (for not so obvious things):
	// - Arms: after equip body, equip head, hair so that hands show when doing U slash (we subtract arms by equip_body)
	renderOrder := []string{"body", "legs", "equip_legs", "equip_feet", "equip_body", "eyes", "hair", "equip_head", "arms", "equip_weapon", "aux"}
	switch eb.currentDirection {
	case model.Directions.Up:
		// aux first: since facing up, aux items (e.g. torches) will generally be covered by everything
		renderOrder = []string{"aux", "body", "legs", "equip_legs", "equip_feet", "equip_body", "eyes", "hair", "equip_head", "arms", "equip_weapon"}
	case model.Directions.Right:
		// aux after arms: shield may cover part of hands, so aux should render after arms
		renderOrder = []string{"body", "legs", "equip_legs", "equip_feet", "equip_body", "eyes", "hair", "equip_head", "arms", "aux", "equip_weapon"}
	}

	yOff := eb.globalOffsetY

	bodyX := float64(config.TileSize * 2)
	bodyY := float64(config.TileSize)

	equipBodyY := bodyY + yOff + eb.getEquipBodyOffsetY()
	equipFeetY := bodyY + config.TileSize // equip feet tiles are only 16x16

	eyesY := bodyY + (float64(eb.nonBodyYOffset)) + yOff
	hairY := bodyY + (float64(eb.nonBodyYOffset)) + yOff

	weaponY := bodyY - (config.TileSize) + yOff
	weaponX := bodyX - (config.TileSize * 2)

	for _, part := range renderOrder {
		switch part {
		case "body":
			rendering.DrawHSVImage(eb.stagingImg, eb.BodySet.img, eb.BodyHSV.H, eb.BodyHSV.S, eb.BodyHSV.V, bodyX, bodyY, 0)
		case "arms":
			rendering.DrawHSVImage(eb.stagingImg, eb.ArmsSet.img, eb.BodyHSV.H, eb.BodyHSV.S, eb.BodyHSV.V, bodyX, bodyY, 0)
		case "legs":
			rendering.DrawHSVImage(eb.stagingImg, eb.LegsSet.img, eb.BodyHSV.H, eb.BodyHSV.S, eb.BodyHSV.V, bodyX, bodyY, 0)
		case "equip_body":
			if eb.EquipBodySet.img != nil {
				rendering.DrawImage(eb.stagingImg, eb.EquipBodySet.img, bodyX, equipBodyY, 0)
			}
		case "equip_legs":
			if eb.EquipLegsSet.img != nil {
				rendering.DrawImage(eb.stagingImg, eb.EquipLegsSet.img, bodyX, equipBodyY, 0)
			}
		case "eyes":
			if eb.EyesSet.img != nil {
				rendering.DrawHSVImage(eb.stagingImg, eb.EyesSet.img, eb.EyesHSV.H, eb.EyesHSV.S, eb.EyesHSV.V, bodyX, eyesY, 0)
			}
		case "hair":
			if eb.HairSet.img == nil {
				logz.Panicln(eb.Name, "hair img is nil")
			}
			rendering.DrawHSVImage(eb.stagingImg, eb.HairSet.img, eb.HairHSV.H, eb.HairHSV.S, eb.HairHSV.V, bodyX, hairY, 0)
		case "equip_head":
			if eb.EquipHeadSet.img != nil {
				rendering.DrawImage(eb.stagingImg, eb.EquipHeadSet.img, bodyX, hairY, 0)
			}
		case "equip_feet":
			if eb.EquipFeetSet.img != nil {
				rendering.DrawImage(eb.stagingImg, eb.EquipFeetSet.img, bodyX, equipFeetY, 0)
			}
		case "equip_weapon":
			if eb.WeaponSet.img != nil {
				rendering.DrawImage(eb.stagingImg, eb.WeaponSet.img, weaponX, weaponY, 0)
				if eb.WeaponFxSet.img != nil {
					rendering.DrawImage(eb.stagingImg, eb.WeaponFxSet.img, weaponX, weaponY, 0)
				}
			}
		case "aux":
			if eb.AuxItemSet.img != nil {
				rendering.DrawImage(eb.stagingImg, eb.AuxItemSet.img, bodyX, bodyY, 0)
			}
		default:
			panic("unrecognized part name: " + part)
		}
	}

	// put the image on the screen now
	ops := ebiten.DrawImageOptions{}
	if eb.dmgFlicker.show {
		if eb.dmgFlicker.red {
			ops.ColorScale.Scale(10, 1, 1, 1)
		}
	}
	scaledTilesize := config.TileSize * characterScale
	drawX := x - (scaledTilesize * 2)
	drawY := y - scaledTilesize
	rendering.DrawImageWithOps(screen, eb.stagingImg, drawX, drawY, characterScale, &ops)
}

// made this into a function since it will be needed when subtracting arms by equipBody
func (eb EntityBodySet) getEquipBodyOffsetY() float64 {
	if eb.stretchY%2 != 0 {
		// if stretchY is an odd number, offset equip body by -1
		return -1
	}
	return 0
}

func (eb *EntityBodySet) animationFinished() bool {
	if !eb.BodySet.reachedLastFrame {
		return false
	}
	if !eb.ArmsSet.reachedLastFrame {
		return false
	}
	if !eb.LegsSet.reachedLastFrame {
		return false
	}
	if !eb.WeaponSet.PartSrc.None {
		if !eb.WeaponSet.reachedLastFrame {
			return false
		}
	}
	if !eb.WeaponFxSet.PartSrc.None {
		if !eb.WeaponFxSet.reachedLastFrame {
			return false
		}
	}
	if !eb.EquipBodySet.PartSrc.None {
		if !eb.EquipBodySet.reachedLastFrame {
			return false
		}
	}
	if !eb.EquipLegsSet.PartSrc.None {
		if !eb.EquipLegsSet.reachedLastFrame {
			return false
		}
	}
	if !eb.AuxItemSet.PartSrc.None {
		if !eb.AuxItemSet.reachedLastFrame {
			return false
		}
	}
	return true
}

func (eb *EntityBodySet) resetCurrentAnimation() {
	eb.BodySet.animIndex = 0
	eb.EyesSet.animIndex = 0
	eb.HairSet.animIndex = 0
	eb.ArmsSet.animIndex = 0
	eb.LegsSet.animIndex = 0
	eb.EquipBodySet.animIndex = 0
	eb.EquipLegsSet.animIndex = 0
	eb.EquipHeadSet.animIndex = 0
	eb.EquipFeetSet.animIndex = 0
	eb.WeaponSet.animIndex = 0
	eb.WeaponFxSet.animIndex = 0
	eb.AuxItemSet.animIndex = 0

	eb.BodySet.reachedLastFrame = false
	eb.EyesSet.reachedLastFrame = false
	eb.HairSet.reachedLastFrame = false
	eb.ArmsSet.reachedLastFrame = false
	eb.LegsSet.reachedLastFrame = false
	eb.EquipBodySet.reachedLastFrame = false
	eb.EquipLegsSet.reachedLastFrame = false
	eb.EquipHeadSet.reachedLastFrame = false
	eb.EquipFeetSet.reachedLastFrame = false
	eb.WeaponSet.reachedLastFrame = false
	eb.WeaponFxSet.reachedLastFrame = false
	eb.AuxItemSet.reachedLastFrame = false
}

func (eb *EntityBodySet) Update() {
	// FOR DEBUG TICK-BY-TICK
	//
	// if ebiten.IsKeyPressed(ebiten.KeyShiftLeft) {
	// 	speed := 500
	// 	logz.Println("SLOW UPDATE TICK", "tick ms:", speed)
	// 	fmt.Println(eb.GetDebugString())
	// 	time.Sleep(time.Millisecond * time.Duration(speed))
	// }
	if eb.animationTickCount == 0 {
		logz.Panic("animationTickCount appears to be unset")
	}
	eb.ticks++
	if eb.ticks > eb.animationTickCount {
		// SETS: next frame
		eb.ticks = 0
		eb.BodySet.nextFrame(eb.animation)
		eb.ArmsSet.nextFrame(eb.animation)
		eb.LegsSet.nextFrame(eb.animation)
		eb.EquipBodySet.nextFrame(eb.animation)
		eb.EquipLegsSet.nextFrame(eb.animation)
		eb.EquipHeadSet.nextFrame(eb.animation)
		eb.EquipFeetSet.nextFrame(eb.animation)
		eb.WeaponSet.nextFrame(eb.animation)
		eb.WeaponFxSet.nextFrame(eb.animation)
		eb.AuxItemSet.nextFrame(eb.animation)
	}
	// check for a queued animation; and if we are idle, switch to that
	if eb.animation == AnimIdle && eb.nextAnimation != "" {
		res := eb.SetAnimation(eb.nextAnimation, SetAnimationOps{})
		if res.FailedToSet {
			panic("failed to set next animation?")
		}
		if eb.animation != eb.nextAnimation {
			panic("next animation wasn't set?")
		}
		eb.nextAnimation = ""
	}

	// SETS: get current frame
	eb.BodySet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.EyesSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.HairSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.ArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.LegsSet.setCurrentFrame(eb.currentDirection, eb.animation)

	eb.EquipBodySet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.EquipLegsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.EquipHeadSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.EquipFeetSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.WeaponSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.WeaponFxSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.AuxItemSet.setCurrentFrame(eb.currentDirection, eb.animation)

	eb.validate()
	// Warning: Keep this immediately after the above setCurrentFrame calls! This must be set based on whatever image is actually showing.
	// (there was a bug where the body appeared out of place for a single update tick, and the cause was this being after resetCurrentAnimation below)
	eb.nonBodyYOffset = eb.BodySet.getCurrentYOffset(eb.animation, eb.currentDirection)

	// detect end of animation
	if eb.animationFinished() {
		eb.resetCurrentAnimation()
		if eb.stopAnimationOnCompletion {
			eb.StopAnimation()
			eb.stopAnimationOnCompletion = false
		}
	}

	eb.dmgFlicker.update()
}

type SetAnimationOps struct {
	Force     bool // if the body is not idle (already doing another animation) use this option to forcibly override the existing animation
	QueueNext bool // if the body is not idle, use this option to queue the animation to run when the current one is finished
	DoOnce    bool // use this option to specifically only do one iteration of the animation (ex: for sword slashes)
}

type SetAnimationResult struct {
	AlreadySet  bool // this animation is already set
	FailedToSet bool // this animation failed to set for some reason
	Queued      bool // this animation was queued up for next
	Success     bool // this animation successfully set
}

func (res SetAnimationResult) String() string {
	return fmt.Sprintf("%#v", res)
}

// SetAnimation sets an animation. returns if animation was successfully set.
func (eb *EntityBodySet) SetAnimation(animation string, ops SetAnimationOps) SetAnimationResult {
	validateAnimation(animation)
	if animation == eb.animation {
		return SetAnimationResult{AlreadySet: true}
	}
	// if we aren't currently idle and not using the force option, then consider if it should be queued
	if eb.animation != AnimIdle && !ops.Force {
		if ops.QueueNext && eb.nextAnimation == "" {
			eb.nextAnimation = animation
			logz.Println(eb.Name, "next animation queued:", animation)
			return SetAnimationResult{Queued: true}
		}
		// logz.Println(eb.Name, "Force:", ops.Force)
		// logz.Println(eb.Name, "attempted to set animation:", animation, "animation already set:", eb.animation)
		return SetAnimationResult{FailedToSet: true}
	}
	eb.stopAnimationOnCompletion = ops.DoOnce
	eb.animation = animation
	eb.resetCurrentAnimation()
	return SetAnimationResult{Success: true}
}

func (eb *EntityBodySet) StopAnimation() {
	if eb.animation == AnimIdle {
		if eb.nextAnimation != "" {
			logz.Warnln(eb.Name, "stop animation: next animation exists - should we be clearing this??")
		}
		return
	}
	res := eb.SetAnimation(AnimIdle, SetAnimationOps{
		Force: true,
	})
	if res.FailedToSet {
		logz.Println(eb.Name, res)
		panic("failed to stop animation?")
	}
	if eb.animation != AnimIdle {
		panic("animation is not stopped?")
	}
}

func (eb *EntityBodySet) RotateLeft() {
	switch eb.currentDirection {
	case 'L':
		eb.SetDirection('U')
	case 'U':
		eb.SetDirection('R')
	case 'R':
		eb.SetDirection('D')
	case 'D':
		eb.SetDirection('L')
	}
}

func (eb *EntityBodySet) RotateRight() {
	switch eb.currentDirection {
	case 'L':
		eb.SetDirection('D')
	case 'D':
		eb.SetDirection('R')
	case 'R':
		eb.SetDirection('U')
	case 'U':
		eb.SetDirection('L')
	}
}

func (eb *EntityBodySet) SetDirection(dir byte) {
	if dir == eb.currentDirection {
		return
	}
	if eb.IsAttacking() {
		// can't change directions while attacking
		return
	}

	eb._initializeDirection(dir)
}

// Warning: Only use within SetDirection or Load!
// does all the direction changing logic, without the checks to quit early.
func (eb *EntityBodySet) _initializeDirection(dir byte) {
	eb.currentDirection = dir

	eb.BodySet.animIndex = 0
	eb.EyesSet.animIndex = 0
	eb.HairSet.animIndex = 0
	eb.ArmsSet.animIndex = 0
	eb.LegsSet.animIndex = 0

	eb.EquipBodySet.animIndex = 0
	eb.EquipLegsSet.animIndex = 0
	eb.EquipHeadSet.animIndex = 0
	eb.EquipFeetSet.animIndex = 0
	eb.WeaponSet.animIndex = 0
	eb.WeaponFxSet.animIndex = 0
	eb.AuxItemSet.animIndex = 0

	eb.BodySet.setCurrentFrame(dir, AnimWalk)
	eb.EyesSet.setCurrentFrame(dir, AnimWalk)
	eb.HairSet.setCurrentFrame(dir, AnimWalk)
	eb.ArmsSet.setCurrentFrame(dir, AnimWalk)
	eb.LegsSet.setCurrentFrame(dir, AnimWalk)
	eb.EquipBodySet.setCurrentFrame(dir, AnimWalk)
	eb.EquipLegsSet.setCurrentFrame(dir, AnimWalk)
	eb.EquipHeadSet.setCurrentFrame(dir, AnimWalk)
	eb.EquipFeetSet.setCurrentFrame(dir, AnimWalk)
	eb.WeaponSet.setCurrentFrame(dir, AnimWalk)
	eb.WeaponFxSet.setCurrentFrame(dir, AnimWalk)
	eb.AuxItemSet.setCurrentFrame(dir, AnimWalk)
}
