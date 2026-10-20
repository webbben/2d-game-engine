package world

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/webbben/2d-game-engine/combat"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/entity"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
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

// CombatProfile returns a combatant's resolved profile within a session: their fighting ability, their
// condition, their personality drives, what they're trying to accomplish, and whether they're running.
//
// The cache lives on the session rather than on any one observer, so five NPCs fighting the same guard
// resolve that guard once between them rather than five times over. Skills, attributes, drives, and
// intent are fixed for the length of a fight and are resolved once. Health and Fleeing are live and are
// refreshed on every read, because both change while the fight runs -- and Fleeing in particular flips
// at exactly the moment a pursuit decision has to be made.
//
// A session that isn't found, or a character who isn't in it, yields a zero profile rather than an
// error. Both are ordinary: sessions are pruned once ended, so a fight can finish between one caller
// resolving a profile and another using it.
//
// A corpse is a third case of the same thing -- a character who is present in the session but no longer
// participating -- and it is handled differently on purpose. Every caller here is asking a combat
// decision question (should this npc break off, would it chase that one), which has no meaningful
// answer for someone already dead, so this panics rather than answering. Returning a zero profile
// instead would be worse than an error: combatConfidence would score it 0 and fleeThreat would compute
// nonsense, and a real flee decision would be made from that garbage, silently. Callers must filter
// dead characters out first, with characterstate.IsDeadOrDying -- which is the predicate this assert
// itself uses, so the two cannot disagree about what "dead" means.
func (w *World) CombatProfile(sessionID combat.CombatSessionID, charStateID id.CharacterStateID) defs.CombatProfile {
	session := w.concreteSessionByID(sessionID)
	if session == nil {
		return defs.CombatProfile{ID: charStateID}
	}

	if characterstate.IsDeadOrDying(charStateID, w.Dataman) {
		logz.PanicCtx("CombatProfile", "a dead character was passed in; callers must filter dead characters out first.", charStateID, "sessionID:", sessionID)
	}

	charState := w.Dataman.GetCharacterState(charStateID)

	// Read live before the cache check, so both paths see the same value.
	fleeing := w.IsFleeing(charStateID)

	if cached, ok := session.CachedProfile(charStateID); ok {
		cached.Health = charState.Health
		cached.MaxHealth = charState.MaxHealth
		cached.Fleeing = fleeing
		session.StoreProfile(charStateID, cached)
		return cached
	}

	skills, attrs := characterstate.GetSkillsAndAttributes(charStateID, w.Dataman)

	// Cloned deliberately. GetSkillsAndAttributes hands back maps held in the DataManager's cache,
	// shared with every other reader, and a profile outliving this fight would keep them writable after
	// that cache entry moved on. One clone per combatant per fight is cheap.
	profile := defs.CombatProfile{
		ID:              charStateID,
		SkillLevels:     maps.Clone(skills),
		AttributeLevels: maps.Clone(attrs),
		Drives:          maps.Clone(characterstate.GetNetDrives(charState.Traits, w.Dataman)),
		Health:          charState.Health,
		MaxHealth:       charState.MaxHealth,
		Fleeing:         fleeing,
		// Snapshotted by the session at join, so empty means this character has no recorded reason to
		// be fighting -- which is the normal case for the player.
		Intent: session.IntentOrEmpty(charStateID),
	}
	session.StoreProfile(charStateID, profile)
	return profile
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
func (w *World) StartCombat(initiator, target id.CharacterStateID, initiatorIntent, targetIntent defs.CombatIntent) combat.SessionView {
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
	w.endSession(session)
}

// endSession retires a finished fight: it drops the session so nothing can find it again.
//
// This is the one place a session leaves the world. Both ways out of a fight come through here, because
// HasEnded is judged by membership and callers leave for different reasons -- dying, walking off the
// map, accepting a surrender, or standing down. A fight that could be retired by only one of those
// paths would be a bug waiting to be found, so nothing gets to retire a session on its own.
//
// Note what this deliberately does NOT do: work out who won. It can't, by the time it runs. See
// WonCelebratableFight.
func (w *World) endSession(session *combat.CombatSession) {
	// fight is over: drop the session so nothing can find it again
	for i, s := range w.combatSessions {
		if s == session {
			w.combatSessions = append(w.combatSessions[:i], w.combatSessions[i+1:]...)
			logz.Println("endSession", "fight ended:", session.SessionID())
			return
		}
	}
	logz.Panicln("endSession", "session reported ended but wasn't in the world's session list",
		"session:", session.SessionID())
}

// fightQualifiesForCelebration reports whether charID won a fight big enough to earn a victory lap.
//
// The winner is established by the caller standing down with nothing left to fight, which has to be
// asked *before* it leaves: a combatant that gives up its seat in the session is how a fight ends, so
// by the time the session retires the winning side has emptied itself and the only side left standing
// is the losing one. Asking "which side survived?" too late picks the wrong side.
//
// Three conditions, all required:
//
//   - charID is still in the fight and able to stand there, which is what rules out a corpse or a
//     mutual annihilation asking to celebrate.
//   - Every enemy is dead or dying. Running isn't losing to them, so a rout where the other side
//     fled or surrendered earns nothing.
//   - Both sides reached the size threshold at their peak. Judged on the peak rather than the finish, so
//     a fight that ground down to a last survivor still counts as a battle -- see MaxSizeOf.
//
// Free function rather than a method so the rule can be tested without standing up a World or a
// DataManager; it needs nothing but the session.
func fightQualifiesForCelebration(session *combat.CombatSession, charID id.CharacterStateID, minSideSize int) bool {
	if !session.IsCombatant(charID) || session.IsDeadOrDying(charID) {
		return false
	}
	if len(session.FightingEnemiesOf(charID)) != 0 {
		return false
	}
	// AllSides rather than Sides: the defeated side may already be empty, so it can be missing from
	// Sides and still be one whose size has to be checked.
	for _, side := range session.AllSides() {
		if session.MaxSizeOf(side) < minSideSize {
			return false
		}
	}
	return true
}

// WonCelebratableFight reports whether charID has just won a fight worth celebrating.
//
// Asked by a combatant as it stands down, so this has to be answered while the fight is still on record.
func (w *World) WonCelebratableFight(charID id.CharacterStateID) bool {
	session := w.concreteSessionFor(charID)
	if session == nil {
		// The fight is already over and gone -- someone else's departure retired it.
		return false
	}
	if !fightQualifiesForCelebration(session, charID, config.MinCelebrationSideSize) {
		return false
	}
	logz.Println("WonCelebratableFight", charID, "won a fight worth celebrating. session:", session.SessionID())
	return true
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

// concreteSessionByID finds a session by its ID, without the read-only interface. Same rule as
// concreteSessionFor: only this file uses it, for the calls that need to mutate.
func (w *World) concreteSessionByID(sessionID combat.CombatSessionID) *combat.CombatSession {
	for _, s := range w.combatSessions {
		if s.SessionID() == sessionID {
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
//
// Deliberately narrower than characterstate.IsDeadOrDying and Entity.IsDead, both of which also treat
// "health already at zero" as dead. This one asks the settled question, because its only consumer is
// CombatSession.LivingMembers, where the answer is about who has actually been removed from the fight
// -- a character at zero health whose corpse has not been processed yet is still a combatant.
//
// So this is the wrong check for "can this character still act" or "should this be a candidate
// target". Use IsDeadOrDying for that; combat does, and so should anything else walking that line.
func (w *World) IsDead(charID id.CharacterStateID) bool {
	utils.PanicAssert(charID != "", "id was empty")
	return w.Dataman.GetCharacterState(charID).Dead
}

// IsDeadOrDying reports whether a character is past being able to fight.
//
// The question World.IsDead is documented as being the wrong one for. IsDead is settled: it goes true
// once a corpse has actually been processed. This one goes true the moment health hits zero, which is
// the tick or so earlier.
//
// That gap is small and easy to dismiss, and it is exactly what makes the difference when asking whether
// anyone's left in a fight: a body still sitting in the session isn't someone who can still hit you.
func (w *World) IsDeadOrDying(charID id.CharacterStateID) bool {
	utils.PanicAssert(charID != "", "id was empty")
	return characterstate.IsDeadOrDying(charID, w.Dataman)
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
	// Cloned because endSession removes from w.combatSessions, and doing that while walking it would
	// both skip sessions and read past the end of the slice. Ending fights is its job; this only decides
	// which ones are ready for it.
	for _, s := range slices.Clone(w.combatSessions) {
		if s.HasEnded() {
			logz.Println("pruneEndedCombatSessions", "dropping ended fight:", s.SessionID())
			w.endSession(s)
		}
	}
}

func (w *World) CombatDebugInfo(sb *strings.Builder) {
	sb.WriteString("COMBAT SESSIONS:\n")
	for _, cs := range w.combatSessions {
		cs.WriteDebugInfo(sb)
	}
}
