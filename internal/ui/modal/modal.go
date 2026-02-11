// Package modal provides UI modals
package modal

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/ui/box"
	"github.com/webbben/2d-game-engine/internal/ui/button"
	"github.com/webbben/2d-game-engine/internal/ui/textfield"
	"golang.org/x/image/font"
)

type TextInputModal struct {
	titleFont, bodyFont font.Face
	windowBoxImg        *ebiten.Image
	numericOnly         bool

	done bool // if true, this modal is finished and no longer active

	// ui components

	textInput  textfield.TextField
	confirmBtn *button.Button

	// actual data

	titleText string
}

type TextInputModalParams struct {
	BoxTilesetSrc  string
	BoxOriginIndex int
	TitleFont      font.Face
	BodyFont       font.Face

	// display values

	TitleText         string
	ConfirmButtonText string // OPT: defaults to "confirm"

	// text field params

	AllowSpecial  bool
	NumericOnly   bool
	MaxCharLength int // defaults to 20
}

func (m TextInputModal) Dimensions() (dx, dy int) {
	if m.windowBoxImg == nil {
		panic("tried to get dimensions before input modal window image was made")
	}
	bounds := m.windowBoxImg.Bounds()
	return bounds.Dx(), bounds.Dy()
}

func NewTextInputModal(params TextInputModalParams) TextInputModal {
	if params.TitleFont == nil {
		params.TitleFont = config.DefaultTitleFont
	}
	if params.BodyFont == nil {
		params.BodyFont = config.DefaultFont
	}
	if params.TitleText == "" {
		params.TitleText = "Input Text"
	}
	if params.ConfirmButtonText == "" {
		params.ConfirmButtonText = "Confirm"
	}
	if params.MaxCharLength == 0 {
		params.MaxCharLength = 20
	}

	m := TextInputModal{
		titleFont:   params.TitleFont,
		bodyFont:    params.BodyFont,
		titleText:   params.TitleText,
		numericOnly: params.NumericOnly,
	}

	// ensure dimensions are by tileSize
	tileSize := int(config.GetScaledTilesize())
	minWidth := tileSize * 6
	height := tileSize * 6

	titleWidth, _, _ := text.GetStringSize(params.TitleText, params.TitleFont)
	titleWidth += tileSize * 2
	width := max(titleWidth, minWidth)

	b := box.NewBox(params.BoxTilesetSrc, params.BoxOriginIndex)
	m.windowBoxImg = b.BuildBoxImage(width, height)

	textFieldParams := textfield.TextFieldParams{
		WidthPx:            150,
		FontFace:           params.BodyFont,
		AllowSpecial:       params.AllowSpecial,
		NumericOnly:        params.NumericOnly,
		TextColor:          color.White,
		BgColor:            color.Black,
		BorderColor:        color.White,
		MaxCharacterLength: params.MaxCharLength,
	}

	m.textInput = *textfield.NewTextField(textFieldParams)

	m.confirmBtn = button.NewButton(params.ConfirmButtonText, m.bodyFont, 0, 0)

	return m
}

type TextModalResponse struct {
	Done        bool
	InputText   string
	InputNumber int
}

func (m *TextInputModal) Update() TextModalResponse {
	if m.done {
		return TextModalResponse{Done: true}
	}

	m.textInput.Update()
	if m.confirmBtn.Update().Clicked {
		if m.textInput.GetText() != "" {
			resp := TextModalResponse{
				Done: true,
			}
			if m.numericOnly {
				resp.InputNumber = m.textInput.GetNumber()
			} else {
				resp.InputText = m.textInput.GetText()
			}
			return resp
		}
	}

	return TextModalResponse{}
}

func (m *TextInputModal) Draw(screen *ebiten.Image, x, y float64) {
	if m.done {
		return
	}
	rendering.DrawImage(screen, m.windowBoxImg, x, y, 0)

	tileSize := config.GetScaledTilesize()

	titleX := x + tileSize
	titleY := y + tileSize

	text.DrawShadowText(screen, m.titleText, m.titleFont, int(titleX), int(titleY), nil, nil, 0, 0)

	inputX := titleX
	inputY := titleY + tileSize*2

	m.textInput.Draw(screen, inputX, inputY)
	textInputDx, _ := m.textInput.Dimensions()

	btnX := inputX + float64(textInputDx) + 20
	btnY := inputY

	m.confirmBtn.Draw(screen, int(btnX), int(btnY))
}
