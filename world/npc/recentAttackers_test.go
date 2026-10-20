package npc

import (
	"testing"

	"github.com/webbben/2d-game-engine/data/id"
)

// Tests for the recent-attacker bookkeeping that feeds targetCandidate.AttackedMe.
//
// ebiten.Tick() reads 0 outside a running game, so "now" is pinned here. Age is then controlled by
// writing negative tick values into the entries, which is how the expiry cases below express "this hit
// landed N ticks ago" without needing a running loop.

// attackerIDs is just the record contents, in list order, for readable failures.
func attackerIDs(n *NPC) []id.CharacterStateID {
	out := make([]id.CharacterStateID, 0, len(n.recentAttackers))
	for _, a := range n.recentAttackers {
		out = append(out, a.id)
	}
	return out
}

func TestRecordAttackerKeepsOneEntryPerAttacker(t *testing.T) {
	n := &NPC{}

	n.recordAttacker("bandit_a")
	n.recordAttacker("bandit_a")
	n.recordAttacker("bandit_a")

	if got := len(n.recentAttackers); got != 1 {
		t.Errorf("expected a repeated attacker to occupy one slot, got %d: %v", got, attackerIDs(n))
	}
	if !n.recentlyAttackedBy("bandit_a") {
		t.Error("expected bandit_a to count as a recent attacker")
	}
}

func TestRecordAttackerPutsMostRecentFirst(t *testing.T) {
	n := &NPC{}

	n.recordAttacker("bandit_a")
	n.recordAttacker("bandit_b")
	n.recordAttacker("bandit_c")

	// Most recent first, so truncating to the cap drops the stalest attackers rather than the newest.
	got := attackerIDs(n)
	want := []id.CharacterStateID{"bandit_c", "bandit_b", "bandit_a"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestRecordAttackerRefreshingMovesToFront(t *testing.T) {
	n := &NPC{}

	n.recordAttacker("bandit_a")
	n.recordAttacker("bandit_b")
	n.recordAttacker("bandit_a") // a takes another swing

	got := attackerIDs(n)
	if len(got) != 2 || got[0] != "bandit_a" {
		t.Errorf("expected a refreshed attacker at the front, got %v", got)
	}
}

// The cap exists so a long fight can't grow the list without bound. Truncating from the front drops
// the least recently struck, which is the right thing to lose.
func TestRecordAttackerCapsTheList(t *testing.T) {
	n := &NPC{}

	for _, id := range []id.CharacterStateID{"a", "b", "c", "d", "e", "f"} {
		n.recordAttacker(id)
	}

	if got := len(n.recentAttackers); got != maxRecentAttackers {
		t.Fatalf("expected the list to cap at %d, got %d: %v", maxRecentAttackers, got, attackerIDs(n))
	}
	if n.recentAttackers[0].id != "f" {
		t.Errorf("expected the newest attacker at the front, got %v", attackerIDs(n))
	}
	for _, evicted := range []id.CharacterStateID{"a", "b"} {
		if n.recentlyAttackedBy(evicted) {
			t.Errorf("expected %s to have been evicted by the cap, but it still reads as recent", evicted)
		}
	}
}

func TestRecentlyAttackedByExpires(t *testing.T) {
	n := &NPC{}
	n.recordAttacker("bandit_a")

	if !n.recentlyAttackedBy("bandit_a") {
		t.Fatal("a hit recorded this tick should read as recent")
	}

	// now is 0, so a tick of -200 means the hit landed 200 ticks ago: past the window.
	n.recentAttackers[0].tick = -200

	if n.recentlyAttackedBy("bandit_a") {
		t.Error("expected a hit older than the window to stop counting as recent")
	}
	if n.hasRecentAttacker() {
		t.Error("hasRecentAttacker should also report false once every entry has expired")
	}
}

func TestHasRecentAttacker(t *testing.T) {
	n := &NPC{}

	if n.hasRecentAttacker() {
		t.Error("an NPC nobody has hit should have no recent attackers")
	}
	n.recordAttacker("bandit_a")
	if !n.hasRecentAttacker() {
		t.Error("expected a recent attacker to be reported")
	}
	if n.recentlyAttackedBy("bandit_b") {
		t.Error("a different character should not read as an attacker just because someone was hit")
	}
}

// Ticks wrap. A negative age means the counter overflowed, and reading that as "recent" would let a
// very old hit keep an attacker in the priority tier indefinitely.
func TestTickAgeTreatsWraparoundAsNotRecent(t *testing.T) {
	if tickAgeIsWithin(10, 50, recentAttackerWindowTicks) {
		t.Error("a negative age (tick wraparound) should not count as within the window")
	}
	if !tickAgeIsWithin(10, 10, recentAttackerWindowTicks) {
		t.Error("a hit recorded this tick should be within the window")
	}
	if tickAgeIsWithin(1000, 0, recentAttackerWindowTicks) {
		t.Error("an age past the window should not count as within it")
	}
}
