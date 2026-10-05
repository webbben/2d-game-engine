package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
)

// SatisfiesObjectOwnership checks if the NPC can use this object without violating ownership rules (roles or owner IDs)
func (n NPC) SatisfiesObjectOwnership(obj object.Object) bool {
	if obj.OwnerID == "" && obj.RoleID == "" {
		return true
	}
	if obj.OwnerID != "" {
		return obj.OwnerID == n.CharacterStateRef.ID
	}
	return n.CharacterStateRef.Roles[obj.RoleID]
}

// npcLockIDs returns the lock IDs this npc is able to unlock, based on the keys it is carrying.
func npcLockIDs(n *NPC) []string {
	return characterstate.GetLockIDs(*n.CharacterStateRef, n.dataman)
}

// unopenableGateTiles returns the tiles blocked by closed gates that this npc has no way to open.
//
// Gates are deliberately left out of the shared cost map (see ActiveMap.buildCostMap) because npcs can
// usually open them on contact, so a computed route can lead straight through one. For a gate this npc
// can't open, that's not an option, so callers overlay these tiles as walls before searching.
//
// Main loop only: object open/close state is main-loop state, so this must not be called from a
// background goroutine.
func unopenableGateTiles(n *NPC) []model.Coords {
	tiles := []model.Coords{}
	// FollowTask can re-request a path every frame while tracking a moving target, so the NPC's keys are
	// only looked up if we actually hit a closed gate. Most maps have none, and then this costs just the
	// object scan.
	var lockIDs []string
	haveLockIDs := false

	for _, obj := range n.ActiveMapCtx.GetAllObjects() {
		if obj.Type != object.TypeGate {
			continue
		}
		// IsCollidable is false for an open gate, so this also skips gates we could just walk through.
		if !obj.IsCollidable() {
			continue
		}
		if !haveLockIDs {
			lockIDs = npcLockIDs(n)
			haveLockIDs = true
		}
		if obj.CanBeOpenedBy(lockIDs) {
			continue
		}
		tiles = append(tiles, obj.GetRect().GetOverlappingTiles()...)
	}
	return tiles
}

// costMapWithBlockedTiles returns a copy of costMap with the given tiles forced impassable.
//
// The copy is load-bearing, not defensive: GetPathfindingSnapshot hands back the slice that the whole
// engine shares, so mutating it in place would corrupt pathfinding for every other consumer, and race
// with the main loop storing a fresh snapshot.
func costMapWithBlockedTiles(costMap [][]int, tiles []model.Coords) [][]int {
	if len(costMap) == 0 {
		return costMap
	}
	out := make([][]int, len(costMap))
	for i, row := range costMap {
		out[i] = make([]int, len(row))
		copy(out[i], row)
	}
	for _, c := range tiles {
		// bounds-checked because this can run on the background goroutine, where a panic takes down the game
		if c.Y < 0 || c.Y >= len(out) || c.X < 0 || c.X >= len(out[c.Y]) {
			continue
		}
		out[c.Y][c.X] = path_finding.BlockThreshold
	}
	return out
}

// call this if a task should have an initial sighting speech bubble.
func (n *NPC) initialPlayerSightingSpeechBubble() {
	if !n.initialPlayerSightingThisTick {
		return
	}

	dialogProfile := n.dataman.GetDialogProfile(n.dialogProfileID)
	for _, reac := range dialogProfile.PlayerNoticeReactions {
		msg := reac.Reaction(defs.Event{}, n.speechBubbleCtx)
		if msg != "" {
			// found a valid reaction; show it in a speech bubble
			n.Entity.ShowSpeechBubble(msg, n.defaultSpeechBubbleParams())
			logz.Println(n.ID(), "showing speech bubble:", msg)
			return
		}
	}
}
