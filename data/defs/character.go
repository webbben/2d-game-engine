package defs

type (
	EntityDefID    string
	CharacterDefID string
)

// CharacterDef represents a definition of a single character, and how that character starts off in the world when first introduced.
// Note: it is NOT Character State
type CharacterDef struct {
	ID                   CharacterDefID
	BodyDef              BodyDef
	DefaultFootstepSFXID string // TODO: once we've centralized sound effects, it should be set here
	DisplayName          string
	InitialInventory     StandardInventory

	DialogProfileID DialogProfileID

	BaseAttributes map[AttributeID]int // Base attribute levels (not including modifiers from traits, etc)
	BaseSkills     map[SkillID]int     // Base skill levels (not including modifiers from traits, etc)
	InitialTraits  []TraitID
	BaseVitals     Vitals
}
