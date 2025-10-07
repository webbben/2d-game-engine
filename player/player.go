package player

import (
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/item"
)

type Player struct {
	Entity *entity.Entity // the entity that represents this player in a map

	InventoryItems []item.ItemInstance
}

// needed for sorting renderables
func (p Player) Y() float64 {
	return p.Entity.Y
}
