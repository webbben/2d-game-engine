package npc

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/webbben/2d-game-engine/data/id"
)

// Recent-attacker tracking.
//
// "I was just hit by X" is a fact about this NPC rather than about whatever task it happens to be
// running, so it lives on the NPC. That placement is load-bearing rather than tidier: a hit normally
// arrives before a FightTask exists at all (OnAttacked runs against whatever schedule task the NPC was
// on, and StartCombat only opens the session), so task-scoped state would be dropped at exactly the
// moment it mattered and the FightTask that responds would never learn who struck first.
//
// FightTask reads this via recentlyAttackedBy to build targetCandidate.AttackedMe.

const (
	// ticksPerSecond is ebiten's fixed tick rate, used to express windows in seconds. Mirrors the
	// constant in entity/combat.go; it's unexported there, so it can't be shared.
	ticksPerSecond = 60

	// recentAttackerWindowTicks is how long a hit keeps an attacker relevant. Long enough to hold a
	// target through a wind-up, a swing, and the recovery after it; short enough that priority decays
	// once someone genuinely stops engaging.
	recentAttackerWindowTicks = 2 * ticksPerSecond

	// maxRecentAttackers bounds the list. Only the ones that matter are the handful actually hitting
	// this NPC right now, and a fixed cap means the record can't grow without limit in a long fight.
	maxRecentAttackers = 4
)

type recentAttacker struct {
	id   id.CharacterStateID
	tick int64
}

// recordAttacker notes that attackerID just hit this NPC.
func (n *NPC) recordAttacker(attackerID id.CharacterStateID) {
	now := ebiten.Tick()
	// Drop expired entries. An attacker already on the list is dropped too, and re-added at the front
	// below, so a repeat hit refreshes its timestamp and its place in the list rather than occupying
	// a second slot.
	kept := n.recentAttackers[:0]
	for _, a := range n.recentAttackers {
		if a.id != attackerID && tickAgeIsWithin(now, a.tick, recentAttackerWindowTicks) {
			kept = append(kept, a)
		}
	}
	n.recentAttackers = append([]recentAttacker{{id: attackerID, tick: now}}, kept...)
	if len(n.recentAttackers) > maxRecentAttackers {
		n.recentAttackers = n.recentAttackers[:maxRecentAttackers]
	}
}

// recentlyAttackedBy reports whether attackerID landed a hit on this NPC within the recent window.
func (n *NPC) recentlyAttackedBy(attackerID id.CharacterStateID) bool {
	now := ebiten.Tick()
	for _, a := range n.recentAttackers {
		if a.id == attackerID && tickAgeIsWithin(now, a.tick, recentAttackerWindowTicks) {
			return true
		}
	}
	return false
}

// hasRecentAttacker reports whether anyone has hit this NPC inside the window. FightTask uses this to
// skip its cheap "keep what I have" path: while nobody has been hit, every candidate is in the same
// priority tier and the cheap answer is provably the one the full policy would give.
func (n *NPC) hasRecentAttacker() bool {
	now := ebiten.Tick()
	for _, a := range n.recentAttackers {
		if tickAgeIsWithin(now, a.tick, recentAttackerWindowTicks) {
			return true
		}
	}
	return false
}

// tickAgeIsWithin reports whether now - then is inside window, treating a negative age as out of range.
// A negative age means the tick counter wrapped (see FinishMeleeAttack, which handles the same thing);
// reading that as "not recent" fails toward keeping the current target, which is the safe direction.
func tickAgeIsWithin(now, then int64, window int64) bool {
	age := now - then
	return age >= 0 && age <= window
}
