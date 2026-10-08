package combat

import (
	"slices"
	"testing"

	"github.com/webbben/2d-game-engine/data/id"
)

// stubResolver answers from fixed sets, standing in for the world/npc packages this package can't import.
type stubResolver struct {
	dead     map[id.CharacterStateID]bool
	offering map[id.CharacterStateID]bool
	fleeing  map[id.CharacterStateID]bool
}

func (s stubResolver) IsDead(c id.CharacterStateID) bool              { return s.dead[c] }
func (s stubResolver) IsOfferingSurrender(c id.CharacterStateID) bool { return s.offering[c] }
func (s stubResolver) IsFleeing(c id.CharacterStateID) bool           { return s.fleeing[c] }

// side0 and side1 name the two sides in these tests by who founded them.
var (
	side0 = NewSideID("player")
	side1 = NewSideID("bandit_a")
)

func newTestSession(t *testing.T) CombatSession {
	t.Helper()
	s := NewCombatSession("s1", "player", IntentSelfDefense, stubResolver{})
	// the initiator already joined as the founder of side0 in NewCombatSession
	s.FoundSide("bandit_a", IntentRob)
	s.Join("bandit_b", side1, IntentKill)
	s.Join("guard", side0, IntentProtect)
	return s
}

// sameIDs compares as sets. The session's query helpers deliberately don't sort -- Go randomizes map
// iteration and these are called often enough that the extra work isn't worth it -- so tests must not
// depend on the order they come back in.
func sameIDs(got, want []id.CharacterStateID) bool {
	sorted := slices.Clone(got)
	slices.Sort(sorted)
	return slices.Equal(sorted, want)
}

func TestNewSessionAutoJoinsInitiator(t *testing.T) {
	s := NewCombatSession("s1", "player", IntentSelfDefense, stubResolver{})

	if !s.IsCombatant("player") {
		t.Error("initiator did not join the session they started")
	}
	if got, ok := s.SideOf("player"); !ok || got != side0 {
		t.Errorf("SideOf(player) = %v, %v; want %v, true", got, ok, side0)
	}
	if got := s.AggressorSide(); got != side0 {
		t.Errorf("AggressorSide() = %v, want %v (the initiator's own side)", got, side0)
	}
	// a session with only the initiator hasn't really started a fight yet
	if !s.HasEnded() {
		t.Error("HasEnded() = false with a single combatant; want true")
	}
	if s.Size() != 1 {
		t.Errorf("Size() = %d, want 1", s.Size())
	}
}

func TestFoundSide(t *testing.T) {
	s := NewCombatSession("s1", "player", IntentSelfDefense, stubResolver{})
	got := s.FoundSide("bandit_a", IntentRob)

	if got != side1 {
		t.Errorf("FoundSide returned %v, want %v (named for its founder)", got, side1)
	}
	if side, ok := s.SideOf("bandit_a"); !ok || side != side1 {
		t.Errorf("SideOf(bandit_a) = %v, %v; want %v, true", side, ok, side1)
	}
	if s.HasEnded() {
		t.Error("HasEnded() = true with two sides; want false")
	}
	if !s.IsAggressorSide(side0) || s.IsAggressorSide(side1) {
		t.Error("IsAggressorSide reported the wrong side")
	}
}

func TestFoundSideRejectsThird(t *testing.T) {
	s := newTestSession(t)
	defer func() {
		if recover() == nil {
			t.Error("founding a third side did not panic")
		}
	}()
	s.FoundSide("bandit_c", IntentRob)
}

func TestJoinAndQueries(t *testing.T) {
	s := newTestSession(t)

	if got := s.Size(); got != 4 {
		t.Errorf("Size() = %d, want 4", got)
	}
	if sides := s.Sides(); len(sides) != 2 || !slices.Contains(sides, side0) || !slices.Contains(sides, side1) {
		t.Errorf("Sides() = %v, want both %v and %v", sides, side0, side1)
	}
	if _, ok := s.SideOf("nobody"); ok {
		t.Error("SideOf(nobody) reported ok; want false")
	}

	if got, want := s.EnemiesOf("player"), []id.CharacterStateID{"bandit_a", "bandit_b"}; !sameIDs(got, want) {
		t.Errorf("EnemiesOf(player) = %v, want %v", got, want)
	}
	if got, want := s.AlliesOf("bandit_a"), []id.CharacterStateID{"bandit_b"}; !sameIDs(got, want) {
		t.Errorf("AlliesOf(bandit_a) = %v, want %v", got, want)
	}
	// a non-combatant has no enemies in this fight
	if got := s.EnemiesOf("nobody"); got != nil {
		t.Errorf("EnemiesOf(nobody) = %v, want nil", got)
	}
	if s.HasEnded() {
		t.Error("HasEnded() = true; want false")
	}
}

