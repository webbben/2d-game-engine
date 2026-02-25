package cmd

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/ui/text"
	"github.com/webbben/2d-game-engine/ui/button"
	"github.com/webbben/2d-game-engine/ui/dropdown"
	"github.com/webbben/2d-game-engine/ui/textfield"
)

type infoScreen struct {
	DisplayNameInput textfield.TextField
	FullNameInput    textfield.TextField
	CharacterIDInput textfield.TextField
	ClassNameInput   textfield.TextField

	DialogProfileSelector dropdown.OptionSelect
	ScheduleSelector      dropdown.OptionSelect

	SaveButton *button.Button
}

func (bg *builderGame) setupInfoScreen() {
	dialogProfileOptions := []string{}
	for _, op := range GetDialogProfiles() {
		dialogProfileOptions = append(dialogProfileOptions, string(op.ProfileID))
	}
	scheduleOptions := []string{}
	for _, sched := range GetAllSchedules() {
		scheduleOptions = append(scheduleOptions, string(sched.ID))
	}

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
		DialogProfileSelector: dropdown.NewOptionSelect(dropdown.OptionSelectParams{
			Font:                  config.DefaultFont,
			Options:               dialogProfileOptions,
			InitialOptionIndex:    0,
			TilesetSrc:            "ui/ui-components.tsj",
			OriginIndex:           288,
			DropDownBoxTilesetSrc: "boxes/boxes.tsj",
			DropDownBoxOrigin:     128,
		}, &bg.popupMgr),
		ScheduleSelector: dropdown.NewOptionSelect(dropdown.OptionSelectParams{
			Font:                  config.DefaultFont,
			Options:               scheduleOptions,
			InitialOptionIndex:    0,
			TilesetSrc:            "ui/ui-components.tsj",
			OriginIndex:           288,
			DropDownBoxTilesetSrc: "boxes/boxes.tsj",
			DropDownBoxOrigin:     128,
		}, &bg.popupMgr),

		SaveButton: button.NewLinearBoxButton("Save Character", "ui/ui-components.tsj", 352, config.DefaultTitleFont),
	}
	bg.scrInfo.ClassNameInput.SetText("(TODO)")
}

func (bg *builderGame) updateInfoScreen() {
	bg.scrInfo.DisplayNameInput.Update()
	bg.scrInfo.FullNameInput.Update()
	bg.scrInfo.CharacterIDInput.Update()
	bg.scrInfo.ClassNameInput.Update()
	bg.scrInfo.DialogProfileSelector.Update()
	bg.scrInfo.ScheduleSelector.Update()
	if bg.scrInfo.SaveButton.Update().Clicked {
		bg.saveCharacter()
	}
}

func (bg *builderGame) drawInfoScreen(screen *ebiten.Image) {
	tileSize := config.GetScaledTilesize()
	x := bg.windowX + int(tileSize)
	y := bg.windowY + int(tileSize) + 10

	topY := y

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

	y = topY
	x += int(tileSize) * 5

	text.DrawShadowText(screen, "Dialog Profile", config.DefaultTitleFont, x, y, nil, nil, 0, 0)
	y += 20
	bg.scrInfo.DialogProfileSelector.Draw(screen, float64(x), float64(y), bg.om)

	y = topY
	x += int(tileSize) * 7

	text.DrawShadowText(screen, "Schedule", config.DefaultTitleFont, x, y, nil, nil, 0, 0)
	y += 20
	bg.scrInfo.ScheduleSelector.Draw(screen, float64(x), float64(y), bg.om)

	dx, dy := bg.scrInfo.SaveButton.Width, bg.scrInfo.SaveButton.Height

	y = bg.windowY + bg.windowHeight - int(tileSize) - dy
	x = bg.windowX + bg.windowWidth - int(tileSize) - dx
	bg.scrInfo.SaveButton.Draw(screen, x, y)
}
