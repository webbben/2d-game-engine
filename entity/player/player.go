package player

import (
	"time"

	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/world/npc"
)

type Player struct {
	Entity            *entity.Entity
	CharacterStateRef *state.CharacterState
	MovementMechanics

	dataman *datamanager.DataManager

	World WorldContext

	LastUserInput time.Time // tracks when the user has last made some kind of input (movement, attack, etc)
}

type WorldContext interface {
	GetOverlayManager() *overlay.OverlayManager
	TogglePlayerMenu()
	GetNearbyNPCs(x, y, radius float64) []*npc.NPC
	ActivateArea(r model.Rect, originX, originY float64) bool
	HandleObjectUpdate(result object.ObjectUpdateResult, obj *object.Object)
	GetHoverTarget() (*npc.NPC, *object.Object)
}

// Y is needed for sorting renderables
func (p Player) Y() float64 {
	return p.Entity.Y
}

func (p Player) X() float64 {
	return p.Entity.X
}

func NewPlayer(dataman *datamanager.DataManager, ent *entity.Entity) Player {
	if ent == nil {
		panic("player must have entity")
	}

	charState := dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))

	return Player{
		CharacterStateRef: charState,
		dataman:           dataman,
		Entity:            ent,
	}
}