// TestAggressorSideOutlivesInitiator covers the reason a side is named for its founder rather than
// numbered: whoever is still fighting on the side that started the fight is part of that side, even
// after the initiator is gone.
func TestAggressorSideOutlivesInitiator(t *testing.T) {
	s := newTestSession(t)

	s.Leave("player")
	if got := s.AggressorSide(); got != side0 {
		t.Errorf("AggressorSide() after initiator left = %v, want %v", got, side0)
	}
	guardSide := sideOfMust(t, s, "guard")
	if !s.IsAggressorSide(guardSide) {
		t.Errorf("a survivor of the initiator's side (%v) no longer counts as an aggressor", guardSide)
	}
}

func sideOfMust(t *testing.T, s CombatSession, charID id.CharacterStateID) SideID {
	t.Helper()
	side, ok := s.SideOf(charID)
	if !ok {
		t.Fatalf("SideOf(%s) not found", charID)
	}
	return side
}

func TestJoinRejectsDuplicate(t *testing.T) {
	s := newTestSession(t)
	defer func() {
		if recover() == nil {
			t.Error("rejoining an existing combatant did not panic")
		}
		if s.Size() != 4 {
			t.Errorf("Size() = %d after a rejected join, want 4 (session must be unchanged)", s.Size())
		}
	}()
	s.Join("bandit_a", side0, IntentKill) // already a combatant, on side1
}

func TestLeaveAndEnd(t *testing.T) {
	s := newTestSession(t)

	if !s.Leave("guard") {
		t.Error("Leave(guard) = false; want true")
	}
	if s.Leave("guard") {
		t.Error("second Leave(guard) = true; want false")
	}
	if s.HasEnded() {
		t.Error("HasEnded() = true with both sides still populated; want false")
	}

	// last member of a side leaves: the fight is over even though the other side still has people
	if !s.Leave("bandit_a") || !s.Leave("bandit_b") {
		t.Error("expected to remove both bandits")
	}
	if !s.HasEnded() {
		t.Error("HasEnded() = false with only one side left; want true")
	}
}

func TestLivingMembersUsesResolver(t *testing.T) {
	r := stubResolver{dead: map[id.CharacterStateID]bool{"bandit_a": true}}
	s := NewCombatSession("s1", "player", IntentSelfDefense, r)
	s.FoundSide("bandit_a", IntentRob)
	s.Join("bandit_b", side1, IntentKill)
	s.Join("guard", side0, IntentProtect)

	if got, want := s.LivingMembers(side1), []id.CharacterStateID{"bandit_b"}; !sameIDs(got, want) {
		t.Errorf("LivingMembers(bandit side) = %v, want %v", got, want)
	}
	if got, want := s.LivingMembers(side0), []id.CharacterStateID{"guard", "player"}; !sameIDs(got, want) {
		t.Errorf("LivingMembers(player side) = %v, want %v", got, want)
	}
}

// TestOfferingSurrenderIsStillFighting pins down the distinction that made IsSurrendered ambiguous.
// A pending offer means the player hasn't answered yet, so the character is still alive, still in the
// session, and still a valid enemy. Only an *accepted* surrender ends the fight, and that's handled by
// leaving the session rather than by a query.
func TestOfferingSurrenderIsStillFighting(t *testing.T) {
	r := stubResolver{
		offering: map[id.CharacterStateID]bool{"bandit_a": true},
		fleeing:  map[id.CharacterStateID]bool{"bandit_b": true},
	}
	s := NewCombatSession("s1", "player", IntentSelfDefense, r)
	s.FoundSide("bandit_a", IntentRob)
	s.Join("bandit_b", side1, IntentKill)

	if !s.IsCombatant("bandit_a") {
		t.Error("offering surrenderee is no longer a combatant; they should still be fighting")
	}
	if s.HasEnded() {
		t.Error("HasEnded() = true while one combatant is merely offering surrender; want false")
	}
	if got, want := s.EnemiesOf("player"), []id.CharacterStateID{"bandit_a", "bandit_b"}; !sameIDs(got, want) {
		t.Errorf("EnemiesOf(player) = %v, want %v (an offering surrenderee is still a target)", got, want)
	}
	if got, want := s.LivingMembers(side1), []id.CharacterStateID{"bandit_a", "bandit_b"}; !sameIDs(got, want) {
		t.Errorf("LivingMembers(bandit side) = %v, want %v", got, want)
	}
}

func TestValidate(t *testing.T) {
	t.Run("valid session", func(t *testing.T) {
		newTestSession(t).Validate()
	})
	t.Run("ended session without initiator is legal", func(t *testing.T) {
		s := NewCombatSession("s1", "player", IntentSelfDefense, stubResolver{})
		s.Leave("player")
		s.Validate()
	})
	t.Run("missing resolver", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("Validate accepted a session with no resolver")
			}
		}()
		NewCombatSession("s1", "player", IntentSelfDefense, nil).Validate()
	})
	t.Run("initiator missing from an ongoing session", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("Validate accepted an ongoing session missing its initiator")
			}
		}()
		// guard shares the initiator's side, so removing the initiator leaves the fight ongoing
		// with no aggressor side -- which should be caught rather than silently tolerated.
		s := newTestSession(t)
		s.Leave("player")
		if s.HasEnded() {
			t.Fatal("test setup: expected the session to still be ongoing")
		}
		s.Validate()
	})
}
