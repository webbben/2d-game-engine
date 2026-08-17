// Package body contains all drawing and update logic for entity bodies for moving in worlds
package body

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
)

var Default defs.HSV = defs.HSV{H: 0.5, S: 0.5, V: 0.5}

type EntityBodySet struct {
	Name string

	decreaseHeight int // number of pixels by which to decrease body height

	// TODO: do we need these `json:"-"` parts? I don't think these would get put into a json file anyway, right?

	animation                 string `json:"-"`
	nextAnimation             string `json:"-"`
	stopAnimationOnCompletion bool   `json:"-"`
	holdLastFrame             bool
	animationTickCount        int  `json:"-"` // the "duration" of ticks until the next animation frame should trigger
	ticks                     int  `json:"-"` // number of ticks elapsed
	currentDirection          byte `json:"-"` // L R U D

	dmgFlicker damageFlickerFX `json:"-"`

	stretchX, stretchY int `json:"-"` // amount to stretch certain body parts - set by body set

	// actual body definition - not including equiped items

	stagingImg *ebiten.Image `json:"-"` // just for putting everything together before drawing to screen (for adding flicker fx)

	// body parts

	BodyHSV defs.HSV
	BodySet BodyPartSet
	EyesHSV defs.HSV
	EyesSet BodyPartSet
	HairHSV defs.HSV
	HairSet BodyPartSet
	ArmsSet BodyPartSet
	LegsSet BodyPartSet

	// currently equiped items

	EquipBodySet BodyPartSet // An equiped piece of body armor or shirt, on the entity's torso and arms.
	EquipArmsSet BodyPartSet // The corresponding arms equipment for the body set
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
	s += eb.EquipArmsSet.animationDebugString() + "\n"
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
	eb.SetEquipBody(eb.EquipBodySet.PartSrc, eb.EquipArmsSet.PartSrc)
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
	// NOTE: we don't validate legs anymore, since legs are combined with body (but we still allow a legs set to exist, it's just not really used anymore.)
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
	eb.EquipArmsSet.validate()
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
	eb.BodyHSV = defs.HSV{h, s, v}
}

func (eb EntityBodySet) GetBodyHSV() (h, s, v float64) {
	return eb.BodyHSV.H, eb.BodyHSV.S, eb.BodyHSV.V
}

func (eb *EntityBodySet) SetEyesHSV(h, s, v float64) {
	eb.EyesHSV = defs.HSV{h, s, v}
}

func (eb EntityBodySet) GetEyesHSV() (h, s, v float64) {
	return eb.EyesHSV.H, eb.EyesHSV.S, eb.EyesHSV.V
}

func (eb *EntityBodySet) SetHairHSV(h, s, v float64) {
	eb.HairHSV = defs.HSV{h, s, v}
}

func (eb EntityBodySet) GetHairHSV() (h, s, v float64) {
	return eb.HairHSV.H, eb.HairHSV.S, eb.HairHSV.V
}

// NewHumanBodyFramework returns an empty baseline for a human body; mostly for use in places like character builder
func NewHumanBodyFramework() EntityBodySet {
	bodySet := NewBodyPartSet(BodyPartSetParams{
		Name:   "bodySet",
		IsBody: true,
		HasUp:  true,
	})
	armsSet := NewBodyPartSet(BodyPartSetParams{
		Name:  "armsSet",
		HasUp: true,
	})
	legsSet := NewBodyPartSet(BodyPartSetParams{
		Name:  "legsSet",
		HasUp: true,
	})
	eyesSet := NewBodyPartSet(BodyPartSetParams{Name: "eyesSet"})
	hairSet := NewBodyPartSet(BodyPartSetParams{HasUp: true, Name: "hairSet"})
	equipBodySet := NewBodyPartSet(BodyPartSetParams{
		Name:        "equipBodySet",
		HasUp:       true,
		IsRemovable: true,
	})
	equipArmsSet := NewBodyPartSet(BodyPartSetParams{
		Name:        "equipArmsSet",
		HasUp:       true,
		IsRemovable: true,
	})
	equipHeadSet := NewBodyPartSet(BodyPartSetParams{
		HasUp:       true,
		Name:        "equipHeadSet",
		IsRemovable: true,
	})
	equipFeetSet := NewBodyPartSet(BodyPartSetParams{
		HasUp:       true,
		Name:        "equipFeetSet",
		IsRemovable: true,
	})
	weaponSet := NewBodyPartSet(BodyPartSetParams{
		Name:        "weaponSet",
		HasUp:       true,
		IsRemovable: true,
	})
	weaponFxSet := NewBodyPartSet(BodyPartSetParams{
		Name:        "weaponFxSet",
		HasUp:       true,
		IsRemovable: true,
	})
	auxSet := NewBodyPartSet(BodyPartSetParams{
		Name:        "auxSet",
		HasUp:       true,
		IsRemovable: true,
	})

	entBody := NewEntityBodySet(bodySet, armsSet, legsSet, hairSet, eyesSet, equipHeadSet, equipFeetSet, equipBodySet, equipArmsSet, weaponSet, weaponFxSet, auxSet, nil, nil, nil)
	return entBody
}

