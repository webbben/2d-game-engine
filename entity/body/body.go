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

	// currently equiped items

	EquipBodySet BodyPartSet
	EquipHeadSet BodyPartSet
	WeaponSet    BodyPartSet
	WeaponFxSet  BodyPartSet
	AuxItemSet   BodyPartSet

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
	s += eb.BodySet.animationDebugString(eb.animation, eb.currentDirection) + "\n"
	s += eb.ArmsSet.animationDebugString(eb.animation, eb.currentDirection) + "\n"
	s += eb.EyesSet.animationDebugString(eb.animation, eb.currentDirection) + "\n"
	s += eb.HairSet.animationDebugString(eb.animation, eb.currentDirection) + "\n"
	s += eb.EquipBodySet.animationDebugString(eb.animation, eb.currentDirection) + "\n"
	s += eb.EquipHeadSet.animationDebugString(eb.animation, eb.currentDirection) + "\n"
	s += eb.WeaponSet.animationDebugString(eb.animation, eb.currentDirection) + "\n"
	return s
}

// Load is for loading all body parts, assuming that they all already have PartSrc set. E.g. for after loading from JSON.
func (eb *EntityBodySet) Load() {
	// load body first, since it dictates stretchX and stretchY (which impact several sets)
	eb.SetBody(eb.BodySet.PartSrc, eb.ArmsSet.PartSrc)
	// load head second, since it impacts the hair set
	eb.SetEquipHead(eb.EquipHeadSet.PartSrc)
	eb.SetHair(eb.HairSet.PartSrc)
	eb.SetEyes(eb.EyesSet.PartSrc)
	eb.SetEquipBody(eb.EquipBodySet.PartSrc)
	eb.SetWeapon(eb.WeaponSet.PartSrc, eb.WeaponFxSet.PartSrc)
	eb.SetAuxiliary(eb.AuxItemSet.PartSrc)

	// set an initial direction and ensure img is set
	eb.animation = ANIM_IDLE
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
	eb.EquipHeadSet.validate()
	eb.BodySet.validate()
	eb.WeaponSet.validate()
	eb.WeaponFxSet.validate()
	eb.AuxItemSet.validate()

	eb.validateAuxFrames()
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
func NewEntityBodySet(bodySet, armsSet, hairSet, eyesSet, equipHeadSet, equipBodySet, weaponSet, weaponFxSet, auxSet BodyPartSet, bodyHSV, eyesHSV, hairHSV *HSV) EntityBodySet {
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
		animation:          ANIM_IDLE,
		animationTickCount: 15,
		currentDirection:   'D',
		BodySet:            bodySet,
		BodyHSV:            *bodyHSV,
		ArmsSet:            armsSet,
		HairSet:            hairSet,
		HairHSV:            *hairHSV,
		EyesSet:            eyesSet,
		EyesHSV:            *eyesHSV,
		EquipBodySet:       equipBodySet,
		EquipHeadSet:       equipHeadSet,
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

func (eb *EntityBodySet) SetBody(bodyDef, armDef SelectedPartDef) {
	if bodyDef.None {
		panic("body must be defined")
	}
	if armDef.None {
		panic("arms must be defined")
	}

	eb.BodySet.setImageSource(bodyDef, 0, 0)

	// reload any body parts that are influenced by stretch properties
	// ensure these stretch values are set before calling subtract arms, since it uses equipBodyStretchY
	eb.stretchX = bodyDef.StretchX
	eb.stretchY = bodyDef.StretchY
	if eb.HairSet.HasLoaded() {
		eb.HairSet.load(eb.stretchX, 0)
	}
	if eb.EquipHeadSet.HasLoaded() {
		eb.EquipHeadSet.load(eb.stretchX, 0)
	}
	if eb.EquipBodySet.HasLoaded() {
		eb.EquipBodySet.load(eb.stretchX, eb.stretchY)
	}

	// ensure this is set before calling subtractArms, since it uses this value
	eb.globalOffsetY = float64(bodyDef.OffsetY)

	// arms are directly set with body
	eb.ArmsSet.setImageSource(armDef, 0, 0)
	if eb.EquipBodySet.sourceSet && !eb.EquipBodySet.PartSrc.None {
		// subtract arms by equip body image (remove parts hidden by it)
		eb.subtractArms()
	}
}

func (eb *EntityBodySet) SetEyes(def SelectedPartDef) {
	if def.None {
		panic("eyes must be defined")
	}
	eb.EyesSet.setImageSource(def, 0, 0)
}

func (eb *EntityBodySet) SetHair(def SelectedPartDef) {
	eb.HairSet.setImageSource(def, eb.stretchX, 0)
	if eb.shouldCropHair() {
		eb.cropHair()
	}
}

func (eb *EntityBodySet) ReloadHair() {
	if !eb.HairSet.HasLoaded() {
		logz.Panicln(eb.Name, "tried to reload hair, but hair hasn't been loaded yet")
	}

	eb.HairSet.load(eb.stretchX, 0)
	if eb.shouldCropHair() {
		eb.cropHair()
	}

	eb.HairSet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
}

func (eb *EntityBodySet) SetEquipHead(def SelectedPartDef) {
	eb.EquipHeadSet.setImageSource(def, eb.stretchX, 0)

	// always reload hair when equiping head, since it could either need to crop or un-crop the hair
	if eb.HairSet.HasLoaded() {
		eb.ReloadHair()
	}
	// if we are already in game (animation has been defined) then ensure first frame is set.
	// We do this here for a couple reasons: firstly, so that in the inventory screen, the change is visible immediately.
	// But also, for sets like Hair, if it's nil in draw we panic. So, this ensures that it's not ever nil when the draw function is called.
	if eb.animation != "" {
		eb.EquipHeadSet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
	}
}

func (eb *EntityBodySet) ReloadArms() {
	if !eb.ArmsSet.HasLoaded() {
		logz.Panicln(eb.Name, "trying to reload arms, but they haven't been loaded yet")
	}

	eb.ArmsSet.load(0, 0)
	if !eb.EquipBodySet.PartSrc.None {
		eb.subtractArms()
	}

	if eb.animation != "" {
		eb.ArmsSet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
	}
}

func (eb *EntityBodySet) SetEquipBody(def SelectedPartDef) {
	eb.EquipBodySet.setImageSource(def, eb.stretchX, eb.stretchY)

	// redo the arms subtraction
	if eb.ArmsSet.HasLoaded() {
		eb.ReloadArms()
	}

	if eb.animation != "" {
		eb.EquipBodySet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
	}
}

func (eb *EntityBodySet) SetAuxiliary(def SelectedPartDef) {
	eb.AuxItemSet.setImageSource(def, 0, 0)
}

// IsAuxEquipped determines if an aux item is currently equiped
func (eb EntityBodySet) IsAuxEquipped() bool {
	return !eb.AuxItemSet.PartSrc.None
}

func (eb *EntityBodySet) SetWeapon(weaponDef, weaponFxDef SelectedPartDef) {
	if weaponDef.None != weaponFxDef.None {
		logz.Panicln("SetWeapon", "weapon and weaponFx should have the same None value (so they always equip or unequip together)", "weapon:", weaponDef.None, "weaponFx:", weaponFxDef.None)
	}
	eb.WeaponSet.setImageSource(weaponDef, 0, 0)
	eb.WeaponFxSet.setImageSource(weaponFxDef, 0, 0)
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
	None                           bool // if true, this part will not be shown
	TilesetSrc                     string
	RStart, LStart, UStart, DStart int
	FlipRForL                      bool // if true, instead of using an L source, we just flip the frames for right

	// body-specific props

	StretchX int // amount to stretch hair and equip body on X axis. Defined here, this represents the value that is applied to ALL (applicable) parts - not to this one.
	StretchY int // amount to stretch equip body on the Y axis. Defined here, this represents the value that is applied to ALL (applicable) parts - not to this one.
	OffsetY  int // amount to offset positions of hair, eyes, equip body, etc on the Y axis

	// headwear-specific props

	CropHairToHead bool // set to have hair not go outside the head image. used for helmets or certain hats.

	// aux-specific props

	// if body has aux enabled, this field indicates the step (from origin) to get the first aux frame.
	// for context: when aux is enabled, we replace the 0-index frame with a different frame,
	// since aux animations only have a different first frame from the regular animations.
	// If set to 0, effectively nothing happens, and no aux frame is built.
	AuxFirstFrameStep int
}

// IsEqual checks if the two are equal. mainly used for validation.
func (def SelectedPartDef) IsEqual(other SelectedPartDef) bool {
	// Q: should we factor None into this? if None is true (for both) do we still want to check other fields?
	if def.TilesetSrc != other.TilesetSrc {
		return false
	}
	if !(def.RStart == other.RStart && def.LStart == other.LStart && def.UStart == other.UStart && def.DStart == other.DStart) {
		return false
	}
	if def.FlipRForL != other.FlipRForL {
		return false
	}
	if !(def.StretchX == other.StretchX && def.StretchY == other.StretchY && def.OffsetY == other.OffsetY) {
		return false
	}
	if def.CropHairToHead != other.CropHairToHead {
		return false
	}
	if def.AuxFirstFrameStep != other.AuxFirstFrameStep {
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
	cropper(&eb.HairSet.IdleAnimation)
}

func (eb *EntityBodySet) subtractArms() {
	if eb.EquipBodySet.PartSrc.None {
		panic("trying to subtract arms, but no bodywear is set")
	}
	fmt.Println("subtract arms")
	cropper := func(a *Animation, subtractorA Animation) {
		equipBodyOffsetY := int(eb.globalOffsetY + eb.getEquipBodyOffsetY())
		for i, img := range a.L {
			equipBodyImg := subtractorA.L[i]
			a.L[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg, 0, equipBodyOffsetY)
		}
		for i, img := range a.R {
			equipBodyImg := subtractorA.R[i]
			a.R[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg, 0, equipBodyOffsetY)
		}
		for i, img := range a.U {
			equipBodyImg := subtractorA.U[i]
			a.U[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg, 0, equipBodyOffsetY)
		}
		for i, img := range a.D {
			equipBodyImg := subtractorA.D[i]
			a.D[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg, 0, equipBodyOffsetY)
		}
		a.leftAux = rendering.SubtractImageByOtherImage(a.leftAux, subtractorA.leftAux, 0, equipBodyOffsetY)
		a.rightAux = rendering.SubtractImageByOtherImage(a.rightAux, subtractorA.rightAux, 0, equipBodyOffsetY)
		a.upAux = rendering.SubtractImageByOtherImage(a.upAux, subtractorA.upAux, 0, equipBodyOffsetY)
		a.downAux = rendering.SubtractImageByOtherImage(a.downAux, subtractorA.downAux, 0, equipBodyOffsetY)
	}

	cropper(&eb.ArmsSet.WalkAnimation, eb.EquipBodySet.WalkAnimation)
	cropper(&eb.ArmsSet.RunAnimation, eb.EquipBodySet.RunAnimation)
	cropper(&eb.ArmsSet.SlashAnimation, eb.EquipBodySet.SlashAnimation)
	cropper(&eb.ArmsSet.BackslashAnimation, eb.EquipBodySet.BackslashAnimation)
	cropper(&eb.ArmsSet.IdleAnimation, eb.EquipBodySet.IdleAnimation)

	eb.validateAuxFrames()
}

func (eb EntityBodySet) validateAuxFrames() {
	if eb.ArmsSet.IdleAnimation.leftAux == nil ||
		eb.ArmsSet.IdleAnimation.rightAux == nil ||
		eb.ArmsSet.IdleAnimation.upAux == nil ||
		eb.ArmsSet.IdleAnimation.downAux == nil {
		logz.Panicln(eb.Name, "one or more arms aux frames are nil")
	}
	if !eb.EquipBodySet.PartSrc.None {
		if eb.EquipBodySet.IdleAnimation.leftAux == nil ||
			eb.EquipBodySet.IdleAnimation.rightAux == nil ||
			eb.EquipBodySet.IdleAnimation.upAux == nil ||
			eb.EquipBodySet.IdleAnimation.downAux == nil {
			logz.Panicln(eb.Name, "one or more equipBody aux frames are nil")
		}
	}
}

func (eb *EntityBodySet) Draw(screen *ebiten.Image, x, y, characterScale float64) {
	// Warning: Do not use characterScale anywhere except the bottom - where we draw stagingImg onto screen!
	// we first make a "staging image" which is drawn without scale, and then we draw that image into screen using characterScale.
	eb.stagingImg.Clear()
	// eb.stagingImg.Fill(color.RGBA{100, 0, 0, 50})  // for testing

	// render order decisions (for not so obvious things):
	// - Arms: after equip body, equip head, hair so that hands show when doing U slash (we subtract arms by equip_body)
	renderOrder := []string{"body", "equip_body", "eyes", "hair", "equip_head", "arms", "equip_weapon", "aux"}
	if eb.currentDirection == 'U' {
		renderOrder = []string{"aux", "body", "equip_body", "eyes", "hair", "equip_head", "arms", "equip_weapon"}
	}

	yOff := eb.globalOffsetY

	bodyX := float64(config.TileSize * 2)
	bodyY := float64(config.TileSize)

	equipBodyY := bodyY + yOff + eb.getEquipBodyOffsetY()

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
		case "equip_body":
			if eb.EquipBodySet.img != nil {
				rendering.DrawImage(eb.stagingImg, eb.EquipBodySet.img, bodyX, equipBodyY, 0)
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
	if !eb.WeaponSet.PartSrc.None {
		if !eb.WeaponSet.reachedLastFrame {
			return false
		}
		if !eb.WeaponFxSet.reachedLastFrame {
			return false
		}
	}
	if !eb.EquipBodySet.PartSrc.None {
		if !eb.EquipBodySet.reachedLastFrame {
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
	eb.EquipBodySet.animIndex = 0
	eb.EquipHeadSet.animIndex = 0
	eb.WeaponSet.animIndex = 0
	eb.WeaponFxSet.animIndex = 0
	eb.AuxItemSet.animIndex = 0

	eb.BodySet.reachedLastFrame = false
	eb.EyesSet.reachedLastFrame = false
	eb.HairSet.reachedLastFrame = false
	eb.ArmsSet.reachedLastFrame = false
	eb.EquipBodySet.reachedLastFrame = false
	eb.EquipHeadSet.reachedLastFrame = false
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
		eb.EquipBodySet.nextFrame(eb.animation)
		eb.EquipHeadSet.nextFrame(eb.animation)
		eb.WeaponSet.nextFrame(eb.animation)
		eb.WeaponFxSet.nextFrame(eb.animation)
		eb.AuxItemSet.nextFrame(eb.animation)
	}
	// check for a queued animation; and if we are idle, switch to that
	if eb.animation == ANIM_IDLE && eb.nextAnimation != "" {
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
	eb.BodySet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
	eb.EyesSet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
	eb.HairSet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
	eb.ArmsSet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())

	eb.EquipBodySet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
	eb.EquipHeadSet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
	eb.WeaponSet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
	eb.WeaponFxSet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())
	eb.AuxItemSet.setCurrentFrame(eb.currentDirection, eb.animation, eb.IsAuxEquipped())

	eb.validate()
	// Warning: Keep this immediately after the above setCurrentFrame calls! This must be set based on whatever image is actually showing.
	// (there was a bug where the body appeared out of place for a single update tick, and the cause was this being after resetCurrentAnimation below)
	eb.nonBodyYOffset = eb.BodySet.getCurrentYOffset(eb.animation)

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
	Force     bool
	QueueNext bool
	DoOnce    bool
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
	if eb.animation != ANIM_IDLE && !ops.Force {
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
	if eb.animation == ANIM_IDLE {
		if eb.nextAnimation != "" {
			logz.Warnln(eb.Name, "stop animation: next animation exists - should we be clearing this??")
		}
		return
	}
	res := eb.SetAnimation(ANIM_IDLE, SetAnimationOps{
		Force: true,
	})
	if res.FailedToSet {
		logz.Println(eb.Name, res)
		panic("failed to stop animation?")
	}
	if eb.animation != ANIM_IDLE {
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

	eb.EquipBodySet.animIndex = 0
	eb.EquipHeadSet.animIndex = 0
	eb.WeaponSet.animIndex = 0
	eb.WeaponFxSet.animIndex = 0
	eb.AuxItemSet.animIndex = 0

	eb.BodySet.setCurrentFrame(dir, ANIM_WALK, eb.IsAuxEquipped())
	eb.EyesSet.setCurrentFrame(dir, ANIM_WALK, eb.IsAuxEquipped())
	eb.HairSet.setCurrentFrame(dir, ANIM_WALK, eb.IsAuxEquipped())
	eb.ArmsSet.setCurrentFrame(dir, ANIM_WALK, eb.IsAuxEquipped())
	eb.EquipBodySet.setCurrentFrame(dir, ANIM_WALK, eb.IsAuxEquipped())
	eb.EquipHeadSet.setCurrentFrame(dir, ANIM_WALK, eb.IsAuxEquipped())
	eb.WeaponSet.setCurrentFrame(dir, ANIM_WALK, eb.IsAuxEquipped())
	eb.WeaponFxSet.setCurrentFrame(dir, ANIM_WALK, eb.IsAuxEquipped())
	eb.AuxItemSet.setCurrentFrame(dir, ANIM_WALK, eb.IsAuxEquipped())
}
