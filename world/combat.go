package world

import (
	"fmt"

	"github.com/webbben/2d-game-engine/combat"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/utils"
)

// Combat sessions live here, on the World rather than on the ActiveMap, for two reasons: the player
// moving between maps shouldn't dissolve a fight, and NPCs that are off the player's map still have
// to be findable. Every entry point into combat funnels through StartCombat below, so who is fighting
// whom is decided in exactly one place -- an NPC reacting to being attacked can't tell whether it's
// starting a new fight or joining one, and isn't asked to.
//
// Design rationale is in issue #190.

// SessionFor returns the fight a character is currently involved in, or nil if they're not fighting
// anyone. The result is a read-only view: consumers can ask who's on which side, but only the World
// can change membership.
func (w *World) SessionFor(charID id.CharacterStateID) combat.SessionView {
	for _, s := range w.combatSessions {
		if s.IsCombatant(charID) {
			return s
		}
	}
	return nil
}

// IsInCombat reports whether a character is currently fighting anyone.
func (w *World) IsInCombat(charID id.CharacterStateID) bool {
	return w.SessionFor(charID) != nil
}

// EntityFor resolves a character to the entity driving them right now, or nil if they aren't in the
// world. Covers the player as well as NPCs.
//
// Combat needs this because the tasks that pick targets work in character IDs (that's what a session
// records) but the melee code works in entities: it needs collision rects and pixel positions, and
// EntityInfo carries neither. Resolving here rather than in the tasks keeps the player/npc split in
// one place instead of at every call site.
func (w *World) EntityFor(charID id.CharacterStateID) *entity.Entity {
	if charID == id.PlayerStateID {
		if w.Player == nil {
			return nil
		}
		return w.Player.Entity
	}
	n := w.getInWorldNPC(charID)
	if n == nil {
		return nil
	}
	return n.Entity
}

// StartCombat brings initiator and target into a fight, creating a session if one doesn't exist yet,
// and returns it (or the session they ended up in). Returns nil when nothing changed, which happens
// when the initiator is already committed elsewhere: a character may only be in one fight, so an
// attack by someone already engaged can't pull them into another.
//
// The cases, per the design in #190:
//
//   - initiator uncommitted, target committed     -> initiator joins the target's side of that fight
//   - initiator uncommitted, target uncommitted   -> new session; target founds the second side
//   - initiator committed                         -> ignored. The damage still landed; whether the
//     victim chooses to get involved is their decision, and there is no decision logic yet, so nothing
//     is recorded. See #190 for why merging two sessions is never the answer here.
func (w *World) StartCombat(initiator, target id.CharacterStateID, initiatorIntent, targetIntent combat.CombatIntent) combat.SessionView {
	if initiator == target {
		logz.Panicln("StartCombat", "character tried to start a fight with themselves", "character:", initiator)
	}

	initiatorSession := w.concreteSessionFor(initiator)
	targetSession := w.concreteSessionFor(target)

	switch {
	case initiatorSession == nil && targetSession == nil:
		// Nobody is committed yet: this is a new fight. The initiator is its aggressor, and the target
		// founds the opposing side.
		session := combat.NewCombatSession(w.nextCombatSessionID(), initiator, initiatorIntent, w)
		session.FoundSide(target, targetIntent)
		session.Validate()
		w.combatSessions = append(w.combatSessions, &session)
		logz.Println("StartCombat", "new fight:", session.SessionID(),
			"initiator:", initiator, "target:", target)
		return &session

	case targetSession == nil:
		// The initiator is already fighting and the target isn't, so the target gets drawn into the fight
		// already in progress. This is the ordinary case for one swing hitting several enemies: the first
		// hit starts a fight, and each later hit has to add its victim to that same fight.
		initiatorSession.Join(target, opposingSide(initiatorSession, initiator), targetIntent)
		initiatorSession.Validate()
		logz.Println("StartCombat", "target drawn into an existing fight:", target,
			"session:", initiatorSession.SessionID())
		return initiatorSession

	case initiatorSession == nil:
		// Symmetric to the above: a free character walks into a fight already underway, on the side
		// opposite whoever they came for.
		targetSession.Join(initiator, opposingSide(targetSession, target), initiatorIntent)
		targetSession.Validate()
		logz.Println("StartCombat", "initiator joined an existing fight:", initiator,
			"session:", targetSession.SessionID())
		return targetSession

	default:
		// Both are already committed, so this is two people who are already fighting coming to blows.
		// The fight doesn't change -- the victim decides for themselves whether to get involved, and
		// there's no decision logic yet. Merging their two fights is never the answer; see #190.
		logz.Println("StartCombat", "both characters are already fighting; recording nothing",
			"initiator:", initiator, "target:", target)
		return nil
	}
}

