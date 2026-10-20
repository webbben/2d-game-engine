package world

import (
	"fmt"
	"slices"
	"testing"

	"github.com/webbben/2d-game-engine/combat"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
)

// StartCombat's bookkeeping is reachable without a fully built World: it only touches the session
// list and its sequence counter. The Resolver methods need a DataManager and NPCs, but nothing here
// calls them, so a bare World is enough to exercise the join matrix.

func sameIDs(got, want []id.CharacterStateID) bool {
	sorted := slices.Clone(got)
	slices.Sort(sorted)
	return slices.Equal(sorted, want)
}

func TestStartCombatCreatesSession(t *testing.T) {
	w := &World{}

	session := w.StartCombat("player", "bandit_a", defs.IntentKill, defs.IntentSelfDefense)
	if session == nil {
		t.Fatal("StartCombat returned nil; want a new session")
	}
	if !w.IsInCombat("player") || !w.IsInCombat("bandit_a") {
		t.Error("both combatants should be findable in the fight")
	}
	if got := session.Sides(); len(got) != 2 {
		t.Errorf("Sides() = %v, want two sides", got)
	}
	if !session.IsAggressorSide(combat.NewSideID("player")) {
		t.Error("the initiator's side should be the aggressor side")
	}
	// the defender founds the opposing side, so they're not on the aggressor's side
	if side, _ := session.SideOf("bandit_a"); side == session.AggressorSide() {
		t.Error("defender ended up on the aggressor's side")
	}
	// and each sees the other as an enemy
	if got, want := session.EnemiesOf("player"), []id.CharacterStateID{"bandit_a"}; !sameIDs(got, want) {
		t.Errorf("EnemiesOf(player) = %v, want %v", got, want)
	}
}

func TestStartCombatJoinsExistingSession(t *testing.T) {
	w := &World{}
	existing := w.StartCombat("player", "bandit_a", defs.IntentKill, defs.IntentSelfDefense)

	// a bystander attacks one of the bandits: pulled into the fight, opposite side from their victim
	joined := w.StartCombat("guard", "bandit_a", defs.IntentProtect, defs.IntentKill)
	if joined == nil {
		t.Fatal("StartCombat returned nil for an uncommitted attacker; want the existing session")
	}
	if joined.SessionID() != existing.SessionID() {
		t.Errorf("guard joined session %v, want the existing %v", joined.SessionID(), existing.SessionID())
	}
	if got, want := joined.MembersOf(joined.AggressorSide()),
		[]id.CharacterStateID{"guard", "player"}; !sameIDs(got, want) {
		t.Errorf("aggressor side members = %v, want %v (guard joined them)", got, want)
	}
	if got, want := joined.AlliesOf("player"), []id.CharacterStateID{"guard"}; !sameIDs(got, want) {
		t.Errorf("player's allies = %v, want %v", got, want)
	}
	// and the guard is now a second enemy for the bandit, who still has exactly one
	if got, want := joined.EnemiesOf("bandit_a"), []id.CharacterStateID{"guard", "player"}; !sameIDs(got, want) {
		t.Errorf("bandit_a's enemies = %v, want %v", got, want)
	}
}

func TestStartCombatDrawsFreeTargetIntoExistingFight(t *testing.T) {
	w := &World{}
	w.StartCombat("player", "bandit_a", defs.IntentKill, defs.IntentSelfDefense)
	before := len(w.combatSessions)

	// The player swings again and hits someone new. The player is already committed, but that must not
	// stop the new victim joining: they're now an enemy too.
	got := w.StartCombat("player", "bandit_b", defs.IntentKill, defs.IntentKill)
	if got == nil {
		t.Fatal("a free target hit by an already-fighting character should join that fight")
	}
	if len(w.combatSessions) != before {
		t.Errorf("session count changed to %d, want %d (should join, not start a new fight)", len(w.combatSessions), before)
	}
	if !w.IsInCombat("bandit_b") {
		t.Error("bandit_b should be in the fight after being hit")
	}
	// and on the same side as the bandit they were already fighting alongside
	if side, _ := got.SideOf("bandit_b"); side != combat.NewSideID("bandit_a") {
		t.Errorf("bandit_b joined side %v, want the same side as bandit_a (%v)", side, combat.NewSideID("bandit_a"))
	}
	if want := []id.CharacterStateID{"bandit_a", "bandit_b"}; !sameIDs(got.EnemiesOf("player"), want) {
		t.Errorf("player's enemies = %v, want %v", got.EnemiesOf("player"), want)
	}
}

