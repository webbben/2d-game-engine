package skills

import (
	"github.com/webbben/2d-game-engine/internal/logz"
)

type TraitID string

// Trait : represents part of an entity's personality, background, or some other piece of information about it that has associated buffs and debuffs.
// The concept of "traits" is more or less from Crusader Kings 2, but we are running with the idea and expanding on it in some ways.
// But anyway, a trait can boost or decrease an entity's skills, change opinion/disposition of other entities, etc.
// Since we want traits to be very flexible, we will define it as an interface. That way, it can be defined to really do whatever you want.
type Trait interface {
	GetID() TraitID
	GetName() string
	GetDescription() string
	GetTilesetSrc() string
	GetTileID() int
	GetConflictTraitIDs() []string // traits that this trait conflicts with and cannot be applied together on a single entity (e.g. Greedy and Generous)
	GetSkillChanges() map[SkillID]int
	GetAttributeChanges() map[AttributeID]int
	GetOpinionChangeToTraitHolder(factors OpinionFactors) int // how this trait changes another character's opinion of the trait holder
	GetOpinionChangeToOther(factors OpinionFactors) int       // how this trait changes the holder's opinion of another character
}

// OpinionFactors are factors that can be considered when calculating opinion modifiers.
type OpinionFactors struct {
	TraitIDs   []TraitID
	CultureID  string
	Skills     map[SkillID]int
	Attributes map[AttributeID]int
}

type TraitParams struct {
	ID               TraitID
	Name             string
	Description      string
	TilesetSrc       string
	TileID           int
	ConflictTraitIDs []string
}

type traitBasic struct {
	TraitParams
	SkillChanges     map[SkillID]int
	AttributeChanges map[AttributeID]int
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

func (tb traitBasic) GetID() TraitID {
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

func (tb traitBasic) GetSkillChanges() map[SkillID]int {
	return tb.SkillChanges
}

func (tb traitBasic) GetAttributeChanges() map[AttributeID]int {
	return tb.AttributeChanges
}

// SkillTrait is a Trait that modifies skills or attributes.
type SkillTrait struct {
	traitBasic
}

func (st SkillTrait) GetOpinionChangeToTraitHolder(factors OpinionFactors) int {
	return 0
}

func (st SkillTrait) GetOpinionChangeToOther(factors OpinionFactors) int {
	return 0
}

func NewSkillTrait(params TraitParams, skillChanges map[SkillID]int, attrChanges map[AttributeID]int) SkillTrait {
	// interface compile check
	_ = append([]Trait{}, SkillTrait{})

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
