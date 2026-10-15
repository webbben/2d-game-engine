package characterstate

import (
	"testing"

	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/pubsub"
)

// minimalDataman builds a DataManager with the class, culture, and character def that resolving a
// character's levels requires. Resolving levels walks charDef -> classDef and charDef -> cultureDef, so
// the chain has to exist even for a test that only cares about trait math.
func minimalDataman(t *testing.T) *datamanager.DataManager {
	t.Helper()
	dataman := datamanager.NewDataManager()

	dataman.LoadLevelSys(&defs.LevelSystemParameters{
		XPToNextSkillLevel: func(currentLevel int) int { return 100 },
		SkillLevelCap:      100,
		MiscCount:          1,
		MiscRate:           1,
		MiscWeight:         1,
	})
	dataman.LoadClassDef(defs.ClassDef{ID: testClassID, SkillCategories: map[defs.SkillID]defs.SkillCategory{}})
	dataman.LoadCultureDef(defs.CultureDef{
		ID: testCultureID, DisplayName: "Test Culture", Description: "culture used by these tests",
	})
	dataman.LoadCharacterDef(defs.CharacterDef{
		ID:          testCharDefID,
		DisplayName: "Test Char",
		ClassDefID:  testClassID,
		CultureID:   testCultureID,
		Version:     defs.CharacterDefVersion,
	})
	return dataman
}

const (
	testCharDefID = id.CharacterDefID("test_char_def")
	testClassID   = defs.ClassDefID("test_class")
	testCultureID = defs.CultureID("test_culture")
)

// A level-up must invalidate the resolved-levels cache.
//
// This is the whole reason the cache is dangerous: a stale entry doesn't crash, it just keeps reporting
// the skills and attributes a character had before they leveled. Combat math would silently carry on
// using the old numbers, and nothing would say so. So the test is deliberately end-to-end through
// AddSkillXP rather than testing MarkCharacterLevelsDirty directly -- the bug that mattered was a
// missing call in the mutation path, not broken cache logic, and only the full path exercises that.
func TestLevelUpInvalidatesCachedLevels(t *testing.T) {
	const (
		charID  = id.CharacterStateID("test_char")
		skillID = defs.SkillID("BLUNT")
		attrID  = defs.AttributeID("MARTIAL")
	)

	dataman := minimalDataman(t)
	eventBus := pubsub.NewEventBus()

	// SkillCategories is empty, so this skill counts as MISC and its weight is too low to trigger a
	// character level-up. That would set PendingLevelUp and make any further AddSkillXP a no-op.
	dataman.LoadCharacterState(&state.CharacterState{
		ID:             charID,
		DefID:          testCharDefID,
		BaseSkills:     map[defs.SkillID]int{skillID: 10},
		BaseAttributes: map[defs.AttributeID]int{attrID: 30},
		SkillXP:        map[defs.SkillID]int{skillID: 0},
	})

	// prime the cache, exactly as the per-frame visibility path does
	skills, attrs := GetSkillsAndAttributes(charID, dataman)
	if skills[skillID] != 10 {
		t.Fatalf("setup: skill level = %d, want 10", skills[skillID])
	}
	if attrs[attrID] != 30 {
		t.Fatalf("setup: attribute level = %d, want 30", attrs[attrID])
	}

	// a plain read must not disturb the cached value
	if _, _ = GetSkillsAndAttributes(charID, dataman); skills[skillID] != 10 {
		t.Fatalf("cached skill level changed on a plain re-read: %d", skills[skillID])
	}

	AddSkillXP(charID, skillID, 150, dataman, eventBus)

	charState := dataman.GetCharacterState(charID)
	if charState.BaseSkills[skillID] != 11 {
		t.Fatalf("base skill didn't actually level up: %d, want 11", charState.BaseSkills[skillID])
	}

	// the assertion that matters: without invalidation this still reads 10
	skills, _ = GetSkillsAndAttributes(charID, dataman)
	if skills[skillID] != 11 {
		t.Errorf("cached skill level after level-up = %d, want 11 (stale cache would report 10)", skills[skillID])
	}
}

// Resolved levels include trait and culture modifiers, so those inputs have to invalidate too. This
// exercises the trait half directly, since traits are appended to a live character at creation rather
// than through one of the engine's own mutators.
func TestTraitChangeInvalidatesCachedLevels(t *testing.T) {
	const (
		charID  = id.CharacterStateID("trait_char")
		skillID = defs.SkillID("BLUNT")
	)

	dataman := minimalDataman(t)
	dataman.LoadTraitDef(defs.Trait{
		ID: "BRAVE", Name: "Brave", Description: "d",
		SkillChanges: map[defs.SkillID]int{skillID: 15},
	})
	dataman.LoadCharacterState(&state.CharacterState{
		ID:         charID,
		DefID:      testCharDefID,
		BaseSkills: map[defs.SkillID]int{skillID: 10},
	})

	skills, _ := GetSkillsAndAttributes(charID, dataman)
	if skills[skillID] != 10 {
		t.Fatalf("setup: skill level = %d, want 10", skills[skillID])
	}

	// what the character builder and class questionnaire do to a real character
	charState := dataman.GetCharacterState(charID)
	charState.Traits = append(charState.Traits, "BRAVE")
	dataman.MarkCharacterLevelsDirty(charID)

	skills, _ = GetSkillsAndAttributes(charID, dataman)
	if skills[skillID] != 25 {
		t.Errorf("skill level after gaining a +15 trait = %d, want 25 (stale cache would report 10)", skills[skillID])
	}
}

// The uncached path must stay usable on its own: it takes a def, not a state, so the character builder
// uses it to preview modifiers before a character exists.
func TestCalculateSkillsAndAttributesBypassesTheCache(t *testing.T) {
	const (
		charID  = id.CharacterStateID("preview_char")
		skillID = defs.SkillID("BLUNT")
	)

	dataman := minimalDataman(t)
	dataman.LoadTraitDef(defs.Trait{
		ID: "BRAVE", Name: "Brave", Description: "d",
		SkillChanges: map[defs.SkillID]int{skillID: 15},
	})
	dataman.LoadCharacterState(&state.CharacterState{
		ID:         charID,
		DefID:      testCharDefID,
		BaseSkills: map[defs.SkillID]int{skillID: 10},
	})

	// cache a traitless result first
	if skills, _ := GetSkillsAndAttributes(charID, dataman); skills[skillID] != 10 {
		t.Fatalf("setup: skill level = %d, want 10", skills[skillID])
	}

	// adding a trait without invalidating leaves the cache stale...
	charState := dataman.GetCharacterState(charID)
	charState.Traits = append(charState.Traits, "BRAVE")

	// ...but the uncached path always reflects the current state
	skills, _ := CalculateSkillsAndAttributes(charID, dataman)
	if skills[skillID] != 25 {
		t.Errorf("uncached skill level = %d, want 25", skills[skillID])
	}
}
