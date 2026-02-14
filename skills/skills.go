// Package skills defines skills and attributes structs and concepts
package skills

import (
	"math"
	"math/rand"
	"slices"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/definitions"
)

const (
	SkillCategoryMisc  defs.SkillCategory = "misc"
	SkillCategoryMinor defs.SkillCategory = "minor"
	SkillCategoryMajor defs.SkillCategory = "major"
)

// CalculateK calculates a "K Constant" for the level calculation formula.
// It is derived from the rate in which skill levels increase, and how much their increase influence character level progress gain.
// "If skills grow at the expected rates, how much weighted progress equals one level?"
func CalculateK(
	majorCount int,
	minorCount int,
	miscCount int,
	majorRate float64,
	minorRate float64,
	miscRate float64,
	majorWeight float64,
	minorWeight float64,
	miscWeight float64,
) float64 {
	majorGain := float64(majorCount) * majorRate * majorWeight
	minorGain := float64(minorCount) * minorRate * minorWeight
	miscGain := float64(miscCount) * miscRate * miscWeight

	return majorGain + minorGain + miscGain
}

func calculateBaseSkillTotal(
	majorCount, minorCount, miscCount int,
	majorBase, minorBase, miscBase int,
	majorWeight, minorWeight, miscWeight float64,
) float64 {
	major := float64(majorCount*majorBase) * majorWeight
	minor := float64(minorCount*minorBase) * minorWeight
	misc := float64(miscCount*miscBase) * miscWeight
	return major + minor + misc
}

// CalculateLevelFromSkills calculates a character's level based on the levels of its skills.
//
// skillCategories: map of skillID to their category ("is this skill major, minor, or misc?")
//
// weights: how much weight to give a skill of each category, when calculating the weighted total.
func CalculateLevelFromSkills(
	skillLevels map[defs.SkillID]int,
	skillCategories map[defs.SkillID]defs.SkillCategory,
	lvlParams defs.LevelSystemParameters,
) int {
	var weightedTotal float64

	baseSkillTotal := calculateBaseSkillTotal(
		lvlParams.MajorCount, lvlParams.MinorCount, lvlParams.MiscCount,
		lvlParams.MajorBase, lvlParams.MinorBase, lvlParams.MiscBase,
		lvlParams.MajorWeight, lvlParams.MinorWeight, lvlParams.MiscWeight,
	)

	k := CalculateK(
		lvlParams.MajorCount, lvlParams.MinorCount, lvlParams.MiscCount,
		lvlParams.MajorRate, lvlParams.MinorRate, lvlParams.MiscRate,
		lvlParams.MajorWeight, lvlParams.MinorWeight, lvlParams.MiscWeight,
	)

	for skillID, val := range skillLevels {
		category := skillCategories[skillID]
		switch category {
		case SkillCategoryMajor:
			weightedTotal += float64(val) * lvlParams.MajorWeight
		case SkillCategoryMinor:
			weightedTotal += float64(val) * lvlParams.MinorWeight
		case SkillCategoryMisc:
			weightedTotal += float64(val) * lvlParams.MiscWeight
		}
	}

	level := 1 + math.Floor((weightedTotal-float64(baseSkillTotal))/k)

	return int(level)
}

// GenerateSkillsAndAttributes calculates skill and attribute levels for a given character level and class.
// See the LvlParams for details on what the actual leveling system parameters do.
//
// levelAttributeRate: say, 0.5 ~ 1? Not sure yet.
//
// randomnessSkill: ~ 0.1 (+/- 10%)
//
// randomnessAttr: ~ 0.05 (+/- 5%)
func GenerateSkillsAndAttributes(
	level int,
	class defs.ClassDef,
	lvlParams defs.LevelSystemParameters,
	defMgr *definitions.DefinitionManager,

	// randomness params
	randomnessSkill float64,
	randomnessAttr float64,
	rng *rand.Rand,

	// attribute params
	levelAttributeRate float64,
) (map[defs.SkillID]int, map[defs.AttributeID]int) {
	skills := generateSkillsForLevel(
		level, class, lvlParams, randomnessSkill, rng,
	)

	attributes := generateAttributesFromSkills(
		level, skills, class, lvlParams,
		levelAttributeRate, randomnessAttr, rng, defMgr,
	)

	return skills, attributes
}

