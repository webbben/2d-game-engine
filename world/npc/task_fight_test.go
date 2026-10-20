package npc

import (
	"testing"

	"github.com/webbben/2d-game-engine/data/id"
)

// Tests for the target-selection policy of FightTask (selectTargetID / targetIsValid).
//
// These exist to pin down the decisions we're confident shouldn't change silently. They're plain-data
// tests on purpose: the policy is deliberately kept in terms of plain candidate values so it can be
// exercised without standing up an npc, a world, or physics. See targetCandidate.
//
// Each case carries a comment explaining *why* it's correct, so that when we later extend the policy
// (traits influencing who we chase, etc.) we can tell which cases are load-bearing rules and which
// were incidental. When a rule genuinely changes on purpose, change it here too -- the point of these
// tests is to make that a deliberate act, not to freeze current behavior no matter what.

// valid builds a candidate that is unambiguously worth fighting, so each test case only has to state
// the one field it actually cares about.
func valid(id id.CharacterStateID, dist float64) targetCandidate {
	return targetCandidate{ID: id, Dist: dist, Alive: true, InSession: true, InWorld: true}
}

func TestTargetIsValid(t *testing.T) {
	tests := []struct {
		name      string
		candidate targetCandidate
		want      bool
	}{
		{
			name:      "a living enemy still in the fight and still on the map is valid",
			candidate: valid("bandit", 10),
			want:      true,
		},
		{
			// We stop fighting the moment someone dies, so a corpse is never a candidate -- even if the
			// session hasn't caught up and still lists them. Selecting one would mean the npc stands
			// there swinging at nothing.
			name:      "a dead enemy is not valid",
			candidate: targetCandidate{ID: "bandit", Dist: 1, Alive: false, InSession: true, InWorld: true},
			want:      false,
		},
		{
			// Someone who fled, surrendered and was released, or whose fight ended leaves the session.
			// Chasing them would pull the npc out of the fight it was told to be in.
			name:      "an enemy who left the session is not valid",
			candidate: targetCandidate{ID: "bandit", Dist: 1, Alive: true, InSession: false, InWorld: true},
			want:      false,
		},
		{
			// An enemy can be unloaded from the map without ever leaving the session (map change, corpse
			// cleanup, despawn). Nothing is there to hit, so we must not pick them.
			name:      "an enemy who left the world without leaving the session is not valid",
			candidate: targetCandidate{ID: "bandit", Dist: 1, Alive: true, InSession: true, InWorld: false},
			want:      false,
		},
		{
			// Deliberate: yielding is not disqualifying. An enemy who offers surrender is still a
			// legitimate target, because we're free to refuse them. If this ever becomes false, an npc
			// would abandon a duel the instant the loser yielded, which we don't want.
			name:      "an enemy offering to surrender is still valid (no surrender flag exists yet)",
			candidate: valid("bandit", 10),
			want:      true,
		},
		{
			// Someone who broke off and is running is only worth following if this npc would pursue
			// them. Both fields must be set together: WorthPursuing on its own is meaningless, because a
			// standing enemy never has the question asked.
			name:      "a fleeing enemy we're willing to pursue is still valid",
			candidate: targetCandidate{ID: "bandit", Dist: 10, Alive: true, InSession: true, InWorld: true, Fleeing: true, WorthPursuing: true},
			want:      true,
		},
		{
			// The whole point of pursuit. When nobody wants to chase a runner, that runner stops being
			// a target -- and once this npc has no other enemy, the fight is over for them.
			name:      "a fleeing enemy we're unwilling to pursue is not valid",
			candidate: targetCandidate{ID: "bandit", Dist: 10, Alive: true, InSession: true, InWorld: true, Fleeing: true, WorthPursuing: false},
			want:      false,
		},
		{
			// Pursuit can't resurrect someone already gone. Deciding not to chase a fleeing enemy adds a
			// second reason to drop them, not a replacement for the first.
			name:      "a dead fleeing enemy is not valid regardless of willingness to pursue",
			candidate: targetCandidate{ID: "bandit", Dist: 10, Alive: false, InSession: true, InWorld: true, Fleeing: true, WorthPursuing: true},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := targetIsValid(tt.candidate); got != tt.want {
				t.Errorf("targetIsValid(%+v) = %v, want %v", tt.candidate, got, tt.want)
			}
		})
	}
}

