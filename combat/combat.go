// Package combat organizes ongoing fights.
//
// A CombatSession is the shared record of who is fighting whom. It exists because the decisions an
// NPC makes in combat -- who to attack, whether to break off, whether to chase a retreating enemy --
// only make sense in terms of the whole fight rather than one opponent. Fleeing from three bandits
// is a different problem than fleeing from one.
//
// Design rationale, including the open questions, is in issue #190.
package combat

import (
	"slices"

	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/logz"
)

// CombatSessionID uniquely identifies one session.
type CombatSessionID string

// SideID identifies one of a session's two sides. Characters sharing a side are allies; everyone on
// the other side is an enemy.
//
// A side is named for the character who founded it, so SideID is that character's ID. That keeps the
// identity free of any counter or allocation state, makes it legible in logs and crash reports, and
// has a useful consequence: the aggressor side is always the initiator's own side, so the fact "who
// started this" survives the initiator leaving the fight.
//
// A founder may later leave or die without invalidating their side. Anyone still fighting on it is
// part of the side that started this -- a follower who joined the player's fight stays an aggressor
// after the player walks away.
//
// Deliberately a distinct type from id.CharacterStateID rather than an alias, so the two can't be
// mixed up by accident. It carries no ordering or arithmetic meaning; it's only ever compared.
type SideID id.CharacterStateID

// NewSideID returns the ID of a side founded by charID.
func NewSideID(founder id.CharacterStateID) SideID {
	return SideID(founder)
}

// CombatIntent describes what an individual combatant is trying to accomplish by fighting.
//
// Per-combatant rather than per-side, on purpose: a victim defending themselves and a bystander who
// intervened to help them are on the same side of the same fight with genuinely different goals, and
// should behave differently as a result. A defending peasant has no reason to chase a fleeing
// attacker; a bandit attempting a robbery has every reason to.
type CombatIntent string

const (
	IntentSelfDefense CombatIntent = "self-defense" // defending themselves; has no reason to pursue
	IntentProtect     CombatIntent = "protect"      // protecting someone else
	IntentKill        CombatIntent = "kill"         // intends to kill
	IntentRob         CombatIntent = "rob"          // intends to rob; fights until killed or surrendered
)

// Combatant is one character's place in a session.
//
// It holds only what can't be asked of the character itself. Who their enemies are, whether they're
// still standing, whether they've given up -- all of that is a query, either against this session or
// against the Resolver. Nothing here duplicates state that has another home, so there's nothing to
// fall out of sync.
type Combatant struct {
	Side   SideID
	Intent CombatIntent
}

// Resolver gives a session read-only access to live information about the characters in it.
//
// The combat package deliberately does not import the world or npc packages: those need to reach
// sessions in order to pick targets, so importing them here would be a cycle. Anything a session
// can't answer from its own data is asked for through this instead.
//
// Note what's absent: an accepted surrender. By then the player has taken the deal and the character
// has left the fight, so that's a departure from the session rather than a state to query. The two
// other parts of surrender are here, and they mean opposite things -- see IsOfferingSurrender.
type Resolver interface {
	IsDead(charID id.CharacterStateID) bool
	// IsOfferingSurrender reports a live offer awaiting the player's answer. Such a character has NOT
	// given up: the player can still refuse and kill them, so they remain a valid target and are not
	// fleeing. This is deliberately not the same question as "has this character's surrender been
	// accepted", which ends the fight instead of describing it.
	IsOfferingSurrender(charID id.CharacterStateID) bool
	// IsFleeing covers a character running from this session's enemies. A fleeing character is still
	// a valid combatant -- they can be caught and are still their side's problem -- but many NPCs
	// will deprioritize them, or refuse to chase them at all.
	IsFleeing(charID id.CharacterStateID) bool
}

// SessionView is the read-only view of a session: everything a consumer needs in order to ask who's
// fighting, and nothing it needs in order to change that.
//
// This exists because the mutating operations can't be kept off a *CombatSession -- they mutate a map
// through value receivers, so anyone who can read a session could also rewire the fight. World holds
// the concrete session and is the only thing that joins or removes combatants; tasks and calculations
// are handed one of these instead.
type SessionView interface {
	SessionID() CombatSessionID
	Initiator() id.CharacterStateID
	AggressorSide() SideID
	IsAggressorSide(side SideID) bool
	Sides() []SideID
	IsCombatant(charID id.CharacterStateID) bool
	SideOf(charID id.CharacterStateID) (SideID, bool)
	MembersOf(side SideID) []id.CharacterStateID
	EnemiesOf(charID id.CharacterStateID) []id.CharacterStateID
	AlliesOf(charID id.CharacterStateID) []id.CharacterStateID
	LivingMembers(side SideID) []id.CharacterStateID
	HasEnded() bool
}