// NewEntityBodySet creates a base body set, without anything equiped
func NewEntityBodySet(bodySet, armsSet, legsSet, hairSet, eyesSet, equipHeadSet, equipFeetSet, equipBodySet, equipArmsSet, weaponSet, weaponFxSet, auxSet BodyPartSet, bodyHSV, eyesHSV, hairHSV *defs.HSV) EntityBodySet {
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
		EquipArmsSet:       equipArmsSet,
		EquipHeadSet:       equipHeadSet,
		EquipFeetSet:       equipFeetSet,
		WeaponSet:          weaponSet,
		WeaponFxSet:        weaponFxSet,
		AuxItemSet:         auxSet,
		stagingImg:         ebiten.NewImage(config.TileSize*5, config.TileSize*5),
	}

	return eb
}

func (eb *EntityBodySet) DecreaseHeight() {
	if eb.decreaseHeight != 0 {
		return
	}
	eb.decreaseHeight = 1
	// when decreasing height, we need to compress the height of the following parts:
	// - body
	// - arms?
	// - equipped body
	// - equipped arms?
	//
	// for the body, we want to clip off some of the vertical space in the **middle**, not at the bottom.
	// otherwise, we'd be clipping off part of the feet.
	// same for the arms, since otherwise we'd be clipping off invisible space (since the arms image has transparency around the actual arms image data)
	//
	// so, we need to decide the exact places in the middle of these parts to cut out pixel rows.
	//
	// for equipped body and arms, we can trim in the exact same locations as we do to the corresponding body and arms.
	//
	// for arms, we only trim the idle position. this is because arms go in different angles and directions during some animations, like when swinging a weapon.
	// so it's not as simple as just cutting off some horizontal pixel rows.

	// TODO: for now we are hardcoding these values assuming a 16x16 tilesize and 2 tiles tall body.
	// to make this flexible for other body sizes, we'd need to dynamically decide based on the height of the body.

	bodyTrimY := []int{26}
	armsTrimY := []int{19}

	// configure the trim on the body and equipped body: all animations except dead
	bodyTrimRows := make(map[string][]int)
	armsTrimRows := make(map[string][]int)
	for _, animName := range AllAnimations() {
		if animName == AnimDead {
			continue // skip dead frame
		}
		bodyTrimRows[animName] = bodyTrimY
		armsTrimRows[animName] = armsTrimY
	}

	eb.BodySet.trimRows = bodyTrimRows
	eb.EquipBodySet.trimRows = bodyTrimRows

	eb.ArmsSet.trimRows = armsTrimRows
	eb.EquipArmsSet.trimRows = armsTrimRows

	// rebuild the frames so the trims are applied; load() re-applies the configured trimRows
	eb.BodySet.load(0, 0, eb.IsAuxEquipped())
	eb.EquipBodySet.load(eb.stretchX, eb.stretchY, eb.IsAuxEquipped())
	eb.ArmsSet.load(0, 0, eb.IsAuxEquipped())
	eb.EquipArmsSet.load(eb.stretchX, eb.stretchY, eb.IsAuxEquipped())

	// refresh the current frame images (the body can be drawn without a preceding Update)
	if eb.animation != "" && eb.currentDirection != 0 {
		eb.BodySet.setCurrentFrame(eb.currentDirection, eb.animation)
		eb.EquipBodySet.setCurrentFrame(eb.currentDirection, eb.animation)
		eb.ArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
		eb.EquipArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}
}

