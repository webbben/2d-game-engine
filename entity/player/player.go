package player

import (
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/entity/npc"
	"github.com/webbben/2d-game-engine/internal/model"
)

type Player struct {
	Entity *entity.Entity
	MovementMechanics

	defMgr *definitions.DefinitionManager

	World WorldContext

	LastUserInput time.Time // tracks when the user has last made some kind of input (movement, attack, etc)
}

type WorldContext interface {
	GetNearbyNPCs(x, y, radius float64) []*npc.NPC
	ActivateArea(r model.Rect) bool
	HandleMouseClick(mouseX, mouseY int) bool
}

// Y is needed for sorting renderables
func (p Player) Y() float64 {
	return p.Entity.Y
}

func NewPlayer(defMgr *definitions.DefinitionManager, ent *entity.Entity) Player {
	if ent == nil {
		panic("player must have entity")
	}

	return Player{
		defMgr: defMgr,
		Entity: ent,
	}
}

// TODO: delete? just use the entity one since there is no difference now
func (p *Player) AddItemToInventory(invItem defs.InventoryItem) (bool, defs.InventoryItem) {
	return characterstate.AddItemToInventory(p.Entity.CharacterStateRef, invItem)
}

// TODO: delete?
func (p *Player) RemoveItemFromInventory(itemToRemove defs.InventoryItem) (bool, defs.InventoryItem) {
	return characterstate.RemoveItemFromInventory(p.Entity.CharacterStateRef, itemToRemove)
}
