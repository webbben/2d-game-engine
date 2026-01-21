package cmd

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/logz"
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

const noneOp = "< None >"

var loadFile string

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
		characterBuilder(loadFile)
	},
}

func init() {
	characterBuilderCmd.PersistentFlags().StringVar(&loadFile, "load", "", "set this flag to load an existing character json file. directory path and extension (.json) should be omitted - just the filename (character ID).")

	rootCmd.AddCommand(characterBuilderCmd)
}

type builderGame struct {
	bgImg *ebiten.Image

	// body part options (characterBuilder only)

	bodySetOptions, eyesSetOptions, hairSetOptions, armsSetOptions, legsSetOptions []body.SelectedPartDef
	bodySetIndex, eyesSetIndex, hairSetIndex, armsSetIndex                         int
	bodywearItems, headwearItems                                                   []item.ItemDef
	equipedBodywear, equipedHeadwear, equipedWeapon, equipedAux                    string // IDs of equiped items

	weaponItems []item.ItemDef

	auxItems []item.ItemDef

	// character info and entity body

	characterData entity.CharacterData
	defMgr        *definitions.DefinitionManager

	// UI components

	popupMgr popup.Manager

	turnLeft          *button.Button
	turnRight         *button.Button
	speedSlider       slider.Slider
	scaleSlider       slider.Slider
	animationSelector dropdown.OptionSelect
	auxiliarySelector dropdown.OptionSelect
	bodywearSelector  dropdown.OptionSelect
	headwearSelector  dropdown.OptionSelect
	weaponSelector    dropdown.OptionSelect

	bodyCtl stepper.Stepper
	hairCtl stepper.Stepper
	eyesCtl stepper.Stepper

	bodyColorSliders slider.SliderGroup
	hairColorSliders slider.SliderGroup
	eyeColorSliders  slider.SliderGroup

	saveButton *button.Button

	nameField textfield.TextField
	idField   textfield.TextField
}

type weaponOption struct {
	weaponPartDef body.SelectedPartDef
	weaponFxDef   body.SelectedPartDef
}

type equipBodyOption struct {
	bodyDef body.SelectedPartDef
	legsDef body.SelectedPartDef
}

