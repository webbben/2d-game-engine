/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui/button"
	"github.com/webbben/2d-game-engine/internal/ui/dropdown"
	"github.com/webbben/2d-game-engine/internal/ui/slider"
	"github.com/webbben/2d-game-engine/internal/ui/stepper"
)

// characterBuilderCmd represents the characterBuilder command
var characterBuilderCmd = &cobra.Command{
	Use:   "characterBuilder",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("CHARACTER BUILDER")

		display.SetupGameDisplay("CHARACTER BUILDER", false)

		config.GameDataPathOverride = "/Users/benwebb/dev/personal/ancient-rome"
		config.DefaultFont = image.LoadFont("ashlander-pixel.ttf", 22, 0)
		config.DefaultTitleFont = image.LoadFont("ashlander-pixel.ttf", 28, 0)

		err := config.InitFileStructure()
		if err != nil {
			panic(err)
		}
		characterBuilder()
	},
}

func init() {
	rootCmd.AddCommand(characterBuilderCmd)
}

type builderGame struct {
	animation          string
	animationTickCount int
	ticks              int
	currentDirection   byte // L R U D
	bodyImg            *ebiten.Image
	eyesImg            *ebiten.Image
	hairImg            *ebiten.Image
	armsImg            *ebiten.Image
	equipBodyImg       *ebiten.Image
	equipHeadImg       *ebiten.Image
	weaponImg          *ebiten.Image
	weaponFxImg        *ebiten.Image
	nonBodyYOffset     int

	bodySet         bodyPartSet
	bodyOptionIndex int
	bodyOptions     []SelectedPartDef
	eyesSet         bodyPartSet
	eyesOptionIndex int
	eyesOptions     []SelectedPartDef
	hairSet         bodyPartSet
	hairOptionIndex int
	hairOptions     []SelectedPartDef
	armsSet         bodyPartSet
	armsOptionIndex int
	armsOptions     []SelectedPartDef

	weaponSet         bodyPartSet
	weaponOptionIndex int
	weaponOptions     []SelectedPartDef
	weaponFxSet       bodyPartSet

	equipBodySet         bodyPartSet
	equipBodyOptionIndex int
	equipBodyOptions     []SelectedPartDef
	equipHeadSet         bodyPartSet
	equipHeadOptionIndex int
	equipHeadOptions     []SelectedPartDef

	// UI components

	turnLeft          *button.Button
	turnRight         *button.Button
	speedSlider       slider.Slider
	scaleSlider       slider.Slider
	animationSelector dropdown.OptionSelect

	hairCtl      stepper.Stepper
	eyesCtl      stepper.Stepper
	equipBodyCtl stepper.Stepper
	equipHeadCtl stepper.Stepper
}

const (
	anim_walk  = "walk"
	anim_run   = "run"
	anim_slash = "slash"
	anim_stab  = "stab"
)

