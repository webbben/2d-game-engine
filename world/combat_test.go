package world

import (
	"slices"
	"testing"

	"github.com/webbben/2d-game-engine/combat"
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

	session := w.StartCombat("player", "bandit_a", combat.IntentKill, combat.IntentSelfDefense)
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
	existing := w.StartCombat("player", "bandit_a", combat.IntentKill, combat.IntentSelfDefense)

	// a bystander attacks one of the bandits: pulled into the fight, opposite side from their victim
	joined := w.StartCombat("guard", "bandit_a", combat.IntentProtect, combat.IntentKill)
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
	w.StartCombat("player", "bandit_a", combat.IntentKill, combat.IntentSelfDefense)
	before := len(w.combatSessions)

	// The player swings again and hits someone new. The player is already committed, but that must not
	// stop the new victim joining: they're now an enemy too.
	got := w.StartCombat("player", "bandit_b", combat.IntentKill, combat.IntentKill)
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
	w.StartCombat("player", "bandit_a", combat.IntentKill, combat.IntentSelfDefense)
	w.StartCombat("guard", "bandit_a", combat.IntentProtect, combat.IntentKill)

	// player and bandit_a are both already committed; they're coming to blows, but the session doesn't
	// change -- the victim decides for themselves whether to get involved, and that logic doesn't exist yet.
	before := len(w.combatSessions)
	got := w.StartCombat("bandit_a", "player", combat.IntentSelfDefense, combat.IntentKill)
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
	w.StartCombat("player", "player", combat.IntentKill, combat.IntentSelfDefense)
}

func TestLeaveCombatEndsSession(t *testing.T) {
	w := &World{}
	w.StartCombat("player", "bandit_a", combat.IntentKill, combat.IntentSelfDefense)
	w.StartCombat("guard", "bandit_a", combat.IntentProtect, combat.IntentKill)

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
	live := w.StartCombat("player", "bandit_a", combat.IntentKill, combat.IntentSelfDefense)

	// a session whose second side never got filled has already ended
	w.StartCombat("guard", "merchant", combat.IntentProtect, combat.IntentKill)
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
	first := w.StartCombat("player", "bandit_a", combat.IntentKill, combat.IntentSelfDefense)
	w.LeaveCombat("player")
	w.LeaveCombat("bandit_a")
	second := w.StartCombat("player", "bandit_b", combat.IntentKill, combat.IntentSelfDefense)

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
	w.StartCombat("player", "bandit_a", combat.IntentKill, combat.IntentSelfDefense)
	w.StartCombat("player", "bandit_b", combat.IntentKill, combat.IntentSelfDefense)

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