func TestStartCombatIgnoresWhenBothCommitted(t *testing.T) {
	w := &World{}
	w.StartCombat("player", "bandit_a", defs.IntentKill, defs.IntentSelfDefense)
	w.StartCombat("guard", "bandit_a", defs.IntentProtect, defs.IntentKill)

	// player and bandit_a are both already committed; they're coming to blows, but the session doesn't
	// change -- the victim decides for themselves whether to get involved, and that logic doesn't exist yet.
	before := len(w.combatSessions)
	got := w.StartCombat("bandit_a", "player", defs.IntentSelfDefense, defs.IntentKill)
	if got != nil {
		t.Error("StartCombat returned a session when both characters were already committed; want nil")
	}
	if len(w.combatSessions) != before {
		t.Errorf("session count changed to %d, want %d", len(w.combatSessions), before)
	}
}

func TestStartCombatRejectsSelfCombat(t *testing.T) {
	w := &World{}
	defer func() {
		if recover() == nil {
			t.Error("a character fighting themselves did not panic")
		}
	}()
	w.StartCombat("player", "player", defs.IntentKill, defs.IntentSelfDefense)
}

func TestLeaveCombatEndsSession(t *testing.T) {
	w := &World{}
	w.StartCombat("player", "bandit_a", defs.IntentKill, defs.IntentSelfDefense)
	w.StartCombat("guard", "bandit_a", defs.IntentProtect, defs.IntentKill)

	// removing a combatant who isn't in a fight is a no-op
	w.LeaveCombat("nobody")

	w.LeaveCombat("guard")
	if !w.IsInCombat("player") {
		t.Error("player should still be in the fight")
	}

	// taking out the last member of the aggressor side ends the fight
	w.LeaveCombat("player")
	if w.IsInCombat("bandit_a") {
		t.Error("bandit_a should no longer be findable once the fight ended")
	}
	if len(w.combatSessions) != 0 {
		t.Errorf("%d sessions remain, want 0", len(w.combatSessions))
	}
}

func TestPruneEndedCombatSessions(t *testing.T) {
	w := &World{}
	live := w.StartCombat("player", "bandit_a", defs.IntentKill, defs.IntentSelfDefense)

	// a session whose second side never got filled has already ended
	w.StartCombat("guard", "merchant", defs.IntentProtect, defs.IntentKill)
	orphan := w.concreteSessionFor("merchant")
	orphan.Leave("merchant")

	if w.SessionFor("guard") == nil {
		t.Fatal("test setup: guard should still be in the orphan session")
	}

	w.pruneEndedCombatSessions()
	if len(w.combatSessions) != 1 {
		t.Fatalf("%d sessions remain, want 1", len(w.combatSessions))
	}
	if got := w.combatSessions[0].SessionID(); got != live.SessionID() {
		t.Errorf("kept session %v, want the live one %v", got, live.SessionID())
	}
	if w.IsInCombat("guard") {
		t.Error("the ended fight should no longer be findable")
	}
}

func TestSessionIDsAreOrdered(t *testing.T) {
	w := &World{}
	first := w.StartCombat("player", "bandit_a", defs.IntentKill, defs.IntentSelfDefense)
	w.LeaveCombat("player")
	w.LeaveCombat("bandit_a")
	second := w.StartCombat("player", "bandit_b", defs.IntentKill, defs.IntentSelfDefense)

	if first.SessionID() == second.SessionID() {
		t.Error("two fights shared a session ID")
	}
	if first.SessionID() != "combat-1" || second.SessionID() != "combat-2" {
		t.Errorf("session IDs = %v, %v; want combat-1, combat-2", first.SessionID(), second.SessionID())
	}
}

