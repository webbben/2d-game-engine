/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui/button"
	"github.com/webbben/2d-game-engine/internal/ui/dropdown"
	"github.com/webbben/2d-game-engine/internal/ui/popup"
	"github.com/webbben/2d-game-engine/internal/ui/slider"
	"github.com/webbben/2d-game-engine/internal/ui/stepper"
	"github.com/webbben/2d-game-engine/internal/ui/textfield"
	"github.com/webbben/2d-game-engine/item"
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
	equipBodySetOptions, equipHeadSetOptions                       []body.SelectedPartDef
	equipBodySetIndex, equipHeadSetIndex, weaponSetIndex           int

	weaponOptions []weaponOption

	equipedAux string
	auxTorch   body.SelectedPartDef

	// character info and entity body

	characterData entity.CharacterData

	// UI components

	popupMgr popup.Manager

	turnLeft          *button.Button
	turnRight         *button.Button
	speedSlider       slider.Slider
	scaleSlider       slider.Slider
	animationSelector dropdown.OptionSelect
	auxiliarySelector dropdown.OptionSelect

	bodyCtl      stepper.Stepper
	hairCtl      stepper.Stepper
	eyesCtl      stepper.Stepper
	equipBodyCtl stepper.Stepper
	equipHeadCtl stepper.Stepper

	bodyColorSliders slider.SliderGroup
	hairColorSliders slider.SliderGroup
	eyeColorSliders  slider.SliderGroup

	textField  textfield.TextField
	saveButton *button.Button
}

type weaponOption struct {
	weaponPartDef body.SelectedPartDef
	weaponFxDef   body.SelectedPartDef
}

