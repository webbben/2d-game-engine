package cmd

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/spf13/cobra"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/image"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/ui/box"
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
		config.DefaultUIBox = config.DefaultBox{
			TilesetSrc:  "boxes/boxes.tsj",
			OriginIndex: 16,
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
	eyesSetOptions, hairSetOptions                                               []defs.SelectedPartDef
	bodySetIndex, eyesSetIndex, hairSetIndex                                     int // index of body/skin option
	bodywearItems, headwearItems, footwearItems                                  []defs.ItemDef
	equipedBodywear, equipedHeadwear, equipedFootwear, equipedWeapon, equipedAux string // IDs of equiped items

	weaponItems []defs.ItemDef

	auxItems []defs.ItemDef

	// character info and entity body

	CharacterDef *defs.CharacterDef // The Character Definition that is produced when we save a new character

	// Just used for showing the actual character and its body - not saved
	// EntRef *entity.Entity
	Body body.EntityBodySet

	defMgr *definitions.DefinitionManager

	// UI components

	lastOpenedTab string // name (display ID) of the last opened tab. used for detecting tab changes.

	windowWidth, windowHeight int
	windowX, windowY          int

	windowBox *ebiten.Image
	titleBox  box.BoxTitle

	om *overlay.OverlayManager

	scrAppearance appearanceScreen
	scrAttributes attributesScreen
	scrInventory  inventoryScreen
	scrInfo       infoScreen

	popupMgr popup.Manager

	tabControl tab.TabControl
}

type weaponOption struct {
	weaponPartDef defs.SelectedPartDef
	weaponFxDef   defs.SelectedPartDef
}

type equipBodyOption struct {
	bodyDef defs.SelectedPartDef
	legsDef defs.SelectedPartDef
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

	bodywearItems := []defs.ItemDef{}
	headwearItems := []defs.ItemDef{}
	footwearItems := []defs.ItemDef{}
	weaponItems := []defs.ItemDef{}
	auxItems := []defs.ItemDef{}

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

	var characterDef defs.CharacterDef
	var charBody body.EntityBodySet
	if fileToLoad != "" {
		// TODO: load a character def from an ID
		// for now, "fileToLoad" is actually going to just be a characterDefID
	} else {
		characterDef, charBody = getNewCharacter()
	}

	g := builderGame{
		bodySkinSets:   bodySkins,
		eyesSetOptions: eyeSets,
		hairSetOptions: hairSets,
		bodywearItems:  bodywearItems,
		headwearItems:  headwearItems,
		footwearItems:  footwearItems,
		weaponItems:    weaponItems,
		auxItems:       auxItems,

		defMgr: defMgr,

		CharacterDef: &characterDef,
		Body:         charBody,
	}

	// do this if using a new, empty slate character body
	if fileToLoad == "" {
		// Set each bodyPartSet with their initial data.
		// We do this in a "weird" way here since this is the character builder screen.
		// In the actual game, we use the Load function instead, since all the PartSrc's are already set (from the JSON data).
		g.SetBodyIndex(0)
		g.SetHairIndex(0)
		g.SetEyesIndex(0)
		g.Body.AuxItemSet.Hide()
	}

	// run this just to confirm that the regular loading process also still works (as used in the actual game)
	// TODO: confirm this function does what we want
	g.Body.Load()

	g.om = &overlay.OverlayManager{}

	tileSize := int(config.TileSize * config.UIScale)

	width := display.SCREEN_WIDTH
	width -= width % tileSize // round it to the size of the box tile
	height := display.SCREEN_HEIGHT - (tileSize * 2)
	height -= height % tileSize

	g.windowWidth = width
	g.windowHeight = height
	g.windowX = (display.SCREEN_WIDTH - width) / 2
	g.windowY = tileSize * 2

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
		{
			DisplayName: "Info",
			ImgTileID:   68,
		},
	})

	b := box.NewBox(config.DefaultUIBox.TilesetSrc, config.DefaultUIBox.OriginIndex)
	g.windowBox = b.BuildBoxImage(g.windowWidth, g.windowHeight)

	longestTitle := text.GetLongestString([]string{"Appearance", "Attributes", "Inventory"}, config.DefaultTitleFont)
	g.titleBox = box.NewBoxTitle(config.DefaultUIBox.TilesetSrc, 111, longestTitle, config.DefaultTitleFont)

	g.setupAppearanceScreen()
	g.setupAttributesPage()
	g.setupInventoryPage()
	g.setupInfoScreen()

	if err := ebiten.RunGame(&g); err != nil {
		panic(err)
	}
}

