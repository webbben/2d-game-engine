package player

import (
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/npc"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/item"
)

type Player struct {
	Entity *entity.Entity
	MovementMechanics

	defMgr *definitions.DefinitionManager

	World WorldContext
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
func (p *Player) AddItemToInventory(invItem item.InventoryItem) (bool, item.InventoryItem) {
	return p.Entity.AddItemToInventory(invItem)
}

// TODO: delete?
func (p *Player) RemoveItemFromInventory(itemToRemove item.InventoryItem) (bool, item.InventoryItem) {
	return p.Entity.RemoveItemFromInventory(itemToRemove)
}