func characterBuilder() {
	bodyTileset := "entities/parts/body.tsj"
	armsTileset := "entities/parts/arms.tsj"
	eyesTileset := "entities/parts/eyes.tsj"
	hairTileset := "entities/parts/hair.tsj"
	equipBodyTileset := "items/equiped_body_01.tsj"
	equipHeadTileset := "items/equiped_head_01.tsj"
	equipWeaponTileset := "items/weapon_frames.tsj"
	weaponFxTileset := "items/weapon_fx_frames.tsj"

	bodyOptions := []SelectedPartDef{
		{
			TilesetSrc: bodyTileset,
			DStart:     0,
			RStart:     13,
			LStart:     26,
			UStart:     39,
		},
	}
	armsOptions := []SelectedPartDef{
		{
			TilesetSrc: armsTileset,
			DStart:     0,
			RStart:     13,
			LStart:     26,
			UStart:     39,
		},
	}
	eyesOptions := []SelectedPartDef{}
	for i := range 14 {
		eyesOptions = append(eyesOptions, SelectedPartDef{
			TilesetSrc: eyesTileset,
			DStart:     i * 32,
			RStart:     (i * 32) + 1,
			FlipRForL:  true,
		})
	}
	hairOptions := []SelectedPartDef{}
	for i := range 7 {
		hairOptions = append(hairOptions, SelectedPartDef{
			TilesetSrc: hairTileset,
			DStart:     i * 32,
			RStart:     (i * 32) + 1,
			LStart:     (i * 32) + 2,
			UStart:     (i * 32) + 3,
		})
	}
	equipBodyOptions := []SelectedPartDef{}
	for i := range 4 {
		equipBodyOptions = append(equipBodyOptions, SelectedPartDef{
			TilesetSrc: equipBodyTileset,
			DStart:     (i * 52),
			RStart:     (i * 52) + 13,
			LStart:     (i * 52) + 26,
			UStart:     (i * 52) + 39,
		})
	}
	equipHeadOptions := []SelectedPartDef{{None: true}}
	for i := range 2 {
		index := i * 4
		cropHair, found := tiled.GetTileBoolProperty(equipHeadTileset, index, "COVER_HAIR")
		fmt.Println("found:", found, "crophair:", cropHair)
		equipHeadOptions = append(equipHeadOptions, SelectedPartDef{
			TilesetSrc:     equipHeadTileset,
			DStart:         (i * 4),
			RStart:         (i * 4) + 1,
			LStart:         (i * 4) + 2,
			UStart:         (i * 4) + 3,
			CropHairToHead: found && cropHair,
		})
	}

	weaponOptions := []SelectedPartDef{
		{
			TilesetSrc: equipWeaponTileset,
			DStart:     0,
			RStart:     13,
			LStart:     26,
			UStart:     39,
		},
	}

	g := builderGame{
		animation:          "run",
		animationTickCount: 15,
		currentDirection:   'D',
		bodySet: bodyPartSet{
			WalkAnimation: Animation{
				Name:         "body/walk",
				TileSteps:    []int{0, 2, 0, 4},
				StepsOffsetY: []int{0, 1, 0, 1},
			},
			RunAnimation: Animation{
				Name:         "body/run",
				TileSteps:    []int{0, 1, 2, 0, 3, 4},
				StepsOffsetY: []int{0, 0, 1, 0, 0, 1},
			},
			SlashAnimation: Animation{
				Name:         "body/slash",
				TileSteps:    []int{0, 5, 6, 7, 8},
				StepsOffsetY: []int{0, 1, 2, 2, 2},
			},
			HasUp: true,
		},
		bodyOptions: bodyOptions,
		armsSet: bodyPartSet{
			WalkAnimation: Animation{
				Name:      "arms/walk",
				TileSteps: []int{0, 2, 0, 4},
			},
			RunAnimation: Animation{
				Name:      "arms/run",
				TileSteps: []int{0, 1, 2, 0, 3, 4},
			},
			SlashAnimation: Animation{
				Name:      "arms/slash",
				TileSteps: []int{0, 5, 6, 7, 8},
			},
			HasUp: true,
		},
		armsOptions: armsOptions,
		eyesSet:     bodyPartSet{},
		eyesOptions: eyesOptions,
		hairSet: bodyPartSet{
			HasUp: true,
		},
		hairOptions: hairOptions,
		equipBodySet: bodyPartSet{
			WalkAnimation: Animation{
				Name:      "equipBody/walk",
				TileSteps: []int{0, 2, 0, 4},
			},
			RunAnimation: Animation{
				Name:      "equipBody/run",
				TileSteps: []int{0, 1, 2, 0, 3, 4},
			},
			SlashAnimation: Animation{
				Name:      "equipBody/slash",
				TileSteps: []int{0, 5, 6, 7, 8},
			},
			HasUp: true,
		},
		equipBodyOptions: equipBodyOptions,
		equipHeadSet: bodyPartSet{
			HasUp: true,
		},
		equipHeadOptions: equipHeadOptions,
		weaponSet: bodyPartSet{
			WalkAnimation: Animation{
				Name:      "weapon/walk",
				TileSteps: []int{0, 2, 0, 4},
			},
			RunAnimation: Animation{
				Name:      "weapon/run",
				TileSteps: []int{0, 1, 2, 0, 3, 4},
			},
			SlashAnimation: Animation{
				Name:      "weapon/slash",
				TileSteps: []int{0, 5, 6, 7, 8},
			},
			HasUp: true,
		},
		weaponOptions: weaponOptions,
		weaponFxSet: bodyPartSet{
			SelectedPartDef: SelectedPartDef{
				TilesetSrc: weaponFxTileset,
				DStart:     0,
				RStart:     6,
				LStart:     12,
				UStart:     18,
			},
			SlashAnimation: Animation{
				Name:      "weaponFx/slash",
				TileSteps: []int{-1, -1, 0, 1, 2}, // -1 = skip a frame (nil image)
			},
			WalkAnimation: Animation{Name: "weaponFx/walk", Skip: true},
			RunAnimation:  Animation{Name: "weaponFx/run", Skip: true},
			HasUp:         true,
		},
	}

	// SETS: load data
	g.SetBodyIndex(0)
	g.SetArmsIndex(0)
	g.SetHairIndex(0)
	g.SetEyesIndex(0)
	g.SetEquipBodyIndex(0)
	g.SetEquipHeadIndex(0)
	g.SetWeaponIndex(0)

	// for now, weaponFx doesn't have options
	g.weaponFxSet.Load()

	g.setDirection('D')

	turnLeftImg := tiled.GetTileImage("ui/ui-components.tsj", 224)
	turnLeftImg = rendering.ScaleImage(turnLeftImg, config.UIScale)
	turnRightImg := tiled.GetTileImage("ui/ui-components.tsj", 225)
	turnRightImg = rendering.ScaleImage(turnRightImg, config.UIScale)
	g.turnLeft = button.NewImageButton("", nil, turnLeftImg)
	g.turnRight = button.NewImageButton("", nil, turnRightImg)

	g.speedSlider = slider.NewSlider(slider.SliderParams{
		TilesetSrc:    "ui/ui-components.tsj",
		TilesetOrigin: 256,
		TileWidth:     4,
		MinVal:        5,
		MaxVal:        20,
		InitialValue:  8,
		StepSize:      1,
	})

	g.scaleSlider = slider.NewSlider(slider.SliderParams{
		TilesetSrc:    "ui/ui-components.tsj",
		TilesetOrigin: 256,
		TileWidth:     4,
		MinVal:        3,
		MaxVal:        10,
		InitialValue:  8,
		StepSize:      1,
	})

	g.animationSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               []string{"", anim_walk, anim_run, anim_slash, anim_stab},
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	})

	g.hairCtl = stepper.NewStepper(stepper.StepperParams{
		MinVal:               0,
		MaxVal:               len(hairOptions) - 1,
		Font:                 config.DefaultTitleFont,
		FontFg:               color.White,
		FontBg:               color.Black,
		DecrementButtonImage: turnLeftImg,
		IncrementButtonImage: turnRightImg,
	})
	g.eyesCtl = stepper.NewStepper(stepper.StepperParams{
		MinVal:               0,
		MaxVal:               len(eyesOptions) - 1,
		Font:                 config.DefaultTitleFont,
		FontFg:               color.White,
		FontBg:               color.Black,
		DecrementButtonImage: turnLeftImg,
		IncrementButtonImage: turnRightImg,
	})
	g.equipBodyCtl = stepper.NewStepper(stepper.StepperParams{
		MinVal:               0,
		MaxVal:               len(equipBodyOptions) - 1,
		Font:                 config.DefaultTitleFont,
		FontFg:               color.White,
		FontBg:               color.Black,
		DecrementButtonImage: turnLeftImg,
		IncrementButtonImage: turnRightImg,
	})
	g.equipHeadCtl = stepper.NewStepper(stepper.StepperParams{
		MinVal:               0,
		MaxVal:               len(equipHeadOptions) - 1,
		Font:                 config.DefaultTitleFont,
		FontFg:               color.White,
		FontBg:               color.Black,
		DecrementButtonImage: turnLeftImg,
		IncrementButtonImage: turnRightImg,
	})

	if err := ebiten.RunGame(&g); err != nil {
		panic(err)
	}
}

