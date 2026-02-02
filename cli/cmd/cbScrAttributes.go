package cmd

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/text"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/internal/ui/textfield"
	"github.com/webbben/2d-game-engine/internal/ui/textwindow"
	"github.com/webbben/2d-game-engine/skills"
)

type attributesScreen struct {
	AttributeSetters    []attributeSetter
	CombatSkillSetters  []attributeSetter
	StealthSkillSetters []attributeSetter
	MagicSkillSetters   []attributeSetter

	traitsMasterList []skills.Trait

	entityTraits []traitIcon
	openTraits   []traitIcon
}

type attributeSetter struct {
	ID          string
	DisplayName string
	Input       textfield.TextField
	ModVal      int // if a modifier is active (from a trait), set it here
}

func newAttributeSetter(id, displayName string) attributeSetter {
	setter := attributeSetter{
		ID:          id,
		DisplayName: displayName,
		Input: *textfield.NewTextField(textfield.TextFieldParams{
			NumericOnly:        true,
			WidthPx:            50,
			FontFace:           config.DefaultFont,
			TextColor:          color.White,
			BorderColor:        color.White,
			BgColor:            color.Black,
			MaxCharacterLength: 3,
		}),
	}
	return setter
}

func (as *attributeSetter) Update() {
	as.Input.Update()
}

func (as *attributeSetter) Draw(screen *ebiten.Image, drawX, drawY float64, maxAttrWidth int) {
	title := as.DisplayName
	inputFieldDx, inputFieldDy := as.Input.Dimensions()
	_, ty := text.CenterTextInRect(title, config.DefaultFont, model.NewRect(drawX, drawY, 10, float64(inputFieldDy)))
	text.DrawShadowText(screen, title, config.DefaultFont, int(drawX), ty, nil, nil, 0, 0)
	as.Input.Draw(screen, drawX+float64(maxAttrWidth), drawY)

	// show the actual value, including modifiers
	if as.ModVal != 0 {
		text.DrawShadowText(screen, fmt.Sprintf("(%v)", as.ModVal+as.Input.GetNumber()), config.DefaultFont, int(drawX)+maxAttrWidth+inputFieldDx+10, ty, nil, nil, 0, 0)
	}
}

type traitIcon struct {
	trait                skills.Trait
	traitImg             *ebiten.Image
	hwContentPlaceholder *ebiten.Image
	hoverWindow          textwindow.CustomHoverWindow

	plusIcon, minusIcon *ebiten.Image

	lineWriter text.LineWriter

	mouseBehavior mouse.MouseBehavior

	x, y, w, h int
}

func newTraitIcon(traitDef skills.Trait, defMgr *definitions.DefinitionManager) traitIcon {
	ti := traitIcon{
		trait: traitDef,
	}

	// load change icon
	ti.plusIcon = tiled.GetTileImage("ui/ui-components.tsj", 195, true)
	ti.minusIcon = tiled.GetTileImage("ui/ui-components.tsj", 196, true)
	if ti.plusIcon == nil {
		panic("plus icon is nil")
	}
	if ti.minusIcon == nil {
		panic("minus icon is nil")
	}

	ti.traitImg = tiled.GetTileImage(traitDef.GetTilesetSrc(), traitDef.GetTileID(), true)
	bounds := ti.traitImg.Bounds()
	ti.w = int(float64(bounds.Dx()) * config.UIScale)
	ti.h = int(float64(bounds.Dy()) * config.UIScale)

	tileSize := config.GetScaledTilesize()

	bodyWidth := int(tileSize * 5)
	bodyHeight := int(tileSize * 7)

	// setup linewriter so it doesn't need updates (ready to write all text at once)
	ti.lineWriter = text.NewLineWriter(bodyWidth, bodyHeight, config.DefaultFont, nil, nil, true, true)
	ti.lineWriter.SetSourceText(ti.trait.GetDescription())
	ti.lineWriter.Update() // do one update, to cause the "write immediately" to take effect

	ti.hwContentPlaceholder = ebiten.NewImage(bodyWidth, bodyHeight)

	title := ti.trait.GetName()
	if title == "" {
		panic("no trait name found")
	}

	ti.hoverWindow = textwindow.NewCustomHoverWindow(title, config.DefaultFont, "boxes/boxes.tsj", 20)
	hoverBodyContent := ti.buildBodyContent(defMgr)
	ti.hoverWindow.SetCustomBodyContent(hoverBodyContent)

	return ti
}

