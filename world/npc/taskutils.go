package npc

import (
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
