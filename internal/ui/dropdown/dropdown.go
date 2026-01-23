package dropdown

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui/box"
	"github.com/webbben/2d-game-engine/internal/ui/button"
	"github.com/webbben/2d-game-engine/internal/ui/popup"
	"golang.org/x/image/font"
)

type OptionSelect struct {
	f             font.Face
	tiles         []*ebiten.Image
	barImg        *ebiten.Image
	dropDownBox   *ebiten.Image
	optionButtons []*button.Button

	drawX, drawY int

	topBarMouseBehavior mouse.MouseBehavior
	optionWindowOpen    bool

	dropDownWindow *DropDownWindow

	selectedOptionIndex int

	popupMgr *popup.Manager
}

func (os OptionSelect) GetCurrentValue() string {
	return os.optionButtons[os.selectedOptionIndex].ButtonText
}

func (os *OptionSelect) SetSelectedIndex(i int) {
	if i < 0 || i >= len(os.optionButtons) {
		panic("index out of range of options list")
	}
	os.selectedOptionIndex = i
}

type OptionSelectParams struct {
	Font                  font.Face
	Options               []string
	InitialOptionIndex    int
	TilesetSrc            string
	OriginIndex           int
	DropDownBoxTilesetSrc string
	DropDownBoxOrigin     int
}

func NewOptionSelect(params OptionSelectParams, popupMgr *popup.Manager) OptionSelect {
	if len(params.Options) == 0 {
		panic("no options given")
	}
	if popupMgr == nil {
		panic("popup manager is nil")
	}
	if params.Font == nil {
		params.Font = config.DefaultFont
	}

	os := OptionSelect{
		f:        params.Font,
		popupMgr: popupMgr,
	}
	os.tiles = append(os.tiles, tiled.GetTileImage(params.TilesetSrc, params.OriginIndex, true))
	os.tiles = append(os.tiles, tiled.GetTileImage(params.TilesetSrc, params.OriginIndex+1, true))
	os.tiles = append(os.tiles, tiled.GetTileImage(params.TilesetSrc, params.OriginIndex+2, true))

	// get max width of text options
	maxWidth := 0
	for _, s := range params.Options {
		dx, _, _ := text.GetStringSize(s, params.Font)
		maxWidth = max(maxWidth, dx)
	}

	tileSize := int(config.TileSize * config.UIScale)

	maxWidth += tileSize / 2
	maxWidth -= maxWidth % tileSize
	totalWidth := maxWidth + (tileSize * 2)

	// draw bar image
	linearBox := box.NewLinearBox(box.LinearBoxParams{
		TilesetSrc:  params.TilesetSrc,
		OriginIndex: params.OriginIndex,
		TileWidth:   totalWidth / tileSize,
	})
	os.barImg = linearBox.Image()

	// drop down window for text options
	dropDownBox := box.NewBox(params.DropDownBoxTilesetSrc, params.DropDownBoxOrigin)
	os.dropDownBox = dropDownBox.BuildBoxImage(totalWidth, tileSize*len(params.Options))

	// make buttons for each option
	os.optionButtons = make([]*button.Button, 0)
	buttonWidth := totalWidth - (tileSize / 2)
	buttonHeight := tileSize - (tileSize / 8)
	for _, s := range params.Options {
		os.optionButtons = append(os.optionButtons, button.NewButton(s, config.DefaultFont, buttonWidth, buttonHeight))
	}

	return os
}

func (os *OptionSelect) Update() {
	// detect click on main bar
	b := os.barImg.Bounds()
	os.topBarMouseBehavior.Update(os.drawX, os.drawY, b.Dx(), b.Dy(), false)
	if os.topBarMouseBehavior.LeftClick.ClickReleased {
		os.optionWindowOpen = !os.optionWindowOpen
		if os.optionWindowOpen {
			os.dropDownWindow = os.getDropDownWindow()
			os.popupMgr.SetPopup(os.dropDownWindow)
		} else {
			os.dropDownWindow = nil
		}
	}

	// detect button clicks
	if os.optionWindowOpen {
		if os.dropDownWindow == nil {
			panic("drop down box is nil")
		}
		if os.dropDownWindow.IsClosed() {
			os.optionWindowOpen = false
		}
	}
}

func (os *OptionSelect) Draw(screen *ebiten.Image, x, y float64, om *overlay.OverlayManager) {
	os.drawX = int(x)
	os.drawY = int(y)

	tileSize := config.TileSize * config.UIScale
	rendering.DrawImage(screen, os.barImg, x, y, 0)

	// draw current value
	dy, _ := text.GetRealisticFontMetrics(os.f)
	sY := int(y+(tileSize/2)) + (dy / 2)
	sX := x + (tileSize / 2)
	text.DrawShadowText(screen, os.GetCurrentValue(), os.f, int(sX), sY, nil, nil, 0, 0)
}

// DropDownWindow is a window that drops down and lets you select an option
type DropDownWindow struct {
	optionSelectRef *OptionSelect

	x, y          int
	dropDownBox   *ebiten.Image
	optionButtons []*button.Button

	closed bool
}

func (ddw DropDownWindow) ZzCompileCheck() {
	_ = append([]popup.Popupable{}, &ddw)
}

func (ddw *DropDownWindow) Close() {
	ddw.closed = true
}

func (ddw DropDownWindow) IsClosed() bool {
	return ddw.closed
}

func (ddw *DropDownWindow) Update() {
	if ddw.closed {
		return
	}
	for i, optionButton := range ddw.optionButtons {
		if optionButton.Update().Clicked {
			ddw.optionSelectRef.SetSelectedIndex(i)
			ddw.closed = true
			break
		}
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		// a click must've happened outside the the drop down window
		ddw.closed = true
	}
}

func (ddw DropDownWindow) Draw(screen *ebiten.Image) {
	tileSize := int(config.TileSize * config.UIScale)

	rendering.DrawImage(screen, ddw.dropDownBox, float64(ddw.x), float64(ddw.y), 0)
	for i, optionButton := range ddw.optionButtons {
		marginHeight := (tileSize / 8) / 2
		bY := ddw.y + (i * tileSize) + marginHeight
		optionButton.Draw(screen, int(ddw.x+(tileSize/4)), int(bY))
	}
}

func (os *OptionSelect) getDropDownWindow() *DropDownWindow {
	tileSize := config.TileSize * config.UIScale

	ddw := DropDownWindow{
		optionSelectRef: os,
		x:               os.drawX,
		y:               os.drawY + int(tileSize),
		dropDownBox:     os.dropDownBox,
		optionButtons:   os.optionButtons,
	}

	return &ddw
}