// SETS: Set Index Functions

func (bg *builderGame) SetBodyIndex(i int) {
	if i < 0 || i > len(bg.bodyOptions) {
		panic("out of bounds")
	}
	bg.bodyOptionIndex = i
	bg.bodySet.SelectedPartDef = bg.bodyOptions[i]
	bg.bodySet.Load()
}
func (bg *builderGame) SetArmsIndex(i int) {
	if i < 0 || i > len(bg.armsOptions) {
		panic("out of bounds")
	}
	bg.armsOptionIndex = i
	bg.armsSet.SelectedPartDef = bg.armsOptions[i]
	bg.armsSet.Load()
}
func (bg *builderGame) SetEyesIndex(i int) {
	if i < 0 || i > len(bg.eyesOptions) {
		panic("out of bounds")
	}
	bg.eyesOptionIndex = i
	bg.eyesSet.SelectedPartDef = bg.eyesOptions[i]
	bg.eyesSet.Load()
}
func (bg *builderGame) SetHairIndex(i int) {
	if i < 0 || i > len(bg.hairOptions) {
		panic("out of bounds")
	}
	bg.hairOptionIndex = i
	bg.hairSet.SelectedPartDef = bg.hairOptions[i]
	bg.hairSet.Load()
	if bg.equipHeadSet.CropHairToHead {
		bg.cropHairToHead()
	}
}
func (bg *builderGame) SetEquipBodyIndex(i int) {
	if i < 0 || i > len(bg.equipBodyOptions) {
		panic("out of bounds")
	}
	bg.equipBodyOptionIndex = i
	bg.equipBodySet.SelectedPartDef = bg.equipBodyOptions[i]
	bg.equipBodySet.Load()
}
func (bg *builderGame) SetEquipHeadIndex(i int) {
	if i < 0 || i > len(bg.equipHeadOptions) {
		panic("out of bounds")
	}
	bg.equipHeadOptionIndex = i
	bg.equipHeadSet.SelectedPartDef = bg.equipHeadOptions[i]
	bg.equipHeadSet.Load()

	// since some head equipment may cause hair to be cropped, always reload hair when head is equiped
	bg.hairSet.Load()

	if bg.equipHeadSet.SelectedPartDef.CropHairToHead {
		bg.cropHairToHead()
	}
}
func (bg *builderGame) SetWeaponIndex(i int) {
	if i < 0 || i > len(bg.weaponOptions) {
		panic("out of bounds")
	}
	bg.weaponOptionIndex = i
	bg.weaponSet.SelectedPartDef = bg.weaponOptions[i]
	bg.weaponSet.Load()
}