func characterBuilder() {
	bodyTileset := "entities/parts/body.tsj"
	armsTileset := "entities/parts/arms.tsj"
	eyesTileset := "entities/parts/eyes.tsj"
	hairTileset := "entities/parts/hair.tsj"

	bgTileset := "buildings/walls.tsj"

	bodyRowLength := 68
	bodyRStart := 17
	bodyLStart := 34
	bodyUStart := 51

	itemDefs := GetItemDefs()

	equipBodyOptions := []body.SelectedPartDef{}
	equipHeadOptions := []body.SelectedPartDef{{None: true}}
	weaponOptions := []weaponOption{}
	auxOptions := []body.SelectedPartDef{}

	for _, itemDef := range itemDefs {
		bodyPartDef := itemDef.GetBodyPartDef()
		if bodyPartDef == nil {
			continue
		}

		switch itemDef.GetItemType() {
		case item.TypeBodywear:
			equipBodyOptions = append(equipBodyOptions, *bodyPartDef)
		case item.TypeHeadwear:
			equipHeadOptions = append(equipHeadOptions, *bodyPartDef)
		case item.TypeWeapon:
			asWeaponDef, ok := itemDef.(*item.WeaponDef)
			if !ok {
				panic("failed to assert to weapon def struct")
			}

			weaponOptions = append(weaponOptions, weaponOption{
				weaponPartDef: *bodyPartDef,
				weaponFxDef:   *asWeaponDef.FxPartDef,
			})
		case item.TypeAuxiliary:
			auxOptions = append(auxOptions, *bodyPartDef)
		}
	}
	bodyOptions := []body.SelectedPartDef{
		{
			TilesetSrc:        bodyTileset,
			DStart:            0,
			RStart:            bodyRStart,
			LStart:            bodyLStart,
			UStart:            bodyUStart,
			AuxFirstFrameStep: 1,
		},
		{
			TilesetSrc:        bodyTileset,
			DStart:            0 + (bodyRowLength),
			RStart:            bodyRStart + (bodyRowLength),
			LStart:            bodyLStart + (bodyRowLength),
			UStart:            bodyUStart + (bodyRowLength),
			StretchY:          -1,
			OffsetY:           2,
			AuxFirstFrameStep: 1,
		},
		{
			TilesetSrc:        bodyTileset,
			DStart:            0 + (bodyRowLength * 2),
			RStart:            bodyRStart + (bodyRowLength * 2),
			LStart:            bodyLStart + (bodyRowLength * 2),
			UStart:            bodyUStart + (bodyRowLength * 2),
			StretchX:          2,
			AuxFirstFrameStep: 1,
		},
		{
			TilesetSrc:        bodyTileset,
			DStart:            0 + (bodyRowLength * 3),
			RStart:            bodyRStart + (bodyRowLength * 3),
			LStart:            bodyLStart + (bodyRowLength * 3),
			UStart:            bodyUStart + (bodyRowLength * 3),
			StretchX:          2,
			StretchY:          -1,
			OffsetY:           2,
			AuxFirstFrameStep: 1,
		},
	}
	armsOptions := []body.SelectedPartDef{
		{
			TilesetSrc:        armsTileset,
			DStart:            0,
			RStart:            bodyRStart,
			LStart:            bodyLStart,
			UStart:            bodyUStart,
			AuxFirstFrameStep: 1,
		},
		{
			TilesetSrc:        armsTileset,
			DStart:            0 + (bodyRowLength),
			RStart:            bodyRStart + (bodyRowLength),
			LStart:            bodyLStart + (bodyRowLength),
			UStart:            bodyUStart + (bodyRowLength),
			AuxFirstFrameStep: 1,
		},
		{
			TilesetSrc:        armsTileset,
			DStart:            0 + (bodyRowLength * 2),
			RStart:            bodyRStart + (bodyRowLength * 2),
			LStart:            bodyLStart + (bodyRowLength * 2),
			UStart:            bodyUStart + (bodyRowLength * 2),
			AuxFirstFrameStep: 1,
		},
		{
			TilesetSrc:        armsTileset,
			DStart:            0 + (bodyRowLength * 3),
			RStart:            bodyRStart + (bodyRowLength * 3),
			LStart:            bodyLStart + (bodyRowLength * 3),
			UStart:            bodyUStart + (bodyRowLength * 3),
			AuxFirstFrameStep: 1,
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

	walkTileSteps := []int{0, 3, 0, 5}
	runTileSteps := []int{0, 2, 3, 0, 4, 5}
	slashTileSteps := []int{0, 6, 7, 8, 9}
	backslashTileSteps := []int{9, 10, 11, 12}
	weaponWalkTileSteps := []int{0, 2, 0, 4}
	weaponRunTileSteps := []int{0, 1, 2, 0, 3, 4}
	weaponSlashTileSteps := []int{0, 5, 6, 7, 8}
	weaponBackslashTileSteps := []int{8, 9, 10, 11}

	bodySet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:   "bodySet",
		IsBody: true,
		HasUp:  true,
		WalkParams: body.AnimationParams{
			TileSteps:    walkTileSteps,
			StepsOffsetY: []int{0, 1, 0, 1},
		},
		RunParams: body.AnimationParams{
			TileSteps:    runTileSteps,
			StepsOffsetY: []int{0, 0, 1, 0, 0, 1},
		},
		SlashParams: body.AnimationParams{
			TileSteps:    slashTileSteps,
			StepsOffsetY: []int{0, 1, 2, 2, 2},
		},
		BackslashParams: body.AnimationParams{
			TileSteps:    backslashTileSteps,
			StepsOffsetY: []int{2, 2, 1, 1},
		},
	})
	armsSet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:  "armsSet",
		HasUp: true,
		WalkParams: body.AnimationParams{
			TileSteps: walkTileSteps,
		},
		RunParams: body.AnimationParams{
			TileSteps: runTileSteps,
		},
		SlashParams: body.AnimationParams{
			TileSteps: slashTileSteps,
		},
		BackslashParams: body.AnimationParams{
			TileSteps: backslashTileSteps,
		},
	})
	eyesSet := body.NewBodyPartSet(body.BodyPartSetParams{Name: "eyesSet"})
	hairSet := body.NewBodyPartSet(body.BodyPartSetParams{HasUp: true, Name: "hairSet"})
	equipBodySet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:  "equipBodySet",
		HasUp: true,
		WalkParams: body.AnimationParams{
			TileSteps: walkTileSteps,
		},
		RunParams: body.AnimationParams{
			TileSteps: runTileSteps,
		},
		SlashParams: body.AnimationParams{
			TileSteps: slashTileSteps,
		},
		BackslashParams: body.AnimationParams{
			TileSteps: backslashTileSteps,
		},
	})
	equipHeadSet := body.NewBodyPartSet(body.BodyPartSetParams{HasUp: true, Name: "equipHeadSet"})
	weaponSet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:        "weaponSet",
		HasUp:       true,
		IsRemovable: true,
		WalkParams: body.AnimationParams{
			TileSteps: weaponWalkTileSteps,
		},
		RunParams: body.AnimationParams{
			TileSteps: weaponRunTileSteps,
		},
		SlashParams: body.AnimationParams{
			TileSteps: weaponSlashTileSteps,
		},
		BackslashParams: body.AnimationParams{
			TileSteps: weaponBackslashTileSteps,
		},
	})
	weaponFxSet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:        "weaponFxSet",
		HasUp:       true,
		IsRemovable: true,
		WalkParams:  body.AnimationParams{Skip: true},
		RunParams:   body.AnimationParams{Skip: true},
		IdleParams:  body.AnimationParams{Skip: true},
		SlashParams: body.AnimationParams{
			TileSteps: []int{-1, -1, 0, 1, 2}, // -1 = skip a frame (nil image)
		},
		BackslashParams: body.AnimationParams{
			TileSteps: []int{-1, 3, 4, 5},
		},
	})
	auxSet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:        "auxSet",
		HasUp:       true,
		IsRemovable: true,
		IdleParams: body.AnimationParams{
			TileSteps: []int{0, 1, 2, 3},
		},
		WalkParams: body.AnimationParams{
			TileSteps: []int{0, 5, 0, 7},
		},
		RunParams: body.AnimationParams{
			TileSteps: []int{0, 4, 5, 0, 6, 7},
		},
		SlashParams: body.AnimationParams{
			TileSteps: []int{0, 8, 9, 10, 10},
		},
		BackslashParams: body.AnimationParams{
			TileSteps: []int{11, 12, 13, 14},
		},
	})

	entBody := body.NewEntityBodySet(bodySet, armsSet, hairSet, eyesSet, equipHeadSet, equipBodySet, weaponSet, weaponFxSet, auxSet, nil, nil, nil)

	g := builderGame{
		bodySetOptions:      bodyOptions,
		armsSetOptions:      armsOptions,
		eyesSetOptions:      eyesOptions,
		hairSetOptions:      hairOptions,
		equipBodySetOptions: equipBodyOptions,
		equipHeadSetOptions: equipHeadOptions,
		weaponOptions:       weaponOptions,
		characterData: entity.CharacterData{
			Body: entBody,
		},

		auxTorch: auxOptions[0],
	}

	// Set each bodyPartSet with their initial data.
	// We do this in a "weird" way here since this is the character builder screen.
	// In the actual game, we use the Load function instead, since all the PartSrc's are already set (from the JSON data).
	g.SetBodyIndex(0)
	g.SetEquipHeadIndex(0)
	g.SetHairIndex(0)
	g.SetEyesIndex(0)
	g.SetEquipBodyIndex(0)
	g.SetWeaponIndex(0)
	g.characterData.Body.AuxItemSet.Hide()

	// run this just to confirm that the regular loading process also still works (as used in the actual game)
	g.characterData.Body.Load()

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

	g.characterData.Body.SetDirection('D')

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

	g.popupMgr = popup.NewPopupManager()

	g.animationSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               []string{body.ANIM_IDLE, body.ANIM_WALK, body.ANIM_RUN, body.ANIM_SLASH, body.ANIM_BACKSLASH},
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)
	g.auxiliarySelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               []string{"None", "Torch"},
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)

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

	g.textField = *textfield.NewTextField(textfield.TextFieldParams{
		FontFace:     config.DefaultFont,
		WidthPx:      500,
		AllowSpecial: true,
		TextColor:    color.White,
		BorderColor:  color.White,
		BgColor:      color.Black,
	})

	g.saveButton = button.NewLinearBoxButton("Save JSON", "ui/ui-components.tsj", 352, config.DefaultTitleFont)

	if err := ebiten.RunGame(&g); err != nil {
		panic(err)
	}
}