func characterBuilder(fileToLoad string) {
	bodyTileset := "entities/parts/human_entity_parts.tsj"
	armsTileset := "entities/parts/human_entity_parts.tsj"
	eyesTileset := "entities/parts/eyes.tsj"
	hairTileset := "entities/parts/hair.tsj"

	bgTileset := "buildings/walls.tsj"

	defMgr := definitions.NewDefinitionManager()
	itemDefs := GetItemDefs()
	defMgr.LoadItemDefs(itemDefs)

	bodywearItems := []item.ItemDef{}
	headwearItems := []item.ItemDef{}
	weaponItems := []item.ItemDef{}
	auxItems := []item.ItemDef{}

	for _, itemDef := range itemDefs {
		switch itemDef.GetItemType() {
		case item.TypeBodywear:
			bodywearItems = append(bodywearItems, itemDef)
		case item.TypeHeadwear:
			headwearItems = append(headwearItems, itemDef)
		case item.TypeWeapon:
			weaponItems = append(weaponItems, itemDef)
		case item.TypeAuxiliary:
			auxItems = append(auxItems, itemDef)
		}
	}
	bodyOptions := []body.SelectedPartDef{
		body.NewPartDef(body.PartDefParams{
			Idle: &body.AnimationParams{
				TilesetSrc:   bodyTileset,
				TilesDown:    []int{73},
				TilesRight:   []int{92},
				TilesLeft:    []int{110},
				TilesUp:      []int{128},
				StepsOffsetY: []int{0},
			},
			Walk: &body.AnimationParams{
				TilesetSrc:   bodyTileset,
				TilesDown:    []int{75, 73, 77, 73},
				TilesRight:   []int{94, 92, 94, 92},
				TilesLeft:    []int{112, 110, 112, 110},
				TilesUp:      []int{130, 128, 130, 128},
				StepsOffsetY: []int{1, 0, 1, 0},
			},
			Run: &body.AnimationParams{
				TilesetSrc:   bodyTileset,
				TilesDown:    []int{74, 75, 73, 76, 77, 73},
				TilesRight:   []int{93, 94, 92, 93, 94, 92},
				TilesLeft:    []int{111, 112, 110, 111, 112, 110},
				TilesUp:      []int{129, 130, 128, 129, 130, 128},
				StepsOffsetY: []int{0, 1, 0, 0, 1, 0},
			},
			Slash: &body.AnimationParams{
				TilesetSrc:     bodyTileset,
				TilesDown:      []int{78, 79, 79, 79},
				TilesRight:     []int{94, 95, 95, 95},
				TilesLeft:      []int{112, 113, 113, 113},
				TilesUp:        []int{128, 131, 131, 131},
				StepsOffsetY:   []int{1, 2, 2, 2},
				StepsOffsetYUp: []int{0, 1, 1, 1},
			},
			Backslash: &body.AnimationParams{
				TilesetSrc:     bodyTileset,
				TilesDown:      []int{79, 79, 78, 78},
				TilesRight:     []int{95, 95, 94, 94},
				TilesLeft:      []int{113, 113, 112, 112},
				TilesUp:        []int{131, 131, 128, 128},
				StepsOffsetY:   []int{2, 2, 1, 1},
				StepsOffsetYUp: []int{1, 1, 0, 0},
			},
			Shield: &body.AnimationParams{
				TilesetSrc:   bodyTileset,
				TilesDown:    []int{78},
				TilesRight:   []int{94},
				TilesLeft:    []int{112},
				TilesUp:      []int{128},
				StepsOffsetY: []int{1},
			},
		}),
	}

	armsOptions := []body.SelectedPartDef{
		body.NewPartDef(body.PartDefParams{
			Idle: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{146},
				TilesRight: []int{165},
				TilesLeft:  []int{183},
				TilesUp:    []int{201},
			},
			Walk: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{150, 146, 152, 146},
				TilesRight: []int{169, 165, 171, 165},
				TilesLeft:  []int{187, 183, 189, 183},
				TilesUp:    []int{205, 201, 207, 201},
			},
			Run: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{149, 150, 146, 151, 152, 146},
				TilesRight: []int{168, 169, 165, 170, 171, 165},
				TilesLeft:  []int{186, 187, 183, 188, 189, 183},
				TilesUp:    []int{204, 205, 201, 206, 207, 201},
			},
			Slash: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{153, 154, 155, 155},
				TilesRight: []int{172, 173, 174, 174},
				TilesLeft:  []int{190, 191, 192, 192},
				TilesUp:    []int{208, 207, 209, 209},
			},
			Backslash: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{155, 154, 153, 153},
				TilesRight: []int{174, 173, 172, 172},
				TilesLeft:  []int{192, 191, 190, 190},
				TilesUp:    []int{209, 207, 208, 208},
			},
			Shield: &body.AnimationParams{
				TilesetSrc: armsTileset,
				TilesDown:  []int{157},
				TilesRight: []int{175},
				TilesLeft:  []int{193},
				TilesUp:    []int{204},
			},
		}),
	}
	legsOptions := []body.SelectedPartDef{
		body.NewPartDef(body.PartDefParams{
			Idle: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{80},
				TilesRight: []int{99},
				TilesLeft:  []int{117},
				TilesUp:    []int{135},
			},
			Walk: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{84, 80, 86, 80},
				TilesRight: []int{103, 99, 105, 99},
				TilesLeft:  []int{121, 117, 123, 117},
				TilesUp:    []int{139, 135, 141, 135},
			},
			Run: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{83, 84, 80, 85, 86, 80},
				TilesRight: []int{102, 103, 99, 104, 105, 99},
				TilesLeft:  []int{120, 121, 117, 122, 123, 117},
				TilesUp:    []int{138, 139, 135, 140, 141, 135},
			},
			Slash: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{87, 87, 87, 87},
				TilesRight: []int{103, 106, 106, 106},
				TilesLeft:  []int{123, 124, 124, 124},
				TilesUp:    []int{142, 142, 142, 142},
			},
			Backslash: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{87, 87, 87, 87},
				TilesRight: []int{106, 106, 103, 103},
				TilesLeft:  []int{124, 124, 123, 123},
				TilesUp:    []int{142, 142, 142, 142},
			},
			Shield: &body.AnimationParams{
				TilesetSrc: bodyTileset,
				TilesDown:  []int{87},
				TilesRight: []int{106},
				TilesLeft:  []int{124},
				TilesUp:    []int{142},
			},
		}),
	}

	eyesOptions := []body.SelectedPartDef{}
	for i := range 14 {
		numCols := 32 // number of colums in the tileset
		animParams := body.AnimationParams{
			TilesetSrc: eyesTileset,
			TilesDown:  []int{numCols * i},
			TilesRight: []int{1 + (numCols * i)},
		}
		eyesOptions = append(eyesOptions, body.NewPartDef(body.PartDefParams{
			FlipRForL: true,
			Idle:      &animParams,
			Walk:      &animParams,
			Run:       &animParams,
			Slash:     &animParams,
			Backslash: &animParams,
			Shield:    &animParams,
		}))
	}
	hairOptions := []body.SelectedPartDef{}
	for i := range 7 {
		numCols := 32
		animParams := body.AnimationParams{
			TilesetSrc: hairTileset,
			TilesDown:  []int{numCols * i},
			TilesRight: []int{(numCols * i) + 1},
			TilesLeft:  []int{(numCols * i) + 2},
			TilesUp:    []int{(numCols * i) + 3},
		}
		hairOptions = append(hairOptions, body.NewPartDef(body.PartDefParams{
			Idle:      &animParams,
			Walk:      &animParams,
			Run:       &animParams,
			Slash:     &animParams,
			Backslash: &animParams,
			Shield:    &animParams,
		}))
	}

	var characterData entity.CharacterData
	if fileToLoad != "" {
		var err error
		fileToLoad = resolveCharacterJSONPath(fileToLoad)
		characterData, err = entity.LoadCharacterDataJSON(fileToLoad, defMgr)
		if err != nil {
			logz.Panicln("Character Builder", "load character data:", err)
		}
	} else {
		characterData = getNewCharacter()
	}

	g := builderGame{
		bodySetOptions: bodyOptions,
		armsSetOptions: armsOptions,
		legsSetOptions: legsOptions,
		eyesSetOptions: eyesOptions,
		hairSetOptions: hairOptions,
		bodywearItems:  bodywearItems,
		headwearItems:  headwearItems,
		weaponItems:    weaponItems,
		characterData:  characterData,
		auxItems:       auxItems,

		defMgr: defMgr,
	}

	// do this if using a new, empty slate character body
	if fileToLoad == "" {
		// Set each bodyPartSet with their initial data.
		// We do this in a "weird" way here since this is the character builder screen.
		// In the actual game, we use the Load function instead, since all the PartSrc's are already set (from the JSON data).
		g.SetBodyIndex(0)
		g.SetHairIndex(0)
		g.SetEyesIndex(0)
		g.characterData.Body.AuxItemSet.Hide()
	}

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
		MaxVal:        8,
		InitialValue:  8,
		StepSize:      1,
	})

	g.popupMgr = popup.NewPopupManager()

	g.animationSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               []string{body.AnimIdle, body.AnimWalk, body.AnimRun, body.AnimSlash, body.AnimBackslash, body.AnimShield},
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)

	auxOptions := []string{noneOp}
	for _, auxItem := range g.auxItems {
		auxOptions = append(auxOptions, auxItem.GetID())
	}
	g.auxiliarySelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               auxOptions,
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)

	headwearOptions := []string{noneOp}
	for _, i := range g.headwearItems {
		headwearOptions = append(headwearOptions, i.GetID())
	}
	g.headwearSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               headwearOptions,
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)
	bodywearOptions := []string{noneOp}
	for _, i := range g.bodywearItems {
		bodywearOptions = append(bodywearOptions, i.GetID())
	}
	g.bodywearSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               bodywearOptions,
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)
	weaponOptions := []string{noneOp}
	for _, i := range g.weaponItems {
		weaponOptions = append(weaponOptions, i.GetID())
	}
	g.weaponSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               weaponOptions,
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)

	if len(bodyOptions) > 1 {
		g.bodyCtl = stepper.NewStepper(stepper.StepperParams{
			MinVal:               0,
			MaxVal:               len(bodyOptions) - 1,
			Font:                 config.DefaultTitleFont,
			FontFg:               color.White,
			FontBg:               color.Black,
			DecrementButtonImage: turnLeftImg,
			IncrementButtonImage: turnRightImg,
		})
	}

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

	g.nameField = *textfield.NewTextField(textfield.TextFieldParams{
		FontFace:     config.DefaultFont,
		WidthPx:      200,
		AllowSpecial: true,
		TextColor:    color.White,
		BorderColor:  color.White,
		BgColor:      color.Black,
	})
	g.idField = *textfield.NewTextField(textfield.TextFieldParams{
		FontFace:     config.DefaultFont,
		WidthPx:      200,
		AllowSpecial: true,
		TextColor:    color.White,
		BorderColor:  color.White,
		BgColor:      color.Black,
	})

	g.saveButton = button.NewLinearBoxButton("Save", "ui/ui-components.tsj", 352, config.DefaultTitleFont)

	if err := ebiten.RunGame(&g); err != nil {
		panic(err)
	}
}