type Animation struct {
	Name         string
	Skip         bool // if true, this animation does not get defined
	L            []*ebiten.Image
	R            []*ebiten.Image
	U            []*ebiten.Image
	D            []*ebiten.Image
	TileSteps    []int
	StepsOffsetY []int
}

// represents either the head, body, eyes, or hair
type bodyPartSet struct {
	animIndex int
	SelectedPartDef
	WalkAnimation  Animation
	RunAnimation   Animation
	SlashAnimation Animation
	HasUp          bool
}

// represents the currently selected body part and it's individual definition
type SelectedPartDef struct {
	None                           bool // if true, this part will not be shown
	TilesetSrc                     string
	RStart, LStart, UStart, DStart int
	FlipRForL                      bool // if true, instead of using an L source, we just flip the frames for right

	// headwear-specific props

	CropHairToHead bool // set to have hair not go outside the head image. used for helmets or certain hats.
}

func (a Animation) getFrame(dir byte, animationIndex int) *ebiten.Image {
	switch dir {
	case 'L':
		if len(a.L) == 0 {
			logz.Println(a.Name, "no left frames; returning nil")
			return nil
		}
		if animationIndex >= len(a.L) {
			logz.Println(a.Name, "past left frames; returning last frame", "animIndex:", animationIndex)
			return a.L[len(a.L)-1]
		}
		return a.L[animationIndex]
	case 'R':
		if len(a.R) == 0 {
			logz.Println(a.Name, "no right frames; returning nil")
			return nil
		}
		if animationIndex >= len(a.R) {
			logz.Println(a.Name, "past right frames; returning last frame", "animIndex:", animationIndex)
			return a.R[len(a.R)-1]
		}
		return a.R[animationIndex]
	case 'U':
		if len(a.U) == 0 {
			logz.Println(a.Name, "no up frames; returning nil")
			return nil
		}
		if animationIndex >= len(a.U) {
			logz.Println(a.Name, "past up frames; returning last frame", "animIndex:", animationIndex)
			return a.U[len(a.U)-1]
		}
		return a.U[animationIndex]
	case 'D':
		if len(a.D) == 0 {
			logz.Println(a.Name, "no down frames; returning nil")
			return nil
		}
		if animationIndex >= len(a.D) {
			logz.Println(a.Name, "past down frames; returning last frame", "animIndex:", animationIndex)
			return a.D[len(a.D)-1]
		}
		return a.D[animationIndex]
	}
	panic("unrecognized direction")
}

