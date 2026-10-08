package world

import (
	"slices"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/logz"
)

// Death has two distinct jobs, and they're separated on purpose:
//
//	On death      -- do the things that must happen immediately and exactly once: stop them fighting, stop
//	                 them acting. The body stays, because the player may not have looted it yet.
//	On expiry     -- take the body away. Done on a timer rather than immediately, because a corpse the
//	                 player meant to loot can easily outlive the fight they died in.
//
// Keeping them separate is also what lets a dead character keep existing as state. A defined character who
// dies stays in Dataman.CharacterStates with Dead set, so quest and dialog conditions can still ask
// whether they're alive. Only characters that were never meant to be permanent go away entirely.

// handleEntityDeath runs the immediate cleanup for a character who just died.
//
// Driven by SysEventEntityDied rather than called from Entity.Kill directly, so that it lands on the main
// loop and so the entity package doesn't have to know the world exists.
func (w *World) handleEntityDeath(charID id.CharacterStateID) {
	logz.Println("World", "character died:", charID)

	// Stop being a combatant. This is what makes a fight end when one side is wiped out: HasEnded is
	// judged by side membership, so without this a losing side that died still "has members" and the
	// session never ends. The player dies through the same path as anyone else, and this is as much as
	// applies to them.
	w.LeaveCombat(charID)

	if charID == id.PlayerStateID {
		// The player's death is a game-over state, not a body to be cleaned up. The rest of this is about
		// corpses: stopping an NPC's tasks, and taking the body away later. Neither means anything for the
		// player, and there's no NPC for them anyway.
		return
	}

	// Stop what they were doing. A corpse that keeps working its schedule is how you end up with dead
	// NPCs wandering around the map; finishing the task also runs its cleanup, so anything the task was
	// holding onto (a reserved object, a flee path) is released.
	w.stopDeadNPC(charID)

	// Schedule the body to be taken away later.
	w.scheduleCorpseExpiry(charID)
}

// stopDeadNPC finishes a dead NPC's current task and prevents the decision loop from assigning another.
func (w *World) stopDeadNPC(charID id.CharacterStateID) {
	n := w.getInWorldNPC(charID)
	if n == nil {
		// a temp scenario npc, or one already cleaned up
		return
	}
	n.ClearCurrentTask()
}

// scheduleCorpseExpiry queues the removal of a body once it has had long enough to be looted.
func (w *World) scheduleCorpseExpiry(charID id.CharacterStateID) {
	if w.Clock == nil {
		return
	}
	expiry := w.Clock.GetFutureGameTime(config.CorpseExpiryHours)
	w.EventBus.ScheduleFutureEvent(defs.Event{
		Type: sysEventCorpseExpired,
		Data: map[string]any{
			"charID": charID,
		},
	}, expiry)
	logz.Println("World", "corpse expiry scheduled:", charID, "at", expiry)
}

// sysEventCorpseExpired is an internal-only event type used to schedule body removal. It's not a sys-event
// anyone subscribes to; it's queued into the future-event schedule and dispatched back through here.
const sysEventCorpseExpired defs.EventType = "SYS_CORPSE_EXPIRED"

// expireCorpse removes a dead NPC's body from the world.
//
// Deliberately tolerant of the body already being gone: expiry is queued at death time, and a corpse can
// be removed by other means in between, so a missing NPC isn't an error.
func (w *World) expireCorpse(charID id.CharacterStateID) {
	n := w.getInWorldNPC(charID)
	if n == nil {
		return
	}

	logz.Println("World", "corpse expired:", charID)
	w.LeaveCombat(charID)
	if w.ActiveMap != nil {
		w.ActiveMap.RemoveNPCFromMap(charID)
	}

	// MapOccupancy decides which NPCs get loaded when the player enters a map, so a body left here would
	// come back next time the player walked into the room.
	w.removeMapOccupancy(charID, n.CharacterStateRef.CurrentMap)
	delete(w.NPCs, charID)

	// Defined characters are kept: their dead flag is something quests and dialog can ask about. Only
	// temporary ones (scenario npcs) were never meant to outlive the session.
	if cs, ok := w.Dataman.CharacterStates[charID]; ok && characterStateIsTransient(cs) {
		w.Dataman.RemoveCharacterState(charID)
	}
}

// characterStateIsTransient reports whether a character was only ever meant to exist for a while, and so
// should stop existing entirely once its corpse is gone.
//
// Scenario characters set Temp, which already means "not meant to be saved long term" (see
// state.CharacterState.Temp). Characters produced by a map's character generator are *not* currently
// marked: GenerateCharacter leaves Temp false, so they're indistinguishable from authored characters and
// are treated the same way. See issue #192.
func characterStateIsTransient(cs *state.CharacterState) bool {
	return cs.Temp
}

// removeMapOccupancy drops a character from a map's roster.
func (w *World) removeMapOccupancy(charID id.CharacterStateID, mapID defs.MapID) {
	occupants, ok := w.MapOccupancy[mapID]
	if !ok {
		return
	}
	idx := slices.Index(occupants, charID)
	if idx < 0 {
		return
	}
	w.MapOccupancy[mapID] = slices.Delete(occupants, idx, idx+1)
}

// handleCorpseExpired is the queued follow-up to a death, run once the body has outlived its usefulness.
func (w *World) handleCorpseExpired(charID id.CharacterStateID) {
	w.expireCorpse(charID)
}
