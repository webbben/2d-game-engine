package dialogv2

import "github.com/webbben/2d-game-engine/data/defs"

// ConditionCulture checks if a character has the specified culture
type ConditionCulture struct {
	CharDefID defs.CharacterDefID
	IsCulture defs.CultureID
}

func (c ConditionCulture) ZzInterfaceCheck() {
	_ = append([]defs.DialogCondition{}, c)
}

func (c ConditionCulture) IsMet(ctx defs.ConditionContext) bool {
	charDef := ctx.GetCharacterDef(c.CharDefID)
	return charDef.CultureID == c.IsCulture
}