func (ti *traitIcon) Update() bool {
	ti.mouseBehavior.Update(ti.x, ti.y, ti.w, ti.h, false)
	return ti.mouseBehavior.LeftClick.ClickReleased
}

func (ti *traitIcon) Draw(screen *ebiten.Image, x, y float64, om *overlay.OverlayManager) {
	ti.x = int(x)
	ti.y = int(y)

	if ti.mouseBehavior.IsHovering {
		ti.hoverWindow.Draw(om)
	}

	rendering.DrawImage(screen, ti.traitImg, x, y, config.UIScale)
}

func (ti traitIcon) buildBodyContent(defMgr *definitions.DefinitionManager) *ebiten.Image {
	ti.hwContentPlaceholder.Clear()

	belowDescY := ti.lineWriter.Draw(ti.hwContentPlaceholder, 0, 0)

	belowDescY += 10

	drawX := float64(0)
	drawY := float64(belowDescY)

	tileSize := config.GetScaledTilesize()

	drawChange := func(screen *ebiten.Image, change int, name string, nameColor color.Color, x, y int) {
		var icon *ebiten.Image
		if change > 0 {
			icon = ti.plusIcon
		} else {
			icon = ti.minusIcon
		}
		if icon == nil {
			panic("icon is nil")
		}

		num := fmt.Sprintf("%v", change)
		numX := int(tileSize*1.5) + x
		numX = int(text.CenterTextOnXPos(num, config.DefaultFont, float64(numX)))
		// if change < 0 {
		// 	// move back to ignore the negative sign and line up the numbers instead
		// 	dashDx, _, _ := text.GetStringSize("-", config.DefaultFont)
		// 	numX -= dashDx
		// }
		dispNameX := int(tileSize*2) + x
		rendering.DrawImage(screen, icon, drawX, drawY, config.UIScale)
		// get Y position for text to be centered
		_, ty := text.CenterTextInRect("ABC", config.DefaultFont, model.NewRect(float64(x), float64(y), 10, tileSize))
		text.DrawShadowText(screen, num, config.DefaultFont, numX, ty, nil, nil, 0, 0)
		text.DrawShadowText(screen, name, config.DefaultFont, dispNameX, ty, nameColor, nil, 0, 0)
	}

	// draw attribute changes
	attrChanges := ti.trait.GetAttributeChanges()
	for attrID, change := range attrChanges {
		// get attribute details
		attrDef := defMgr.GetAttributeDef(attrID)
		dispName := attrDef.DisplayName

		drawChange(ti.hwContentPlaceholder, change, dispName, color.RGBA{0, 0, 150, 255}, int(drawX), int(drawY))

		drawY += tileSize
	}

	skillChanges := ti.trait.GetSkillChanges()
	for skillID, change := range skillChanges {
		// get attribute details
		skillDef := defMgr.GetSkillDef(skillID)
		dispName := skillDef.DisplayName

		drawChange(ti.hwContentPlaceholder, change, dispName, color.Black, int(drawX), int(drawY))

		drawY += tileSize
	}

	return ti.hwContentPlaceholder
}

func (bg *builderGame) setupAttributesPage() {
	attributes := GetAllAttributes()
	for _, attr := range attributes {
		bg.defMgr.LoadAttributeDef(attr)
	}
	allSkills := GetAllSkills()
	for _, sk := range allSkills {
		bg.defMgr.LoadSkillDef(sk)
	}
	traits := GetAllTraits()
	for _, trait := range traits {
		bg.defMgr.LoadTraitDef(trait)
	}

	combatSkillIDs := []skills.SkillID{
		Blade, Blunt, Axe, Spear, Marksmanship, Repair, HeavyArmor, LightArmor,
	}
	stealthSkillIDs := []skills.SkillID{
		Security, Sneak, Speechcraft, Mercantile,
	}
	magicSkillIDs := []skills.SkillID{
		Alchemy, Incantation,
	}

	scr := attributesScreen{}

	for _, attr := range attributes {
		scr.AttributeSetters = append(scr.AttributeSetters, newAttributeSetter(string(attr.ID), attr.DisplayName))
	}
	for _, id := range combatSkillIDs {
		sk := bg.defMgr.GetSkillDef(id)
		scr.CombatSkillSetters = append(scr.CombatSkillSetters, newAttributeSetter(string(sk.ID), sk.DisplayName))
	}
	for _, id := range stealthSkillIDs {
		sk := bg.defMgr.GetSkillDef(id)
		scr.StealthSkillSetters = append(scr.StealthSkillSetters, newAttributeSetter(string(sk.ID), sk.DisplayName))
	}
	for _, id := range magicSkillIDs {
		sk := bg.defMgr.GetSkillDef(id)
		scr.MagicSkillSetters = append(scr.MagicSkillSetters, newAttributeSetter(string(sk.ID), sk.DisplayName))
	}

	scr.traitsMasterList = GetAllTraits()

	for _, t := range scr.traitsMasterList {
		scr.openTraits = append(scr.openTraits, newTraitIcon(t, bg.defMgr))
	}

	bg.scrAttributes = scr
}

