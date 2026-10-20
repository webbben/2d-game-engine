package combat

import (
	"slices"
	"testing"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
)

// stubResolver answers from fixed sets, standing in for the world/npc packages this package can't import.
type stubResolver struct {
	dead     map[id.CharacterStateID]bool
	dying    map[id.CharacterStateID]bool
	offering map[id.CharacterStateID]bool
	fleeing  map[id.CharacterStateID]bool
}

func (s stubResolver) IsDead(c id.CharacterStateID) bool { return s.dead[c] }

// IsDeadOrDying treats the dying set as dead-or-dying but not dead, which is the whole point of having
// both: it lets a test say "this one just died and hasn't been processed yet" and see the two questions
// disagree, the way they do for real.
func (s stubResolver) IsDeadOrDying(c id.CharacterStateID) bool { return s.dead[c] || s.dying[c] }
func (s stubResolver) IsOfferingSurrender(c id.CharacterStateID) bool {
	return s.offering[c]
}
func (s stubResolver) IsFleeing(c id.CharacterStateID) bool { return s.fleeing[c] }

// side0 and side1 name the two sides in these tests by who founded them.
var (
	side0 = NewSideID("player")
	side1 = NewSideID("bandit_a")
)

func newTestSession(t *testing.T) CombatSession {
	t.Helper()
	s := NewCombatSession("s1", "player", defs.IntentSelfDefense, stubResolver{})
	// the initiator already joined as the founder of side0 in NewCombatSession
	s.FoundSide("bandit_a", defs.IntentRob)
	s.Join("bandit_b", side1, defs.IntentKill)
	s.Join("guard", side0, defs.IntentProtect)
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
	s := NewCombatSession("s1", "player", defs.IntentSelfDefense, stubResolver{})

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
	s := NewCombatSession("s1", "player", defs.IntentSelfDefense, stubResolver{})
	got := s.FoundSide("bandit_a", defs.IntentRob)

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
	s.FoundSide("bandit_c", defs.IntentRob)
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
	s.Join("bandit_a", side0, defs.IntentKill) // already a combatant, on side1
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
	s := NewCombatSession("s1", "player", defs.IntentSelfDefense, r)
	s.FoundSide("bandit_a", defs.IntentRob)
	s.Join("bandit_b", side1, defs.IntentKill)
	s.Join("guard", side0, defs.IntentProtect)

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
	s := NewCombatSession("s1", "player", defs.IntentSelfDefense, r)
	s.FoundSide("bandit_a", defs.IntentRob)
	s.Join("bandit_b", side1, defs.IntentKill)

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
		s := NewCombatSession("s1", "player", defs.IntentSelfDefense, stubResolver{})
		s.Leave("player")
		s.Validate()
	})
	t.Run("missing resolver", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Error("Validate accepted a session with no resolver")
			}
		}()
		NewCombatSession("s1", "player", defs.IntentSelfDefense, nil).Validate()
	})
	t.Run("initiator may leave an ongoing session", func(t *testing.T) {
		// An initiator who dies or walks away does not end the fight -- their side-mates carry on
		// without them. The aggressor side survives that, because a SideID is just the founder's char
		// id and the side-mates are still standing on it. Validate must not read a departed initiator
		// as a broken session.
		s := newTestSession(t)
		s.Leave("player")
		if s.HasEnded() {
			t.Fatal("test setup: expected the session to still be ongoing")
		}

		// The guard shares the initiator's side, so the fight's aggressor is still recoverable from
		// them. This is what makes it safe for the initiator to be gone.
		guardSide, ok := s.SideOf("guard")
		if !ok {
			t.Fatal("expected the guard to still be a combatant after the initiator left")
		}
		if !s.IsAggressorSide(guardSide) {
			t.Error("the initiator's side stopped being the aggressor side once the initiator left")
		}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Validate rejected an ongoing session whose initiator had left: %v", r)
			}
		}()
		s.Validate()
	})
}

func TestMaxSizeOfTracksJoinHighWaterMark(t *testing.T) {
	s := NewCombatSession("s1", "player", defs.IntentSelfDefense, stubResolver{})
	s.FoundSide("bandit_a", defs.IntentRob)
	s.Join("bandit_b", side1, defs.IntentKill)
	s.Join("bandit_c", side1, defs.IntentKill)

	if got := s.MaxSizeOf(side1); got != 3 {
		t.Errorf("MaxSizeOf(%s) = %d; want 3", side1, got)
	}
	if got := s.MaxSizeOf(side0); got != 1 {
		t.Errorf("MaxSizeOf(%s) = %d; want 1 (only the initiator)", side0, got)
	}
}

