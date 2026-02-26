package skills

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

type TraitParams struct {
	ID               defs.TraitID
	Name             string
	Description      string
	TilesetSrc       string
	TileID           int
	ConflictTraitIDs []string
}

type traitBasic struct {
	TraitParams
	SkillChanges     map[defs.SkillID]int
	AttributeChanges map[defs.AttributeID]int
}

func (tb traitBasic) validate() {
	if tb.ID == "" {
		panic("id is empty")
	}
	if tb.Name == "" {
		panic("name is empty")
	}
	if tb.Description == "" {
		logz.Panicln(tb.Name, "description is empty")
	}
	if tb.TilesetSrc == "" {
		logz.Panicln(tb.Name, "tilesetSrc is empty")
	}
	if tb.TileID < 0 {
		logz.Panicln(tb.Name, "tileID is negative...")
	}
}

func (tb traitBasic) GetID() defs.TraitID {
	return tb.ID
}

func (tb traitBasic) GetName() string {
	return tb.Name
}

func (tb traitBasic) GetDescription() string {
	return tb.Description
}

func (tb traitBasic) GetTilesetSrc() string {
	return tb.TilesetSrc
}

func (tb traitBasic) GetTileID() int {
	return tb.TileID
}

func (tb traitBasic) GetConflictTraitIDs() []string {
	return tb.ConflictTraitIDs
}

func (tb traitBasic) GetSkillChanges() map[defs.SkillID]int {
	return tb.SkillChanges
}

func (tb traitBasic) GetAttributeChanges() map[defs.AttributeID]int {
	return tb.AttributeChanges
}

// SkillTrait is a Trait that modifies skills or attributes.
type SkillTrait struct {
	traitBasic
}

func (st SkillTrait) GetOpinionChangeToTraitHolder(factors defs.OpinionFactors) int {
	return 0
}

func (st SkillTrait) GetOpinionChangeToOther(factors defs.OpinionFactors) int {
	return 0
}

func NewSkillTrait(params TraitParams, skillChanges map[defs.SkillID]int, attrChanges map[defs.AttributeID]int) SkillTrait {
	// interface compile check
	_ = append([]defs.Trait{}, SkillTrait{})

	st := SkillTrait{
		traitBasic: traitBasic{
			TraitParams:      params,
			SkillChanges:     skillChanges,
			AttributeChanges: attrChanges,
		},
	}

	st.validate()

	return st
}