func getNewCharacter() entity.CharacterData {
	bodySet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:   "bodySet",
		IsBody: true,
		HasUp:  true,
	})
	armsSet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:  "armsSet",
		HasUp: true,
	})
	legsSet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:  "legsSet",
		HasUp: true,
	})
	eyesSet := body.NewBodyPartSet(body.BodyPartSetParams{Name: "eyesSet"})
	hairSet := body.NewBodyPartSet(body.BodyPartSetParams{HasUp: true, Name: "hairSet"})
	equipBodySet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:        "equipBodySet",
		HasUp:       true,
		IsRemovable: true,
	})
	equipLegsSet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:        "equipLegsSet",
		HasUp:       true,
		IsRemovable: true,
	})
	equipHeadSet := body.NewBodyPartSet(body.BodyPartSetParams{
		HasUp:       true,
		Name:        "equipHeadSet",
		IsRemovable: true,
	})
	weaponSet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:        "weaponSet",
		HasUp:       true,
		IsRemovable: true,
	})
	weaponFxSet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:        "weaponFxSet",
		HasUp:       true,
		IsRemovable: true,
	})
	auxSet := body.NewBodyPartSet(body.BodyPartSetParams{
		Name:        "auxSet",
		HasUp:       true,
		IsRemovable: true,
	})

	entBody := body.NewEntityBodySet(bodySet, armsSet, legsSet, hairSet, eyesSet, equipHeadSet, equipBodySet, equipLegsSet, weaponSet, weaponFxSet, auxSet, nil, nil, nil)

	// Setting these various fields just to prevent validation errors (e.g. WalkSpeed). But, these values are eventually overwritten
	// when used in the actual game.
	cd := entity.CharacterData{
		ID:             "newCharacter",
		DisplayName:    "newCharacter",
		Body:           entBody,
		WalkSpeed:      1,
		RunSpeed:       1,
		InventoryItems: make([]*item.InventoryItem, 1),
	}
	return cd
}

