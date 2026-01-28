package cmd

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui/button"
	"github.com/webbben/2d-game-engine/internal/ui/dropdown"
	"github.com/webbben/2d-game-engine/internal/ui/popup"
	"github.com/webbben/2d-game-engine/internal/ui/slider"
	"github.com/webbben/2d-game-engine/internal/ui/stepper"
	"github.com/webbben/2d-game-engine/internal/ui/textfield"
	"github.com/webbben/2d-game-engine/item"
)

type appearanceScreen struct {
	bgImg             *ebiten.Image
	turnLeft          *button.Button
	turnRight         *button.Button
	speedSlider       slider.Slider
	scaleSlider       slider.Slider
	animationSelector dropdown.OptionSelect
	auxiliarySelector dropdown.OptionSelect
	bodywearSelector  dropdown.OptionSelect
	headwearSelector  dropdown.OptionSelect
	footwearSelector  dropdown.OptionSelect
	weaponSelector    dropdown.OptionSelect

	bodyCtl stepper.Stepper
	hairCtl stepper.Stepper
	eyesCtl stepper.Stepper

	bodyColorSliders slider.SliderGroup
	hairColorSliders slider.SliderGroup
	eyeColorSliders  slider.SliderGroup

	saveButton *button.Button

	nameField textfield.TextField
	idField   textfield.TextField
}