var _ SessionView = (*CombatSession)(nil)

// CombatSession tracks one ongoing fight: who is on which side, and why each of them is fighting.
//
// Sides and side membership are deliberately not stored. A side is just the tag on each combatant, so
// MembersOf finds members by scanning, which is cheap at these sizes and can't disagree with the
// combatants map about who's on which side.
//
// A character may be in at most one session, and one session never merges with another. When an
// uncommitted character is drawn into an ongoing fight they join one of its sides; when two
// committed characters come to blows, the fight doesn't change -- the victim decides for themselves
// whether to get involved.
//
// Every field is unexported so that the only way to change who is fighting is through the methods
// below, and the only way to read it is through SessionView.
type CombatSession struct {
	id         CombatSessionID
	initiator  id.CharacterStateID
	resolver   Resolver
	combatants map[id.CharacterStateID]Combatant
}

// NewCombatSession starts a session begun by initiator, who joins it immediately as the founder of the
// aggressor side. Starting a fight and being in it aren't separate facts: whoever attacked is part of
// the fight they started, and making that automatic means a session always has an aggressor side.
func NewCombatSession(sessionID CombatSessionID, initiator id.CharacterStateID, intent CombatIntent, resolver Resolver) CombatSession {
	return CombatSession{
		id:        sessionID,
		initiator: initiator,
		resolver:  resolver,
		combatants: map[id.CharacterStateID]Combatant{
			initiator: {Side: NewSideID(initiator), Intent: intent},
		},
	}
}

// FoundSide starts a new side with charID as its founder and only member, and returns that side.
//
// This is how the second side of a fight comes into existence: the first opponent to commit founds
// it. Panics if the session already has two sides, since more would mean a three-way fight, which is
// a case the design doesn't support yet.
//
// The returned side ID is NewSideID(charID), so callers can name it without tracking a return value.
func (s CombatSession) FoundSide(charID id.CharacterStateID, intent CombatIntent) SideID {
	if len(s.Sides()) >= 2 {
		logz.Panicln("CombatSession", "cannot found a third side; sessions have at most two sides",
			"session:", s.id, "character:", charID)
	}
	side := NewSideID(charID)
	if _, alreadyHere := s.combatants[charID]; alreadyHere {
		logz.Panicln("CombatSession", "cannot found a side for a character already in the session",
			"session:", s.id, "character:", charID)
	}
	s.combatants[charID] = Combatant{Side: side, Intent: intent}
	return side
}

// Join adds a combatant to an existing side.
//
// Panics if the character is already in this session. A character may only be in one session at a
// time, and only once, so a duplicate join means the caller lost track of where they were -- which is
// the kind of state error that would silently corrupt the fight if we let it slide. Callers who don't
// already know whether the character is committed elsewhere should check IsCombatant first; this only
// catches the within-session case.
//
// The receiver is a value, but the map inside is shared with every copy, so this mutates the session.
func (s CombatSession) Join(charID id.CharacterStateID, side SideID, intent CombatIntent) {
	if _, alreadyHere := s.combatants[charID]; alreadyHere {
		logz.Panicln("CombatSession", "cannot join: character is already a combatant in this session",
			"session:", s.id, "character:", charID)
	}
	s.combatants[charID] = Combatant{Side: side, Intent: intent}
}

// Leave removes a combatant, reporting whether they were in the session. Callers should check
// HasEnded afterwards: removing the last member of a side ends the fight.
func (s CombatSession) Leave(charID id.CharacterStateID) bool {
	if _, present := s.combatants[charID]; !present {
		return false
	}
	delete(s.combatants, charID)
	return true
}

