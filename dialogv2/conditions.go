package dialogv2

import (
	"math/rand"

	"github.com/webbben/2d-game-engine/data/defs"
)

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

type ConditionHasGold struct {
	Amount int
}

func (c ConditionHasGold) IsMet(ctx defs.ConditionContext) bool {
	playerGold := ctx.GetPlayerGold()
	return playerGold >= c.Amount
}

type ConditionDialogProfile struct {
	ProfileID defs.DialogProfileID
}

func (c ConditionDialogProfile) IsMet(ctx defs.ConditionContext) bool {
	return c.ProfileID == ctx.GetNPCDialogProfileID()
}

// ConditionRand lets you set a random chance that this condition is true.
type ConditionRand struct {
	Percent float32 // should be in range (0, 1) - doesn't panic if it isn't though. returns true if rand.Float32 produces a value less than this.
}

func (c ConditionRand) IsMet(ctx defs.ConditionContext) bool {
	return rand.Float32() < c.Percent
}
