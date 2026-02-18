package game

import (
	"github.com/webbben/2d-game-engine/data/defs"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/pubsub"
)

func (g *Game) DialogCtxAddGold(amount int) {
	characterstate.EarnMoney(&g.Player.Entity.CharacterStateRef.StandardInventory, amount, g.DefinitionManager)
}

func (g *Game) AssignTaskToNPC(id defs.CharacterDefID, taskDef defs.TaskDef) {
	// confirm this is the ID of a unique characterDef
	charDef := g.DefinitionManager.GetCharacterDef(id)
	if !charDef.Unique {
		logz.Panicln("AssignTaskToNPC", "characterDef of ID given is not unique; can only assign tasks to specific characters if they are unique.")
	}

	// send an event to the NPC, assuming he exists...
	g.EventBus.Publish(pubsub.NPCAssignTask(string(id), taskDef))
}