func TestSelectTargetIDKeepsCurrentTarget(t *testing.T) {
	tests := []struct {
		name       string
		current    id.CharacterStateID
		candidates []targetCandidate
		want       id.CharacterStateID
	}{
		{
			// The core anti-thrash rule. An enemy stepping a little closer is not a reason to walk away
			// mid-swing; without this, npcs dither between two similar enemies and read as broken.
			name:       "keeps the current target even when a nearer enemy exists",
			current:    "bandit_a",
			candidates: []targetCandidate{valid("bandit_a", 40), valid("bandit_b", 5)},
			want:       "bandit_a",
		},
		{
			// Same rule, even when the current target is the farthest thing around. Distance is only
			// consulted when we have no target, never as a reason to abandon a valid one.
			name:       "keeps the current target even when it is the farthest enemy",
			current:    "bandit_a",
			candidates: []targetCandidate{valid("bandit_a", 500), valid("bandit_b", 5)},
			want:       "bandit_a",
		},
		{
			// Nothing was being fought yet (fresh task, or the target pointer is empty), so the nearest
			// enemy wins. This is the normal way a target gets chosen.
			name:       "with no current target, picks the nearest valid enemy",
			current:    "",
			candidates: []targetCandidate{valid("bandit_a", 90), valid("bandit_b", 12)},
			want:       "bandit_b",
		},
		{
			// The current target died, so we must move on -- but only to someone actually worth hitting.
			// This is the common case after killing one bandit in a group.
			name:       "current target died, switches to the nearest survivor",
			current:    "bandit_a",
			candidates: []targetCandidate{targetCandidate{ID: "bandit_a", Alive: false, InSession: true, InWorld: true}, valid("bandit_b", 30), valid("bandit_c", 60)},
			want:       "bandit_b",
		},
		{
			// The current target gave up and was released, so they're out of the fight. This is what ends
			// a duel cleanly once the player accepts a surrender.
			name:       "current target left the session, switches to the nearest survivor",
			current:    "bandit_a",
			candidates: []targetCandidate{targetCandidate{ID: "bandit_a", Alive: true, InSession: false, InWorld: true}, valid("bandit_b", 30)},
			want:       "bandit_b",
		},
		{
			// This is the original bug this policy exists to prevent: a single swing hitting several
			// bandits must leave every one of them as a candidate, and once we drop one we must still
			// see the rest rather than falling back to "last one hit".
			name:       "after dropping one of several enemies, another is still chosen",
			current:    "bandit_a",
			candidates: []targetCandidate{targetCandidate{ID: "bandit_a", Alive: false, InSession: true, InWorld: true}, valid("bandit_b", 80), valid("bandit_c", 45)},
			want:       "bandit_c",
		},
		{
			// The moment a target breaks off, the npc reconsiders rather than finishing the duel. This
			// is the pursuit decision arriving as a mid-fight switch, and it's the reason eligibility
			// isn't just a starting condition.
			name:    "current target starts fleeing and we won't pursue, switches to a standing enemy",
			current: "bandit_a",
			candidates: []targetCandidate{
				{ID: "bandit_a", Dist: 5, Alive: true, InSession: true, InWorld: true, Fleeing: true, WorthPursuing: false},
				valid("bandit_b", 80),
			},
			want: "bandit_b",
		},
		{
			// Same situation, opposite answer. Someone who'll give chase keeps hitting the person who's
			// running, even with a closer alternative available.
			name:    "current target starts fleeing but we will pursue, keeps hitting them",
			current: "bandit_a",
			candidates: []targetCandidate{
				{ID: "bandit_a", Dist: 5, Alive: true, InSession: true, InWorld: true, Fleeing: true, WorthPursuing: true},
				valid("bandit_b", 80),
			},
			want: "bandit_a",
		},
		{
			// Everyone left is running and none of them are worth chasing. Returning empty is how the
			// task learns to stand down rather than chase someone out of the map.
			name:    "every enemy is fleeing and none are worth pursuing",
			current: "bandit_a",
			candidates: []targetCandidate{
				{ID: "bandit_a", Dist: 5, Alive: true, InSession: true, InWorld: true, Fleeing: true, WorthPursuing: false},
				{ID: "bandit_b", Dist: 7, Alive: true, InSession: true, InWorld: true, Fleeing: true, WorthPursuing: false},
			},
			want: "",
		},
		{
			// A closer corpse must never win just for being closest -- that's how an npc ends up swinging
			// at a body. Distance only gets compared among valid candidates.
			name:       "skips a nearer invalid candidate in favour of the nearest valid one",
			current:    "",
			candidates: []targetCandidate{targetCandidate{ID: "corpse", Dist: 2, Alive: false, InSession: true, InWorld: true}, valid("bandit_a", 70)},
			want:       "bandit_a",
		},
		{
			// An enemy unloaded from the world can still be listed in the session. Treat it as absent
			// rather than as a target at distance zero.
			name:       "skips a candidate that left the world without leaving the session",
			current:    "",
			candidates: []targetCandidate{targetCandidate{ID: "ghost", Dist: 3, Alive: true, InSession: true, InWorld: false}, valid("bandit_a", 70)},
			want:       "bandit_a",
		},
		{
			// No candidates at all: everyone is dead or gone. Returning "" is how pickTarget reports
			// "no target", which is how the fight task knows to end itself instead of spinning forever.
			name:       "returns empty when there are no candidates",
			current:    "bandit_a",
			candidates: nil,
			want:       "",
		},
		{
			// Same signal, reached a different way: candidates exist, but none of them are worth hitting.
			// This is the normal way a fight ends after the last enemy dies or flees.
			name:       "returns empty when every candidate is invalid",
			current:    "bandit_a",
			candidates: []targetCandidate{targetCandidate{ID: "bandit_a", Alive: false, InSession: true, InWorld: true}},
			want:       "",
		},
		{
			// An empty candidate list means the fight is over, so even a "current" target must not be
			// kept. Returning it here would leave the npc fighting a session that no longer exists.
			name:       "does not keep a current target that is absent from the candidate list",
			current:    "bandit_a",
			candidates: []targetCandidate{valid("bandit_b", 30)},
			want:       "bandit_b",
		},
		{
			// Two candidates at exactly the same distance. Ties are broken by whichever comes first in
			// the list; the session's ordering is arbitrary, so we don't claim a winner. This case
			// exists to document that a tie is not an error rather than to lock in a particular choice.
			name:       "ties are resolved by candidate order, not by preferring a lower ID",
			current:    "",
			candidates: []targetCandidate{valid("bandit_b", 25), valid("bandit_a", 25)},
			want:       "bandit_b",
		},
		{
			// Being hit has to outrank "closest". The previous rule kept any valid target, so an npc
			// would walk past a bandit swinging at it to keep chasing the nearest stranger. Being struck
			// is a settled fact; distance is only where someone happens to be in their approach.
			name:    "an attacker outranks a nearer enemy we haven't been hit by",
			current: "",
			candidates: []targetCandidate{
				{ID: "bandit_a", Dist: 120, Alive: true, InSession: true, InWorld: true, AttackedMe: true},
				valid("bandit_b", 12),
			},
			want: "bandit_a",
		},
		{
			// The reported bug: hit from behind while chasing a runaway. Chasing someone who is leaving
			// is discretionary, and ignoring a landed hit to keep doing it is the behaviour that made no
			// sense in testing.
			name:    "hit while chasing a fleeing target, turns on the attacker",
			current: "bandit_a",
			candidates: []targetCandidate{
				{ID: "bandit_a", Dist: 40, Alive: true, InSession: true, InWorld: true, Fleeing: true, WorthPursuing: true},
				{ID: "bandit_b", Dist: 200, Alive: true, InSession: true, InWorld: true, AttackedMe: true},
			},
			want: "bandit_b",
		},
		{
			// The same interruption while dueling someone who is actually fighting back. We stop ignoring
			// the hits entirely, not just when the current target happens to be running.
			name:    "hit while fighting a standing enemy, turns on the attacker",
			current: "bandit_a",
			candidates: []targetCandidate{
				valid("bandit_a", 16),
				{ID: "bandit_b", Dist: 30, Alive: true, InSession: true, InWorld: true, AttackedMe: true},
			},
			want: "bandit_b",
		},
		{
			// The anti-thrash rule, and the whole reason stickiness is scoped to the top tier rather than
			// dropped. Outnumbered and being hit from two sides, the npc switches to the first attacker
			// and then holds them: the second attacker joins the tier but doesn't displace a target
			// already in it. Without this, alternating hits would ping-pong the target every swing.
			name:    "a second attacker joining the tier does not displace our current attacker",
			current: "bandit_b",
			candidates: []targetCandidate{
				{ID: "bandit_a", Dist: 16, Alive: true, InSession: true, InWorld: true, AttackedMe: true},
				{ID: "bandit_b", Dist: 90, Alive: true, InSession: true, InWorld: true, AttackedMe: true},
			},
			want: "bandit_b",
		},
		{
			// Validity is decided before tiering, so a corpse that hit us a moment ago can't form a tier
			// by itself and leave us with nobody to fight while someone alive is standing right there.
			name:    "a dead attacker does not form an empty tier over a living enemy",
			current: "",
			candidates: []targetCandidate{
				{ID: "bandit_a", Dist: 10, Alive: false, InSession: true, InWorld: true, AttackedMe: true},
				valid("bandit_b", 40),
			},
			want: "bandit_b",
		},
		{
			// Tiering must not resurrect targets we already refuse to chase. A fleeing enemy we won't
			// pursue is invalid regardless of whether they're also the one hitting us.
			name:    "a fleeing attacker we won't pursue is still not a target",
			current: "",
			candidates: []targetCandidate{
				{ID: "bandit_a", Dist: 10, Alive: true, InSession: true, InWorld: true, Fleeing: true, WorthPursuing: false, AttackedMe: true},
				valid("bandit_b", 90),
			},
			want: "bandit_b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := selectTargetID(tt.current, tt.candidates)
			if got != tt.want {
				t.Errorf("selectTargetID(current=%q, %+v) = %q, want %q", tt.current, tt.candidates, got, tt.want)
			}
		})
	}
}