// TestLeaveCombatEndsFightWhenSideWiped covers the bookkeeping that motivates death cleanup: HasEnded is
// judged by side membership, so every combatant on a side having to actually leave is what ends a fight.
// Without that, a losing side that died still counted as having members and the session lingered.
func TestLeaveCombatEndsFightWhenSideWiped(t *testing.T) {
	w := &World{}
	// the player hits both bandits in one swing
	w.StartCombat("player", "bandit_a", defs.IntentKill, defs.IntentSelfDefense)
	w.StartCombat("player", "bandit_b", defs.IntentKill, defs.IntentSelfDefense)

	session := w.SessionFor("bandit_a")
	if session == nil {
		t.Fatal("test setup: bandit_a should be in the fight")
	}
	if banditBSide, _ := session.SideOf("bandit_b"); banditBSide != combat.NewSideID("bandit_a") {
		t.Fatalf("test setup: bandit_b is on side %v, want the same side as bandit_a", banditBSide)
	}

	// one bandit dies, but the other is still standing, so the fight continues
	w.LeaveCombat("bandit_a")
	if !w.IsInCombat("bandit_b") {
		t.Error("the fight should continue while bandit_b is still standing")
	}

	// last one dies: the fight is over
	w.LeaveCombat("bandit_b")
	if w.IsInCombat("player") {
		t.Error("the player should no longer be in a fight once every enemy is gone")
	}
	if len(w.combatSessions) != 0 {
		t.Errorf("%d sessions remain, want 0", len(w.combatSessions))
	}
}

// Whether a finished fight earns a victory celebration depends only on the session's own history --
// how big each side got, and whether exactly one was left standing. None of that needs a World, so the
// rule is a free function and these tests drive real sessions through StartCombat.
// Whether a combatant won a fight worth celebrating depends on the session's history plus one liveness
// question, and the liveness question is the part that needs a resolver. So these build sessions directly
// against a stub rather than through StartCombat, which would wire up a real World whose Dataman is nil
// here -- IsDeadOrDying panics on that by design.

const minCelebrationSideSize = 3

// celebDown is the shared set of combatants that are dead or dying.
//
// One map, created before the session is built and shared with it. CombatSession stores its resolver by
// value, so a resolver handed a map sees later writes to it, while one built from a map literal it owns
// would freeze the moment it was constructed -- and these tests decide who is down after the fight is
// built.
type celebDown map[id.CharacterStateID]bool

type celebResolver struct{ down celebDown }

func (r celebResolver) IsDead(c id.CharacterStateID) bool              { return r.down[c] }
func (r celebResolver) IsDeadOrDying(c id.CharacterStateID) bool       { return r.down[c] }
func (r celebResolver) IsOfferingSurrender(c id.CharacterStateID) bool { return false }
func (r celebResolver) IsFleeing(c id.CharacterStateID) bool           { return false }

// celebFight returns a session with perSide combatants on each side, plus the sides themselves. The first
// side is the initiator's, the second was founded by the first opponent to commit.
func celebFight(t *testing.T, down celebDown, perSide int) (*combat.CombatSession, combat.SideID, combat.SideID) {
	t.Helper()
	s := combat.NewCombatSession("s1", "player", defs.IntentSelfDefense, celebResolver{down: down})
	side0 := s.AggressorSide()
	side1 := s.FoundSide("bandit_a", defs.IntentRob)
	for i := 1; i < perSide; i++ {
		s.Join(id.CharacterStateID(fmt.Sprintf("guard%d", i)), side0, defs.IntentProtect)
		s.Join(id.CharacterStateID(fmt.Sprintf("bandit%d", i)), side1, defs.IntentKill)
	}
	return &s, side0, side1
}

// knockDown marks everyone on a side as no longer able to fight.
func knockDown(t *testing.T, down celebDown, session *combat.CombatSession, side combat.SideID) {
	t.Helper()
	for _, memberID := range session.MembersOf(side) {
		down[memberID] = true
	}
}

func TestCelebrationQualifiesOnEvenBigFight(t *testing.T) {
	down := celebDown{}
	session, side0, side1 := celebFight(t, down, 3)
	knockDown(t, down, session, side1)

	if got := len(session.MembersOf(side0)); got != 3 {
		t.Fatalf("test setup: winning side has %d members, want 3", got)
	}
	if !fightQualifiesForCelebration(session, "player", minCelebrationSideSize) {
		t.Error("a 3v3 that one side won should earn a celebration")
	}
}