func TestMaxSizeOfSeedsTheInitiator(t *testing.T) {
	// The initiator is written straight into the combatants map by NewCombatSession, so nothing ever
	// calls Join on them. If the constructor didn't seed their side's mark, it would read 0 forever
	// until some other combatant joined the aggressor side.
	s := NewCombatSession("s1", "player", defs.IntentSelfDefense, stubResolver{})
	if got := s.MaxSizeOf(side0); got != 1 {
		t.Errorf("MaxSizeOf(%s) = %d on a fresh session; want 1", side0, got)
	}
	s.FoundSide("bandit_a", defs.IntentRob)
	if got := s.MaxSizeOf(side0); got != 1 {
		t.Errorf("MaxSizeOf(%s) = %d after the other side was founded; want 1", side0, got)
	}
	// founding a side counts as putting someone in it, even though it doesn't go through Join
	if got := s.MaxSizeOf(side1); got != 1 {
		t.Errorf("MaxSizeOf(%s) = %d for a side that was only founded, never joined; want 1", side1, got)
	}
}

func TestMaxSizeOfSurvivesCombatantsLeaving(t *testing.T) {
	// newTestSession leaves both sides holding two.
	s := newTestSession(t)
	for _, side := range []SideID{side0, side1} {
		if got := s.MaxSizeOf(side); got != 2 {
			t.Errorf("MaxSizeOf(%s) = %d on a fresh two-a-side; want 2", side, got)
		}
	}

	s.Leave("guard")
	if got := len(s.MembersOf(side0)); got != 1 {
		t.Fatalf("test setup: side0 has %d members after the guard left, want 1", got)
	}
	if got := s.MaxSizeOf(side0); got != 2 {
		t.Errorf("MaxSizeOf(%s) = %d after a member left; want 2 (a high-water mark, not a count)", side0, got)
	}
}

func TestMaxSizeOfIgnoresARejoin(t *testing.T) {
	// Leaving and rejoining the same fight is rare but legal. A combatant coming back is the same
	// person who was there before, not a new one.
	//
	// The side is taken well below its peak first so the rejoining join lands *under* the high-water
	// mark. A rejoining back up to exactly the peak would pass even if the mark were overwritten
	// unconditionally rather than only raised, so this is the case that actually pins the behavior down.
	s := newTestSession(t) // side0: player, guard
	s.Join("warden", side0, defs.IntentProtect)
	if got := s.MaxSizeOf(side0); got != 3 {
		t.Fatalf("test setup: MaxSizeOf(%s) = %d, want 3", side0, got)
	}
	s.Leave("guard")
	s.Leave("warden")
	if got := len(s.MembersOf(side0)); got != 1 {
		t.Fatalf("test setup: side0 has %d members, want 1", got)
	}

	s.Join("guard", side0, defs.IntentProtect)
	if got := len(s.MembersOf(side0)); got != 2 {
		t.Fatalf("test setup: side0 has %d members after the rejoin, want 2", got)
	}
	if got := s.MaxSizeOf(side0); got != 3 {
		t.Errorf("MaxSizeOf(%s) = %d after a rejoin took it under the mark; want 3", side0, got)
	}
}

func TestMaxSizeOfIsZeroForAnUnusedSide(t *testing.T) {
	s := newTestSession(t)
	if got := s.MaxSizeOf(NewSideID("nobody")); got != 0 {
		t.Errorf("MaxSizeOf(a side nobody joined) = %d; want 0", got)
	}
}

func TestAllSidesOutlivesEmptiedSides(t *testing.T) {
	// Sides() shrinks as their last combatant leaves, and goes empty on a finished fight. AllSides is
	// how a caller still sees both sides of a fight after it ended.
	s := newTestSession(t)
	if got := s.Sides(); len(got) != 2 {
		t.Fatalf("test setup: Sides() = %v, want 2 sides", got)
	}

	// wipe out side1 entirely
	for _, charID := range s.MembersOf(side1) {
		s.Leave(charID)
	}
	if !s.HasEnded() {
		t.Fatal("test setup: expected the session to have ended once side1 emptied")
	}
	if got := s.Sides(); len(got) != 1 || got[0] != side0 {
		t.Errorf("Sides() = %v after side1 emptied; want just [%s]", got, side0)
	}

	// sorted, not declaration order: side IDs are their founders' character ids, so "bandit_a" precedes
	// "player" here. The point of the test is that both sides are present, not which way round.
	got := s.AllSides()
	want := []SideID{side1, side0}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("AllSides() = %v after side1 emptied; want %v (both sides remembered)", got, want)
	}
	// and the emptied side still knows how big it got
	if n := s.MaxSizeOf(side1); n != 2 {
		t.Errorf("MaxSizeOf(%s) = %d for the emptied side; want 2", side1, n)
	}
}