func (set bodyPartSet) getCurrentFrame(dir byte, animationName string) *ebiten.Image {
	if set.None {
		return nil
	}

	switch animationName {
	case anim_walk:
		return set.WalkAnimation.getFrame(dir, set.animIndex)
	case anim_run:
		return set.RunAnimation.getFrame(dir, set.animIndex)
	case anim_slash:
		return set.SlashAnimation.getFrame(dir, set.animIndex)
	case "":
		return set.WalkAnimation.getFrame(dir, 0)
	default:
		panic("unrecognized animation name: " + animationName)
	}
}

func (set bodyPartSet) getCurrentYOffset(animationName string) int {
	if set.animIndex == 0 {
		return 0
	}
	switch animationName {
	case anim_walk:
		if len(set.WalkAnimation.StepsOffsetY) > 0 {
			return set.WalkAnimation.StepsOffsetY[set.animIndex]
		}
	case anim_run:
		if len(set.RunAnimation.StepsOffsetY) > 0 {
			return set.RunAnimation.StepsOffsetY[set.animIndex]
		}
	case anim_slash:
		if len(set.SlashAnimation.StepsOffsetY) > 0 {
			return set.SlashAnimation.StepsOffsetY[set.animIndex]
		}
	}

	return 0
}

func (set *bodyPartSet) nextFrame(animationName string) {
	if set.None {
		return
	}

	set.animIndex++
	switch animationName {
	case anim_walk:
		if set.animIndex >= len(set.WalkAnimation.TileSteps) {
			set.animIndex = 0
		}
	case anim_run:
		if set.animIndex >= len(set.RunAnimation.TileSteps) {
			set.animIndex = 0
		}
	case anim_slash:
		if set.animIndex >= len(set.SlashAnimation.TileSteps) {
			set.animIndex = 0
		}
	}
}

func (a *Animation) reset() {
	a.L = make([]*ebiten.Image, 0)
	a.R = make([]*ebiten.Image, 0)
	a.U = make([]*ebiten.Image, 0)
	a.D = make([]*ebiten.Image, 0)
}

func (bg *builderGame) cropHairToHead() {
	leftHead := ebiten.NewImage(config.TileSize, config.TileSize)
	leftHead.DrawImage(bg.bodySet.WalkAnimation.L[0], nil)
	rightHead := ebiten.NewImage(config.TileSize, config.TileSize)
	rightHead.DrawImage(bg.bodySet.WalkAnimation.R[0], nil)
	upHead := ebiten.NewImage(config.TileSize, config.TileSize)
	upHead.DrawImage(bg.bodySet.WalkAnimation.U[0], nil)
	downHead := ebiten.NewImage(config.TileSize, config.TileSize)
	downHead.DrawImage(bg.bodySet.WalkAnimation.D[0], nil)

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

	cropper(&bg.hairSet.WalkAnimation)
	cropper(&bg.hairSet.RunAnimation)
	cropper(&bg.hairSet.SlashAnimation)
}

