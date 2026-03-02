package defs

type (
	AttributeID string
	SkillID     string
	TraitID     string
	CultureID   string
)

// LevelSystemParameters defines all the parameters that are used in calculating anything to do with our level system, such as:
//
// - How character levels are calculated from a character's skill levels.
//
// - Base attribute values, and how attributes grow alongside their governed skills.
//
// There are, of course, other things that can boost or decrease these things, but these are the basic parameters that are involved
// at the system level, and before anything else (culture buffs, etc) are factored in.
// Here, we are mainly concerned with deterministic parts of calculating levels and/or approximating level values for generating leveled NPCs.
type LevelSystemParameters struct {
	// Skill Parameters

	// "Skill Counts": how many total skills there are in each category. This is not something you would "toggle"; You just plug in the actual values for
	// how many major skills and minor skills there are, and the rest would be misc skills.
	MajorCount, MinorCount, MiscCount int

	// "Rates": an actual "expectation" of the rate at which skills of different categories would increase per level.
	// Important: This represents a change on *each skill* in a category - NOT a total number of skill increases.
	// So, if you set this to X, each skill in the category would be expected to increase by X.
	// To formulate in terms of "total skill increases in a category", calculate by dividing that number by the number of skills in the category.
	// We don't actually apply these rates directly - they are just used in calculating the K constant, which describes essentially a "standard XP per level".
	// So, these control "how much weighted gain equals one level".
	// Higher rates => slower leveling up, lower rates => faster leveling up.
	MajorRate, MinorRate, MiscRate float64

	// "Weights": defines the "weight" of a skill increase for a skill category. A higher weight means that skill increase gives more progress towards a level-up.
	// These don't need to add up to 1; they can be different values, like 1, 3, and 5. they just serve to give different amounts of influences to different skill categories.
	// "How fast different categories level you up" and "How strongly they influence attributes"
	MajorWeight, MinorWeight, MiscWeight float64

	// "Bases": the base skill level for a skill of this class (at character level 1)
	MajorBase, MinorBase, MiscBase int

	// Attribute Parameters

	// What the baseline is for all attributes (at level 1)
	AttributeBase int

	// A rate in which an attribute "grows" as one of its related skills increases. Should be a small value, since we don't want attributes to increase with its
	// skills on a 1-to-1 basis. Planned to be used in giving the player extra points they could assign to attributes on level-up (like how morrowind gives extra points).
	// If too low, attributes may lag behind skills. If too high, attributes may inflate too fast.
	// Note: a skill being "major" or "minor" has no impact on this attribute growth process.
	// Suggestion: 0.3 ~ 0.5
	AttributeGrowth float64

	// How much of a bonus is given to an attribute when it is designated as "favored" in a class.
	FavoredBonus int
}

type (
	SkillCategory string
	ClassDefID    string
)

type ClassDef struct {
	ID                ClassDefID
	Name              string
	SkillCategories   map[SkillID]SkillCategory
	FavoredAttributes []AttributeID
	AboutMe           string
}

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
	Description string
}

type SkillDef struct {
	ID                  SkillID
	DisplayName         string
	GoverningAttributes []AttributeID
	Description         string
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

type CultureDef struct {
	ID          CultureID
	DisplayName string
	Description string
	AttrMods    map[AttributeID]int
	SkillMods   map[SkillID]int
}
