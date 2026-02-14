package cmd

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/ui/button"
	"github.com/webbben/2d-game-engine/internal/ui/textfield"
)

type infoScreen struct {
	DisplayNameInput textfield.TextField
	FullNameInput    textfield.TextField
	CharacterIDInput textfield.TextField
	ClassNameInput   textfield.TextField

	SaveButton *button.Button
}

func (bg *builderGame) setupInfoScreen() {
	bg.scrInfo = infoScreen{
		DisplayNameInput: *textfield.NewTextField(textfield.TextFieldParams{
			WidthPx:            200,
			AllowSpecial:       false,
			FontFace:           config.DefaultFont,
			TextColor:          color.White,
			BorderColor:        color.White,
			BgColor:            color.Black,
			MaxCharacterLength: 20,
		}),
		FullNameInput: *textfield.NewTextField(textfield.TextFieldParams{
			WidthPx:            200,
			AllowSpecial:       false,
			FontFace:           config.DefaultFont,
			TextColor:          color.White,
			BorderColor:        color.White,
			BgColor:            color.Black,
			MaxCharacterLength: 30,
		}),
		CharacterIDInput: *textfield.NewTextField(textfield.TextFieldParams{
			WidthPx:            200,
			AllowSpecial:       true,
			FontFace:           config.DefaultFont,
			TextColor:          color.White,
			BorderColor:        color.White,
			BgColor:            color.Black,
			MaxCharacterLength: 30,
		}),
		// TODO: this doesn't do anything yet; need to make some system where user defined classes can exist. maybe as a "state"?
		// but, as of now, class defs wouldn't be able to be "overwritten" here.
		ClassNameInput: *textfield.NewTextField(textfield.TextFieldParams{
			WidthPx:            200,
			AllowSpecial:       false,
			FontFace:           config.DefaultFont,
			TextColor:          color.White,
			BorderColor:        color.White,
			BgColor:            color.Black,
			MaxCharacterLength: 20,
		}),
		SaveButton: button.NewLinearBoxButton("Save Character", "ui/ui-components.tsj", 352, config.DefaultTitleFont),
	}
	bg.scrInfo.ClassNameInput.SetText("(TODO)")
}

func (bg *builderGame) updateInfoScreen() {
	bg.scrInfo.DisplayNameInput.Update()
	bg.scrInfo.FullNameInput.Update()
	bg.scrInfo.CharacterIDInput.Update()
	bg.scrInfo.ClassNameInput.Update()
	if bg.scrInfo.SaveButton.Update().Clicked {
		bg.saveCharacter()
	}
}

func (bg *builderGame) drawInfoScreen(screen *ebiten.Image) {
	tileSize := config.GetScaledTilesize()
	x := bg.windowX + int(tileSize)
	y := bg.windowY + int(tileSize)

	text.DrawShadowText(screen, "Display Name", config.DefaultTitleFont, x, y, nil, nil, 0, 0)
	y += 20
	bg.scrInfo.DisplayNameInput.Draw(screen, float64(x), float64(y))

	y += int(tileSize) * 2
	text.DrawShadowText(screen, "Full Name", config.DefaultTitleFont, x, y, nil, nil, 0, 0)
	y += 20
	bg.scrInfo.FullNameInput.Draw(screen, float64(x), float64(y))

	y += int(tileSize) * 2
	text.DrawShadowText(screen, "Character ID", config.DefaultTitleFont, x, y, nil, nil, 0, 0)
	y += 20
	bg.scrInfo.CharacterIDInput.Draw(screen, float64(x), float64(y))

	y += int(tileSize) * 2
	text.DrawShadowText(screen, "Class Name", config.DefaultTitleFont, x, y, nil, nil, 0, 0)
	y += 20
	bg.scrInfo.ClassNameInput.Draw(screen, float64(x), float64(y))

	y += int(tileSize) * 2
	bg.scrInfo.SaveButton.Draw(screen, x, y)
}
