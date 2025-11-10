package body

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	preventAnimationSkip      bool   `json:"-"` // if set, the current animation cannot be skipped
	stopAnimationOnCompletion bool   `json:"-"`
	animationTickCount        int    `json:"-"`
	ticks                     int    `json:"-"` // number of ticks elapsed
	currentDirection          byte   `json:"-"` // L R U D
	cropHairToHead            bool   `json:"-"`

	dmgFlicker damageFlickerFX `json:"-"`

	// actual body definition - not including equiped items

	stagingImg *ebiten.Image // just for putting everything together before drawing to screen (for adding flicker fx)

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

	globalOffsetY  float64 `json:"-"` // amount to offset placement of (non-body) parts by, when body is taller or shorter
	nonBodyYOffset int     `json:"-"`
}

func (eb *EntityBodySet) Load() {
	eb.SetHair(eb.HairSet.PartSrc)
	eb.SetEyes(eb.EyesSet.PartSrc)
	eb.SetEquipBody(eb.EquipBodySet.PartSrc)
	eb.SetEquipHead(eb.EquipHeadSet.PartSrc)
	eb.SetBody(eb.BodySet.PartSrc, eb.ArmsSet.PartSrc)
	eb.SetWeapon(eb.WeaponSet.PartSrc, eb.WeaponFxSet.PartSrc)

	// set an initial direction and ensure img is set
	eb.animation = ""
	eb.SetDirection(model.Directions.Down)
	if eb.BodySet.img == nil {
		panic("body image is nil!")
	}

	// make sure everything looks correct
	eb.HairSet.validate()
	eb.EyesSet.validate()
	eb.EquipBodySet.validate()
	eb.EquipHeadSet.validate()
	eb.BodySet.validate()
	eb.WeaponSet.validate()
	eb.WeaponFxSet.validate()

	tilesize := config.TileSize
	eb.stagingImg = ebiten.NewImage(tilesize*5, tilesize*5)
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

func (eb EntityBodySet) WriteToJSON(outputFilePath string) error {
	if !filepath.IsAbs(outputFilePath) {
		return fmt.Errorf("given path is not abs (%s); please pass an absolute path", outputFilePath)
	}

	data, err := json.MarshalIndent(eb, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return os.WriteFile(outputFilePath, data, 0644)
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

// creates a base body set, without anything equiped
func NewEntityBodySet(bodySet, armsSet, hairSet, eyesSet, equipHeadSet, equipBodySet, weaponSet, weaponFxSet BodyPartSet, bodyHSV, eyesHSV, hairHSV *HSV) EntityBodySet {
	if bodyHSV == nil {
		bodyHSV = &Default
	}
	if eyesHSV == nil {
		eyesHSV = &Default
	}
	if hairHSV == nil {
		hairHSV = &Default
	}
	if bodySet.None {
		panic("body must not be none")
	}
	if armsSet.None {
		panic("arms must not be none")
	}
	if eyesSet.None {
		panic("eyes must not be none")
	}

	eb := EntityBodySet{
		animation:          "",
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

	eb.BodySet.setImageSource(bodyDef)

	// arms are directly set with body
	eb.ArmsSet.setImageSource(armDef)
	if eb.EquipBodySet.sourceSet {
		// subtract arms by equip body image (remove parts hidden by it)
		eb.subtractArms()
	}

	eb.globalOffsetY = float64(bodyDef.OffsetY)

	// reload any body parts that are influenced by stretch properties
	eb.HairSet.stretchX, eb.HairSet.stretchY = bodyDef.StretchX, 0
	eb.HairSet.load()
	eb.EquipHeadSet.stretchX, eb.EquipHeadSet.stretchY = bodyDef.StretchX, 0
	eb.EquipHeadSet.load()
	eb.EquipBodySet.stretchX, eb.EquipBodySet.stretchY = bodyDef.StretchX, bodyDef.StretchY
	eb.EquipBodySet.load()
}

func (eb *EntityBodySet) SetEyes(def SelectedPartDef) {
	if def.None {
		panic("eyes must be defined")
	}
	eb.EyesSet.setImageSource(def)
}

func (eb *EntityBodySet) SetHair(def SelectedPartDef) {
	eb.HairSet.setImageSource(def)
	if eb.cropHairToHead {
		eb.cropHair()
	}
}

func (eb *EntityBodySet) SetEquipHead(def SelectedPartDef) {
	eb.EquipHeadSet.setImageSource(def)

	// since some head equipment may cause hair to be cropped, always reload hair when head is equiped
	if eb.HairSet.sourceSet {
		eb.HairSet.load()

		if def.CropHairToHead {
			eb.cropHairToHead = true
			eb.cropHair()
		} else {
			eb.cropHairToHead = false
		}
	}
}

func (eb *EntityBodySet) SetEquipBody(def SelectedPartDef) {
	eb.EquipBodySet.setImageSource(def)

	// redo the arms subtraction
	if eb.ArmsSet.sourceSet {
		eb.ArmsSet.load()
		eb.subtractArms()
	}
}

func (eb *EntityBodySet) SetWeapon(weaponDef, weaponFxDef SelectedPartDef) {
	if eb.WeaponSet.None {
		return
	}
	if eb.WeaponFxSet.None {
		panic("no weaponFx set when setting weapon!")
	}
	eb.WeaponSet.setImageSource(weaponDef)
	eb.WeaponFxSet.setImageSource(weaponFxDef)
}

func (eb EntityBodySet) GetCurrentAnimation() string {
	return eb.animation
}

func (eb *EntityBodySet) SetAnimationTickCount(tickCount int) {
	eb.animationTickCount = tickCount
}

// TODO choose a better name. Maybe BodyPartDef?
// represents the currently selected body part and it's individual definition
type SelectedPartDef struct {
	None                           bool // if true, this part will not be shown
	TilesetSrc                     string
	RStart, LStart, UStart, DStart int
	FlipRForL                      bool // if true, instead of using an L source, we just flip the frames for right

	// body-specific props

	StretchX int // amount to stretch hair and equip body on X axis
	StretchY int // amount to stretch equip body on the Y axis
	OffsetY  int // amount to offset positions of hair, eyes, equip body, etc on the Y axis

	// headwear-specific props

	CropHairToHead bool // set to have hair not go outside the head image. used for helmets or certain hats.
}

func (eb *EntityBodySet) cropHair() {
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
}

func (eb *EntityBodySet) subtractArms() {
	fmt.Println("subtract arms")
	cropper := func(a *Animation, subtractorA Animation) {
		for i, img := range a.L {
			equipBodyImg := subtractorA.L[i]
			a.L[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg)
		}
		for i, img := range a.R {
			equipBodyImg := subtractorA.R[i]
			a.R[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg)
		}
		for i, img := range a.U {
			equipBodyImg := subtractorA.U[i]
			a.U[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg)
		}
		for i, img := range a.D {
			equipBodyImg := subtractorA.D[i]
			a.D[i] = rendering.SubtractImageByOtherImage(img, equipBodyImg)
		}
	}

	cropper(&eb.ArmsSet.WalkAnimation, eb.EquipBodySet.WalkAnimation)
	cropper(&eb.ArmsSet.RunAnimation, eb.EquipBodySet.RunAnimation)
	cropper(&eb.ArmsSet.SlashAnimation, eb.EquipBodySet.SlashAnimation)
}

func (eb *EntityBodySet) Draw(screen *ebiten.Image, x, y, characterScale float64) {
	eb.stagingImg.Clear()
	//eb.stagingImg.Fill(color.RGBA{100, 0, 0, 50})  // for testing
	// render order decisions (for not so obvious things):
	// - Arms: after equip body, equip head, hair so that hands show when doing U slash
	renderOrder := []string{"body", "equip_body", "eyes", "hair", "equip_head", "arms", "equip_weapon"}

	yOff := eb.globalOffsetY

	bodyX := float64(config.TileSize * 2)
	bodyY := float64(config.TileSize)

	equipBodyYOffset := 0.0
	if eb.EquipBodySet.stretchY%2 != 0 {
		// if stretchY is an odd number, offset equip body by -1
		equipBodyYOffset = -characterScale
	}
	equipBodyY := bodyY + yOff + equipBodyYOffset

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
			rendering.DrawImage(eb.stagingImg, eb.EquipBodySet.img, bodyX, equipBodyY, 0)
		case "eyes":
			if eb.EyesSet.img != nil {
				rendering.DrawHSVImage(eb.stagingImg, eb.EyesSet.img, eb.EyesHSV.H, eb.EyesHSV.S, eb.EyesHSV.V, bodyX, eyesY, 0)
			}
		case "hair":
			if eb.HairSet.img == nil {
				panic("hair img is nil")
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

func (eb *EntityBodySet) animationFinished() bool {
	if !eb.BodySet.reachedLastFrame {
		return false
	}
	if !eb.ArmsSet.reachedLastFrame {
		return false
	}
	if !eb.WeaponSet.None {
		if !eb.WeaponSet.reachedLastFrame {
			return false
		}
		if !eb.WeaponFxSet.reachedLastFrame {
			return false
		}
	}
	if !eb.EquipBodySet.None {
		if !eb.EquipBodySet.reachedLastFrame {
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

	eb.BodySet.reachedLastFrame = false
	eb.EyesSet.reachedLastFrame = false
	eb.HairSet.reachedLastFrame = false
	eb.ArmsSet.reachedLastFrame = false
	eb.EquipBodySet.reachedLastFrame = false
	eb.EquipHeadSet.reachedLastFrame = false
	eb.WeaponSet.reachedLastFrame = false
	eb.WeaponFxSet.reachedLastFrame = false
}

func (eb *EntityBodySet) Update() {
	if eb.animation != "" {
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
		}
	} else {
		if eb.nextAnimation != "" {
			res := eb.SetAnimation(eb.nextAnimation, SetAnimationOps{})
			if res.FailedToSet {
				panic("failed to set next animation?")
			}
			if eb.animation != eb.nextAnimation {
				panic("next animation wasn't set?")
			}
			eb.nextAnimation = ""
		}
	}

	// SETS: get current frame
	eb.BodySet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.EyesSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.HairSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.ArmsSet.setCurrentFrame(eb.currentDirection, eb.animation)

	eb.EquipBodySet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.EquipHeadSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.WeaponSet.setCurrentFrame(eb.currentDirection, eb.animation)
	eb.WeaponFxSet.setCurrentFrame(eb.currentDirection, eb.animation)

	// detect end of animation
	if eb.animationFinished() {
		eb.resetCurrentAnimation()
		eb.preventAnimationSkip = false
		if eb.stopAnimationOnCompletion {
			eb.StopAnimation()
			eb.stopAnimationOnCompletion = false
		}
	}

	eb.dmgFlicker.update()

	eb.nonBodyYOffset = eb.BodySet.getCurrentYOffset(eb.animation)
}

type SetAnimationOps struct {
	Force       bool
	QueueNext   bool
	PreventSkip bool
	DoOnce      bool
}

type SetAnimationResult struct {
	AlreadySet  bool
	FailedToSet bool
	Queued      bool
	Success     bool
}

// sets an animation. returns if animation was successfully set.
func (eb *EntityBodySet) SetAnimation(animation string, ops SetAnimationOps) SetAnimationResult {
	if animation == eb.animation {
		return SetAnimationResult{AlreadySet: true}
	}
	if eb.animation != "" && (!ops.Force || eb.preventAnimationSkip) {
		if ops.QueueNext && eb.nextAnimation == "" {
			eb.nextAnimation = animation
			logz.Println(eb.Name, "next animation queued:", animation)
			return SetAnimationResult{Queued: true}
		}
		logz.Println(eb.Name, "animation already set, and force is not enabled")
		return SetAnimationResult{FailedToSet: true}
	}
	eb.preventAnimationSkip = ops.PreventSkip
	eb.stopAnimationOnCompletion = ops.DoOnce
	eb.animation = animation
	eb.resetCurrentAnimation()
	return SetAnimationResult{Success: true}
}

func (eb *EntityBodySet) StopAnimation() {
	if eb.animation == "" {
		panic("trying to stop animation, but animation already unset?")
	}
	res := eb.SetAnimation("", SetAnimationOps{
		Force: true,
	})
	if res.FailedToSet {
		panic("failed to stop animation?")
	}
	if eb.animation != "" {
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

	// SETS: reset animation index
	eb.BodySet.animIndex = 0
	eb.EyesSet.animIndex = 0
	eb.HairSet.animIndex = 0
	eb.ArmsSet.animIndex = 0

	eb.EquipBodySet.animIndex = 0
	eb.EquipHeadSet.animIndex = 0
	eb.WeaponSet.animIndex = 0
	eb.WeaponFxSet.animIndex = 0

	eb.currentDirection = dir

	// SETS: set to first frame of walking animation
	eb.BodySet.setCurrentFrame(dir, ANIM_WALK)
	eb.EyesSet.setCurrentFrame(dir, ANIM_WALK)
	eb.HairSet.setCurrentFrame(dir, ANIM_WALK)
	eb.ArmsSet.setCurrentFrame(dir, ANIM_WALK)
	eb.EquipBodySet.setCurrentFrame(dir, ANIM_WALK)
	eb.EquipHeadSet.setCurrentFrame(dir, ANIM_WALK)
	eb.WeaponSet.setCurrentFrame(dir, ANIM_WALK)
	eb.WeaponFxSet.setCurrentFrame(dir, ANIM_WALK)
}
