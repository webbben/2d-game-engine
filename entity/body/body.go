package body

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

type HSV struct {
	H, S, V float64
}

var Default HSV = HSV{0.5, 0.5, 0.5}

type EntityBodySet struct {
	animation          string `json:"-"`
	animationTickCount int    `json:"-"`
	ticks              int    `json:"-"` // number of ticks elapsed
	currentDirection   byte   `json:"-"` // L R U D
	cropHairToHead     bool   `json:"-"`

	// actual body definition - not including equiped items

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
	eb.HairSet.load()

	if def.CropHairToHead {
		eb.cropHair()
	}
}

func (eb *EntityBodySet) SetEquipBody(def SelectedPartDef) {
	eb.EquipBodySet.setImageSource(def)
}

func (eb *EntityBodySet) SetWeapon(weaponDef, weaponFxDef SelectedPartDef) {
	if eb.WeaponSet.None {
		panic("no weapon set!")
	}
	if eb.WeaponFxSet.None {
		panic("no weaponFx set!")
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

// represents either the head, body, eyes, or hair of an entity.
//
// Defines the animation patterns for each body part, so this is required to be defined for each entity.
// The actual body part definitions (which tiles to show for hair, eyes, etc) are defined by the TilesetSrc and start indices, and can be set
// using the set functions.
type BodyPartSet struct {
	// tileset and image source definitions
	PartSrc SelectedPartDef

	TilesetSrc                     string
	RStart, LStart, UStart, DStart int
	FlipRForL                      bool
	None                           bool
	stretchX, stretchY             int

	// animation definitions

	animIndex          int
	WalkAnimation      Animation
	RunAnimation       Animation
	SlashAnimation     Animation
	BackslashAnimation Animation
	HasUp              bool

	img *ebiten.Image `json:"-"`
}

func (bps BodyPartSet) validate() {
	if bps.None {
		return
	}
	fmt.Println(bps.TilesetSrc)
	bps.WalkAnimation.validate()
	bps.RunAnimation.validate()
	bps.SlashAnimation.validate()
}

// for no body part
var NONE BodyPartSet = BodyPartSet{None: true}

func (bps *BodyPartSet) unsetAllImages() {
	bps.WalkAnimation.reset()
	bps.RunAnimation.reset()
	bps.SlashAnimation.reset()
	bps.img = nil
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

func (bps *BodyPartSet) setImageSource(def SelectedPartDef) {
	bps.TilesetSrc = def.TilesetSrc
	bps.LStart = def.LStart
	bps.RStart = def.RStart
	bps.UStart = def.UStart
	bps.DStart = def.DStart
	bps.FlipRForL = def.FlipRForL
	bps.None = def.None

	// set this so the part can be reloaded later
	bps.PartSrc = def

	bps.load()
}

func (set *BodyPartSet) load() {
	set.WalkAnimation.reset()
	set.RunAnimation.reset()
	set.SlashAnimation.reset()

	if set.None {
		return
	}

	if set.TilesetSrc == "" {
		panic("no TilesetSrc set in BodyPartSet. has an option been set yet?")
	}

	set.WalkAnimation.Name = fmt.Sprintf("%s/walk", set.TilesetSrc)
	set.RunAnimation.Name = fmt.Sprintf("%s/run", set.TilesetSrc)
	set.SlashAnimation.Name = fmt.Sprintf("%s/slash", set.TilesetSrc)
	set.BackslashAnimation.Name = fmt.Sprintf("%s/backslash", set.TilesetSrc)

	set.WalkAnimation.loadFrames(set.TilesetSrc, set.RStart, set.LStart, set.UStart, set.DStart, set.stretchX, set.stretchY, set.FlipRForL, set.HasUp)
	set.RunAnimation.loadFrames(set.TilesetSrc, set.RStart, set.LStart, set.UStart, set.DStart, set.stretchX, set.stretchY, set.FlipRForL, set.HasUp)
	set.SlashAnimation.loadFrames(set.TilesetSrc, set.RStart, set.LStart, set.UStart, set.DStart, set.stretchX, set.stretchY, set.FlipRForL, set.HasUp)
	set.BackslashAnimation.loadFrames(set.TilesetSrc, set.RStart, set.LStart, set.UStart, set.DStart, set.stretchX, set.stretchY, set.FlipRForL, set.HasUp)
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

func (set *BodyPartSet) setCurrentFrame(dir byte, animationName string) {
	if set.None {
		set.img = nil
		return
	}
	if dir == 'U' && !set.HasUp {
		set.img = nil
		return
	}

	switch animationName {
	case ANIM_WALK:
		set.img = set.WalkAnimation.getFrame(dir, set.animIndex)
	case ANIM_RUN:
		set.img = set.RunAnimation.getFrame(dir, set.animIndex)
	case ANIM_SLASH:
		set.img = set.SlashAnimation.getFrame(dir, set.animIndex)
	case ANIM_BACKSLASH:
		set.img = set.BackslashAnimation.getFrame(dir, set.animIndex)
	case "":
		set.img = set.WalkAnimation.getFrame(dir, 0)
	default:
		panic("unrecognized animation name: " + animationName)
	}
}

func (set BodyPartSet) getCurrentYOffset(animationName string) int {
	switch animationName {
	case ANIM_WALK:
		if len(set.WalkAnimation.StepsOffsetY) > 0 {
			return set.WalkAnimation.StepsOffsetY[set.animIndex]
		}
	case ANIM_RUN:
		if len(set.RunAnimation.StepsOffsetY) > 0 {
			return set.RunAnimation.StepsOffsetY[set.animIndex]
		}
	case ANIM_SLASH:
		if len(set.SlashAnimation.StepsOffsetY) > 0 {
			return set.SlashAnimation.StepsOffsetY[set.animIndex]
		}
	case ANIM_BACKSLASH:
		if len(set.BackslashAnimation.StepsOffsetY) > 0 {
			return set.BackslashAnimation.StepsOffsetY[set.animIndex]
		}
	}

	return 0
}

func (set *BodyPartSet) nextFrame(animationName string) {
	if set.None {
		return
	}

	set.animIndex++
	switch animationName {
	case ANIM_WALK:
		if set.animIndex >= len(set.WalkAnimation.TileSteps) {
			set.animIndex = 0
		}
	case ANIM_RUN:
		if set.animIndex >= len(set.RunAnimation.TileSteps) {
			set.animIndex = 0
		}
	case ANIM_SLASH:
		if set.animIndex >= len(set.SlashAnimation.TileSteps) {
			set.animIndex = 0
		}
	case ANIM_BACKSLASH:
		if set.animIndex >= len(set.BackslashAnimation.TileSteps) {
			set.animIndex = 0
		}
	}
}

func (eb *EntityBodySet) Draw(screen *ebiten.Image, x, y, characterScale float64) {
	bodyX := x
	bodyY := y
	yOff := eb.globalOffsetY * characterScale
	characterTileSize := config.TileSize * characterScale
	// Body
	rendering.DrawHSVImage(screen, eb.BodySet.img, eb.BodyHSV.H, eb.BodyHSV.S, eb.BodyHSV.V, bodyX, bodyY, characterScale)
	// Arms
	rendering.DrawHSVImage(screen, eb.ArmsSet.img, eb.BodyHSV.H, eb.BodyHSV.S, eb.BodyHSV.V, bodyX, bodyY, characterScale)
	// Equip Body
	equipBodyYOffset := 0.0
	if eb.EquipBodySet.stretchY%2 != 0 {
		// if stretchY is an odd number, offset equip body by -1
		equipBodyYOffset = -characterScale
	}
	rendering.DrawImage(screen, eb.EquipBodySet.img, bodyX, bodyY+yOff+equipBodyYOffset, characterScale)

	// Eyes
	eyesX := bodyX
	eyesY := bodyY + (float64(eb.nonBodyYOffset) * characterScale) + yOff
	if eb.EyesSet.img != nil {
		rendering.DrawHSVImage(screen, eb.EyesSet.img, eb.EyesHSV.H, eb.EyesHSV.S, eb.EyesHSV.V, eyesX, eyesY, characterScale)
	}
	// Hair
	hairY := bodyY + (float64(eb.nonBodyYOffset) * characterScale) + yOff
	if eb.HairSet.img == nil {
		panic("hair img is nil")
	}
	rendering.DrawHSVImage(screen, eb.HairSet.img, eb.HairHSV.H, eb.HairHSV.S, eb.HairHSV.V, bodyX, hairY, characterScale)

	// Equip Head
	if eb.EquipHeadSet.img != nil {
		rendering.DrawImage(screen, eb.EquipHeadSet.img, bodyX, hairY, characterScale)
	}

	// Equip Weapon
	if eb.WeaponSet.img != nil {
		// weapons are in 80x80 (5 tiles width & height) tiles
		// this is to accomodate for the extra space they need for their swings and stuff
		weaponY := bodyY - (characterTileSize) + yOff
		weaponX := bodyX - (characterTileSize * 2)
		rendering.DrawImage(screen, eb.WeaponSet.img, weaponX, weaponY, characterScale)
		if eb.WeaponFxSet.img != nil {
			rendering.DrawImage(screen, eb.WeaponFxSet.img, weaponX, weaponY, characterScale)
		}
	}
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

	eb.nonBodyYOffset = eb.BodySet.getCurrentYOffset(eb.animation)
}

func (eb *EntityBodySet) SetAnimation(animation string) {
	if animation == eb.animation {
		return
	}

	eb.animation = animation

	// SETS: reset animation index
	eb.BodySet.animIndex = 0
	eb.EyesSet.animIndex = 0
	eb.HairSet.animIndex = 0
	eb.ArmsSet.animIndex = 0
	eb.EquipBodySet.animIndex = 0
	eb.EquipHeadSet.animIndex = 0
	eb.WeaponSet.animIndex = 0
	eb.WeaponFxSet.animIndex = 0
}

func (eb *EntityBodySet) StopAnimation() {
	eb.SetAnimation("")
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