func (g *builderGame) setupAppearanceScreen() {
	bgTileset := "buildings/walls.tsj"

	// create the backdrop
	t := float64(config.TileSize)
	g.scrAppearance.bgImg = ebiten.NewImage(int(t*3), int(t*3))
	rendering.DrawImage(g.scrAppearance.bgImg, tiled.GetTileImage(bgTileset, 150, true), 0, 0, 0)
	rendering.DrawImage(g.scrAppearance.bgImg, tiled.GetTileImage(bgTileset, 151, true), t, 0, 0)
	rendering.DrawImage(g.scrAppearance.bgImg, tiled.GetTileImage(bgTileset, 152, true), t*2, 0, 0)

	rendering.DrawImage(g.scrAppearance.bgImg, tiled.GetTileImage(bgTileset, 182, true), 0, t, 0)
	rendering.DrawImage(g.scrAppearance.bgImg, tiled.GetTileImage(bgTileset, 183, true), t, t, 0)
	rendering.DrawImage(g.scrAppearance.bgImg, tiled.GetTileImage(bgTileset, 184, true), t*2, t, 0)

	rendering.DrawImage(g.scrAppearance.bgImg, tiled.GetTileImage(bgTileset, 214, true), 0, t*2, 0)
	rendering.DrawImage(g.scrAppearance.bgImg, tiled.GetTileImage(bgTileset, 215, true), t, t*2, 0)
	rendering.DrawImage(g.scrAppearance.bgImg, tiled.GetTileImage(bgTileset, 216, true), t*2, t*2, 0)

	g.characterData.Body.SetDirection('D')

	turnLeftImg := tiled.GetTileImage("ui/ui-components.tsj", 224, true)
	turnLeftImg = rendering.ScaleImage(turnLeftImg, config.UIScale, config.UIScale)
	turnRightImg := tiled.GetTileImage("ui/ui-components.tsj", 225, true)
	turnRightImg = rendering.ScaleImage(turnRightImg, config.UIScale, config.UIScale)
	g.scrAppearance.turnLeft = button.NewImageButton("", nil, turnLeftImg)
	g.scrAppearance.turnRight = button.NewImageButton("", nil, turnRightImg)

	g.scrAppearance.speedSlider = slider.NewSlider(slider.SliderParams{
		TilesetSrc:    "ui/ui-components.tsj",
		TilesetOrigin: 256,
		TileWidth:     4,
		MinVal:        5,
		MaxVal:        20,
		InitialValue:  8,
		StepSize:      1,
	})

	g.scrAppearance.scaleSlider = slider.NewSlider(slider.SliderParams{
		TilesetSrc:    "ui/ui-components.tsj",
		TilesetOrigin: 256,
		TileWidth:     4,
		MinVal:        3,
		MaxVal:        8,
		InitialValue:  8,
		StepSize:      1,
	})

	g.popupMgr = popup.NewPopupManager()

	g.scrAppearance.animationSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               []string{body.AnimIdle, body.AnimWalk, body.AnimRun, body.AnimSlash, body.AnimBackslash, body.AnimShield},
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)

	auxOptions := []string{noneOp}
	for _, auxItem := range g.auxItems {
		auxOptions = append(auxOptions, auxItem.GetID())
	}
	g.scrAppearance.auxiliarySelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               auxOptions,
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)

	headwearOptions := []string{noneOp}
	for _, i := range g.headwearItems {
		headwearOptions = append(headwearOptions, i.GetID())
	}
	g.scrAppearance.headwearSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               headwearOptions,
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)

	footwearOptions := []string{noneOp}
	for _, i := range g.footwearItems {
		footwearOptions = append(footwearOptions, i.GetID())
	}
	g.scrAppearance.footwearSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               footwearOptions,
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)

	bodywearOptions := []string{noneOp}
	for _, i := range g.bodywearItems {
		bodywearOptions = append(bodywearOptions, i.GetID())
	}
	g.scrAppearance.bodywearSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               bodywearOptions,
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)
	weaponOptions := []string{noneOp}
	for _, i := range g.weaponItems {
		weaponOptions = append(weaponOptions, i.GetID())
	}
	g.scrAppearance.weaponSelector = dropdown.NewOptionSelect(dropdown.OptionSelectParams{
		Font:                  config.DefaultFont,
		Options:               weaponOptions,
		InitialOptionIndex:    0,
		TilesetSrc:            "ui/ui-components.tsj",
		OriginIndex:           288,
		DropDownBoxTilesetSrc: "boxes/boxes.tsj",
		DropDownBoxOrigin:     128,
	}, &g.popupMgr)

	if len(g.bodySkinSets) > 1 {
		g.scrAppearance.bodyCtl = stepper.NewStepper(stepper.StepperParams{
			MinVal:               0,
			MaxVal:               len(g.bodySkinSets) - 1,
			Font:                 config.DefaultTitleFont,
			FontFg:               color.White,
			FontBg:               color.Black,
			DecrementButtonImage: turnLeftImg,
			IncrementButtonImage: turnRightImg,
		})
	}

	g.scrAppearance.hairCtl = stepper.NewStepper(stepper.StepperParams{
		MinVal:               0,
		MaxVal:               len(g.hairSetOptions) - 1,
		Font:                 config.DefaultTitleFont,
		FontFg:               color.White,
		FontBg:               color.Black,
		DecrementButtonImage: turnLeftImg,
		IncrementButtonImage: turnRightImg,
	})
	g.scrAppearance.eyesCtl = stepper.NewStepper(stepper.StepperParams{
		MinVal:               0,
		MaxVal:               len(g.eyesSetOptions) - 1,
		Font:                 config.DefaultTitleFont,
		FontFg:               color.White,
		FontBg:               color.Black,
		DecrementButtonImage: turnLeftImg,
		IncrementButtonImage: turnRightImg,
	})

	// sliders for adjusting colors
	sliderParams := slider.SliderParams{
		TilesetSrc:    "ui/ui-components.tsj",
		TilesetOrigin: 256,
		TileWidth:     4,
		MinVal:        0,
		MaxVal:        256,
		InitialValue:  128,
		StepSize:      1,
	}
	g.scrAppearance.bodyColorSliders = slider.NewSliderGroup(slider.SliderGroupParams{
		LabelFont:    config.DefaultFont,
		LabelColorFg: color.White,
		LabelColorBg: color.Black,
	}, []slider.SliderDef{
		{
			Label:  "H",
			Params: sliderParams,
		},
		{
			Label:  "S",
			Params: sliderParams,
		},
		{
			Label:  "V",
			Params: sliderParams,
		},
	})
	g.scrAppearance.hairColorSliders = slider.NewSliderGroup(slider.SliderGroupParams{
		LabelFont:    config.DefaultFont,
		LabelColorFg: color.White,
		LabelColorBg: color.Black,
	}, []slider.SliderDef{
		{
			Label:  "H",
			Params: sliderParams,
		},
		{
			Label:  "S",
			Params: sliderParams,
		},
		{
			Label:  "V",
			Params: sliderParams,
		},
	})
	g.scrAppearance.eyeColorSliders = slider.NewSliderGroup(slider.SliderGroupParams{
		LabelFont:    config.DefaultFont,
		LabelColorFg: color.White,
		LabelColorBg: color.Black,
	}, []slider.SliderDef{
		{
			Label:  "H",
			Params: sliderParams,
		},
		{
			Label:  "S",
			Params: sliderParams,
		},
		{
			Label:  "V",
			Params: sliderParams,
		},
	})

	g.scrAppearance.nameField = *textfield.NewTextField(textfield.TextFieldParams{
		FontFace:     config.DefaultFont,
		WidthPx:      200,
		AllowSpecial: true,
		TextColor:    color.White,
		BorderColor:  color.White,
		BgColor:      color.Black,
	})
	g.scrAppearance.idField = *textfield.NewTextField(textfield.TextFieldParams{
		FontFace:     config.DefaultFont,
		WidthPx:      200,
		AllowSpecial: true,
		TextColor:    color.White,
		BorderColor:  color.White,
		BgColor:      color.Black,
	})

	g.scrAppearance.saveButton = button.NewLinearBoxButton("Save", "ui/ui-components.tsj", 352, config.DefaultTitleFont)
}