func (bg *builderGame) updateAttributesPage() {
	for i := range bg.scrAttributes.AttributeSetters {
		bg.scrAttributes.AttributeSetters[i].Update()
	}
	for i := range bg.scrAttributes.CombatSkillSetters {
		bg.scrAttributes.CombatSkillSetters[i].Update()
	}
	for i := range bg.scrAttributes.StealthSkillSetters {
		bg.scrAttributes.StealthSkillSetters[i].Update()
	}
	for i := range bg.scrAttributes.MagicSkillSetters {
		bg.scrAttributes.MagicSkillSetters[i].Update()
	}

	// handle moving traits around when clicked
	traitChanged := false
	for i := range bg.scrAttributes.openTraits {
		if bg.scrAttributes.openTraits[i].Update() {
			// trait clicked; move it to character traits
			bg.scrAttributes.entityTraits = append(bg.scrAttributes.entityTraits, bg.scrAttributes.openTraits[i])
			bg.scrAttributes.openTraits = append(bg.scrAttributes.openTraits[:i], bg.scrAttributes.openTraits[i+1:]...)
			traitChanged = true
			break
		}
	}
	for i := range bg.scrAttributes.entityTraits {
		if bg.scrAttributes.entityTraits[i].Update() {
			// trait clicked; move it to open traits
			bg.scrAttributes.openTraits = append(bg.scrAttributes.openTraits, bg.scrAttributes.entityTraits[i])
			bg.scrAttributes.entityTraits = append(bg.scrAttributes.entityTraits[:i], bg.scrAttributes.entityTraits[i+1:]...)
			traitChanged = true
			break
		}
	}

	// save trait changes
	if traitChanged {
		entityTraits := []skills.Trait{}
		for _, trait := range bg.scrAttributes.entityTraits {
			entityTraits = append(entityTraits, trait.trait)
		}
		bg.characterData.Traits = entityTraits

		// reload attribute and skill setters
		bg.updateAttributeSelectors()
	}
}

func (bg *builderGame) updateAttributeSelectors() {
	skillMods, attrMods := bg.characterData.GetTraitModifiers()

	for i := range bg.scrAttributes.AttributeSetters {
		id := skills.AttributeID(bg.scrAttributes.AttributeSetters[i].ID)
		modVal := 0
		if val, exists := attrMods[id]; exists {
			modVal = val
		}
		bg.scrAttributes.AttributeSetters[i].ModVal = modVal
	}
	for i := range bg.scrAttributes.CombatSkillSetters {
		id := skills.SkillID(bg.scrAttributes.CombatSkillSetters[i].ID)
		modVal := 0
		if val, exists := skillMods[id]; exists {
			modVal = val
		}
		bg.scrAttributes.CombatSkillSetters[i].ModVal = modVal
	}
	for i := range bg.scrAttributes.StealthSkillSetters {
		id := skills.SkillID(bg.scrAttributes.StealthSkillSetters[i].ID)
		modVal := 0
		if val, exists := skillMods[id]; exists {
			modVal = val
		}
		bg.scrAttributes.StealthSkillSetters[i].ModVal = modVal
	}
	for i := range bg.scrAttributes.MagicSkillSetters {
		id := skills.SkillID(bg.scrAttributes.MagicSkillSetters[i].ID)
		modVal := 0
		if val, exists := skillMods[id]; exists {
			modVal = val
		}
		bg.scrAttributes.MagicSkillSetters[i].ModVal = modVal
	}
}