func getNewCharacter() (defs.CharacterDef, body.EntityBodySet) {
	charBody := body.NewHumanBodyFramework()

	charDef := defs.CharacterDef{
		ID:               "newCharacter",
		DisplayName:      "New Character",
		DialogProfileID:  ProfileDefault,       // TODO: at some point, we will add a control for setting dialog profile
		FootstepSFXDefID: DefaultFootstepSFXID, // TODO: also for this
		BodyDef:          defs.BodyDef{
			// Note: not setting them here because the setters will handle putting the IDs here
		},
		InitialInventory: defs.StandardInventory{
			InventoryItems: make([]*defs.InventoryItem, 18),
			CoinPurse:      make([]*defs.InventoryItem, 6),
		},
		BaseAttributes: make(map[defs.AttributeID]int),
		BaseSkills:     make(map[defs.SkillID]int),
		InitialTraits:  make([]defs.TraitID, 0),
		BaseVitals:     defs.Vitals{}, // TODO: should vitals be calculated by attributes instead? I think yes.
	}

	return charDef, charBody
}

func (bg builderGame) saveCharacter() {
	bg.CharacterDef.ID = defs.CharacterDefID(bg.scrInfo.CharacterIDInput.GetText())
	bg.CharacterDef.DisplayName = bg.scrInfo.DisplayNameInput.GetText()
	bg.CharacterDef.FullName = bg.scrInfo.FullNameInput.GetText()
	bg.CharacterDef.ClassName = bg.scrInfo.ClassNameInput.GetText()

	// TODO: create UI component for showing error messages, info, etc.
	// I'm picturing a bubble or chip style info box the fades in and slowly slides up to the top of the screen on the right side.
	id := bg.CharacterDef.ID
	if id == "" {
		logz.Warnln("CharacterBuilder", "failed to save: id was empty")
		return
	}
	if bg.CharacterDef.DisplayName == "" {
		logz.Warnln("CharacterBuilder", "failed to save: display name was empty")
		return
	}
	if bg.CharacterDef.ClassName == "" {
		logz.Warnln("CharacterBuilder", "failed to save: class name was empty")
		return
	}

	bg.CharacterDef.BodyDef.BodyHSV = bg.Body.BodyHSV
	bg.CharacterDef.BodyDef.EyesHSV = bg.Body.EyesHSV
	bg.CharacterDef.BodyDef.HairHSV = bg.Body.HairHSV

	// do some quick validations to ensure data isn't unexpectedly missing
	if bg.CharacterDef.BodyDef.BodyID == "" {
		panic("body ID empty")
	}
	if bg.CharacterDef.BodyDef.ArmsID == "" {
		panic("arms ID empty")
	}
	if bg.CharacterDef.BodyDef.LegsID == "" {
		panic("legs ID empty")
	}
	if bg.CharacterDef.BodyDef.EyesID == "" {
		panic("eyes ID empty")
	}
	if bg.CharacterDef.BodyDef.HairID == "" {
		panic("hair ID empty")
	}

	jsonPath := resolveCharacterJSONPath(string(id))
	err := general_util.WriteToJSON(bg.CharacterDef, jsonPath)
	if err != nil {
		logz.Panicln("CharacterBuilder", "failed to save characterDef to JSON:", err)
	}
}

func resolveCharacterJSONPath(id string) string {
	return fmt.Sprintf("/Users/benwebb/dev/personal/ancient-rome/src/data/characters/json/%s.json", id)
}

func (bg *builderGame) Draw(screen *ebiten.Image) {
	tileSize := config.TileSize * config.UIScale
	rendering.DrawImage(screen, bg.windowBox, float64(bg.windowX), float64(bg.windowY), 0)
	titleX := bg.windowX + (bg.windowWidth / 2) - (bg.titleBox.Width() / 2)
	titleY := bg.windowY - int(tileSize)
	bg.titleBox.Draw(screen, float64(titleX), float64(titleY))

	bg.tabControl.Draw(screen, float64(bg.windowX+int(tileSize)), float64(bg.windowY-int(tileSize)), bg.om)

	switch bg.tabControl.GetActiveTab().DisplayName {
	case "Appearance":
		bg.drawAppearancePage(screen)
	case "Attributes":
		bg.drawAttributesPage(screen)
	case "Inventory":
		bg.drawInventoryPage(screen, bg.om)
	case "Info":
		bg.drawInfoScreen(screen)
	}

	bg.popupMgr.Draw(screen)

	bg.om.Draw(screen)
}

func (bg *builderGame) Update() error {
	if bg.popupMgr.IsPopupActive() {
		bg.popupMgr.Update()
		return nil
	}

	bg.tabControl.Update()

	currentTab := bg.tabControl.GetActiveTab().DisplayName

	bg.titleBox.SetTitle(currentTab)

	if currentTab != "Inventory" && bg.lastOpenedTab == "Inventory" {
		// switching away from inventory; save any changes to inventory items
		bg.saveInventory()
	}

	switch currentTab {
	case "Appearance":
		bg.updateAppearanceScreen()
	case "Attributes":
		bg.updateAttributesPage()
	case "Inventory":
		if bg.lastOpenedTab != "Inventory" {
			bg.refreshInventory() // if changing to inventory page, refresh items
		}
		bg.updateInventoryPage()
	case "Info":
		bg.updateInfoScreen()
	}

	bg.lastOpenedTab = currentTab

	return nil
}

func (bg *builderGame) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return display.SCREEN_WIDTH, display.SCREEN_HEIGHT
}