func (bg *builderGame) updateAppearanceScreen() {
	if bg.scrAppearance.turnLeft.Update().Clicked {
		bg.characterData.Body.RotateLeft()
	} else if bg.scrAppearance.turnRight.Update().Clicked {
		bg.characterData.Body.RotateRight()
	}

	if len(bg.bodySkinSets) > 1 {
		bg.scrAppearance.bodyCtl.Update()
		if bg.scrAppearance.bodyCtl.GetValue() != bg.bodySetIndex {
			bg.SetBodyIndex(bg.scrAppearance.bodyCtl.GetValue())
		}
	}

	bg.scrAppearance.hairCtl.Update()
	if bg.scrAppearance.hairCtl.GetValue() != bg.hairSetIndex {
		bg.SetHairIndex(bg.scrAppearance.hairCtl.GetValue())
	}
	bg.scrAppearance.eyesCtl.Update()
	if bg.scrAppearance.eyesCtl.GetValue() != bg.eyesSetIndex {
		bg.SetEyesIndex(bg.scrAppearance.eyesCtl.GetValue())
	}

	bg.scrAppearance.bodyColorSliders.Update()
	bg.scrAppearance.hairColorSliders.Update()
	bg.scrAppearance.eyeColorSliders.Update()

	bg.characterData.Body.SetBodyHSV(
		float64(bg.scrAppearance.bodyColorSliders.GetValue("H"))/256,
		float64(bg.scrAppearance.bodyColorSliders.GetValue("S"))/256,
		float64(bg.scrAppearance.bodyColorSliders.GetValue("V"))/256,
	)
	bg.characterData.Body.SetHairHSV(
		float64(bg.scrAppearance.hairColorSliders.GetValue("H"))/256,
		float64(bg.scrAppearance.hairColorSliders.GetValue("S"))/256,
		float64(bg.scrAppearance.hairColorSliders.GetValue("V"))/256,
	)
	bg.characterData.Body.SetEyesHSV(
		float64(bg.scrAppearance.eyeColorSliders.GetValue("H"))/256,
		float64(bg.scrAppearance.eyeColorSliders.GetValue("S"))/256,
		float64(bg.scrAppearance.eyeColorSliders.GetValue("V"))/256,
	)

	bg.scrAppearance.speedSlider.Update()
	bg.scrAppearance.scaleSlider.Update()

	bg.scrAppearance.animationSelector.Update()
	selectorValue := bg.scrAppearance.animationSelector.GetCurrentValue()
	if selectorValue != bg.characterData.Body.GetCurrentAnimation() {
		bg.characterData.Body.SetAnimation(selectorValue, body.SetAnimationOps{Force: true})
	}
	bg.scrAppearance.auxiliarySelector.Update()
	selectorValue = bg.scrAppearance.auxiliarySelector.GetCurrentValue()
	if selectorValue != bg.equipedAux {
		bg.handleChangeAux(selectorValue)
	}
	bg.scrAppearance.headwearSelector.Update()
	selectorValue = bg.scrAppearance.headwearSelector.GetCurrentValue()
	if selectorValue != bg.equipedHeadwear {
		bg.handleChangeHeadwear(selectorValue)
	}
	bg.scrAppearance.footwearSelector.Update()
	selectorValue = bg.scrAppearance.footwearSelector.GetCurrentValue()
	if selectorValue != bg.equipedFootwear {
		bg.handleChangeFootwear(selectorValue)
	}
	bg.scrAppearance.bodywearSelector.Update()
	selectorValue = bg.scrAppearance.bodywearSelector.GetCurrentValue()
	if selectorValue != bg.equipedBodywear {
		bg.handleChangeBodywear(selectorValue)
	}
	bg.scrAppearance.weaponSelector.Update()
	selectorValue = bg.scrAppearance.weaponSelector.GetCurrentValue()
	if selectorValue != bg.equipedWeapon {
		bg.handleChangeWeapon(selectorValue)
	}

	bg.characterData.Body.SetAnimationTickCount(bg.scrAppearance.speedSlider.GetValue())

	bg.scrAppearance.nameField.Update()
	bg.scrAppearance.idField.Update()

	if bg.scrAppearance.saveButton.Update().Clicked {
		bg.saveCharacter()
	}

	bg.characterData.Body.Update()
}

