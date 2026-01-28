package cmd

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/ui/popup"
	"github.com/webbben/2d-game-engine/internal/ui/tab"
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
		config.DefaultTooltipBox = config.DefaultBox{
			TilesetSrc:  "boxes/boxes.tsj",
			OriginIndex: 132,
		}

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
	// body part options

	bodySkinSets                                                                 []EntityBodySkinSet
	eyesSetOptions, hairSetOptions                                               []body.SelectedPartDef
	bodySetIndex, eyesSetIndex, hairSetIndex                                     int
	bodywearItems, headwearItems, footwearItems                                  []item.ItemDef
	equipedBodywear, equipedHeadwear, equipedFootwear, equipedWeapon, equipedAux string // IDs of equiped items

	weaponItems []item.ItemDef

	auxItems []item.ItemDef

	// character info and entity body

	characterData entity.CharacterData
	defMgr        *definitions.DefinitionManager

	// UI components

	scrAppearance appearanceScreen

	popupMgr popup.Manager

	tabControl tab.TabControl
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
	defMgr := definitions.NewDefinitionManager()
	itemDefs := GetItemDefs()
	defMgr.LoadItemDefs(itemDefs)
	bodySkins, eyeSets, hairSets := GetAllEntityBodyPartSets()
	for _, skin := range bodySkins {
		defMgr.LoadBodyPartDef(skin.Body)
		defMgr.LoadBodyPartDef(skin.Arms)
		defMgr.LoadBodyPartDef(skin.Legs)
	}
	for _, eyes := range eyeSets {
		defMgr.LoadBodyPartDef(eyes)
	}
	for _, hair := range hairSets {
		defMgr.LoadBodyPartDef(hair)
	}

	bodywearItems := []item.ItemDef{}
	headwearItems := []item.ItemDef{}
	footwearItems := []item.ItemDef{}
	weaponItems := []item.ItemDef{}
	auxItems := []item.ItemDef{}

	for _, itemDef := range itemDefs {
		switch itemDef.GetItemType() {
		case item.TypeBodywear:
			bodywearItems = append(bodywearItems, itemDef)
		case item.TypeHeadwear:
			headwearItems = append(headwearItems, itemDef)
		case item.TypeFootwear:
			footwearItems = append(footwearItems, itemDef)
		case item.TypeWeapon:
			weaponItems = append(weaponItems, itemDef)
		case item.TypeAuxiliary:
			auxItems = append(auxItems, itemDef)
		}
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
		bodySkinSets:   bodySkins,
		eyesSetOptions: eyeSets,
		hairSetOptions: hairSets,
		bodywearItems:  bodywearItems,
		headwearItems:  headwearItems,
		footwearItems:  footwearItems,
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

	g.setupAppearanceScreen()

	g.tabControl = tab.NewTabControl("ui/ui-components.tsj", []tab.Tab{
		{
			DisplayName: "Appearance",
			ImgTileID:   67,
			Active:      true,
		},
		{
			DisplayName: "Attributes",
			ImgTileID:   65,
		},
		{
			DisplayName: "Inventory",
			ImgTileID:   64,
		},
	})

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
	equipFeetSet := body.NewBodyPartSet(body.BodyPartSetParams{
		HasUp:       true,
		Name:        "equipFeetSet",
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

	entBody := body.NewEntityBodySet(bodySet, armsSet, legsSet, hairSet, eyesSet, equipHeadSet, equipFeetSet, equipBodySet, equipLegsSet, weaponSet, weaponFxSet, auxSet, nil, nil, nil)

	// Setting these various fields just to prevent validation errors (e.g. WalkSpeed). But, these values are eventually overwritten
	// when used in the actual game.
	cd := entity.CharacterData{
		ID:             "newCharacter",
		DisplayName:    "newCharacter",
		Body:           entBody,
		WalkSpeed:      entity.GetDefaultWalkSpeed(),
		RunSpeed:       entity.GetDefaultRunSpeed(),
		InventoryItems: make([]*item.InventoryItem, 1),
	}
	return cd
}

func (bg builderGame) saveCharacter() {
	id := bg.scrAppearance.idField.GetText()
	if id == "" {
		return
	}
	name := bg.scrAppearance.nameField.GetText()
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

func (bg *builderGame) Draw(screen *ebiten.Image) {
	bg.tabControl.Draw(screen, 50, 50, nil)
	if bg.tabControl.GetActiveTab().DisplayName == "Appearance" {
		bg.drawAppearancePage(screen)
	}
}

func (bg *builderGame) Update() error {
	if bg.popupMgr.IsPopupActive() {
		bg.popupMgr.Update()
		return nil
	}

	bg.tabControl.Update()

	if bg.tabControl.GetActiveTab().DisplayName == "Appearance" {
		bg.updateAppearanceScreen()
	}

	return nil
}

func (bg *builderGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return display.SCREEN_WIDTH, display.SCREEN_HEIGHT
}