func (bg builderGame) saveCharacter() {
	outputFileName := bg.textField.GetText()
	if outputFileName == "" {
		return
	}

	basePath := "/Users/benwebb/dev/personal/ancient-rome/src/data/characters/json/" + outputFileName + ".json"

	bg.characterData.Body.WriteToJSON(basePath)
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

	bg.characterData.Body.SetBody(bodyDef, armDef)
}
func (bg *builderGame) SetEyesIndex(i int) {
	if i < 0 || i >= len(bg.eyesSetOptions) {
		panic("out of bounds")
	}
	bg.eyesSetIndex = i
	op := bg.eyesSetOptions[i]
	bg.characterData.Body.SetEyes(op)
}
func (bg *builderGame) SetHairIndex(i int) {
	if i < 0 || i >= len(bg.hairSetOptions) {
		panic("out of bounds")
	}
	bg.hairSetIndex = i
	op := bg.hairSetOptions[i]
	bg.characterData.Body.SetHair(op)
}
func (bg *builderGame) SetEquipBodyIndex(i int) {
	if i < 0 || i >= len(bg.equipBodySetOptions) {
		panic("out of bounds")
	}
	bg.equipBodySetIndex = i
	op := bg.equipBodySetOptions[i]
	bg.characterData.Body.SetEquipBody(op)
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
	bg.characterData.Body.SetEquipHead(op)
}
func (bg *builderGame) SetWeaponIndex(i int) {
	if i < 0 || i > len(bg.weaponOptions) {
		panic("out of bounds")
	}
	bg.weaponSetIndex = i
	op := bg.weaponOptions[i]
	bg.characterData.Body.SetWeapon(op.weaponPartDef, op.weaponFxDef)
}

