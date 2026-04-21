package dialogv2

import (
	"math/rand"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
)

type ConditionDialogMemory struct {
	Key string
}

func (c ConditionDialogMemory) IsMet(ctx defs.ConditionContext) bool {
	return ctx.HasMemory(c.Key)
}

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

type ConditionMapID struct {
	MapID defs.MapID
}

func (c ConditionMapID) IsMet(ctx defs.ConditionContext) bool {
	mapDef := ctx.GetActiveMapDef()
	return c.MapID == mapDef.ID
}

type ConditionRegion struct {
	RegionID defs.RegionID
}

func (c ConditionRegion) IsMet(ctx defs.ConditionContext) bool {
	mapDef := ctx.GetActiveMapDef()
	return c.RegionID == mapDef.Region
}

type ConditionNOT struct {
	Arg defs.DialogCondition
}

func (c ConditionNOT) IsMet(ctx defs.ConditionContext) bool {
	return !c.Arg.IsMet(ctx)
}

// ConditionOR lets you do OR logic over a list of conditions. If any of them are true, then the whole condition is true.
type ConditionOR struct {
	Args []defs.DialogCondition
}

func (c ConditionOR) IsMet(ctx defs.ConditionContext) bool {
	for _, arg := range c.Args {
		if arg.IsMet(ctx) {
			return true
		}
	}
	return false
}

type ConditionSocialRank struct {
	Player bool            // if set, will judge player's rank rather than NPC's rank (default is NPC)
	Rank   defs.SocialRank // the rank level for this condition
	GEQ    bool            // "greater than or equal to"
	LEQ    bool            // "less than or equal to"
}

func (c ConditionSocialRank) IsMet(ctx defs.ConditionContext) bool {
	charStateID := ctx.GetNPCCharStateID()
	if c.Player {
		charStateID = id.CharacterStateID(defs.PlayerID)
	}
	rank := ctx.GetCharacterSocialRank(charStateID)

	if c.GEQ {
		return rank >= c.Rank
	}
	if c.LEQ {
		return rank <= c.Rank
	}
	return rank == c.Rank
}

type ConditionHasRole struct {
	Player bool // if set, will check player's roles instead of the NPC's (default is NPC).
	RoleID defs.RoleID
}

func (c ConditionHasRole) IsMet(ctx defs.ConditionContext) bool {
	charID := ctx.GetNPCCharStateID()
	if c.Player {
		charID = id.CharacterStateID(defs.PlayerID)
	}
	return ctx.CharacterHasRole(charID, c.RoleID)
}

// ConditionQuestStage checks if the given quest is at the given stage
type ConditionQuestStage struct {
	QuestID                       defs.QuestID
	NotStarted, Completed, Failed bool
	StageID                       defs.QuestStageID
}

func (c ConditionQuestStage) IsMet(ctx defs.ConditionContext) bool {
	started, comp, fail, sid := ctx.GetQuestStage(c.QuestID)
	if c.NotStarted {
		return !started
	}
	if c.Completed {
		return comp
	}
	if c.Failed {
		return fail
	}
	if sid == "" {
		panic("stageID wasn't set, but neither were the other flags")
	}
	return sid == c.StageID
}

type ConditionItemEquipped struct {
	ItemID defs.ItemID
}

func (c ConditionItemEquipped) IsMet(ctx defs.ConditionContext) bool {
	return ctx.IsItemEquipped(c.ItemID)
}

type ConditionKnowledge struct {
	TopicID defs.TopicID
}

func (c ConditionKnowledge) IsMet(ctx defs.ConditionContext) bool {
	return ctx.PlayerHasKnowledge(c.TopicID)
}
