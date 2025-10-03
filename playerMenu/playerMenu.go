package playermenu

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/ui"
)

type PlayerMenu struct {
	x, y          int
	width, height int

	ui.BoxDef
	boxTilesetSource string
	boxID            string // ID of box used from box tileset (box_id=)
	boxImage         *ebiten.Image

	pageTabs              ui.TabControl
	pageTabsTilesetSource string // tileset for the tab control ui component
	pageTabsX, pageTabsY  int

	mainContentX, mainContentY          int // start position for tab main content
	mainContentWidth, mainContentHeight int // size of a tab main content section
}

func (pm *PlayerMenu) Load() {
	if pm.boxTilesetSource == "" {
		panic("no box tileset source set")
	}
	if pm.boxID == "" {
		panic("no box ID set")
	}
	if pm.pageTabsTilesetSource == "" {
		panic("no page tabs tileset source set")
	}

	pm.x = 0
	pm.y = 0

	// determine full size of player menu
	pm.width = display.SCREEN_WIDTH
	pm.width -= pm.width % pm.TileWidth // round it to the size of the box tile
	pm.height = display.SCREEN_HEIGHT
	pm.height -= pm.height % pm.TileHeight

	// load menu tabs
	pm.pageTabs = ui.NewTabControl(pm.pageTabsTilesetSource, []ui.Tab{
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

	pm.pageTabsX = pm.x
	pm.pageTabsY = pm.y

	// get size of main content box
	pm.mainContentX = pm.x
	_, pageTabsHeight := pm.pageTabs.Dimensions()
	pm.mainContentY = pm.y + pageTabsHeight
	pm.mainContentHeight = pm.height - (pm.mainContentY - pm.y)
	pm.mainContentWidth = pm.width

	// generate box image for main content area
	pm.BoxDef.LoadBoxTiles(pm.boxTilesetSource, pm.boxID)
	pm.boxImage = pm.BoxDef.CreateBoxImage(pm.mainContentWidth, pm.mainContentHeight)
}
