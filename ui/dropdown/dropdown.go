// Package dropdown provides a dropdown option select UI component
package dropdown

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/ui/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/ui/box"
	"github.com/webbben/2d-game-engine/ui/button"
	"github.com/webbben/2d-game-engine/ui/popup"
	"github.com/webbben/2d-game-engine/ui/textfield"
	"golang.org/x/image/font"
)

// OptionSelect is a UI component that lets users choose from a list of options. It has a drop down window.
type OptionSelect struct {
	f             font.Face
	tiles         []*ebiten.Image
	barImg        *ebiten.Image
	dropDownBox   *ebiten.Image
	optionButtons []*button.Button
	totalWidth    int // used when making buttons. don't change this directly; if you need to, then just recreate the entire OptionSelect.

	drawX, drawY int
	textX, textY int

	topBarMouseBehavior mouse.MouseBehavior
	optionWindowOpen    bool

	dropDownWindow *DropDownWindow

	selectedOptionIndex int

	popupMgr *popup.Manager

	// properties for the input feature

	inputEnabled bool // if true, this option select takes input to filter options
	hideBarText  bool // if set, the bar will not write which option is currently selected. This is to allow the input field to show instead.

	masterOptionsList   []string
	inputField          textfield.TextField
	lastInput           string
	lastFilteredOptions []string
}

func (os OptionSelect) Dimensions() (barDx, barDy, windowDx, windowDy int) {
	if os.barImg == nil {
		panic("bar img is nil")
	}
	if os.dropDownBox == nil {
		panic("drop down box is nil")
	}
	barBounds := os.barImg.Bounds()
	windowBounds := os.dropDownBox.Bounds()
	return barBounds.Dx(), barBounds.Dy(), windowBounds.Dx(), windowBounds.Dy()
}

func (os OptionSelect) GetCurrentValue() string {
	return os.masterOptionsList[os.selectedOptionIndex]
}

func (os *OptionSelect) SetSelectedIndex(i int) {
	if i < 0 || i >= len(os.masterOptionsList) {
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
	InputEnabled          bool
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
		f:                 params.Font,
		popupMgr:          popupMgr,
		masterOptionsList: params.Options,
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

	os.totalWidth = totalWidth
	os.createOptionButtons(params.Options)

	// if input is enabled
	if params.InputEnabled {
		os.inputEnabled = true

		barDx, _, _, _ := os.Dimensions()

		os.inputField = *textfield.NewTextField(textfield.TextFieldParams{
			FontFace:           params.Font,
			TextColor:          color.Black,
			BorderColor:        color.Transparent,
			BgColor:            color.Transparent,
			AllowSpecial:       true,
			WidthPx:            barDx,
			MaxCharacterLength: 50,
		})
	}
	return os
}

// made this into a function since input dropdowns can reset the options list
func (os *OptionSelect) createOptionButtons(options []string) {
	tileSize := int(config.TileSize * config.UIScale)
	// make buttons for each option
	os.optionButtons = make([]*button.Button, 0)
	buttonWidth := os.totalWidth - (tileSize / 2)
	buttonHeight := tileSize - (tileSize / 8)
	for _, s := range options {
		os.optionButtons = append(os.optionButtons, button.NewButton(s, config.DefaultFont, buttonWidth, buttonHeight))
	}
}

func (os *OptionSelect) Update() {
	// detect click on main bar, which triggers the popup manager
	b := os.barImg.Bounds()
	os.topBarMouseBehavior.Update(os.drawX, os.drawY, b.Dx(), b.Dy(), false)
	if os.topBarMouseBehavior.LeftClick.ClickReleased {
		os.optionWindowOpen = !os.optionWindowOpen
		if os.optionWindowOpen {
			os.dropDownWindow = os.getDropDownWindow()
			os.popupMgr.SetPopup(os.dropDownWindow)
			os.inputField.SetText(os.GetCurrentValue())
			os.inputField.Focus()
		} else {
			os.dropDownWindow = nil
		}
	}
	os.hideBarText = os.inputEnabled && os.optionWindowOpen

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
	if os.hideBarText {
		return
	}
	dy, _ := text.GetRealisticFontMetrics(os.f)
	sY := int(y+(tileSize/2)) + (dy / 2)
	sX := int(x + (tileSize / 2))
	os.textX = sX
	os.textY = sY

	text.DrawShadowText(screen, os.GetCurrentValue(), os.f, sX, sY, nil, nil, 0, 0)
}

// DropDownWindow is a window that drops down and lets you select an option. This is the "popupable" implementation.
// Note: this is not the actual UI component! You should get an OptionSelect instead of directly getting this.
type DropDownWindow struct {
	optionSelectRef *OptionSelect

	x, y        int
	dropDownBox *ebiten.Image

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
	if ddw.optionSelectRef.HandleButtonClicks() {
		// button was clicked; new option was selected, so we can close the window
		ddw.closed = true
		return
	}

	// check for input
	if ddw.optionSelectRef.inputEnabled {
		ddw.optionSelectRef.inputField.Update()

		// check if input has changed and we need to filter options
		if ddw.optionSelectRef.inputField.GetText() != ddw.optionSelectRef.lastInput {
			ddw.optionSelectRef.lastInput = ddw.optionSelectRef.inputField.GetText()
			ddw.optionSelectRef.lastFilteredOptions = []string{}
			for _, s := range ddw.optionSelectRef.masterOptionsList {
				if strings.Contains(s, ddw.optionSelectRef.lastInput) {
					ddw.optionSelectRef.lastFilteredOptions = append(ddw.optionSelectRef.lastFilteredOptions, s)
				}
			}
			if len(ddw.optionSelectRef.lastFilteredOptions) != 0 {
				ddw.optionSelectRef.createOptionButtons(ddw.optionSelectRef.lastFilteredOptions)
			}
		}

		// don't close window if input is still focused and hasn't been clicked away yet
		if ddw.optionSelectRef.inputField.IsFocused() {
			return
		}
		// not focused - check if the input has a fully typed valid option, and if so set that as the current value
		for i, option := range ddw.optionSelectRef.masterOptionsList {
			if ddw.optionSelectRef.lastInput == option {
				// found an exact match
				ddw.optionSelectRef.SetSelectedIndex(i)
				break
			}
		}
		ddw.closed = true
	}

	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		// a click must've happened outside the the drop down window
		ddw.closed = true
	}
}

