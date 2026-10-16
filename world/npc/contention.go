package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
)

// Waypoint contention.
//
// When two or more NPCs pick the same next tile, neither can resolve the standoff alone. This file
// only detects and reports that; deciding who goes first is not implemented yet.

// contestedClaim returns the tile this npc is committed to entering next, along with every other live
// NPC on the map that is walking into that same tile. An empty claimant list means there is no
// contention, and the returned tile is then meaningless.
//
// The reason this happens: A* charges only EntityCollision (2) for an occupied tile, which is no more
// than a two-tile detour costs. So when two NPCs want the same waypoint, both paths are individually
// optimal and identically shaped, and the search has no reason to prefer either. They converge, push
// each other apart, and that displacement can knock one off its own route -- its next waypoint then
// sits further away than the dist<=1 salvage range in trySetNextTargetPath, the path is cleared, and
// the NPC re-routes to the same waypoint. That cycle repeats indefinitely.
//
// This is deliberately a property of the NPCs rather than of any one task. "These two are walking
// into the same tile" is true regardless of what either is doing, and an NPC can be committed to a
// waypoint without a task being able to see it. Tasks ask the questions; this answers them for anyone
// who asks, and stays available to whichever task wants to act on it.
//
// Only NPCs are claimants. The player is deliberately excluded, so this answers "is another NPC
// competing for this tile" and not the separate question of player occupancy, which the cost map
// already handles.
func (n *NPC) contestedClaim() (model.Coords, []*NPC) {
	var tile model.Coords
	if next, ok := n.Entity.PathAhead(0); ok {
		tile = next
	} else if n.Entity.IsMoving() {
		tile = n.Entity.TargetTilePos()
	} else {
		// Standing still with no path: not committed to entering anything.
		return model.Coords{}, nil
	}

	var claimants []*NPC
	for _, other := range n.ActiveMapCtx.GetAllNPCs() {
		if other.ID() == n.ID() {
			continue
		}
		// A body lying on the map is still a map occupant but is going nowhere, and an NPC that
		// isn't moving has already arrived wherever it was headed. Note the moving check matters more
		// than it looks: an NPC in a standoff oscillates between moving and stopping as it gets
		// displaced and re-paths, and without it the check would blink out during exactly the frames
		// worth looking at.
		if other.Entity.IsDead() || !other.Entity.IsMoving() {
			continue
		}
		if other.Entity.TargetTilePos().Equals(tile) {
			claimants = append(claimants, other)
		}
	}
	if len(claimants) == 0 {
		return model.Coords{}, nil
	}
	return tile, claimants
}

// logContestedClaim reports detected contention, rate-limited so a standoff that persists doesn't bury
// every other message. The total number of NPCs is reported rather than just "yes/no", because three
// or more competing for one tile needs different handling than a pair, and it's cheaper to find out
// whether that actually happens than to assume it.
func (n *NPC) logContestedClaim() {
	tile, claimants := n.contestedClaim()
	if len(claimants) == 0 {
		return
	}
	if time.Since(n.lastContestedLog) < time.Second*5 {
		return
	}
	n.lastContestedLog = time.Now()

	who := make([]string, 0, len(claimants)+1)
	who = append(who, n.WhoAmI())
	for _, c := range claimants {
		who = append(who, c.WhoAmI())
	}
	logz.Println("contestedClaim:", len(claimants)+1, "NPCs are all walking into tile", tile, who)
}