func TestFightingEnemiesOfExcludesDyingNotJustDead(t *testing.T) {
	// The distinction this method exists for. A combatant whose health has hit zero but whose corpse
	// hasn't been processed is settled-not-dead but already past fighting, so it must not be reported as
	// someone who can still hit you.
	r := stubResolver{
		dying: map[id.CharacterStateID]bool{"bandit_dying": true},
		fleeing: map[id.CharacterStateID]bool{
			// a fleer is alive and still very much a problem
			"bandit_running": true,
		},
	}
	s := NewCombatSession("s1", "player", defs.IntentSelfDefense, r)
	side1 := s.FoundSide("bandit_a", defs.IntentRob)
	s.Join("bandit_dying", side1, defs.IntentKill)
	s.Join("bandit_running", side1, defs.IntentKill)
	s.Join("bandit_fresh", side1, defs.IntentKill)

	if got := r.IsDead("bandit_dying"); got {
		t.Error("test setup: the dying combatant should not be settled-dead")
	}
	if got := r.IsDeadOrDying("bandit_dying"); !got {
		t.Error("test setup: the dying combatant should be dead-or-dying")
	}

	got := s.FightingEnemiesOf("player")
	// bandit_a founded the other side, so they're an enemy too and still very much fighting.
	// Sorted, because sameIDs compares as sets but takes an already-sorted want.
	want := []id.CharacterStateID{"bandit_a", "bandit_fresh", "bandit_running"}
	if !sameIDs(got, want) {
		t.Errorf("FightingEnemiesOf(player) = %v; want %v (the dying one excluded, the fleer kept)", got, want)
	}
}

func TestFightingEnemiesOfIsEmptyWhenAllEnemiesAreDown(t *testing.T) {
	r := stubResolver{
		dead:  map[id.CharacterStateID]bool{"bandit_a": true, "bandit_b": true},
		dying: map[id.CharacterStateID]bool{"bandit_c": true},
	}
	s := NewCombatSession("s1", "player", defs.IntentSelfDefense, r)
	side1 := s.FoundSide("bandit_a", defs.IntentRob)
	s.Join("bandit_b", side1, defs.IntentKill)
	s.Join("bandit_c", side1, defs.IntentKill)

	if got := s.FightingEnemiesOf("player"); len(got) != 0 {
		t.Errorf("FightingEnemiesOf(player) = %v; want none once every enemy is dead or dying", got)
	}
	// they are all still combatants -- that's the trap this method steps around
	if got := s.EnemiesOf("player"); len(got) != 3 {
		t.Errorf("test setup: EnemiesOf(player) = %v; want all 3 still listed as combatants", got)
	}
}

func TestFightingEnemiesOfIsSideRelative(t *testing.T) {
	// "Enemies" is relative to whoever is asking: the dying guard is the player's problem, not the
	// bandit's, because they're on the player's side.
	r := stubResolver{dying: map[id.CharacterStateID]bool{"guard": true}}
	s := NewCombatSession("s1", "player", defs.IntentSelfDefense, r)
	s.FoundSide("bandit_a", defs.IntentRob)
	s.Join("guard", s.AggressorSide(), defs.IntentProtect)

	// the player's opponents are the bandit's side; the guard is their own side-mate and never an enemy
	if got := s.FightingEnemiesOf("player"); len(got) != 1 || got[0] != "bandit_a" {
		t.Errorf("FightingEnemiesOf(player) = %v; want just [bandit_a]", got)
	}
	// and the other way round: the band's opponents are the player's side, minus the dying guard
	if got := s.FightingEnemiesOf("bandit_a"); len(got) != 1 || got[0] != "player" {
		t.Errorf("FightingEnemiesOf(bandit_a) = %v; want just [player], the guard is dying", got)
	}
}