func (set *bodyPartSet) Load() {
	set.WalkAnimation.reset()
	set.RunAnimation.reset()
	set.SlashAnimation.reset()

	if set.None {
		return
	}

	// walk animation
	if !set.WalkAnimation.Skip {
		if set.FlipRForL {
			set.WalkAnimation.L = getAnimationFrames(set.TilesetSrc, set.RStart, set.WalkAnimation.TileSteps, true)
		} else {
			set.WalkAnimation.L = getAnimationFrames(set.TilesetSrc, set.LStart, set.WalkAnimation.TileSteps, false)
		}
		set.WalkAnimation.R = getAnimationFrames(set.TilesetSrc, set.RStart, set.WalkAnimation.TileSteps, false)
		if set.HasUp {
			set.WalkAnimation.U = getAnimationFrames(set.TilesetSrc, set.UStart, set.WalkAnimation.TileSteps, false)
		}
		set.WalkAnimation.D = getAnimationFrames(set.TilesetSrc, set.DStart, set.WalkAnimation.TileSteps, false)
	}

	// run animation
	if !set.RunAnimation.Skip {
		if set.FlipRForL {
			set.RunAnimation.L = getAnimationFrames(set.TilesetSrc, set.RStart, set.RunAnimation.TileSteps, true)
		} else {
			set.RunAnimation.L = getAnimationFrames(set.TilesetSrc, set.LStart, set.RunAnimation.TileSteps, false)
		}
		set.RunAnimation.R = getAnimationFrames(set.TilesetSrc, set.RStart, set.RunAnimation.TileSteps, false)
		if set.HasUp {
			set.RunAnimation.U = getAnimationFrames(set.TilesetSrc, set.UStart, set.RunAnimation.TileSteps, false)
		}
		set.RunAnimation.D = getAnimationFrames(set.TilesetSrc, set.DStart, set.RunAnimation.TileSteps, false)
	}

	// slash animation
	if !set.SlashAnimation.Skip {
		if set.FlipRForL {
			set.SlashAnimation.L = getAnimationFrames(set.TilesetSrc, set.RStart, set.SlashAnimation.TileSteps, true)
		} else {
			set.SlashAnimation.L = getAnimationFrames(set.TilesetSrc, set.LStart, set.SlashAnimation.TileSteps, false)
		}
		set.SlashAnimation.R = getAnimationFrames(set.TilesetSrc, set.RStart, set.SlashAnimation.TileSteps, false)
		if set.HasUp {
			set.SlashAnimation.U = getAnimationFrames(set.TilesetSrc, set.UStart, set.SlashAnimation.TileSteps, false)
		}
		set.SlashAnimation.D = getAnimationFrames(set.TilesetSrc, set.DStart, set.SlashAnimation.TileSteps, false)
	}
}

func getAnimationFrames(tilesetSrc string, startIndex int, indexSteps []int, flip bool) []*ebiten.Image {
	frames := []*ebiten.Image{}

	if len(indexSteps) == 0 {
		// no animation defined; just use the start tile
		img := tiled.GetTileImage(tilesetSrc, startIndex)
		if flip {
			img = rendering.FlipHoriz(img)
		}
		frames = append(frames, img)
	}
	for _, step := range indexSteps {
		if step == -1 {
			// indicates a skip frame
			frames = append(frames, nil)
			continue
		}
		img := tiled.GetTileImage(tilesetSrc, startIndex+step)
		if flip {
			img = rendering.FlipHoriz(img)
		}
		frames = append(frames, img)
	}
	return frames
}