func (bg *builderGame) drawAttributesPage(screen *ebiten.Image) {
	tileSize := int(config.TileSize * config.UIScale)
	drawX := float64(bg.windowX + tileSize)
	topY := float64(bg.windowY+tileSize) + 20
	drawY := topY

	inputFieldDx, inputFieldDy := bg.scrAttributes.AttributeSetters[0].Input.Dimensions()
	skillRowMargin := 5

	// Attributes

	maxAttrWidth := 0
	for _, attrSetter := range bg.scrAttributes.AttributeSetters {
		dx, _, _ := text.GetStringSize(attrSetter.DisplayName, config.DefaultFont)
		if dx > maxAttrWidth {
			maxAttrWidth = dx
		}
	}
	maxAttrWidth += 10
	text.DrawShadowText(screen, "Attributes", config.DefaultTitleFont, int(drawX), int(drawY), nil, nil, 0, 0)
	drawY += 20
	for i := range bg.scrAttributes.AttributeSetters {
		bg.scrAttributes.AttributeSetters[i].Draw(screen, drawX, drawY, maxAttrWidth)
		drawY += float64(inputFieldDy + skillRowMargin)
	}
	drawX += float64(maxAttrWidth) + float64(inputFieldDx) + 70
	drawY = topY

	// Combat Skills

	maxAttrWidth = 0
	text.DrawShadowText(screen, "Combat", config.DefaultTitleFont, int(drawX), int(drawY), nil, nil, 0, 0)
	drawY += 20

	for _, attrSetter := range bg.scrAttributes.CombatSkillSetters {
		dx, _, _ := text.GetStringSize(attrSetter.DisplayName, config.DefaultFont)
		if dx > maxAttrWidth {
			maxAttrWidth = dx
		}
	}
	maxAttrWidth += 10
	for i := range bg.scrAttributes.CombatSkillSetters {
		bg.scrAttributes.CombatSkillSetters[i].Draw(screen, drawX, drawY, maxAttrWidth)
		drawY += float64(inputFieldDy + skillRowMargin)
	}
	drawX += float64(maxAttrWidth) + float64(inputFieldDx) + 70
	drawY = topY

	// Stealth Skills

	maxAttrWidth = 0
	text.DrawShadowText(screen, "Stealth", config.DefaultTitleFont, int(drawX), int(drawY), nil, nil, 0, 0)
	drawY += 20

	for _, attrSetter := range bg.scrAttributes.StealthSkillSetters {
		dx, _, _ := text.GetStringSize(attrSetter.DisplayName, config.DefaultFont)
		if dx > maxAttrWidth {
			maxAttrWidth = dx
		}
	}
	maxAttrWidth += 10
	for i := range bg.scrAttributes.StealthSkillSetters {
		bg.scrAttributes.StealthSkillSetters[i].Draw(screen, drawX, drawY, maxAttrWidth)
		drawY += float64(inputFieldDy + skillRowMargin)
	}
	drawX += float64(maxAttrWidth) + float64(inputFieldDx) + 70
	drawY = topY

	// Magic Skills

	maxAttrWidth = 0
	text.DrawShadowText(screen, "Magic", config.DefaultTitleFont, int(drawX), int(drawY), nil, nil, 0, 0)
	drawY += 20

	for _, attrSetter := range bg.scrAttributes.MagicSkillSetters {
		dx, _, _ := text.GetStringSize(attrSetter.DisplayName, config.DefaultFont)
		if dx > maxAttrWidth {
			maxAttrWidth = dx
		}
	}
	maxAttrWidth += 10
	for i := range bg.scrAttributes.MagicSkillSetters {
		title := bg.scrAttributes.MagicSkillSetters[i].DisplayName
		_, ty := text.CenterTextInRect(title, config.DefaultFont, model.NewRect(drawX, drawY, 10, float64(inputFieldDy)))
		text.DrawShadowText(screen, title, config.DefaultFont, int(drawX), ty, nil, nil, 0, 0)
		bg.scrAttributes.MagicSkillSetters[i].Input.Draw(screen, drawX+float64(maxAttrWidth), drawY)
		drawY += float64(inputFieldDy + skillRowMargin)
	}

	// Traits Section

	drawX = float64(bg.windowX + tileSize)
	midY := topY + float64((inputFieldDy*8)+(tileSize*2)) + 20
	drawY = midY

	text.DrawShadowText(screen, "Character Traits", config.DefaultTitleFont, int(drawX), int(drawY), nil, nil, 0, 0)
	drawY += 15

	for i := range bg.scrAttributes.entityTraits {
		bg.scrAttributes.entityTraits[i].Draw(screen, drawX, drawY, bg.om)
		drawX += float64(tileSize) + 20
	}

	drawY += float64(tileSize * 3)
	drawX = float64(bg.windowX + tileSize)
	text.DrawShadowText(screen, "Choose Traits", config.DefaultTitleFont, int(drawX), int(drawY), nil, nil, 0, 0)
	drawY += 15

	for i := range bg.scrAttributes.openTraits {
		bg.scrAttributes.openTraits[i].Draw(screen, drawX, drawY, bg.om)
		drawX += float64(tileSize) + 20
	}
}