// GenerateSkillsForLevel calculates skills for a given level and class.
func generateSkillsForLevel(
	level int,
	class defs.ClassDef,
	lvlParams defs.LevelSystemParameters,
	randomness float64,
	rng *rand.Rand,
) map[defs.SkillID]int {
	k := CalculateK(
		lvlParams.MajorCount, lvlParams.MinorCount, lvlParams.MiscCount,
		lvlParams.MajorRate, lvlParams.MinorRate, lvlParams.MiscRate,
		lvlParams.MajorWeight, lvlParams.MinorWeight, lvlParams.MiscWeight,
	)

	growthBudget := float64(level-1) * k

	// assign base weight per skill
	effectiveWeights := make(map[defs.SkillID]float64)
	var weightSum float64

	result := make(map[defs.SkillID]int)

	for skillID, category := range class.SkillCategories {
		var baseWeight float64
		var baseLevel int

		switch category {
		case SkillCategoryMajor:
			baseWeight = lvlParams.MajorWeight
			baseLevel = lvlParams.MajorBase
		case SkillCategoryMinor:
			baseWeight = lvlParams.MinorWeight
			baseLevel = lvlParams.MinorBase
		case SkillCategoryMisc:
			baseWeight = lvlParams.MiscWeight
			baseLevel = lvlParams.MiscBase
		}

		if baseLevel == 0 {
			panic("base level is 0?")
		}

		result[skillID] = baseLevel

		// apply randomness
		randomFactor := 1 + (rng.Float64()*2-1)*randomness
		effectiveWeight := baseWeight * randomFactor

		effectiveWeights[skillID] = effectiveWeight
		weightSum += effectiveWeight
	}

	for skillID := range class.SkillCategories {
		proportion := effectiveWeights[skillID] / weightSum
		skillGrowth := min(proportion*growthBudget, 100)
		// TODO: if skill level goes over 100 and gets capped, that means "budget" is being wasted.
		// This will cause "diminishing returns" where NPCs of exceptionally high level will not actually have all the budget
		// applied, and therefore won't really calculate to the level we are setting them to be. Not a problem for players, just for
		// NPC creation with level-based skill generation.
		result[skillID] += int(skillGrowth)
	}

	return result
}

// GenerateAttributesFromSkills generates attribute levels based on a given map of skill levels.
// Tries to approximate a realistic set of attributes for a character of a certain level and with a certain skillset.
//
// attributeGrowth: baseline rate for how much attributes grow when a linked skill levels up
// (not factoring in major/minor/misc though - just a minimal growth standard that affects every attribute).
// should be a value between 0.3 - 0.5; meant to keep attributes from not growing at an equal rate as its skills. that would cause it to blow up.
//
// levelAttributeRate: a baseline expectation for how fast attributes would "theoretically" grow on each level up.
// this doesn't factor anything in; just a baseline expectation that we will blend in to "normalize" attribute growth a bit.
//
// defMgr: used for getting a skill's definition, so we can see what its governing skill is.
func generateAttributesFromSkills(
	level int,
	skills map[defs.SkillID]int,
	class defs.ClassDef,
	lvlParams defs.LevelSystemParameters,
	levelAttributeRate float64,
	randomness float64,
	rng *rand.Rand,
	defMgr *definitions.DefinitionManager,
) map[defs.AttributeID]int {
	// aggregate governing skills
	type agg struct {
		weightedSum float64
		totalWeight float64
	}

	attributeData := make(map[defs.AttributeID]*agg)

	for skillID, skillLevel := range skills {
		weight := 1.0

		var baseLevel int

		category := class.SkillCategories[skillID]
		switch category {
		case SkillCategoryMajor:
			weight = lvlParams.MajorWeight
			baseLevel = lvlParams.MajorBase
		case SkillCategoryMinor:
			weight = lvlParams.MinorWeight
			baseLevel = lvlParams.MinorBase
		case SkillCategoryMisc:
			weight = lvlParams.MiscWeight
			baseLevel = lvlParams.MiscBase
		}

		skillGrowth := max(skillLevel-baseLevel, 0)

		skillDef := defMgr.GetSkillDef(skillID)
		govAttr := skillDef.GoverningAttributes

		if len(govAttr) == 0 {
			panic("skill has no governing attribute?")
		}

		weightedSumAdd := float64(skillGrowth) * weight
		totalWeightAdd := weight

		for _, attrID := range govAttr {
			if _, exists := attributeData[attrID]; !exists {
				attributeData[attrID] = &agg{}
			}

			// since there is more than one governming attribute (usually), split the effect between them.
			// TODO: this implies that, for a skill that only has a single attribute, level ups for that skill affect that attributes growth more.
			attributeData[attrID].weightedSum += weightedSumAdd / float64(len(govAttr))
			attributeData[attrID].totalWeight += totalWeightAdd / float64(len(govAttr))
		}
	}
	result := make(map[defs.AttributeID]int)

	for attrID, data := range attributeData {
		// // Leaving this here in case we want to try averages again.
		// skillAverage := 0.0
		// if data.totalWeight > 0 {
		// 	skillAverage = data.weightedSum / data.totalWeight
		// }

		// use total skill growth instead of averages
		skillTotalGrowth := data.weightedSum

		// bases and bonuses
		attrBase := lvlParams.AttributeBase
		if slices.Contains(class.FavoredAttributes, attrID) {
			attrBase += lvlParams.FavoredBonus
		}

		// derived from skills
		derivedGrowth := skillTotalGrowth * lvlParams.AttributeGrowth

		// level expectation blend
		// expected := float64(lvlParams.AttributeBase) + float64(level)*levelAttributeRate
		// final := 0.7*derived + 0.3*expected

		// small randomness
		randomFactor := 1 + (rng.Float64()*2-1)*randomness
		derivedGrowth *= randomFactor

		final := float64(attrBase) + derivedGrowth

		// cap at 100
		final = min(final, 100)
		result[attrID] = int(final)
	}

	return result
}