func (bg *builderGame) Draw(screen *ebiten.Image) {
	var characterScale float64 = float64(bg.scaleSlider.GetValue())
	characterTileSize := config.TileSize * characterScale

	tileSize := int(config.TileSize * config.UIScale)

	bodyDx, bodyDy := bg.characterData.Body.Dimensions()
	bodyWidth := float64(bodyDx) * characterScale
	bodyHeight := float64(bodyDy) * characterScale

	bodyX := float64(display.SCREEN_WIDTH/2) - (bodyWidth / 2)
	bodyY := float64(display.SCREEN_HEIGHT/2) - (bodyHeight / 2)

	// Backdrop
	rendering.DrawImage(screen, bg.bgImg, bodyX-characterTileSize, bodyY-characterTileSize, characterScale)

	// Character body
	bg.characterData.Body.Draw(screen, bodyX, bodyY, characterScale)

	buttonsY := bodyY + (bodyHeight) + 20
	buttonLX := (display.SCREEN_WIDTH / 2) - bg.turnLeft.Width - 20
	buttonRX := (display.SCREEN_WIDTH / 2) + 20
	bg.turnLeft.Draw(screen, buttonLX, int(buttonsY))
	bg.turnRight.Draw(screen, buttonRX, int(buttonsY))

	// UI controls - Left side
	sliderX := 100
	sliderY := 50
	text.DrawShadowText(screen, "Ticks Per Frame", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	text.DrawShadowText(screen, fmt.Sprintf("%v", bg.speedSlider.GetValue()), config.DefaultFont, sliderX-40, sliderY+(tileSize*2/3), color.White, nil, 0, 0)
	bg.speedSlider.Draw(screen, float64(sliderX), float64(sliderY))

	sliderY += tileSize * 2
	text.DrawShadowText(screen, "Scale", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	text.DrawShadowText(screen, fmt.Sprintf("%v", bg.scaleSlider.GetValue()), config.DefaultFont, sliderX-40, sliderY+(tileSize*2/3), color.White, nil, 0, 0)
	bg.scaleSlider.Draw(screen, float64(sliderX), float64(sliderY))

	sliderY += tileSize * 2
	text.DrawShadowText(screen, "Animation", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.animationSelector.Draw(screen, float64(sliderX), float64(sliderY), nil)
	sliderY += tileSize * 2
	text.DrawShadowText(screen, "Auxiliary", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.auxiliarySelector.Draw(screen, float64(sliderX), float64(sliderY), nil)

	saveX := sliderX
	saveY := display.SCREEN_HEIGHT - tileSize*2
	bg.textField.Draw(screen, float64(saveX), float64(saveY))
	textFieldDx, _ := bg.textField.Dimensions()
	saveX += textFieldDx + tileSize
	bg.saveButton.Draw(screen, saveX, saveY)

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

	bg.popupMgr.Draw(screen)
}

func (bg *builderGame) Update() error {
	if bg.popupMgr.IsPopupActive() {
		bg.popupMgr.Update()
		return nil
	}

	if bg.turnLeft.Update().Clicked {
		bg.characterData.Body.RotateLeft()
	} else if bg.turnRight.Update().Clicked {
		bg.characterData.Body.RotateRight()
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

	bg.characterData.Body.SetBodyHSV(
		float64(bg.bodyColorSliders.GetValue("H"))/256,
		float64(bg.bodyColorSliders.GetValue("S"))/256,
		float64(bg.bodyColorSliders.GetValue("V"))/256,
	)
	bg.characterData.Body.SetHairHSV(
		float64(bg.hairColorSliders.GetValue("H"))/256,
		float64(bg.hairColorSliders.GetValue("S"))/256,
		float64(bg.hairColorSliders.GetValue("V"))/256,
	)
	bg.characterData.Body.SetEyesHSV(
		float64(bg.eyeColorSliders.GetValue("H"))/256,
		float64(bg.eyeColorSliders.GetValue("S"))/256,
		float64(bg.eyeColorSliders.GetValue("V"))/256,
	)

	bg.speedSlider.Update()
	bg.scaleSlider.Update()

	bg.animationSelector.Update()
	selectorValue := bg.animationSelector.GetCurrentValue()
	if selectorValue != bg.characterData.Body.GetCurrentAnimation() {
		bg.characterData.Body.SetAnimation(selectorValue, body.SetAnimationOps{Force: true})
	}
	bg.auxiliarySelector.Update()
	selectorValue = bg.auxiliarySelector.GetCurrentValue()
	if selectorValue != bg.equipedAux {
		bg.handleChangeAux(selectorValue)
	}

	bg.characterData.Body.SetAnimationTickCount(bg.speedSlider.GetValue())

	bg.textField.Update()
	if bg.saveButton.Update().Clicked {
		bg.saveCharacter()
	}

	bg.characterData.Body.Update()

	return nil
}

func (bg *builderGame) handleChangeAux(val string) {
	switch val {
	case "None":
		// remove aux item
		bg.characterData.Body.AuxItemSet.Remove()
	case "Torch":
		bg.characterData.Body.SetAuxiliary(bg.auxTorch)
	default:
		panic("unrecognized aux value: " + val)
	}
}

func (bg *builderGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return display.SCREEN_WIDTH, display.SCREEN_HEIGHT
}
