package defs

type (
	EntityDefID    string
	CharacterDefID string
)

// CharacterDef represents a definition of a single character, and how that character starts off in the world when first introduced.
// Note: it is NOT Character State
type CharacterDef struct {
	ID          CharacterDefID
	BodyDef     BodyDef
	DisplayName string // REQ: the main name this character uses. Somewhat short so it can be used everywhere. Ex: "Scipio Africanus".
	FullName    string // OPT: a longer version of the name. Ex: Roman elites will have multiple names, like "Publius Cornelius Scipio Africanus".

	// REQ: the "class" of this character. Usually describes what type of person they are, their combat style, or whatever is most notable.
	// This is set by the classDef, but the classDef remains as is; this just allows a character to have a customized name, but still use a specific base classDef.
	ClassName        string
	ClassDefID       ClassDefID // the actual class def
	InitialInventory StandardInventory

	DialogProfileID  DialogProfileID
	FootstepSFXDefID FootstepSFXDefID

	BaseAttributes map[AttributeID]int // Base attribute levels (not including modifiers from traits, etc)
	BaseSkills     map[SkillID]int     // Base skill levels (not including modifiers from traits, etc)
	InitialTraits  []TraitID
	BaseVitals     Vitals
}