func (bg builderGame) saveCharacter() {
	id := bg.idField.GetText()
	if id == "" {
		return
	}
	name := bg.nameField.GetText()
	if name == "" {
		return
	}

	basePath := resolveCharacterJSONPath(id)

	bg.characterData.ID = id
	bg.characterData.DisplayName = name

	err := bg.characterData.WriteToJSON(basePath)
	if err != nil {
		logz.Panicln("saveCharacter", "error occurred while saving character data to JSON:", err)
	}
}

func resolveCharacterJSONPath(id string) string {
	return fmt.Sprintf("/Users/benwebb/dev/personal/ancient-rome/src/data/characters/json/%s.json", id)
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

	legDef := bg.legsSetOptions[i]

	bg.characterData.Body.SetBody(bodyDef, armDef, legDef)
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

func (bg *builderGame) Draw(screen *ebiten.Image) {
	characterScale := float64(bg.scaleSlider.GetValue())
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
	sliderX := 50
	sliderY := 50
	text.DrawShadowText(screen, "Ticks Per Frame", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	text.DrawShadowText(screen, fmt.Sprintf("%v", bg.speedSlider.GetValue()), config.DefaultFont, sliderX-30, sliderY+(tileSize*2/3), color.White, nil, 0, 0)
	bg.speedSlider.Draw(screen, float64(sliderX), float64(sliderY))

	sliderY += tileSize * 3 / 2
	text.DrawShadowText(screen, "Scale", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	text.DrawShadowText(screen, fmt.Sprintf("%v", bg.scaleSlider.GetValue()), config.DefaultFont, sliderX-30, sliderY+(tileSize*2/3), color.White, nil, 0, 0)
	bg.scaleSlider.Draw(screen, float64(sliderX), float64(sliderY))

	sliderY += tileSize * 5 / 3
	text.DrawShadowText(screen, "Animation", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.animationSelector.Draw(screen, float64(sliderX), float64(sliderY), nil)
	sliderY += tileSize * 5 / 3
	text.DrawShadowText(screen, "Auxiliary", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.auxiliarySelector.Draw(screen, float64(sliderX), float64(sliderY), nil)

	sliderY += tileSize * 5 / 3
	text.DrawShadowText(screen, "Headwear", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.headwearSelector.Draw(screen, float64(sliderX), float64(sliderY), nil)
	sliderY += tileSize * 5 / 3
	text.DrawShadowText(screen, "Bodywear", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.bodywearSelector.Draw(screen, float64(sliderX), float64(sliderY), nil)
	sliderY += tileSize * 5 / 3
	text.DrawShadowText(screen, "Weapon", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.weaponSelector.Draw(screen, float64(sliderX), float64(sliderY), nil)

	// Character Info Input - Middle Top
	drawX := (display.SCREEN_WIDTH / 3) - 100
	drawY := 50
	text.DrawShadowText(screen, "Character Info", config.DefaultTitleFont, drawX, drawY, color.White, nil, 0, 0)
	drawY += 50
	text.DrawShadowText(screen, "Name", config.DefaultFont, drawX, drawY, color.White, nil, 0, 0)
	drawX += 100
	bg.nameField.Draw(screen, float64(drawX), float64(drawY-25))
	drawX -= 100
	drawY += 50
	text.DrawShadowText(screen, "ID", config.DefaultFont, drawX, drawY, color.White, nil, 0, 0)
	drawX += 100
	bg.idField.Draw(screen, float64(drawX), float64(drawY-25))
	drawX = (display.SCREEN_WIDTH / 3) + 350
	drawY = 150
	bg.saveButton.Draw(screen, drawX, drawY)

	// UI controls - Right side
	ctlX := (display.SCREEN_WIDTH * 3 / 4) - 100
	ctlY := 50
	if len(bg.bodySetOptions) > 1 {
		text.DrawShadowText(screen, "Body", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
		bg.bodyCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	}
	ctlX += 170
	text.DrawShadowText(screen, "Body Color", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	ctlY += tileSize / 8
	bg.bodyColorSliders.Draw(screen, float64(ctlX), float64(ctlY))
	_, dy := bg.bodyColorSliders.Dimensions()
	ctlY += dy + (tileSize)
	ctlX -= 170

	text.DrawShadowText(screen, "Hair", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.hairCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlX += 170
	text.DrawShadowText(screen, "Hair Color", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	ctlY += tileSize / 8
	bg.hairColorSliders.Draw(screen, float64(ctlX), float64(ctlY))
	_, dy = bg.hairColorSliders.Dimensions()
	ctlY += dy + (tileSize)
	ctlX -= 170

	text.DrawShadowText(screen, "Eyes", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.eyesCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlX += 170
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

	if len(bg.bodySetOptions) > 1 {
		bg.bodyCtl.Update()
		if bg.bodyCtl.GetValue() != bg.bodySetIndex {
			bg.SetBodyIndex(bg.bodyCtl.GetValue())
		}
	}

	bg.hairCtl.Update()
	if bg.hairCtl.GetValue() != bg.hairSetIndex {
		bg.SetHairIndex(bg.hairCtl.GetValue())
	}
	bg.eyesCtl.Update()
	if bg.eyesCtl.GetValue() != bg.eyesSetIndex {
		bg.SetEyesIndex(bg.eyesCtl.GetValue())
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
	bg.headwearSelector.Update()
	selectorValue = bg.headwearSelector.GetCurrentValue()
	if selectorValue != bg.equipedHeadwear {
		bg.handleChangeHeadwear(selectorValue)
	}
	bg.bodywearSelector.Update()
	selectorValue = bg.bodywearSelector.GetCurrentValue()
	if selectorValue != bg.equipedBodywear {
		bg.handleChangeBodywear(selectorValue)
	}
	bg.weaponSelector.Update()
	selectorValue = bg.weaponSelector.GetCurrentValue()
	if selectorValue != bg.equipedWeapon {
		bg.handleChangeWeapon(selectorValue)
	}

	bg.characterData.Body.SetAnimationTickCount(bg.speedSlider.GetValue())

	bg.nameField.Update()
	bg.idField.Update()

	if bg.saveButton.Update().Clicked {
		bg.saveCharacter()
	}

	bg.characterData.Body.Update()

	return nil
}

func (bg *builderGame) handleChangeAux(val string) {
	bg.equipedAux = val

	if val == noneOp {
		bg.characterData.EquipedAuxiliary = nil
		bg.characterData.Body.AuxItemSet.Remove()
		return
	}
	for _, auxItem := range bg.auxItems {
		if auxItem.GetID() == val {
			bg.characterData.EquipedAuxiliary = nil // unset the previous value so it's not added to inventory
			bg.characterData.EquipItem(item.NewInventoryItem(auxItem, 1))
			return
		}
	}
	panic("val doesn't seem to match an item ID: " + val)
}

func (bg *builderGame) handleChangeHeadwear(val string) {
	bg.equipedHeadwear = val

	if val == noneOp {
		if bg.characterData.EquipedHeadwear == nil {
			return
		}
		bg.characterData.UnequipHeadwear()
		return
	}
	for _, i := range bg.headwearItems {
		if i.GetID() == val {
			bg.characterData.EquipedHeadwear = nil
			bg.characterData.EquipItem(item.NewInventoryItem(i, 1))
			return
		}
	}
	panic("val doesn't seem to match an item ID: " + val)
}

func (bg *builderGame) handleChangeBodywear(val string) {
	bg.equipedBodywear = val

	if val == noneOp {
		if bg.characterData.EquipedBodywear == nil {
			return
		}
		bg.characterData.UnequipBodywear()
		return
	}
	for _, i := range bg.bodywearItems {
		if i.GetID() == val {
			bg.characterData.EquipedBodywear = nil
			bg.characterData.EquipItem(item.NewInventoryItem(i, 1))
			return
		}
	}
	panic("val doesn't seem to match an item ID: " + val)
}

func (bg *builderGame) handleChangeWeapon(val string) {
	bg.equipedWeapon = val

	if val == noneOp {
		bg.characterData.Body.WeaponSet.Remove()
		bg.characterData.Body.WeaponFxSet.Remove()
		return
	}
	for _, i := range bg.weaponItems {
		if i.GetID() == val {
			bg.characterData.EquipItem(item.NewInventoryItem(i, 1))
			return
		}
	}
	panic("val doesn't seem to match an item ID: " + val)
}

func (bg *builderGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return display.SCREEN_WIDTH, display.SCREEN_HEIGHT
}
