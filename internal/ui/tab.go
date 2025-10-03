package ui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

type Tab struct {
	DisplayName   string
	x, y          float64
	init          bool
	ImgTileId     int // tile ID of tile in the source tileset
	img           *ebiten.Image
	mouseBehavior mouse.MouseBehavior
}

type TabControl struct {
	x, y          float64
	SourceTileset string
	Tabs          []Tab
}

func (tc TabControl) Dimensions() (dx, dy int) {
	if len(tc.Tabs) == 0 {
		panic("no tabs in tab control")
	}
	if !tc.Tabs[0].init {
		panic("tab not initialized")
	}
	if tc.Tabs[0].img == nil {
		panic("tab image is nil")
	}
	return tc.Tabs[0].img.Bounds().Dx(), tc.Tabs[0].img.Bounds().Dy()
}

func NewTabControl(sourceTileset string, tabs []Tab) TabControl {
	tabControl := TabControl{
		SourceTileset: sourceTileset,
		Tabs:          tabs,
	}

	tabControl.Load()

	return tabControl
}

func (tc *TabControl) Load() {
	if tc.SourceTileset == "" {
		panic("no source tileset")
	}

	ts, err := tiled.LoadTileset(tc.SourceTileset)
	if err != nil {
		logz.Panicf("error loading tileset for tab control: %s", err)
	}

	for i, tab := range tc.Tabs {
		img, err := ts.GetTileImage(tab.ImgTileId)
		if err != nil {
			logz.Panicf("error loading tile image: %s", err)
		}
		tc.Tabs[i].img = img
		tc.Tabs[i].init = true
	}
}

func (tc *TabControl) Update() {
	for i, tab := range tc.Tabs {
		tc.Tabs[i].mouseBehavior.Update(int(tab.x), int(tab.y), tab.img.Bounds().Dx(), tab.img.Bounds().Dy(), false)
	}
}
