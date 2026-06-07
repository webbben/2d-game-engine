package dialogv2

import (
	"math/rand"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/quest"
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

type ConditionHasItem struct {
	ItemID defs.ItemID
}

func (c ConditionHasItem) IsMet(ctx defs.ConditionContext) bool {
	return ctx.PlayerHasItem(c.ItemID)
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

	// if neither of the following are set, this evaluates if social rank is exactly equal to Rank

	GEQ bool // "greater than or equal to"
	LEQ bool // "less than or equal to"
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
	stg, status := ctx.GetQuestStage(c.QuestID)
	if c.NotStarted {
		return status == quest.NotStarted
	}
	if c.Completed {
		return status == quest.Completed
	}
	if c.Failed {
		return status == quest.Failed
	}
	if c.StageID == "" {
		panic("stageID wasn't set, but neither were the other flags")
	}
	return stg.ID == c.StageID
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

type ConditionSkillLevel struct {
	SkillID defs.SkillID
	Level   int

	// if neither of the below are set, this evaluates that the skill is exactly equal to level

	GEQ bool // if set, evaluates if skill if greater than or equal to level
	LEQ bool // if set, evaluates if skill is less than or equal to level
}

func (c ConditionSkillLevel) IsMet(ctx defs.ConditionContext) bool {
	lvl := ctx.GetPlayerSkillLevel(c.SkillID)

	if c.GEQ {
		return lvl >= c.Level
	}
	if c.LEQ {
		return lvl <= c.Level
	}
	return lvl == c.Level
}

type ConditionAttributeLevel struct {
	AttrID defs.AttributeID
	Level  int

	// if neither of the below are set, this evaluates that the skill is exactly equal to level

	GEQ bool // if set, evaluates if skill if greater than or equal to level
	LEQ bool // if set, evaluates if skill is less than or equal to level
}

func (c ConditionAttributeLevel) IsMet(ctx defs.ConditionContext) bool {
	lvl := ctx.GetPlayerAttributeLevel(c.AttrID)

	if c.GEQ {
		return lvl >= c.Level
	}
	if c.LEQ {
		return lvl <= c.Level
	}
	return lvl == c.Level
}

// ConditionClass is for the class of the NPC - the class of the player cannot be considered in dialog conditions,
// since it is usually custom anyway.
type ConditionClass struct {
	ClassDefID defs.ClassDefID
}

func (c ConditionClass) IsMet(ctx defs.ConditionContext) bool {
	classDef := ctx.GetNPCClassDef()
	return classDef.ID == c.ClassDefID
}

type ConditionOpinion struct {
	Value int
	GEQ   bool
	LEQ   bool
}

func (c ConditionOpinion) IsMet(ctx defs.ConditionContext) bool {
	opinion := ctx.GetOpinionOfPlayer()

	if c.GEQ {
		return opinion >= c.Value
	}
	if c.LEQ {
		return opinion <= c.Value
	}

	return opinion == c.Value
}
