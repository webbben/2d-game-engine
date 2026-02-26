// Package tab defines a tab control UI component
package tab

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/mouse"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/ui/textwindow"
)

type Tab struct {
	DisplayName   string
	x, y          float64
	init          bool
	ImgTileID     int // tile ID of tile in the source tileset
	img           *ebiten.Image
	mouseBehavior mouse.MouseBehavior
	hoverTooltip  textwindow.HoverTooltip
	Active        bool // flag indicating if this tag is active or not
}

func (t Tab) Dimensions() (dx, dy int) {
	if t.img == nil {
		panic("image is nil")
	}
	return int(float64(t.img.Bounds().Dx()) * config.UIScale), int(float64(t.img.Bounds().Dy()) * config.UIScale)
}

type TabControl struct {
	x, y          float64
	SourceTileset string
	Tabs          []Tab
}

// Dimensions returns the dimensions of the entire tab control
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
	tabWidth, tabHeight := tc.Tabs[0].Dimensions()
	return tabWidth * len(tc.Tabs), tabHeight
}

func NewTabControl(sourceTileset string, tabs []Tab) TabControl {
	if config.DefaultTooltipBox.TilesetSrc == "" {
		panic("no default tooltip box tileset src set. currently it is required to be globally set for using tab controls.")
	}
	for i, tab := range tabs {
		tabs[i].hoverTooltip = textwindow.NewHoverTooltip(tab.DisplayName, config.DefaultTooltipBox.TilesetSrc, config.DefaultTooltipBox.OriginIndex, 1000, -10, -10)
	}

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
		img, err := ts.GetTileImage(tab.ImgTileID, true)
		if err != nil {
			logz.Panicf("error loading tile image: %s", err)
		}
		tc.Tabs[i].img = img
		tc.Tabs[i].init = true
	}
}

func (tc *TabControl) Update() {
	tabWidth, tabHeight := tc.Tabs[0].Dimensions()
	for i, tab := range tc.Tabs {
		tc.Tabs[i].mouseBehavior.Update(int(tab.x), int(tab.y), tabWidth, tabHeight, false)
		if tc.Tabs[i].mouseBehavior.LeftClick.ClickReleased {
			tc.ActivateTab(i)
		}
		tc.Tabs[i].hoverTooltip.Update(tab.x, tab.y, tabWidth, tabHeight)
	}
}

func (tc *TabControl) Draw(screen *ebiten.Image, drawX, drawY float64, om *overlay.OverlayManager) {
	if len(tc.Tabs) == 0 {
		panic("tried to draw a tab control with no tabs")
	}

	tc.x = drawX
	tc.y = drawY

	tabX, tabY := drawX, drawY
	tabWidth, _ := tc.Tabs[0].Dimensions()

	for i, tab := range tc.Tabs {
		tc.Tabs[i].x = tabX
		tc.Tabs[i].y = tabY
		if tab.Active {
			tc.Tabs[i].y += 4 * config.UIScale // lower the tab that is currently active
		}

		op := ebiten.DrawImageOptions{}
		if tab.mouseBehavior.IsHovering {
			op.ColorScale.Scale(1.1, 1.1, 1.1, 1)
		}
		rendering.DrawImageWithOps(screen, tab.img, tc.Tabs[i].x, tc.Tabs[i].y, config.UIScale, &op)

		if om != nil {
			tab.hoverTooltip.Draw(om)
		}

		tabX += float64(tabWidth)
	}
}

func (tc *TabControl) ActivateTab(tabIndex int) {
	if tc.Tabs[tabIndex].Active {
		return
	}
	for i, tab := range tc.Tabs {
		if tab.Active {
			tc.Tabs[i].Active = false
		}
		if i == tabIndex {
			tc.Tabs[i].Active = true
		}
	}
}

func (tc TabControl) GetActiveTab() Tab {
	for _, tab := range tc.Tabs {
		if tab.Active {
			return tab
		}
	}
	panic("no active tab found?")
}
