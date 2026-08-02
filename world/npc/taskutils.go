package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
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

// returns true if the NPC is seeing the player for the first time
func (n *NPC) initialPlayerSighting() bool {
	if n.hasSeenPlayerYet {
		return false
	}
	if !n.playerInSightRange {
		return false
	}
	n.hasSeenPlayerYet = true
	return true
}

// call this if a task should have an initial sighting speech bubble.
func (n *NPC) initialPlayerSightingSpeechBubble() {
	if !n.initialPlayerSighting() {
		return
	}

	dialogProfile := n.dataman.GetDialogProfile(n.dialogProfileID)
	for _, reac := range dialogProfile.PlayerNoticeReactions {
		msg := reac.Reaction(defs.Event{}, n.speechBubbleCtx)
		if msg != "" {
			// found a valid reaction; show it in a speech bubble
			n.Entity.ShowSpeechBubble(msg, n.defaultSpeechBubbleParams())
			return
		}
	}
}
