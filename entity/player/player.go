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
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/world/npc"
)

const (
	sneakXPCheckInterval = 120 // ticks between checks (~every 2s at 60fps)
)

type Player struct {
	Entity            *entity.Entity
	CharacterStateRef *state.CharacterState
	MovementMechanics

	dataman  *datamanager.DataManager
	eventBus *pubsub.EventBus

	World WorldContext

	LastUserInput time.Time // tracks when the user has last made some kind of input (movement, attack, etc)

	sneakXPTicks int // ticks since last passive sneak-XP check
}

func (p Player) GetPlayerInfo() defs.PlayerInfo {
	charDef := p.dataman.GetCharacterDef(id.PlayerDefID)
	return defs.PlayerInfo{
		PlayerName:    p.CharacterStateRef.DisplayName,
		PlayerCulture: charDef.CultureID,
	}
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

func NewPlayer(dataman *datamanager.DataManager, eventBus *pubsub.EventBus, ent *entity.Entity) Player {
	if ent == nil {
		panic("player must have entity")
	}

	charState := dataman.GetCharacterState(id.PlayerStateID)

	return Player{
		CharacterStateRef: charState,
		dataman:           dataman,
		eventBus:          eventBus,
		Entity:            ent,
	}
}
