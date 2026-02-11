package defs

type (
	AttributeID string
	SkillID     string
	TraitID     string
)

type CurMax struct {
	CurrentVal int
	MaxVal     int
}

type Vitals struct {
	Health  CurMax
	Stamina CurMax
}

type AttributeDef struct {
	ID          AttributeID
	DisplayName string
}

type SkillDef struct {
	ID          SkillID
	DisplayName string
}

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
