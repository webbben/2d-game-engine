package defs

import (
	"fmt"

	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/logz"
)

type (
	AttributeID string
	SkillID     string
	TraitID     string
	CultureID   string

	// BaseDamage represents the base damage value an item has, before any skills, protection or other factors have been processed
	BaseDamage float64
	// BaseProtection represents the base protection value an item has, before any skills or other factors have been processed
	BaseProtection float64
	// RealDamage represents derived damage after weapon/character factors are applied, but before armor mitigation.
	RealDamage float64
	// RealProtection represents derived protection after skills and condition are applied.
	RealProtection float64
	// FinalDamage represents the damage actually dealt to a target, after armor mitigation.
	FinalDamage float64
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

	CalculateMaxHealth  func(map[AttributeID]int) int
	CalculateMaxStamina func(map[AttributeID]int) int
}

// CombatSystemCalc includes all necessary functions to handle combat related calculation.
type CombatSystemCalc interface {
	// Functions for calculating damage. Includes:
	// 	- weaponID: for knowing which specific weapon was used (in case specific weapons have special cases)
	// 	- weaponType: in case specific weapon types behave differently in the calculation

	// Calculates how much damage a melee weapon does, before armor mitigation.
	MeleeWeaponDamage(weaponID ItemID, condition float64, mult float64, weaponType SkillID, attrs map[AttributeID]int, skills map[SkillID]int) RealDamage
	// Calculates how much damage a ranged weapon does, before armor mitigation.
	RangedWeaponDamage(weaponID ItemID, condition float64, mult float64, weaponType SkillID, attrs map[AttributeID]int, skills map[SkillID]int) RealDamage
	// Calculates how much protection a single piece of armor provides.
	ArmorProtection(armorID ItemID, condition float64, armorType SkillID, attrs map[AttributeID]int, skills map[SkillID]int) RealProtection
	// Performs the final calculation to determine the damage actually dealt to an entity. totalRealProtection is the sum
	// of all RealProtection values of all worn armor by the defender. The result is post-mitigation damage.
	CalculateFinalDamage(realDamage RealDamage, totalRealProtection RealProtection) FinalDamage

	// Calculates how much durability a weapon loses from landing a hit on a target.
	// targetRealProtection is the total real armor protection of the target that was hit.
	WeaponDurabilityLoss(targetRealProtection RealProtection) float64

	// Calculates whether a single piece of armor takes wear from an unblocked hit, and how much
	// durability it loses if it does. itemBaseProtection is the base (authored) protection of that
	// one armor item, and totalBaseProtection is the sum of base protections over all equipped armor.
	// attackRealDamage is the raw (pre-mitigation) damage of the attack.
	ArmorDurabilityLoss(itemBaseProtection, totalBaseProtection BaseProtection, attackRealDamage RealDamage) (tookWear bool, wearAmount float64)

	// Calculates how much durability a shield loses from a successful active block. attackRealDamage is
	// the raw (pre-mitigation) damage of the blocked attack.
	ShieldBlockDurabilityLoss(attackRealDamage RealDamage) float64
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
type Trait struct {
	ID               TraitID
	Name             string
	Description      string
	TilesetSrc       string
	TileID           int
	ConflictTraitIDs []TraitID // which traits this trait conflicts with (and thus cannot be had together by a single character)
	SkillChanges     map[SkillID]int
	AttributeChanges map[AttributeID]int

	// Opinion modifiers: how other characters' opinions are modified by a character that has this trait
	// (positive = other character likes trait-holder more; negative = other characters like trait-holder less)

	SameTraitOpinionModifier   OpinionModifier             // how opinion is modified for characters that share this same trait
	OtherTraitOpinionModifiers map[TraitID]OpinionModifier // how this trait modifies the opinion of characters with certain traits towards this trait-holder
}

func (t Trait) Validate() {
	if t.ID == "" {
		panic("id is empty")
	}
	if t.Name == "" {
		panic("name is empty")
	}
	if t.Description == "" {
		logz.Panicln(t.Name, "description is empty")
	}
	if t.TilesetSrc == "" {
		logz.Panicln(t.Name, "tilesetSrc is empty")
	}
	if t.TileID < 0 {
		logz.Panicln(t.Name, "tileID is negative...")
	}
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

	// baseline opinion modifiers towards other cultures (from this one)
	OtherCultureOpinions map[CultureID]OpinionModifier
}

type OpinionModifier struct {
	Mod    int
	Reason string
	Until  *clock.GameTime
}

func (om OpinionModifier) String() string {
	return fmt.Sprintf("%+d %s", om.Mod, om.Reason)
}
