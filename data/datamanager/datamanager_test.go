package datamanager

import (
	"testing"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
)

// The resolved-levels cache is the kind of thing that breaks silently: a stale entry doesn't crash, it
// just hands out the wrong skills. So these tests pin the caching contract itself rather than trusting
// that the invalidation call sites are all still there.
//
// The integration half lives in entity/characterState, which checks that an actual level-up invalidates.

func testLevels(n int) (map[defs.SkillID]int, map[defs.AttributeID]int) {
	return map[defs.SkillID]int{defs.SkillID("BLUNT"): n},
		map[defs.AttributeID]int{defs.AttributeID("MARTIAL"): n}
}

// A miss computes; a hit reuses. This is the whole point of the cache, so it's worth pinning that
// compute isn't called again on a hit -- an accidental recompute would silently undo the optimization.
func TestResolveCharacterLevelsCachesUntilDirty(t *testing.T) {
	dataman := NewDataManager()
	const charID = id.CharacterStateID("npc")

	calls := 0
	compute := func() (map[defs.SkillID]int, map[defs.AttributeID]int) {
		calls++
		return testLevels(10)
	}

	skills, _ := dataman.ResolveCharacterLevels(charID, compute)
	if calls != 1 {
		t.Fatalf("first read should compute once, got %d computes", calls)
	}
	if skills[defs.SkillID("BLUNT")] != 10 {
		t.Errorf("skill level = %d, want 10", skills[defs.SkillID("BLUNT")])
	}

	skills, _ = dataman.ResolveCharacterLevels(charID, compute)
	if calls != 1 {
		t.Errorf("second read should be served from cache, but compute ran %d times", calls)
	}

	// now pretend the character changed underneath the cache
	dataman.MarkCharacterLevelsDirty(charID)
	skills, _ = dataman.ResolveCharacterLevels(charID, func() (map[defs.SkillID]int, map[defs.AttributeID]int) {
		calls++
		return testLevels(20)
	})
	if calls != 2 {
		t.Errorf("read after dirty should recompute, compute ran %d times, want 2", calls)
	}
	if skills[defs.SkillID("BLUNT")] != 20 {
		t.Errorf("skill level after invalidation = %d, want 20 (stale cache would report 10)", skills[defs.SkillID("BLUNT")])
	}
}

// The bug this guards against: marking a character dirty used to panic when nothing was cached. It's
// entirely reachable, because a character can change before anything has read their levels -- sneaking
// far enough to earn a level-up qualifies, and nothing in the player's own update reads their skills.
func TestMarkCharacterLevelsDirtyWithNoEntryIsNotAnError(t *testing.T) {
	dataman := NewDataManager()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("marking an uncached character dirty should be a no-op, but it panicked: %v", r)
		}
	}()

	dataman.MarkCharacterLevelsDirty(id.CharacterStateID("never-read-me"))

	// and the character still resolves correctly afterwards
	skills, _ := dataman.ResolveCharacterLevels(id.CharacterStateID("never-read-me"), func() (map[defs.SkillID]int, map[defs.AttributeID]int) {
		return testLevels(5)
	})
	if skills[defs.SkillID("BLUNT")] != 5 {
		t.Errorf("skill level = %d, want 5", skills[defs.SkillID("BLUNT")])
	}
}

// Generated NPCs get deleted at corpse expiry and can be recreated with the same ID. If the cache entry
// outlived the character, the replacement would inherit the dead one's skills.
func TestRemoveCharacterStateDropsCachedLevels(t *testing.T) {
	dataman := NewDataManager()
	const charID = id.CharacterStateID("generated_bandit")

	dataman.LoadCharacterState(&state.CharacterState{ID: charID})

	skills, _ := dataman.ResolveCharacterLevels(charID, func() (map[defs.SkillID]int, map[defs.AttributeID]int) {
		return testLevels(40)
	})
	if skills[defs.SkillID("BLUNT")] != 40 {
		t.Fatalf("setup: skill level = %d, want 40", skills[defs.SkillID("BLUNT")])
	}

	dataman.RemoveCharacterState(charID)

	// same ID, different character: must not inherit the previous one's levels
	dataman.LoadCharacterState(&state.CharacterState{ID: charID})
	skills, _ = dataman.ResolveCharacterLevels(charID, func() (map[defs.SkillID]int, map[defs.AttributeID]int) {
		return testLevels(3)
	})
	if skills[defs.SkillID("BLUNT")] != 3 {
		t.Errorf("recreated character got skill level %d, want 3 (stale cache entry would report 40)", skills[defs.SkillID("BLUNT")])
	}
}

// Each character must have its own entry. A cache shared across IDs would hand one NPC another's skills.
func TestResolveCharacterLevelsIsPerCharacter(t *testing.T) {
	dataman := NewDataManager()
	computeFor := func(level int) func() (map[defs.SkillID]int, map[defs.AttributeID]int) {
		return func() (map[defs.SkillID]int, map[defs.AttributeID]int) {
			return testLevels(level)
		}
	}

	a, _ := dataman.ResolveCharacterLevels(id.CharacterStateID("a"), computeFor(11))
	b, _ := dataman.ResolveCharacterLevels(id.CharacterStateID("b"), computeFor(77))

	if a[defs.SkillID("BLUNT")] != 11 {
		t.Errorf("character a skill level = %d, want 11", a[defs.SkillID("BLUNT")])
	}
	if b[defs.SkillID("BLUNT")] != 77 {
		t.Errorf("character b skill level = %d, want 77", b[defs.SkillID("BLUNT")])
	}

	// invalidating one must not disturb the other
	dataman.MarkCharacterLevelsDirty(id.CharacterStateID("a"))
	b, _ = dataman.ResolveCharacterLevels(id.CharacterStateID("b"), computeFor(77))
	if b[defs.SkillID("BLUNT")] != 77 {
		t.Errorf("invalidating a changed b: got %d, want 77", b[defs.SkillID("BLUNT")])
	}
}

// Guard against a future edit quietly dropping the nil check that every other dataman accessor has.
func TestResolveCharacterLevelsRejectsEmptyID(t *testing.T) {
	dataman := NewDataManager()

	defer func() {
		if recover() == nil {
			t.Error("expected a panic for an empty character ID")
		}
	}()

	_, _ = dataman.ResolveCharacterLevels("", func() (map[defs.SkillID]int, map[defs.AttributeID]int) {
		return testLevels(1)
	})
}
