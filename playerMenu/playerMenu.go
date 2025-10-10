package playermenu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/ui"
	"github.com/webbben/2d-game-engine/inventory"
	"github.com/webbben/2d-game-engine/player"
)

type PlayerMenu struct {
	init          bool
	x, y          int
	width, height int

	playerRef *player.Player

	ui.BoxDef
	BoxTilesetSource    string
	BoxOriginIndex      int // index of the top left tile of this box in the tileset
	BoxTitleOriginIndex int // index of the top left of the box title
	boxImage            *ebiten.Image
	boxTitle            ui.BoxTitle
	boxX, boxY          int // position of the entire box containing the page content

	pageTabs              ui.TabControl
	PageTabsTilesetSource string // tileset for the tab control ui component
	pageTabsX, pageTabsY  int

	InventoryPage InventoryPage

	mainContentBoxX, mainContentBoxY                int // start position for tab main content box
	mainContentBoxWidth, mainContentBoxHeight       int // size of the tab main content box
	mainContentActualWidth, mainContentActualHeight int // actual area main content tabs can take
}

func (pm *PlayerMenu) Load(playerRef *player.Player, defMgr *definitions.DefinitionManager, inventoryParams inventory.InventoryParams) {
	if pm.BoxTilesetSource == "" {
		panic("no box tileset source set")
	}
	if pm.PageTabsTilesetSource == "" {
		panic("no page tabs tileset source set")
	}
	if playerRef == nil {
		panic("player ref is nil")
	}
	pm.playerRef = playerRef

	pm.BoxDef = ui.NewBox(pm.BoxTilesetSource, pm.BoxOriginIndex)
	tileSize := pm.BoxDef.TileSize()

	// determine full size of player menu
	pm.width = display.SCREEN_WIDTH * 2 / 3
	pm.width -= pm.width % tileSize // round it to the size of the box tile
	pm.height = display.SCREEN_HEIGHT * 2 / 3
	pm.height -= pm.height % tileSize

	// center the page on the screen
	pm.x = (display.SCREEN_WIDTH - pm.width) / 2
	pm.y = (display.SCREEN_HEIGHT - pm.height) / 2

	// load menu tabs
	pm.pageTabs = ui.NewTabControl(pm.PageTabsTilesetSource, []ui.Tab{
		{
			ImgTileId:   64,
			DisplayName: "Inventory",
			Active:      true, // start with inventory as active
		},
		{
			ImgTileId:   65,
			DisplayName: "Levels",
		},
		{
			ImgTileId:   66,
			DisplayName: "Map",
		},
		{
			ImgTileId:   67,
			DisplayName: "Pantheon",
		},
		{
			ImgTileId:   68,
			DisplayName: "Quests",
		},
		{
			ImgTileId:   69,
			DisplayName: "Misc Stats",
		},
	})
	pm.pageTabs.Load()

	pm.pageTabsX = pm.x + tileSize
	pm.pageTabsY = pm.y

	// get size of main content box
	_, pageTabsHeight := pm.pageTabs.Dimensions()

	pm.boxX = pm.x
	pm.boxY = pm.y + pageTabsHeight

	pm.mainContentBoxX = pm.boxX + (tileSize / 2) // space the inner content from the outer box by a tile
	pm.mainContentBoxY = pm.boxY + (tileSize / 2)
	pm.mainContentBoxHeight = pm.height - (pm.mainContentBoxY - pm.y)
	pm.mainContentBoxWidth = pm.width

	// generate box image for main content area
	pm.boxImage = pm.BoxDef.BuildBoxImage(pm.mainContentBoxWidth, pm.mainContentBoxHeight)

	longestTitle := ""
	longestTitleWidth := 0
	for _, tab := range pm.pageTabs.Tabs {
		w, _, _ := text.GetStringSize(tab.DisplayName, config.DefaultTitleFont)
		if w > longestTitleWidth {
			longestTitle = tab.DisplayName
			longestTitleWidth = w
		}
	}

	pm.boxTitle = ui.NewBoxTitle(pm.BoxTilesetSource, pm.BoxTitleOriginIndex, longestTitle, nil)

	// load each page
	pm.mainContentActualWidth = pm.mainContentBoxWidth - (tileSize)
	pm.mainContentActualHeight = pm.mainContentBoxHeight - (tileSize)
	pm.InventoryPage.Load(pm.mainContentActualWidth, pm.mainContentActualHeight, pm.playerRef, defMgr, inventoryParams)

	pm.init = true
}

func (pm *PlayerMenu) Draw(screen *ebiten.Image, om *overlay.OverlayManager) {
	if !pm.init {
		panic("player menu drawing before being initialized")
	}
	tileSize := pm.BoxDef.TileSize()

	// menu box
	rendering.DrawImage(screen, pm.boxImage, float64(pm.boxX), float64(pm.boxY), 0)
	pm.boxTitle.Draw(screen, float64(pm.boxX+(pm.width/2)-pm.boxTitle.Width()+100), float64(pm.boxY-tileSize))
	// top level menu tabs
	pm.pageTabs.Draw(screen, float64(pm.pageTabsX), float64(pm.pageTabsY))

	// inventory page (for now, until we implement other tabs)
	pm.InventoryPage.Draw(screen, float64(pm.mainContentBoxX), float64(pm.mainContentBoxY), om)
}

func (pm *PlayerMenu) Update() {
	pm.pageTabs.Update()
	activeTab := pm.pageTabs.GetActiveTab()
	pm.boxTitle.SetTitle(activeTab.DisplayName)

	pm.InventoryPage.Update()
}