// Validate checks the session's invariants and panics on violation. Meant for use where a broken
// session would silently produce wrong combat, rather than just fail a check -- World calls this after
// building a session, as a sanity check on logic that decides who is fighting whom.
//
// The rule enforced here is that a session has at most two sides. A one-sided session is legal --
// it's how a fight ends -- but it has ended, and callers should notice via HasEnded.
func (s CombatSession) Validate() {
	if s.id == "" {
		logz.Panicln("CombatSession", "session has no ID")
	}
	if s.initiator == "" {
		logz.Panicln("CombatSession", "session has no initiator", "session:", s.id)
	}
	if s.resolver == nil {
		logz.Panicln("CombatSession", "session has no resolver", "session:", s.id)
	}
	if s.combatants == nil {
		logz.Panicln("CombatSession", "session combatants map is nil", "session:", s.id)
	}
	if sides := s.Sides(); len(sides) > 2 {
		logz.Panicln("CombatSession", "session has more than two sides", "session:", s.id, "sides:", sides)
	}
	if !s.HasEnded() && !s.IsCombatant(s.initiator) {
		logz.Panicln("CombatSession", "ongoing session does not contain its initiator",
			"session:", s.id, "initiator:", s.initiator)
	}
}

// SessionID returns this session's unique ID.
func (s CombatSession) SessionID() CombatSessionID { return s.id }

// Initiator returns the character who started the fight, regardless of whether they're still in it.
func (s CombatSession) Initiator() id.CharacterStateID { return s.initiator }

// AggressorSide returns the side that started the fight. This is always the initiator's own side, and
// stays meaningful after the initiator has left: anyone still fighting on it is part of the side that
// started this.
func (s CombatSession) AggressorSide() SideID {
	return NewSideID(s.initiator)
}

// IsAggressorSide reports whether a side is the one that started the fight.
func (s CombatSession) IsAggressorSide(side SideID) bool {
	return side == s.AggressorSide()
}

// Size returns the number of combatants in the session.
func (s CombatSession) Size() int { return len(s.combatants) }

// IsCombatant reports whether this character is in the session.
func (s CombatSession) IsCombatant(charID id.CharacterStateID) bool {
	_, present := s.combatants[charID]
	return present
}

// SideOf returns the side a character is fighting on.
func (s CombatSession) SideOf(charID id.CharacterStateID) (SideID, bool) {
	c, present := s.combatants[charID]
	if !present {
		return "", false
	}
	return c.Side, true
}

// Sides returns the distinct sides currently represented
func (s CombatSession) Sides() []SideID {
	var sides []SideID
	for _, c := range s.combatants {
		if !slices.Contains(sides, c.Side) {
			sides = append(sides, c.Side)
		}
	}
	return sides
}

// MembersOf returns the characters on one side
func (s CombatSession) MembersOf(side SideID) []id.CharacterStateID {
	var members []id.CharacterStateID
	for charID, c := range s.combatants {
		if c.Side == side {
			members = append(members, charID)
		}
	}
	return members
}

// EnemiesOf returns everyone on the other side. Returns nil if the character isn't in
// the session -- a non-combatant has no enemies as far as this fight is concerned.
func (s CombatSession) EnemiesOf(charID id.CharacterStateID) []id.CharacterStateID {
	self, present := s.combatants[charID]
	if !present {
		return nil
	}
	var enemies []id.CharacterStateID
	for otherID, c := range s.combatants {
		if otherID != charID && c.Side != self.Side {
			enemies = append(enemies, otherID)
		}
	}
	return enemies
}

// AlliesOf returns the character's own side, excluding the character
func (s CombatSession) AlliesOf(charID id.CharacterStateID) []id.CharacterStateID {
	self, present := s.combatants[charID]
	if !present {
		return nil
	}
	var allies []id.CharacterStateID
	for otherID, c := range s.combatants {
		if otherID != charID && c.Side == self.Side {
			allies = append(allies, otherID)
		}
	}
	return allies
}

// HasEnded reports whether the fight is over, which is true once one side has no combatants left.
//
// This is about side membership, not survival: it's correct only as long as callers remove characters
// from the session as they leave it -- dying, being taken out of the fight, or walking away. Callers
// are responsible for keeping membership honest each tick; LivingMembers is here for the questions
// that need finer-grained liveness than membership gives.
func (s CombatSession) HasEnded() bool {
	return len(s.Sides()) < 2
}

// LivingMembers returns the members of a side who are still alive.
//
// Answers from the session's Resolver: the session knows who belongs to a side, but not what state
// they're in, so anything about standing, fleeing, or surrendering has to be asked of the resolver.
func (s CombatSession) LivingMembers(side SideID) []id.CharacterStateID {
	var living []id.CharacterStateID
	for _, memberID := range s.MembersOf(side) {
		if !s.resolver.IsDead(memberID) {
			living = append(living, memberID)
		}
	}
	return living
}
