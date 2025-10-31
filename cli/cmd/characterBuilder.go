/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
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
	bgImg *ebiten.Image

	// body part options (characterBuilder only)

	bodySetOptions, eyesSetOptions, hairSetOptions, armsSetOptions []body.SelectedPartDef
	bodySetIndex, eyesSetIndex, hairSetIndex, armsSetIndex         int
	equipBodySetOptions, equipHeadSetOptions, weaponSetOptions     []body.SelectedPartDef
	equipBodySetIndex, equipHeadSetIndex, weaponSetIndex           int

	// entity body animation parts (in entity)

	entityBody body.EntityBodySet

	// UI components

	turnLeft          *button.Button
	turnRight         *button.Button
	speedSlider       slider.Slider
	scaleSlider       slider.Slider
	animationSelector dropdown.OptionSelect

	bodyCtl      stepper.Stepper
	hairCtl      stepper.Stepper
	eyesCtl      stepper.Stepper
	equipBodyCtl stepper.Stepper
	equipHeadCtl stepper.Stepper

	bodyColorSliders slider.SliderGroup
	hairColorSliders slider.SliderGroup
	eyeColorSliders  slider.SliderGroup
}

func characterBuilder() {
	bodyTileset := "entities/parts/body.tsj"
	armsTileset := "entities/parts/arms.tsj"
	eyesTileset := "entities/parts/eyes.tsj"
	hairTileset := "entities/parts/hair.tsj"
	equipBodyTileset := "items/equiped_body_01.tsj"
	equipHeadTileset := "items/equiped_head_01.tsj"
	equipWeaponTileset := "items/weapon_frames.tsj"
	weaponFxTileset := "items/weapon_fx_frames.tsj"

	bgTileset := "buildings/walls.tsj"

	bodyOptions := []body.SelectedPartDef{
		{
			TilesetSrc: bodyTileset,
			DStart:     0,
			RStart:     13,
			LStart:     26,
			UStart:     39,
		},
		{
			TilesetSrc: bodyTileset,
			DStart:     0 + (52),
			RStart:     13 + (52),
			LStart:     26 + (52),
			UStart:     39 + (52),
			StretchX:   2,
		},
		{
			TilesetSrc: bodyTileset,
			DStart:     0 + (52 * 2),
			RStart:     13 + (52 * 2),
			LStart:     26 + (52 * 2),
			UStart:     39 + (52 * 2),
			StretchY:   -1,
			OffsetY:    2,
		},
		{
			TilesetSrc: bodyTileset,
			DStart:     0 + (52 * 3),
			RStart:     13 + (52 * 3),
			LStart:     26 + (52 * 3),
			UStart:     39 + (52 * 3),
			StretchX:   2,
			StretchY:   -1,
			OffsetY:    2,
		},
	}
	armsOptions := []body.SelectedPartDef{
		{
			TilesetSrc: armsTileset,
			DStart:     0,
			RStart:     13,
			LStart:     26,
			UStart:     39,
		},
		{
			TilesetSrc: armsTileset,
			DStart:     0 + (52),
			RStart:     13 + (52),
			LStart:     26 + (52),
			UStart:     39 + (52),
		},
		{
			TilesetSrc: armsTileset,
			DStart:     0 + (52 * 2),
			RStart:     13 + (52 * 2),
			LStart:     26 + (52 * 2),
			UStart:     39 + (52 * 2),
		},
		{
			TilesetSrc: armsTileset,
			DStart:     0 + (52 * 3),
			RStart:     13 + (52 * 3),
			LStart:     26 + (52 * 3),
			UStart:     39 + (52 * 3),
		},
	}
	eyesOptions := []body.SelectedPartDef{}
	for i := range 14 {
		eyesOptions = append(eyesOptions, body.SelectedPartDef{
			TilesetSrc: eyesTileset,
			DStart:     i * 32,
			RStart:     (i * 32) + 1,
			FlipRForL:  true,
		})
	}
	hairOptions := []body.SelectedPartDef{}
	for i := range 7 {
		hairOptions = append(hairOptions, body.SelectedPartDef{
			TilesetSrc: hairTileset,
			DStart:     i * 32,
			RStart:     (i * 32) + 1,
			LStart:     (i * 32) + 2,
			UStart:     (i * 32) + 3,
		})
	}
	equipBodyOptions := []body.SelectedPartDef{}
	for i := range 4 {
		equipBodyOptions = append(equipBodyOptions, body.SelectedPartDef{
			TilesetSrc: equipBodyTileset,
			DStart:     (i * 52),
			RStart:     (i * 52) + 13,
			LStart:     (i * 52) + 26,
			UStart:     (i * 52) + 39,
		})
	}
	equipHeadOptions := []body.SelectedPartDef{{None: true}}
	for i := range 2 {
		index := i * 4
		cropHair, found := tiled.GetTileBoolProperty(equipHeadTileset, index, "COVER_HAIR")
		equipHeadOptions = append(equipHeadOptions, body.SelectedPartDef{
			TilesetSrc:     equipHeadTileset,
			DStart:         (i * 4),
			RStart:         (i * 4) + 1,
			LStart:         (i * 4) + 2,
			UStart:         (i * 4) + 3,
			CropHairToHead: found && cropHair,
		})
	}

	weaponOptions := []body.SelectedPartDef{
		{
			TilesetSrc: equipWeaponTileset,
			DStart:     0,
			RStart:     13,
			LStart:     26,
			UStart:     39,
		},
	}

	bodySet := body.BodyPartSet{
		WalkAnimation: body.Animation{
			Name:         "body/walk",
			TileSteps:    []int{0, 2, 0, 4},
			StepsOffsetY: []int{0, 1, 0, 1},
		},
		RunAnimation: body.Animation{
			Name:         "body/run",
			TileSteps:    []int{0, 1, 2, 0, 3, 4},
			StepsOffsetY: []int{0, 0, 1, 0, 0, 1},
		},
		SlashAnimation: body.Animation{
			Name:         "body/slash",
			TileSteps:    []int{0, 5, 6, 7, 8},
			StepsOffsetY: []int{0, 1, 2, 2, 2},
		},
		HasUp: true,
	}
	armsSet := body.BodyPartSet{
		WalkAnimation: body.Animation{
			Name:      "arms/walk",
			TileSteps: []int{0, 2, 0, 4},
		},
		RunAnimation: body.Animation{
			Name:      "arms/run",
			TileSteps: []int{0, 1, 2, 0, 3, 4},
		},
		SlashAnimation: body.Animation{
			Name:      "arms/slash",
			TileSteps: []int{0, 5, 6, 7, 8},
		},
		HasUp: true,
	}
	eyesSet := body.BodyPartSet{}
	hairSet := body.BodyPartSet{HasUp: true}
	equipBodySet := body.BodyPartSet{
		WalkAnimation: body.Animation{
			Name:      "equipBody/walk",
			TileSteps: []int{0, 2, 0, 4},
		},
		RunAnimation: body.Animation{
			Name:      "equipBody/run",
			TileSteps: []int{0, 1, 2, 0, 3, 4},
		},
		SlashAnimation: body.Animation{
			Name:      "equipBody/slash",
			TileSteps: []int{0, 5, 6, 7, 8},
		},
		HasUp: true,
	}
	equipHeadSet := body.BodyPartSet{HasUp: true}
	weaponSet := body.BodyPartSet{
		WalkAnimation: body.Animation{
			Name:      "weapon/walk",
			TileSteps: []int{0, 2, 0, 4},
		},
		RunAnimation: body.Animation{
			Name:      "weapon/run",
			TileSteps: []int{0, 1, 2, 0, 3, 4},
		},
		SlashAnimation: body.Animation{
			Name:      "weapon/slash",
			TileSteps: []int{0, 5, 6, 7, 8},
		},
		HasUp: true,
	}
	weaponFxSet := body.BodyPartSet{
		TilesetSrc: weaponFxTileset,
		DStart:     0,
		RStart:     6,
		LStart:     12,
		UStart:     18,
		SlashAnimation: body.Animation{
			Name:      "weaponFx/slash",
			TileSteps: []int{-1, -1, 0, 1, 2}, // -1 = skip a frame (nil image)
		},
		WalkAnimation: body.Animation{Name: "weaponFx/walk", Skip: true},
		RunAnimation:  body.Animation{Name: "weaponFx/run", Skip: true},
		HasUp:         true,
	}

	entBody := body.NewEntityBodySet(
		bodySet,
		armsSet,
		eyesSet,
		hairSet,
		&equipBodySet,
		&equipHeadSet,
		&weaponSet,
		&weaponFxSet,
	)

	g := builderGame{
		bodySetOptions:      bodyOptions,
		armsSetOptions:      armsOptions,
		eyesSetOptions:      eyesOptions,
		hairSetOptions:      hairOptions,
		equipBodySetOptions: equipBodyOptions,
		equipHeadSetOptions: equipHeadOptions,
		weaponSetOptions:    weaponOptions,
		entityBody:          entBody,
	}

	// SETS: load data
	g.SetHairIndex(0)
	g.SetEyesIndex(0)
	g.SetEquipBodyIndex(0)
	g.SetEquipHeadIndex(0)
	g.SetBodyIndex(0)
	g.SetWeaponIndex(0)

	// create the backdrop
	t := float64(config.TileSize)
	g.bgImg = ebiten.NewImage(int(t*3), int(t*3))
	rendering.DrawImage(g.bgImg, tiled.GetTileImage(bgTileset, 150), 0, 0, 0)
	rendering.DrawImage(g.bgImg, tiled.GetTileImage(bgTileset, 151), t, 0, 0)
	rendering.DrawImage(g.bgImg, tiled.GetTileImage(bgTileset, 152), t*2, 0, 0)

	rendering.DrawImage(g.bgImg, tiled.GetTileImage(bgTileset, 182), 0, t, 0)
	rendering.DrawImage(g.bgImg, tiled.GetTileImage(bgTileset, 183), t, t, 0)
	rendering.DrawImage(g.bgImg, tiled.GetTileImage(bgTileset, 184), t*2, t, 0)

	rendering.DrawImage(g.bgImg, tiled.GetTileImage(bgTileset, 214), 0, t*2, 0)
	rendering.DrawImage(g.bgImg, tiled.GetTileImage(bgTileset, 215), t, t*2, 0)
	rendering.DrawImage(g.bgImg, tiled.GetTileImage(bgTileset, 216), t*2, t*2, 0)

	g.entityBody.SetDirection('D')

	turnLeftImg := tiled.GetTileImage("ui/ui-components.tsj", 224)
	turnLeftImg = rendering.ScaleImage(turnLeftImg, config.UIScale, config.UIScale)
	turnRightImg := tiled.GetTileImage("ui/ui-components.tsj", 225)
	turnRightImg = rendering.ScaleImage(turnRightImg, config.UIScale, config.UIScale)
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
		Options:               []string{"", body.ANIM_WALK, body.ANIM_RUN, body.ANIM_SLASH},
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	})

	g.bodyCtl = stepper.NewStepper(stepper.StepperParams{
		MinVal:               0,
		MaxVal:               len(bodyOptions) - 1,
		Font:                 config.DefaultTitleFont,
		FontFg:               color.White,
		FontBg:               color.Black,
		DecrementButtonImage: turnLeftImg,
		IncrementButtonImage: turnRightImg,
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

	// sliders for adjusting colors
	sliderParams := slider.SliderParams{
		TilesetSrc:    "ui/ui-components.tsj",
		TilesetOrigin: 256,
		TileWidth:     4,
		MinVal:        0,
		MaxVal:        256,
		InitialValue:  128,
		StepSize:      1,
	}
	g.bodyColorSliders = slider.NewSliderGroup(slider.SliderGroupParams{
		LabelFont:    config.DefaultFont,
		LabelColorFg: color.White,
		LabelColorBg: color.Black,
	}, []slider.SliderDef{
		{
			Label:  "H",
			Params: sliderParams,
		},
		{
			Label:  "S",
			Params: sliderParams,
		},
		{
			Label:  "V",
			Params: sliderParams,
		},
	})
	g.hairColorSliders = slider.NewSliderGroup(slider.SliderGroupParams{
		LabelFont:    config.DefaultFont,
		LabelColorFg: color.White,
		LabelColorBg: color.Black,
	}, []slider.SliderDef{
		{
			Label:  "H",
			Params: sliderParams,
		},
		{
			Label:  "S",
			Params: sliderParams,
		},
		{
			Label:  "V",
			Params: sliderParams,
		},
	})
	g.eyeColorSliders = slider.NewSliderGroup(slider.SliderGroupParams{
		LabelFont:    config.DefaultFont,
		LabelColorFg: color.White,
		LabelColorBg: color.Black,
	}, []slider.SliderDef{
		{
			Label:  "H",
			Params: sliderParams,
		},
		{
			Label:  "S",
			Params: sliderParams,
		},
		{
			Label:  "V",
			Params: sliderParams,
		},
	})

	if err := ebiten.RunGame(&g); err != nil {
		panic(err)
	}
}

// SETS: Set Index Functions

// DEPENDS ON:
//
// hairSet, equipHeadSet, equipBodySet
func (bg *builderGame) SetBodyIndex(i int) {
	if i < 0 || i >= len(bg.bodySetOptions) {
		panic("out of bounds")
	}
	bg.bodySetIndex = i
	bodyDef := bg.bodySetOptions[i]

	// arms are directly set with body
	bg.armsSetIndex = i
	armDef := bg.armsSetOptions[i]

	bg.entityBody.SetBody(bodyDef, armDef)
}
func (bg *builderGame) SetEyesIndex(i int) {
	if i < 0 || i >= len(bg.eyesSetOptions) {
		panic("out of bounds")
	}
	bg.eyesSetIndex = i
	op := bg.eyesSetOptions[i]
	bg.entityBody.SetEyes(op)
}
func (bg *builderGame) SetHairIndex(i int) {
	if i < 0 || i >= len(bg.hairSetOptions) {
		panic("out of bounds")
	}
	bg.hairSetIndex = i
	op := bg.hairSetOptions[i]
	bg.entityBody.SetHair(op)
}
func (bg *builderGame) SetEquipBodyIndex(i int) {
	if i < 0 || i >= len(bg.equipBodySetOptions) {
		panic("out of bounds")
	}
	bg.equipBodySetIndex = i
	op := bg.equipBodySetOptions[i]
	bg.entityBody.SetEquipBody(op)
}

// DEPENDS ON:
//
// hairSet
func (bg *builderGame) SetEquipHeadIndex(i int) {
	if i < 0 || i >= len(bg.equipHeadSetOptions) {
		panic("out of bounds")
	}
	bg.equipHeadSetIndex = i
	op := bg.equipHeadSetOptions[i]
	bg.entityBody.SetEquipHead(op)
}
func (bg *builderGame) SetWeaponIndex(i int) {
	if i < 0 || i > len(bg.weaponSetOptions) {
		panic("out of bounds")
	}
	bg.weaponSetIndex = i
	op := bg.weaponSetOptions[i]
	bg.entityBody.SetWeapon(op)
}

func (bg *builderGame) Draw(screen *ebiten.Image) {
	var characterScale float64 = float64(bg.scaleSlider.GetValue())
	characterTileSize := config.TileSize * characterScale

	tileSize := int(config.TileSize * config.UIScale)

	bodyDx, bodyDy := bg.entityBody.Dimensions()
	bodyWidth := float64(bodyDx) * characterScale
	bodyHeight := float64(bodyDy) * characterScale

	bodyX := float64(display.SCREEN_WIDTH/2) - (bodyWidth / 2)
	bodyY := float64(display.SCREEN_HEIGHT/2) - (bodyHeight / 2)

	// Backdrop
	rendering.DrawImage(screen, bg.bgImg, bodyX-characterTileSize, bodyY-characterTileSize, characterScale)

	// Character body
	bg.entityBody.Draw(screen, bodyX, bodyY, characterScale)

	// UI controls - Left side
	buttonsY := bodyY + (bodyHeight) + 20
	buttonLX := (display.SCREEN_WIDTH / 2) - bg.turnLeft.Width - 20
	buttonRX := (display.SCREEN_WIDTH / 2) + 20
	bg.turnLeft.Draw(screen, buttonLX, int(buttonsY))
	bg.turnRight.Draw(screen, buttonRX, int(buttonsY))

	sliderX := 100
	sliderY := 50
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

	// UI controls - Right side
	ctlX := (display.SCREEN_WIDTH * 3 / 4)
	ctlY := 50
	text.DrawShadowText(screen, "Body", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.bodyCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlY += 100
	text.DrawShadowText(screen, "Hair", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.hairCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlX += 200
	text.DrawShadowText(screen, "Eyes", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.eyesCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlY += 100
	ctlX -= 200
	text.DrawShadowText(screen, "Equip Head", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.equipHeadCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlX += 200
	text.DrawShadowText(screen, "Equip Body", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.equipBodyCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlX -= 200

	ctlY += 150
	text.DrawShadowText(screen, "Body Color", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	ctlY += tileSize / 8
	bg.bodyColorSliders.Draw(screen, float64(ctlX), float64(ctlY))
	_, dy := bg.bodyColorSliders.Dimensions()
	ctlY += dy + (tileSize / 2)
	text.DrawShadowText(screen, "Hair Color", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	ctlY += tileSize / 8
	bg.hairColorSliders.Draw(screen, float64(ctlX), float64(ctlY))
	_, dy = bg.hairColorSliders.Dimensions()
	ctlY += dy + (tileSize / 2)
	text.DrawShadowText(screen, "Eye Color", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	ctlY += tileSize / 8
	bg.eyeColorSliders.Draw(screen, float64(ctlX), float64(ctlY))
}

func (bg *builderGame) Update() error {
	if bg.turnLeft.Update().Clicked {
		bg.entityBody.RotateLeft()
	} else if bg.turnRight.Update().Clicked {
		bg.entityBody.RotateRight()
	}

	bg.bodyCtl.Update()
	if bg.bodyCtl.GetValue() != bg.bodySetIndex {
		bg.SetBodyIndex(bg.bodyCtl.GetValue())
	}
	bg.hairCtl.Update()
	if bg.hairCtl.GetValue() != bg.hairSetIndex {
		bg.SetHairIndex(bg.hairCtl.GetValue())
	}
	bg.eyesCtl.Update()
	if bg.eyesCtl.GetValue() != bg.eyesSetIndex {
		bg.SetEyesIndex(bg.eyesCtl.GetValue())
	}
	bg.equipHeadCtl.Update()
	if bg.equipHeadCtl.GetValue() != bg.equipHeadSetIndex {
		bg.SetEquipHeadIndex(bg.equipHeadCtl.GetValue())
	}
	bg.equipBodyCtl.Update()
	if bg.equipBodyCtl.GetValue() != bg.equipBodySetIndex {
		bg.SetEquipBodyIndex(bg.equipBodyCtl.GetValue())
	}
	bg.bodyColorSliders.Update()
	bg.hairColorSliders.Update()
	bg.eyeColorSliders.Update()

	bg.entityBody.BodyHSV.H = float64(bg.bodyColorSliders.GetValue("H")) / 256
	bg.entityBody.BodyHSV.S = float64(bg.bodyColorSliders.GetValue("S")) / 256
	bg.entityBody.BodyHSV.V = float64(bg.bodyColorSliders.GetValue("V")) / 256
	bg.entityBody.HairHSV.H = float64(bg.hairColorSliders.GetValue("H")) / 256
	bg.entityBody.HairHSV.S = float64(bg.hairColorSliders.GetValue("S")) / 256
	bg.entityBody.HairHSV.V = float64(bg.hairColorSliders.GetValue("V")) / 256
	bg.entityBody.EyesHSV.H = float64(bg.eyeColorSliders.GetValue("H")) / 256
	bg.entityBody.EyesHSV.S = float64(bg.eyeColorSliders.GetValue("S")) / 256
	bg.entityBody.EyesHSV.V = float64(bg.eyeColorSliders.GetValue("V")) / 256

	bg.speedSlider.Update()
	bg.scaleSlider.Update()

	bg.animationSelector.Update()
	selectorValue := bg.animationSelector.GetCurrentValue()
	if selectorValue != bg.entityBody.GetCurrentAnimation() {
		bg.entityBody.SetAnimation(selectorValue)
	}

	bg.entityBody.SetAnimationTickCount(bg.speedSlider.GetValue())

	bg.entityBody.Update()

	return nil
}

func (bg *builderGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return display.SCREEN_WIDTH, display.SCREEN_HEIGHT
}
