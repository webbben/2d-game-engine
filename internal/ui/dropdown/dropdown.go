package dropdown

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui/box"
	"github.com/webbben/2d-game-engine/internal/ui/button"
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
	//outsideClick        mouse.MouseBehavior

	selectedOptionIndex int
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

func NewOptionSelect(params OptionSelectParams) OptionSelect {
	if len(params.Options) == 0 {
		panic("no options given")
	}
	if params.Font == nil {
		params.Font = config.DefaultFont
	}

	os := OptionSelect{
		f: params.Font,
	}
	os.tiles = append(os.tiles, tiled.GetTileImage(params.TilesetSrc, params.OriginIndex))
	os.tiles = append(os.tiles, tiled.GetTileImage(params.TilesetSrc, params.OriginIndex+1))
	os.tiles = append(os.tiles, tiled.GetTileImage(params.TilesetSrc, params.OriginIndex+2))

	// get max width of text options
	maxWidth := 0
	for _, s := range params.Options {
		dx, _, _ := text.GetStringSize(s, params.Font)
		maxWidth = max(maxWidth, dx)
	}

	tileSize := int(config.TileSize * config.UIScale)

	maxWidth += tileSize / 2
	maxWidth -= maxWidth % tileSize

	// draw bar image
	totalWidth := maxWidth + (tileSize * 2)
	os.barImg = ebiten.NewImage(totalWidth, tileSize)
	rendering.DrawImage(os.barImg, os.tiles[0], 0, 0, config.UIScale)
	for i := range maxWidth / tileSize {
		rendering.DrawImage(os.barImg, os.tiles[1], float64(i+1*tileSize), 0, config.UIScale)
	}
	rendering.DrawImage(os.barImg, os.tiles[2], float64(maxWidth+tileSize), 0, config.UIScale)

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
	}

	// detect button clicks
	if os.optionWindowOpen {
		for i, optionButton := range os.optionButtons {
			if optionButton.Update().Clicked {
				os.SetSelectedIndex(i)
				os.optionWindowOpen = false
				break
			}
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

	if os.optionWindowOpen {
		rendering.DrawImage(screen, os.dropDownBox, x, y+tileSize, 0)
		for i, optionButton := range os.optionButtons {
			marginHeight := (tileSize / 8) / 2
			bY := int(y+tileSize) + (i * int(tileSize)) + int(marginHeight)
			optionButton.Draw(screen, int(x+(tileSize/4)), int(bY))
		}
	}
}