func (eb *EntityBodySet) ResetNormalHeight() {
	if eb.decreaseHeight == 0 {
		return
	}
	eb.decreaseHeight = 0
	eb.BodySet.trimRows = nil
	eb.EquipBodySet.trimRows = nil
	eb.ArmsSet.trimRows = nil
	eb.EquipArmsSet.trimRows = nil

	// rebuild the frames untrimmed; load() re-applies the (now empty) trimRows
	eb.BodySet.load(0, 0, eb.IsAuxEquipped())
	eb.EquipBodySet.load(eb.stretchX, eb.stretchY, eb.IsAuxEquipped())
	eb.ArmsSet.load(0, 0, eb.IsAuxEquipped())
	eb.EquipArmsSet.load(eb.stretchX, eb.stretchY, eb.IsAuxEquipped())

	// refresh the current frame images (the body can be drawn without a preceding Update)
	if eb.animation != "" && eb.currentDirection != 0 {
		eb.BodySet.setCurrentFrame(eb.currentDirection, eb.animation)
		eb.EquipBodySet.setCurrentFrame(eb.currentDirection, eb.animation)
		eb.ArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
		eb.EquipArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}
}

// trimAnimation removes the given pixel row from every frame of an animation, then re-pads the image
// back to its original height so that dimensions stay consistent (content below the cut stays in place).
func trimAnimation(anim *Animation, y int) {
	trimDir := func(frames []*ebiten.Image, y int) []*ebiten.Image {
		for i, frame := range frames {
			if frame == nil {
				continue
			}
			img := rendering.RemovePixelRow(frame, y)

			// make the dimensions match the original; shift the content down
			img2 := ebiten.NewImage(img.Bounds().Dx(), img.Bounds().Dy()+1)
			rendering.DrawImage(img2, img, 0, 1, 0)

			frames[i] = img2
		}
		return frames
	}
	anim.L = trimDir(anim.L, y)
	anim.R = trimDir(anim.R, y)
	anim.U = trimDir(anim.U, y)
	anim.D = trimDir(anim.D, y)
}

func (eb *EntityBodySet) Dimensions() (dx, dy int) {
	if eb.BodySet.img == nil {
		panic("body image is nil")
	}
	bounds := eb.BodySet.img.Bounds()
	return bounds.Dx(), bounds.Dy()
}

func (eb *EntityBodySet) SetBody(bodyDef, armDef, legDef defs.SelectedPartDef) {
	if bodyDef.None {
		panic("body must be defined")
	}
	if armDef.None {
		panic("arms must be defined")
	}

	eb.BodySet.setImageSource(bodyDef, 0, 0, eb.IsAuxEquipped())

	// reload any body parts that are influenced by stretch properties
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
	// 2026-07-16 still aren't using stretch. should we consider removing?
	if eb.EquipArmsSet.HasLoaded() {
		eb.EquipArmsSet.load(eb.stretchX, eb.stretchY, eb.IsAuxEquipped())
	}
	if eb.EquipFeetSet.HasLoaded() {
		eb.EquipFeetSet.load(eb.stretchX, 0, eb.IsAuxEquipped())
	}

	eb.globalOffsetY = float64(bodyDef.OffsetY)

	// arms are directly set with body
	eb.ArmsSet.setImageSource(armDef, 0, 0, eb.IsAuxEquipped())

	// legs are also set directly with body
	if !legDef.None {
		eb.LegsSet.setImageSource(legDef, 0, 0, eb.IsAuxEquipped())
	}
}

func (eb *EntityBodySet) SetEyes(def defs.SelectedPartDef) {
	if def.None {
		panic("eyes must be defined")
	}
	eb.EyesSet.setImageSource(def, 0, 0, eb.IsAuxEquipped())
}

func (eb *EntityBodySet) SetHair(def defs.SelectedPartDef) {
	eb.HairSet.setImageSource(def, eb.stretchX, 0, eb.IsAuxEquipped())
	if eb.shouldCropHair() {
		eb.cropHair()
	}
}

func (eb *EntityBodySet) ReloadHair() {
	if !eb.HairSet.HasLoaded() {
		logz.Println(eb.Name, "tried to reload hair, but hair hasn't been loaded yet")
		logz.Panicln("Body", "tried to reload hair, but hair hasn't been loaded yet")
	}

	eb.HairSet.load(eb.stretchX, 0, eb.IsAuxEquipped())
	if eb.shouldCropHair() {
		eb.cropHair()
	}

	eb.HairSet.setCurrentFrame(eb.currentDirection, eb.animation)
}