func (bg *builderGame) Draw(screen *ebiten.Image) {
	var characterScale float64 = float64(bg.scaleSlider.GetValue())
	characterTileSize := config.TileSize * characterScale

	tileSize := int(config.TileSize * config.UIScale)

	if bg.bodyImg == nil {
		panic("no body image!")
	}

	bodyWidth := float64(bg.bodyImg.Bounds().Dx()) * characterScale
	bodyHeight := float64(bg.bodyImg.Bounds().Dy()) * characterScale

	bodyX := float64(display.SCREEN_WIDTH/2) - (bodyWidth / 2)
	bodyY := float64(display.SCREEN_HEIGHT/2) - (bodyHeight / 2)

	// Body
	rendering.DrawImage(screen, bg.bodyImg, bodyX, bodyY, characterScale)
	// Arms
	rendering.DrawImage(screen, bg.armsImg, bodyX, bodyY, characterScale)
	// Equip Body
	rendering.DrawImage(screen, bg.equipBodyImg, bodyX, bodyY, characterScale)

	// Eyes
	eyesX := bodyX
	eyesY := bodyY + (float64(bg.nonBodyYOffset) * characterScale)
	if bg.eyesImg != nil {
		rendering.DrawImage(screen, bg.eyesImg, eyesX, eyesY, characterScale)
	}
	// Hair
	hairY := bodyY + (float64(bg.nonBodyYOffset) * characterScale)
	if bg.hairImg == nil {
		panic("hair img is nil")
	}
	rendering.DrawImage(screen, bg.hairImg, bodyX, hairY, characterScale)

	// Equip Head
	if bg.equipHeadImg != nil {
		rendering.DrawImage(screen, bg.equipHeadImg, bodyX, hairY, characterScale)
	}

	// Equip Weapon
	if bg.weaponImg != nil {
		// weapons are in 80x80 (5 tiles width & height) tiles
		// this is to accomodate for the extra space they need for their swings and stuff
		weaponY := bodyY - (characterTileSize)
		weaponX := bodyX - (characterTileSize * 2)
		rendering.DrawImage(screen, bg.weaponImg, weaponX, weaponY, characterScale)
		if bg.weaponFxImg != nil {
			rendering.DrawImage(screen, bg.weaponFxImg, weaponX, weaponY, characterScale)
		}
	}

	buttonsY := bodyY + (bodyHeight) + 20
	buttonLX := (display.SCREEN_WIDTH / 2) - bg.turnLeft.Width - 20
	buttonRX := (display.SCREEN_WIDTH / 2) + 20
	bg.turnLeft.Draw(screen, buttonLX, int(buttonsY))
	bg.turnRight.Draw(screen, buttonRX, int(buttonsY))

	sliderX := 100
	sliderY := 100
	text.DrawShadowText(screen, "Ticks Per Frame", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	text.DrawShadowText(screen, fmt.Sprintf("%v", bg.speedSlider.GetValue()), config.DefaultFont, sliderX-40, sliderY+(tileSize*2/3), color.White, nil, 0, 0)
	bg.speedSlider.Draw(screen, float64(sliderX), float64(sliderY))

	scaleSliderY := 200
	text.DrawShadowText(screen, "Scale", config.DefaultTitleFont, sliderX, scaleSliderY, color.White, nil, 0, 0)
	text.DrawShadowText(screen, fmt.Sprintf("%v", bg.scaleSlider.GetValue()), config.DefaultFont, sliderX-40, scaleSliderY+(tileSize*2/3), color.White, nil, 0, 0)
	bg.scaleSlider.Draw(screen, float64(sliderX), float64(scaleSliderY))

	animationSelectorY := 300
	text.DrawShadowText(screen, "Animation", config.DefaultTitleFont, sliderX, animationSelectorY, color.White, nil, 0, 0)
	bg.animationSelector.Draw(screen, float64(sliderX), float64(animationSelectorY), nil)

	ctlX := (display.SCREEN_WIDTH * 3 / 4)
	ctlY := 100
	text.DrawShadowText(screen, "Hair", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.hairCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlY += 100
	text.DrawShadowText(screen, "Eyes", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.eyesCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlY += 100
	text.DrawShadowText(screen, "Equip Head", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.equipHeadCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlY += 100
	text.DrawShadowText(screen, "Equip Body", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.equipBodyCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
}

func (bg *builderGame) Update() error {
	if bg.turnLeft.Update().Clicked {
		bg.rotateLeft()
	} else if bg.turnRight.Update().Clicked {
		bg.rotateRight()
	}

	bg.hairCtl.Update()
	if bg.hairCtl.GetValue() != bg.hairOptionIndex {
		bg.SetHairIndex(bg.hairCtl.GetValue())
	}
	bg.eyesCtl.Update()
	if bg.eyesCtl.GetValue() != bg.eyesOptionIndex {
		bg.SetEyesIndex(bg.eyesCtl.GetValue())
	}
	bg.equipHeadCtl.Update()
	if bg.equipHeadCtl.GetValue() != bg.equipHeadOptionIndex {
		bg.SetEquipHeadIndex(bg.equipHeadCtl.GetValue())
	}
	bg.equipBodyCtl.Update()
	if bg.equipBodyCtl.GetValue() != bg.equipBodyOptionIndex {
		bg.SetEquipBodyIndex(bg.equipBodyCtl.GetValue())
	}

	bg.speedSlider.Update()
	bg.scaleSlider.Update()

	bg.animationSelector.Update()
	selectorValue := bg.animationSelector.GetCurrentValue()
	if selectorValue != bg.animation {
		bg.setAnimation(selectorValue)
	}

	bg.animationTickCount = bg.speedSlider.GetValue()

	if bg.animation != "" {
		bg.ticks++
		if bg.ticks > bg.animationTickCount {
			// SETS: next frame
			bg.ticks = 0
			bg.bodySet.nextFrame(bg.animation)
			bg.armsSet.nextFrame(bg.animation)
			bg.equipBodySet.nextFrame(bg.animation)
			bg.equipHeadSet.nextFrame(bg.animation)
			bg.weaponSet.nextFrame(bg.animation)
			bg.weaponFxSet.nextFrame(bg.animation)
		}
	}

	// SETS: get current frame
	bg.bodyImg = bg.bodySet.getCurrentFrame(bg.currentDirection, bg.animation)
	bg.eyesImg = bg.eyesSet.getCurrentFrame(bg.currentDirection, bg.animation)
	bg.hairImg = bg.hairSet.getCurrentFrame(bg.currentDirection, bg.animation)
	bg.armsImg = bg.armsSet.getCurrentFrame(bg.currentDirection, bg.animation)

	bg.equipBodyImg = bg.equipBodySet.getCurrentFrame(bg.currentDirection, bg.animation)
	bg.equipHeadImg = bg.equipHeadSet.getCurrentFrame(bg.currentDirection, bg.animation)
	bg.weaponImg = bg.weaponSet.getCurrentFrame(bg.currentDirection, bg.animation)
	bg.weaponFxImg = bg.weaponFxSet.getCurrentFrame(bg.currentDirection, bg.animation)

	bg.nonBodyYOffset = bg.bodySet.getCurrentYOffset(bg.animation)

	return nil
}

func (bg *builderGame) setAnimation(animation string) {
	bg.animation = animation

	// SETS: reset animation index
	bg.bodySet.animIndex = 0
	bg.eyesSet.animIndex = 0
	bg.hairSet.animIndex = 0
	bg.armsSet.animIndex = 0
	bg.equipBodySet.animIndex = 0
	bg.equipHeadSet.animIndex = 0
	bg.weaponSet.animIndex = 0
	bg.weaponFxSet.animIndex = 0
}

func (bg *builderGame) rotateLeft() {
	switch bg.currentDirection {
	case 'L':
		bg.setDirection('U')
	case 'U':
		bg.setDirection('R')
	case 'R':
		bg.setDirection('D')
	case 'D':
		bg.setDirection('L')
	}
}

func (bg *builderGame) rotateRight() {
	switch bg.currentDirection {
	case 'L':
		bg.setDirection('D')
	case 'D':
		bg.setDirection('R')
	case 'R':
		bg.setDirection('U')
	case 'U':
		bg.setDirection('L')
	}
}

func (bg *builderGame) setDirection(dir byte) {
	// SETS: reset animation index
	bg.bodySet.animIndex = 0
	bg.eyesSet.animIndex = 0
	bg.hairSet.animIndex = 0
	bg.armsSet.animIndex = 0

	bg.equipBodySet.animIndex = 0
	bg.equipHeadSet.animIndex = 0
	bg.weaponSet.animIndex = 0
	bg.weaponFxSet.animIndex = 0

	bg.currentDirection = dir

	// SETS: set to first frame of walking animation
	bg.bodyImg = bg.bodySet.WalkAnimation.getFrame(dir, 0)
	bg.eyesImg = bg.eyesSet.WalkAnimation.getFrame(dir, 0)
	bg.hairImg = bg.hairSet.WalkAnimation.getFrame(dir, 0)
	bg.armsImg = bg.armsSet.WalkAnimation.getFrame(dir, 0)
	bg.equipBodyImg = bg.equipBodySet.WalkAnimation.getFrame(dir, 0)
	bg.equipHeadImg = bg.equipHeadSet.WalkAnimation.getFrame(dir, 0)
	bg.weaponImg = bg.weaponSet.WalkAnimation.getFrame(dir, 0)
	bg.weaponFxImg = bg.weaponFxSet.WalkAnimation.getFrame(dir, 0)
}

func (bg *builderGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return display.SCREEN_WIDTH, display.SCREEN_HEIGHT
}