func (os *OptionSelect) HandleButtonClicks() (buttonWasClicked bool) {
	for _, optionButton := range os.optionButtons {
		if optionButton.Update().Clicked {
			for i, option := range os.masterOptionsList {
				if option == optionButton.ButtonText {
					os.SetSelectedIndex(i)
					return true
				}
			}
			logz.Panicln("OptionSelect", "option button clicked, but no match in the option list was found:", optionButton.ButtonText)
		}
	}
	return false
}

func (ddw DropDownWindow) Draw(screen *ebiten.Image) {
	tileSize := int(config.TileSize * config.UIScale)

	if ddw.dropDownBox == nil {
		panic("drop down box is nil")
	}
	if len(ddw.optionSelectRef.optionButtons) == 0 && !ddw.optionSelectRef.inputEnabled {
		panic("no option buttons")
	}

	rendering.DrawImage(screen, ddw.dropDownBox, float64(ddw.x), float64(ddw.y), 0)
	for i, optionButton := range ddw.optionSelectRef.optionButtons {
		if optionButton == nil {
			panic("option button is nil")
		}
		marginHeight := (tileSize / 8) / 2
		bY := ddw.y + (i * tileSize) + marginHeight
		optionButton.Draw(screen, int(ddw.x+(tileSize/4)), int(bY))
	}
	if ddw.optionSelectRef.inputEnabled {
		inputX := ddw.optionSelectRef.drawX + (tileSize / 2)
		inputY := ddw.optionSelectRef.drawY + 4
		ddw.optionSelectRef.inputField.Draw(screen, float64(inputX), float64(inputY))
	}
}

func (os *OptionSelect) getDropDownWindow() *DropDownWindow {
	tileSize := config.TileSize * config.UIScale

	ddw := DropDownWindow{
		optionSelectRef: os,
		x:               os.drawX,
		y:               os.drawY + int(tileSize),
		dropDownBox:     os.dropDownBox,
	}

	return &ddw
}
