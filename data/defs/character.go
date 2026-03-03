package defs

type (
	EntityDefID    string
	CharacterDefID string
	UniquePlayerID string
)

const (
	// This is used to identify the player's state and def. It is the only CharacterDefID that should be defined in the engine.
	PlayerID CharacterDefID = "player"
)

// CharacterDef represents a definition of a single character, and how that character starts off in the world when first introduced.
// Note: it is NOT Character State
type CharacterDef struct {
	// used for identifying the character within places like the DataManager.
	ID CharacterDefID
	// A unique ID for this player, that is unique across playthroughs. allows a specific player's save files to be correctly identified.
	// we can't use display names for this, since more than one character/playthrough might have the same display name.
	// This should not be defined for non-player characters.
	UniquePlayerID UniquePlayerID
	// if true, this character can only be instantiated once at any given time. this enables its characterState to be found by CharacterDefID,
	// since it means we can use CharacterDefID directly without needing to add random strings to it.
	// A non-unique character will have a CharacterStateID that uses CharacterDefID as its base, with randomized string attached at the end.
	Unique      bool
	BodyDef     BodyDef
	DisplayName string // REQ: the main name this character uses. Somewhat short so it can be used everywhere. Ex: "Scipio Africanus".
	FullName    string // OPT: a longer version of the name. Ex: Roman elites will have multiple names, like "Publius Cornelius Scipio Africanus".

	// REQ: the "class" of this character. Usually describes what type of person they are, their combat style, or whatever is most notable.
	// This is set by the classDef, but the classDef remains as is; this just allows a character to have a customized name, but still use a specific base classDef.
	ClassName        string
	ClassDefID       ClassDefID // the actual class def
	CultureID        CultureID
	InitialInventory StandardInventory

	DialogProfileID  DialogProfileID
	FootstepSFXDefID FootstepSFXDefID
	ScheduleID       ScheduleID

	BaseAttributes map[AttributeID]int // Base attribute levels (not including modifiers from traits, etc)
	BaseSkills     map[SkillID]int     // Base skill levels (not including modifiers from traits, etc)
	InitialTraits  []TraitID
	BaseVitals     Vitals
}