// opposingSide returns the side of a session that isn't the given combatant's side.
//
// A live session has exactly two sides, so there's always exactly one other side to name.
func opposingSide(session *combat.CombatSession, of id.CharacterStateID) combat.SideID {
	theirSide, ok := session.SideOf(of)
	if !ok {
		logz.Panicln("StartCombat", "character isn't a combatant of the session we were given",
			"session:", session.SessionID(), "character:", of)
	}
	for _, side := range session.Sides() {
		if side != theirSide {
			return side
		}
	}
	logz.Panicln("StartCombat", "session had no opposing side to return", "session:", session.SessionID(), "character:", of)
	return ""
}

// LeaveCombat takes a character out of whatever fight they're in, and drops the fight entirely once
// one of its sides has been emptied. Callers are responsible for leaving on every way out of a fight
// -- death, being removed from the map, walking away -- since HasEnded is judged by membership.
func (w *World) LeaveCombat(charID id.CharacterStateID) {
	session := w.concreteSessionFor(charID)
	if session == nil {
		return
	}
	session.Leave(charID)
	if !session.HasEnded() {
		return
	}
	// fight is over: drop the session so nothing can find it again
	for i, s := range w.combatSessions {
		if s == session {
			w.combatSessions = append(w.combatSessions[:i], w.combatSessions[i+1:]...)
			logz.Println("LeaveCombat", "fight ended:", session.SessionID())
			return
		}
	}
	logz.Panicln("LeaveCombat", "session reported ended but wasn't in the world's session list",
		"session:", session.SessionID(), "character:", charID)
}

// concreteSessionFor is SessionFor without the read-only interface, for the internal calls that
// actually mutate a session. Only this file should use it.
func (w *World) concreteSessionFor(charID id.CharacterStateID) *combat.CombatSession {
	for _, s := range w.combatSessions {
		if s.IsCombatant(charID) {
			return s
		}
	}
	return nil
}

// nextCombatSessionID mints the next session ID. A plain counter, so IDs sort in the order fights
// started -- "combat-7 started before combat-8" is readable straight off a log -- without carrying
// any context that could go stale or leak into saves.
func (w *World) nextCombatSessionID() combat.CombatSessionID {
	w.combatSessionSeq++
	return combat.CombatSessionID(fmt.Sprintf("combat-%d", w.combatSessionSeq))
}

// The three methods below are how World satisfies combat.Resolver: they answer what state a
// combatant is in, which a session can't work out for itself. They're here rather than on the NPC
// deliberately -- an NPC off the player's map has a nil ActiveMapCtx but a live WorldCtx, so a
// session stays able to ask about it.
//
// An accepted surrender is deliberately not among them: by then the character has left the fight,
// so that's a LeaveCombat call rather than a question to ask.

var _ combat.Resolver = (*World)(nil)

// IsDead reports whether a character has died.
func (w *World) IsDead(charID id.CharacterStateID) bool {
	utils.PanicAssert(charID != "", "id was empty")
	return w.Dataman.GetCharacterState(charID).Dead
}

// IsOfferingSurrender reports whether an NPC has an unaccepted offer on the table. The player never
// does -- they don't surrender to anyone -- and neither does anyone who isn't in the world.
func (w *World) IsOfferingSurrender(charID id.CharacterStateID) bool {
	n := w.getInWorldNPC(charID)
	if n == nil {
		return false
	}
	return n.IsOfferingSurrender()
}

// IsFleeing reports whether an NPC is currently running from their enemies. The player never flees.
func (w *World) IsFleeing(charID id.CharacterStateID) bool {
	n := w.getInWorldNPC(charID)
	if n == nil {
		return false
	}
	return n.IsFleeing()
}

// pruneEndedCombatSessions drops any session that has run out of combatants on one side. Called from
// the world update loop so a fight ends even if the character who ended it never went through
// LeaveCombat -- a dead NPC, for instance.
func (w *World) pruneEndedCombatSessions() {
	kept := w.combatSessions[:0]
	for _, s := range w.combatSessions {
		if s.HasEnded() {
			logz.Println("pruneEndedCombatSessions", "dropping ended fight:", s.SessionID())
			continue
		}
		kept = append(kept, s)
	}
	w.combatSessions = kept
}