func (eb *EntityBodySet) SetEquipHead(def defs.SelectedPartDef) {
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

func (eb *EntityBodySet) SetEquipFeet(def defs.SelectedPartDef) {
	eb.EquipFeetSet.setImageSource(def, eb.stretchX, 0, eb.IsAuxEquipped())

	if eb.animation != "" {
		eb.EquipFeetSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}
}

func (eb *EntityBodySet) ReloadArms() {
	if !eb.ArmsSet.HasLoaded() {
		logz.Println(eb.Name, "trying to reload arms, but they haven't been loaded yet")
		logz.Panicln("Body", "trying to reload arms, but they haven't been loaded yet")
	}

	eb.ArmsSet.load(0, 0, eb.IsAuxEquipped())

	if eb.animation != "" {
		eb.ArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}
}

func (eb *EntityBodySet) EquipBodyItem(i defs.ItemDef) {
	if i.Type != defs.TypeBodywear {
		logz.Println("EquipBodyItem", i.ID)
		logz.Panicln("EquipBodyItem", "item is not bodywear")
	}
	eb.SetEquipBody(*i.BodyPartDef, *i.ArmsPartDef)
}

func (eb *EntityBodySet) EquipHeadItem(i defs.ItemDef) {
	if i.Type != defs.TypeHeadwear {
		logz.Println("EquipHeadItem", i.ID)
		logz.Panicln("EquipHeadItem", "item is not headwear")
	}
	eb.SetEquipHead(*i.BodyPartDef)
}

func (eb *EntityBodySet) EquipAuxItem(i defs.ItemDef) {
	if i.Type != defs.TypeAuxiliary {
		logz.Println("EquipAuxItem", i.ID)
		logz.Panicln("EquipAuxItem", "item is not aux")
	}
	eb.SetAuxiliary(*i.BodyPartDef)
}

func (eb *EntityBodySet) EquipWeaponItem(i defs.ItemDef) {
	if i.Type != defs.TypeWeapon {
		logz.Println("EquipWeaponItem", i.ID)
		logz.Panicln("EquipWeaponItem", "item is not weapon")
	}
	weaponPart, fxPart := item.GetWeaponParts(i)
	eb.SetWeapon(weaponPart, fxPart)
}

func (eb *EntityBodySet) EquipFootItem(i defs.ItemDef) {
	if i.Type != defs.TypeFootwear {
		logz.Println("EquipFootItem", i.ID)
		logz.Panicln("EquipFootItem", "item is not footwear")
	}
	eb.SetEquipFeet(*i.BodyPartDef)
}

func (eb *EntityBodySet) SetEquipBody(bodyDef, armsDef defs.SelectedPartDef) {
	eb.EquipBodySet.setImageSource(bodyDef, eb.stretchX, eb.stretchY, eb.IsAuxEquipped())

	eb.EquipArmsSet.setImageSource(armsDef, eb.stretchX, eb.stretchY, eb.IsAuxEquipped())

	if eb.animation != "" {
		eb.EquipBodySet.setCurrentFrame(eb.currentDirection, eb.animation)
		eb.EquipArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}
}

func (eb *EntityBodySet) SetAuxiliary(def defs.SelectedPartDef) {
	eb.AuxItemSet.setImageSource(def, 0, 0, eb.IsAuxEquipped())

	if eb.animation != "" {
		eb.AuxItemSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}

	eb.reloadAuxAffectedParts()
}

func (eb *EntityBodySet) reloadAuxAffectedParts() {
	// equip body
	eb.EquipBodySet.load(eb.stretchX, eb.stretchY, eb.IsAuxEquipped())
	// arms
	eb.ReloadArms()
}

func (eb *EntityBodySet) RemoveAuxiliary() {
	eb.AuxItemSet.Remove()

	eb.reloadAuxAffectedParts()

	// TODO: why do we need to do this? doesn't the above Remove function already handle setting img to nil?
	if eb.animation != "" {
		eb.EquipBodySet.setCurrentFrame(eb.currentDirection, eb.animation)
		eb.ArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	}

	if eb.IsAuxEquipped() {
		logz.Println(eb.Name, "sanity check: just removed auxiliary, but IsAuxEquipped returned true...")
		logz.Panicln("Body", "sanity check: just removed auxiliary, but IsAuxEquipped returned true...")
	}
}

func (eb *EntityBodySet) RemoveHeadwear() {
	eb.EquipHeadSet.Remove()
	// reload hair too, since it may have been cropped by the previously equiped headwear
	eb.ReloadHair()
}

func (eb *EntityBodySet) RemoveFootwear() {
	eb.EquipFeetSet.Remove()
}

func (eb *EntityBodySet) RemoveBodywear() {
	eb.EquipBodySet.Remove()
	eb.EquipArmsSet.Remove()
}

// IsAuxEquipped determines if an aux item is currently equiped.
// An "Aux" item is an item that is held in the left hand (e.g. a torch.).
func (eb EntityBodySet) IsAuxEquipped() bool {
	return !eb.AuxItemSet.PartSrc.None
}

func (eb *EntityBodySet) SetWeapon(weaponDef, weaponFxDef defs.SelectedPartDef) {
	if weaponDef.None != weaponFxDef.None {
		logz.Println("SetWeapon", weaponDef.None, "weaponFx:", weaponFxDef.None)
		logz.Panicln("SetWeapon", "weapon and weaponFx should have the same None value (so they always equip or unequip together)")
	}

	// as of now, we are assuming that weaponFx will never have an idle animation, so setting it to skip here.
	// this is to prevent the weaponFx frames from showing while idle is active.
	if !weaponDef.None {
		fxIdle := weaponFxDef.Animations[AnimIdle]
		fxIdle.Skip = true
		weaponFxDef.Animations[AnimIdle] = fxIdle
	}

	eb.WeaponSet.setImageSource(weaponDef, 0, 0, eb.IsAuxEquipped())
	eb.WeaponFxSet.setImageSource(weaponFxDef, 0, 0, eb.IsAuxEquipped())
}

func (eb *EntityBodySet) RemoveWeapon() {
	eb.WeaponSet.Remove()
	eb.WeaponFxSet.Remove()
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

type PartDefParams struct {
	ID        defs.BodyPartID
	None      bool
	FlipRForL bool // if true, frames for Right directions will be flipped horizontally and reused for the Left direction.
	Female    bool

	// Animation params keyed by animation name (use body.AnimIdle, body.AnimWalk, etc.)
	Anims map[string]*defs.AnimationParams

	StretchX, StretchY int
	OffsetY            int

	CropHairToHead bool
}

// NewPartDef creates a new SelectedPartDef, which essentially defines a specific body part's animations, visuals, etc.
// Use this function to create a SelectedPartDef, rather than directly making the struct, since this will handle some important validation.
func NewPartDef(params PartDefParams) defs.SelectedPartDef {
	if params.None {
		return defs.SelectedPartDef{None: true}
	}
	def := defs.SelectedPartDef{
		ID:             params.ID,
		Female:         params.Female,
		FlipRForL:      params.FlipRForL,
		StretchX:       params.StretchX,
		StretchY:       params.StretchY,
		OffsetY:        params.OffsetY,
		CropHairToHead: params.CropHairToHead,
		Animations:     make(map[string]defs.AnimationParams),
	}
	// Pre-populate with Skip: true defaults so all animations exist in the map
	for _, name := range AllAnimations() {
		def.Animations[name] = defs.AnimationParams{Skip: true, Name: name}
	}
	// Override with caller-provided animation params
	for name, animParams := range params.Anims {
		if animParams != nil {
			anim := *animParams
			anim.Name = name
			def.Animations[name] = anim
		}
	}
	// Validate that non-skip animations have a tileset source
	for _, name := range AllAnimations() {
		ap := def.Animations[name]
		if !ap.Skip && ap.TilesetSrc == "" {
			panic("tilesetSrc must not be empty for animation " + name)
		}
	}

	return def
}

// Requires BodySet and HairSet to be loaded already
func (eb *EntityBodySet) cropHair() {
	eb.BodySet.validate()
	eb.HairSet.validate()

	walkAnim := eb.BodySet.Animations[AnimWalk]
	leftHead := ebiten.NewImage(config.TileSize, config.TileSize)
	rendering.DrawImage(leftHead, walkAnim.L[0], 0, 0, 0)
	rightHead := ebiten.NewImage(config.TileSize, config.TileSize)
	rendering.DrawImage(rightHead, walkAnim.R[0], 0, 0, 0)
	upHead := ebiten.NewImage(config.TileSize, config.TileSize)
	rendering.DrawImage(upHead, walkAnim.U[0], 0, 0, 0)
	downHead := ebiten.NewImage(config.TileSize, config.TileSize)
	rendering.DrawImage(downHead, walkAnim.D[0], 0, 0, 0)

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

	for _, name := range AllAnimations() {
		a := eb.HairSet.Animations[name]
		cropper(&a)
		eb.HairSet.Animations[name] = a
	}
}

func (eb *EntityBodySet) Draw(screen *ebiten.Image, x, y, characterScale float64) {
	// Warning: Do not use characterScale anywhere except the bottom - where we draw stagingImg onto screen!
	// we first make a "staging image" which is drawn without scale, and then we draw that image into screen using characterScale.
	eb.stagingImg.Clear()
	// eb.stagingImg.Fill(color.RGBA{100, 0, 0, 50})  // for testing

	// render order decisions (for not so obvious things):
	// - Arms: after equip body, equip head, hair so that hands show when doing U slash
	renderOrder := []string{"body", "legs", "equip_feet", "equip_body", "eyes", "hair", "equip_head", "arms", "equip_arms", "equip_weapon", "aux"}
	switch eb.currentDirection {
	case model.Directions.Up:
		// aux first: since facing up, aux items (e.g. torches) will generally be covered by everything
		renderOrder = []string{"aux", "body", "legs", "equip_feet", "equip_body", "eyes", "hair", "equip_head", "arms", "equip_arms", "equip_weapon"}
	case model.Directions.Right:
		// aux after arms: shield may cover part of hands, so aux should render after arms
		renderOrder = []string{"body", "legs", "equip_feet", "equip_body", "eyes", "hair", "equip_head", "arms", "equip_arms", "aux", "equip_weapon"}
	}

	yOff := eb.globalOffsetY

	bodyX := float64(config.TileSize * 2)
	bodyY := float64(config.TileSize)

	equipBodyY := bodyY + yOff + eb.getEquipBodyOffsetY()
	// equipFeetY := bodyY + config.TileSize // equip feet tiles are only 16x16
	equipFeetY := equipBodyY // equip feet tiles are now 32px tall (same as body)

	eyesY := bodyY + (float64(eb.nonBodyYOffset)) + yOff
	hairY := bodyY + (float64(eb.nonBodyYOffset)) + yOff

	// TODO: currently weapon frames are 48x64, but in the past I considered a larger 80x80, if needed for things like spears.
	// change this (weaponX subtract by tilesize*2) if we go back to 80x80 frame size for weapon animations
	weaponY := bodyY - (config.TileSize) + yOff
	weaponX := bodyX - (config.TileSize)

	// if decreaseHeight is set, only the head-tracking parts (eyes, hair, equip head) shift down,
	// since the head content was cut 2px lower in the trimmed body frame. everything else stays:
	// body/arms/equip_body/equip_arms/equip_feet track the torso (which doesn't move), and
	// weapon/aux track the hands (which don't move either).
	if eb.decreaseHeight != 0 {
		eyesY += float64(eb.decreaseHeight)
		hairY += float64(eb.decreaseHeight)
	}

	for _, part := range renderOrder {
		switch part {
		case "body":
			rendering.DrawHSVImage(eb.stagingImg, eb.BodySet.img, eb.BodyHSV.H, eb.BodyHSV.S, eb.BodyHSV.V, bodyX, bodyY, 0)
		case "arms":
			if eb.ArmsSet.img != nil {
				rendering.DrawHSVImage(eb.stagingImg, eb.ArmsSet.img, eb.BodyHSV.H, eb.BodyHSV.S, eb.BodyHSV.V, bodyX, bodyY, 0)
			}
		case "legs":
			if eb.LegsSet.img != nil {
				rendering.DrawHSVImage(eb.stagingImg, eb.LegsSet.img, eb.BodyHSV.H, eb.BodyHSV.S, eb.BodyHSV.V, bodyX, bodyY, 0)
			}
		case "equip_body":
			if eb.EquipBodySet.img != nil {
				rendering.DrawImage(eb.stagingImg, eb.EquipBodySet.img, bodyX, equipBodyY, 0)
			}
		case "equip_arms":
			if eb.EquipArmsSet.img != nil {
				rendering.DrawImage(eb.stagingImg, eb.EquipArmsSet.img, bodyX, equipBodyY, 0)
			}
		case "eyes":
			if eb.EyesSet.img != nil {
				rendering.DrawHSVImage(eb.stagingImg, eb.EyesSet.img, eb.EyesHSV.H, eb.EyesHSV.S, eb.EyesHSV.V, bodyX, eyesY, 0)
			}
		case "hair":
			if eb.HairSet.img != nil {
				rendering.DrawHSVImage(eb.stagingImg, eb.HairSet.img, eb.HairHSV.H, eb.HairHSV.S, eb.HairHSV.V, bodyX, hairY, 0)
			}
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
				rendering.DrawImage(eb.stagingImg, eb.AuxItemSet.img, weaponX, weaponY, 0)
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

// AnimationFinished returns true if all body parts have finished their animation sequences.
// Not really meant to be used unless HoldLastFrame is true and the outside needs to know if it's on the last frame yet.
func (eb *EntityBodySet) AnimationFinished() bool {
	if !eb.BodySet.reachedLastFrame {
		return false
	}
	if !eb.ArmsSet.reachedLastFrame {
		return false
	}
	if !eb.LegsSet.PartSrc.None {
		if !eb.LegsSet.reachedLastFrame {
			return false
		}
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
	if !eb.EquipArmsSet.PartSrc.None {
		if !eb.EquipArmsSet.reachedLastFrame {
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
	eb.EquipArmsSet.animIndex = 0
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
	eb.EquipArmsSet.reachedLastFrame = false
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
		eb.EquipArmsSet.nextFrame(eb.animation)
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
	eb.EquipArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.EquipHeadSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.EquipFeetSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.WeaponSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.WeaponFxSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.AuxItemSet.setCurrentFrame(eb.currentDirection, eb.animation)

	eb.validate()
	// Warning: Keep this immediately after the above setCurrentFrame calls! This must be set based on whatever image is actually showing.
	// (there was a bug where the body appeared out of place for a single update tick, and the cause was this being after resetCurrentAnimation below)
	eb.nonBodyYOffset = eb.BodySet.getCurrentYOffset(eb.animation, eb.currentDirection)

	// detect end of animation and reset, unless supposed to hold last frame
	if eb.AnimationFinished() && !eb.holdLastFrame {
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
	// use this option to specifically only do one iteration of the animation before reverting back to idle (ex: for sword slashes). otherwise, animation loops.
	// Not allowed to be used with HoldLastFrame.
	DoOnce bool
	// if set, animation will hold the last frame and not end, until something else explicitly changes the animation again.
	// Not allowed to be used with DoOnce.
	HoldLastFrame bool
}

type SetAnimationResult struct {
	AlreadySet  bool   // this animation is already set
	FailedToSet bool   // this animation failed to set for some reason
	Queued      bool   // this animation was queued up for next
	Success     bool   // this animation successfully set
	Msg         string // any extra context or information for debugging
}

func (res SetAnimationResult) String() string {
	result := ""
	if res.Success {
		result = "success"
	}
	if res.AlreadySet {
		result = "already set"
	}
	if res.FailedToSet {
		result = "failed to set"
	}
	if res.Queued {
		result = "queued"
	}
	return fmt.Sprintf("%s;%s", result, res.Msg)
}

// SetAnimation sets an animation. returns if animation was successfully set.
func (eb *EntityBodySet) SetAnimation(animation string, ops SetAnimationOps) SetAnimationResult {
	validateAnimation(animation)
	if ops.DoOnce && ops.HoldLastFrame {
		logz.Panic("both DoOnce and HoldLastFrame were set to true")
	}
	if animation == eb.animation {
		return SetAnimationResult{AlreadySet: true, Msg: fmt.Sprintf("current animation: %s", eb.animation)}
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
		return SetAnimationResult{FailedToSet: true, Msg: fmt.Sprintf("current anim: %s, next anim: %s, tried to queue?: %v", eb.animation, eb.nextAnimation, ops.QueueNext)}
	}
	eb.stopAnimationOnCompletion = ops.DoOnce
	eb.holdLastFrame = ops.HoldLastFrame
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
	eb.EquipArmsSet.animIndex = 0
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
	eb.EquipArmsSet.setCurrentFrame(dir, AnimWalk)
	eb.EquipHeadSet.setCurrentFrame(dir, AnimWalk)
	eb.EquipFeetSet.setCurrentFrame(dir, AnimWalk)
	eb.WeaponSet.setCurrentFrame(dir, AnimWalk)
	eb.WeaponFxSet.setCurrentFrame(dir, AnimWalk)
	eb.AuxItemSet.setCurrentFrame(dir, AnimWalk)
}
