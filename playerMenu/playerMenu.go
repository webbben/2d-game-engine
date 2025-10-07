package playermenu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/ui"
)

type PlayerMenu struct {
	x, y          int
	width, height int

	ui.BoxDef
	BoxTilesetSource string
	BoxID            string // ID of box used from box tileset (box_id=)
	boxImage         *ebiten.Image
	boxX, boxY       int // position of the entire box containing the page content

	pageTabs              ui.TabControl
	PageTabsTilesetSource string // tileset for the tab control ui component
	pageTabsX, pageTabsY  int

	InventoryPage InventoryPage

	mainContentX, mainContentY          int // start position for tab main content
	mainContentWidth, mainContentHeight int // size of a tab main content section
}

func (pm *PlayerMenu) Load() {
	if pm.BoxTilesetSource == "" {
		panic("no box tileset source set")
	}
	if pm.BoxID == "" {
		panic("no box ID set")
	}
	if pm.PageTabsTilesetSource == "" {
		panic("no page tabs tileset source set")
	}

	pm.BoxDef.LoadBoxTiles(pm.BoxTilesetSource, pm.BoxID)
	tileWidth, tileHeight := pm.BoxDef.TileDimensions()

	// determine full size of player menu
	pm.width = display.SCREEN_WIDTH * 2 / 3
	pm.width -= pm.width % tileWidth // round it to the size of the box tile
	pm.height = display.SCREEN_HEIGHT * 2 / 3
	pm.height -= pm.height % tileHeight

	// center the page on the screen
	pm.x = (display.SCREEN_WIDTH - pm.width) / 2
	pm.y = (display.SCREEN_HEIGHT - pm.height) / 2

	// load menu tabs
	pm.pageTabs = ui.NewTabControl(pm.PageTabsTilesetSource, []ui.Tab{
		{
			ImgTileId:   64,
			DisplayName: "Inventory",
		},
		{
			ImgTileId:   65,
			DisplayName: "Map",
		},
		{
			ImgTileId:   66,
			DisplayName: "Pantheon",
		},
		{
			ImgTileId:   67,
			DisplayName: "Statistics",
		},
	})
	pm.pageTabs.Load()

	pm.pageTabsX = pm.x + tileWidth
	pm.pageTabsY = pm.y

	// get size of main content box
	_, pageTabsHeight := pm.pageTabs.Dimensions()

	pm.boxX = pm.x
	pm.boxY = pm.y + pageTabsHeight

	pm.mainContentX = pm.boxX + (tileWidth / 2) // space the inner content from the outer box by a tile
	pm.mainContentY = pm.boxY + (tileWidth / 2)
	pm.mainContentHeight = pm.height - (pm.mainContentY - pm.y) - tileHeight
	pm.mainContentWidth = pm.width - tileWidth

	// generate box image for main content area
	pm.boxImage = pm.BoxDef.CreateBoxImage(pm.mainContentWidth, pm.mainContentHeight)

	// load each page
	pm.InventoryPage.Load()
}

func (pm *PlayerMenu) Draw(screen *ebiten.Image) {
	// menu box
	rendering.DrawImage(screen, pm.boxImage, float64(pm.boxX), float64(pm.boxY), 0)
	// top level menu tabs
	pm.pageTabs.Draw(screen, float64(pm.pageTabsX), float64(pm.pageTabsY))

	// inventory page (for now, until we implement other tabs)
	pm.InventoryPage.Draw(screen, float64(pm.mainContentX), float64(pm.mainContentY))
}

func (pm *PlayerMenu) Update() {
	pm.pageTabs.Update()
	pm.InventoryPage.Update()
}