func (bg *builderGame) drawAppearancePage(screen *ebiten.Image) {
	characterScale := float64(bg.scrAppearance.scaleSlider.GetValue())
	characterTileSize := config.TileSize * characterScale

	tileSize := int(config.TileSize * config.UIScale)

	bodyDx, bodyDy := bg.characterData.Body.Dimensions()
	bodyWidth := float64(bodyDx) * characterScale
	bodyHeight := float64(bodyDy) * characterScale

	bodyX := float64(display.SCREEN_WIDTH/2) - (bodyWidth / 2)
	bodyY := float64(display.SCREEN_HEIGHT/2) - (bodyHeight / 2) + 150

	// Backdrop
	rendering.DrawImage(screen, bg.scrAppearance.bgImg, bodyX-characterTileSize, bodyY-characterTileSize, characterScale)

	// Character body
	bg.characterData.Body.Draw(screen, bodyX, bodyY, characterScale)

	buttonsY := bodyY + (bodyHeight) + 20
	buttonLX := (display.SCREEN_WIDTH / 2) - bg.scrAppearance.turnLeft.Width - 20
	buttonRX := (display.SCREEN_WIDTH / 2) + 20
	bg.scrAppearance.turnLeft.Draw(screen, buttonLX, int(buttonsY))
	bg.scrAppearance.turnRight.Draw(screen, buttonRX, int(buttonsY))

	// UI controls - Left side
	sliderX := 50
	sliderY := 150
	text.DrawShadowText(screen, "Ticks Per Frame", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	text.DrawShadowText(screen, fmt.Sprintf("%v", bg.scrAppearance.speedSlider.GetValue()), config.DefaultFont, sliderX-30, sliderY+(tileSize*2/3), color.White, nil, 0, 0)
	bg.scrAppearance.speedSlider.Draw(screen, float64(sliderX), float64(sliderY))

	sliderY += tileSize * 3 / 2
	text.DrawShadowText(screen, "Scale", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	text.DrawShadowText(screen, fmt.Sprintf("%v", bg.scrAppearance.scaleSlider.GetValue()), config.DefaultFont, sliderX-30, sliderY+(tileSize*2/3), color.White, nil, 0, 0)
	bg.scrAppearance.scaleSlider.Draw(screen, float64(sliderX), float64(sliderY))

	sliderY += tileSize * 5 / 3
	text.DrawShadowText(screen, "Animation", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.scrAppearance.animationSelector.Draw(screen, float64(sliderX), float64(sliderY), nil)
	sliderY += tileSize * 5 / 3
	text.DrawShadowText(screen, "Auxiliary", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.scrAppearance.auxiliarySelector.Draw(screen, float64(sliderX), float64(sliderY), nil)

	sliderY += tileSize * 5 / 3
	text.DrawShadowText(screen, "Headwear", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.scrAppearance.headwearSelector.Draw(screen, float64(sliderX), float64(sliderY), nil)
	sliderY += tileSize * 5 / 3
	text.DrawShadowText(screen, "Bodywear", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.scrAppearance.bodywearSelector.Draw(screen, float64(sliderX), float64(sliderY), nil)
	sliderY += tileSize * 5 / 3
	text.DrawShadowText(screen, "Footwear", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.scrAppearance.footwearSelector.Draw(screen, float64(sliderX), float64(sliderY), nil)
	sliderY += tileSize * 5 / 3

	text.DrawShadowText(screen, "Weapon", config.DefaultTitleFont, sliderX, sliderY, color.White, nil, 0, 0)
	bg.scrAppearance.weaponSelector.Draw(screen, float64(sliderX), float64(sliderY), nil)

	// Character Info Input - Middle Top
	drawX := (display.SCREEN_WIDTH / 3) - 100
	drawY := 150
	text.DrawShadowText(screen, "Character Info", config.DefaultTitleFont, drawX, drawY, color.White, nil, 0, 0)
	drawY += 50
	text.DrawShadowText(screen, "Name", config.DefaultFont, drawX, drawY, color.White, nil, 0, 0)
	drawX += 100
	bg.scrAppearance.nameField.Draw(screen, float64(drawX), float64(drawY-25))
	drawX -= 100
	drawY += 50
	text.DrawShadowText(screen, "ID", config.DefaultFont, drawX, drawY, color.White, nil, 0, 0)
	drawX += 100
	bg.scrAppearance.idField.Draw(screen, float64(drawX), float64(drawY-25))
	drawX = (display.SCREEN_WIDTH / 3) + 350
	drawY = 250
	bg.scrAppearance.saveButton.Draw(screen, drawX, drawY)

	// UI controls - Right side
	ctlX := (display.SCREEN_WIDTH * 3 / 4) - 100
	ctlY := 150
	if len(bg.bodySkinSets) > 1 {
		text.DrawShadowText(screen, "Body", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
		bg.scrAppearance.bodyCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	}
	ctlX += 170
	text.DrawShadowText(screen, "Body Color", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	ctlY += tileSize / 8
	bg.scrAppearance.bodyColorSliders.Draw(screen, float64(ctlX), float64(ctlY))
	_, dy := bg.scrAppearance.bodyColorSliders.Dimensions()
	ctlY += dy + (tileSize)
	ctlX -= 170

	text.DrawShadowText(screen, "Hair", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.scrAppearance.hairCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlX += 170
	text.DrawShadowText(screen, "Hair Color", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	ctlY += tileSize / 8
	bg.scrAppearance.hairColorSliders.Draw(screen, float64(ctlX), float64(ctlY))
	_, dy = bg.scrAppearance.hairColorSliders.Dimensions()
	ctlY += dy + (tileSize)
	ctlX -= 170

	text.DrawShadowText(screen, "Eyes", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	bg.scrAppearance.eyesCtl.Draw(screen, float64(ctlX), float64(ctlY+20))
	ctlX += 170
	text.DrawShadowText(screen, "Eye Color", config.DefaultTitleFont, ctlX, ctlY, color.White, nil, 0, 0)
	ctlY += tileSize / 8
	bg.scrAppearance.eyeColorSliders.Draw(screen, float64(ctlX), float64(ctlY))

	bg.popupMgr.Draw(screen)
}

func (bg *builderGame) handleChangeAux(val string) {
	bg.equipedAux = val

	if val == noneOp {
		// catch the initial loading time call where the selector sees a mismatch and tries to set to None
		if bg.characterData.EquipedAuxiliary == nil {
			return
		}
		bg.characterData.UnequipAuxiliary()
		return
	}
	for _, auxItem := range bg.auxItems {
		if auxItem.GetID() == val {
			bg.characterData.EquipedAuxiliary = nil // unset the previous value so it's not added to inventory
			bg.characterData.EquipItem(item.NewInventoryItem(auxItem, 1))
			return
		}
	}
	panic("val doesn't seem to match an item ID: " + val)
}

func (bg *builderGame) handleChangeHeadwear(val string) {
	bg.equipedHeadwear = val

	if val == noneOp {
		if bg.characterData.EquipedHeadwear == nil {
			return
		}
		bg.characterData.UnequipHeadwear()
		return
	}
	for _, i := range bg.headwearItems {
		if i.GetID() == val {
			bg.characterData.EquipedHeadwear = nil
			bg.characterData.EquipItem(item.NewInventoryItem(i, 1))
			return
		}
	}
	panic("val doesn't seem to match an item ID: " + val)
}

func (bg *builderGame) handleChangeFootwear(val string) {
	bg.equipedFootwear = val

	if val == noneOp {
		if bg.characterData.EquipedFootwear == nil {
			return
		}
		bg.characterData.UnequipFootwear()
		return
	}
	for _, i := range bg.footwearItems {
		if i.GetID() == val {
			bg.characterData.EquipedFootwear = nil
			bg.characterData.EquipItem(item.NewInventoryItem(i, 1))
			return
		}
	}
	panic("val doesn't seem to match an item ID:" + val)
}

func (bg *builderGame) handleChangeBodywear(val string) {
	bg.equipedBodywear = val

	if val == noneOp {
		if bg.characterData.EquipedBodywear == nil {
			return
		}
		bg.characterData.UnequipBodywear()
		return
	}
	for _, i := range bg.bodywearItems {
		if i.GetID() == val {
			bg.characterData.EquipedBodywear = nil
			bg.characterData.EquipItem(item.NewInventoryItem(i, 1))
			return
		}
	}
	panic("val doesn't seem to match an item ID: " + val)
}

func (bg *builderGame) handleChangeWeapon(val string) {
	bg.equipedWeapon = val

	if val == noneOp {
		if bg.characterData.EquipedWeapon == nil {
			return
		}
		bg.characterData.UnequipWeapon()
		return
	}
	for _, i := range bg.weaponItems {
		if i.GetID() == val {
			bg.characterData.EquipItem(item.NewInventoryItem(i, 1))
			return
		}
	}
	panic("val doesn't seem to match an item ID: " + val)
}

// SETS: Set Index Functions

// DEPENDS ON:
//
// hairSet, equipHeadSet, equipBodySet
func (bg *builderGame) SetBodyIndex(i int) {
	if i < 0 || i >= len(bg.bodySkinSets) {
		panic("out of bounds")
	}
	bg.bodySetIndex = i
	skin := bg.bodySkinSets[i]
	bodyDef := skin.Body

	// arms are directly set with body
	armDef := skin.Arms

	legDef := skin.Legs

	bg.characterData.Body.SetBody(bodyDef, armDef, legDef)
}

func (bg *builderGame) SetEyesIndex(i int) {
	if i < 0 || i >= len(bg.eyesSetOptions) {
		panic("out of bounds")
	}
	bg.eyesSetIndex = i
	op := bg.eyesSetOptions[i]
	bg.characterData.Body.SetEyes(op)
}

func (bg *builderGame) SetHairIndex(i int) {
	if i < 0 || i >= len(bg.hairSetOptions) {
		panic("out of bounds")
	}
	bg.hairSetIndex = i
	op := bg.hairSetOptions[i]
	bg.characterData.Body.SetHair(op)
}