func TestCelebrationJudgesPeaksNotFinalCounts(t *testing.T) {
	// The rule the whole feature hangs on: a fight is measured by how big it got, not by how many are
	// left when it ends. Here the winner is down to its last combatant and the loser is gone entirely,
	// and it still counts as a battle worth celebrating.
	down := celebDown{"guard1": true, "guard2": true}
	session, side0, side1 := celebFight(t, down, 3)
	// dying and leaving are separate things: a death takes them out of the fight, while the corpse gap
	// (see the test below) is the window where they're down but not yet gone.
	session.Leave("guard1")
	session.Leave("guard2")
	knockDown(t, down, session, side1)

	if got := len(session.MembersOf(side0)); got != 1 {
		t.Fatalf("test setup: winning side has %d members, want 1", got)
	}
	if !fightQualifiesForCelebration(session, "player", minCelebrationSideSize) {
		t.Error("a 3v3 that ground down to a last-man-standing should still qualify")
	}
}

func TestCelebrationSkipsSmallFights(t *testing.T) {
	down := celebDown{}
	session, _, side1 := celebFight(t, down, 2)
	knockDown(t, down, session, side1)

	if fightQualifiesForCelebration(session, "player", minCelebrationSideSize) {
		t.Error("a 2v2 should not earn a victory celebration")
	}
}

func TestCelebrationSkipsOneSidedAmbush(t *testing.T) {
	// Big enough on one side, hopeless on the other: winning a mugging isn't a battle.
	down := celebDown{}
	oneSided := combat.NewCombatSession("s1", "guard_a", defs.IntentKill, celebResolver{down: down})
	side0 := oneSided.AggressorSide()
	oneSided.Join("guard_b", side0, defs.IntentProtect)
	oneSided.Join("warden", side0, defs.IntentProtect)
	side1 := oneSided.FoundSide("bandit_a", defs.IntentRob)
	session := &oneSided
	knockDown(t, down, session, side1)

	if fightQualifiesForCelebration(session, "guard_a", minCelebrationSideSize) {
		t.Error("a 3-on-1 should not earn a victory celebration")
	}
}

func TestCelebrationSkipsWhenEnemiesRanInsteadOfDying(t *testing.T) {
	// Everyone still standing: the fight is being lost because the other side is leaving, which isn't
	// beating them. This is the rout case, and it has to be a no.
	down := celebDown{}
	session, _, _ := celebFight(t, down, 3)

	if fightQualifiesForCelebration(session, "player", minCelebrationSideSize) {
		t.Error("a fight where no enemy is dead or dying should not earn a celebration")
	}
}

func TestCelebrationSkipsForADeadWinner(t *testing.T) {
	// Nobody left standing at all. The winner is down too, so there's nobody to throw their arms up.
	down := celebDown{}
	session, side0, side1 := celebFight(t, down, 3)
	knockDown(t, down, session, side1)
	down["player"] = true

	if got := len(session.MembersOf(side0)); got != 3 {
		t.Fatalf("test setup: winning side has %d members, want 3", got)
	}
	if fightQualifiesForCelebration(session, "player", minCelebrationSideSize) {
		t.Error("a dead combatant should not earn a celebration")
	}
}

func TestCelebrationSkipsNonCombatant(t *testing.T) {
	down := celebDown{}
	session, _, side1 := celebFight(t, down, 3)
	knockDown(t, down, session, side1)

	if fightQualifiesForCelebration(session, "someone_who_was_never_here", minCelebrationSideSize) {
		t.Error("a character outside the fight should not earn a celebration for it")
	}
}

func TestCelebrationSurvivesTheOneTickCorpseGap(t *testing.T) {
	// The bug this rule was rewritten to fix. A combatant whose health has hit zero is already past
	// fighting, but is still a member of the session until their corpse is processed a tick later. So the
	// last enemy is still listed as a combatant at the moment the fight ends -- and the winner must be
	// recognised anyway. A settled-dead check would have said otherwise here.
	down := celebDown{}
	session, _, side1 := celebFight(t, down, 3)
	knockDown(t, down, session, side1)

	if got := len(session.MembersOf(side1)); got != 3 {
		t.Fatalf("test setup: losing side has %d members still in the session, want 3", got)
	}
	if got := len(session.FightingEnemiesOf("player")); got != 0 {
		t.Fatalf("test setup: %d enemies can still fight, want 0", got)
	}
	if !fightQualifiesForCelebration(session, "player", minCelebrationSideSize) {
		t.Error("a winner should be recognised while the defeated side's corpses are still registered")
	}
}
