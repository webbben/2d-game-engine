package game

import (
	"github.com/webbben/2d-game-engine/data/defs"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/internal/pubsub"
)

func (g *Game) DialogCtxAddGold(amount int) {
	characterstate.EarnMoney(&g.Player.Entity.CharacterStateRef.StandardInventory, amount, g.Dataman)
}

func (g *Game) AssignTaskToNPC(id defs.CharacterDefID, taskDef defs.TaskDef) {
	// confirm this is the ID of a unique characterDef
	charDef := g.Dataman.GetCharacterDef(id)
	if !charDef.Unique {
		logz.Panicln("AssignTaskToNPC", "characterDef of ID given is not unique; can only assign tasks to specific characters if they are unique.")
	}

	// send an event to the NPC, assuming he exists...
	g.EventBus.Publish(pubsub.NPCAssignTask(string(id), taskDef))
}

func (g *Game) QueueScenario(id defs.ScenarioID) {
	scenarioDef := g.Dataman.GetScenarioDef(id)

	mapID := scenarioDef.MapID
	if mapID == "" {
		panic("mapID was empty")
	}

	g.EnsureMapStateExists(mapID)

	mapState := g.Dataman.GetMapState(mapID)

	// ensure this scenario is not already queued up
	for _, scenarioID := range mapState.QueuedScenarios {
		if scenarioID == id {
			logz.Panicln("QueueScenario", "tried to queue a scenario, but its ID was already in the scenario queue for this map:", id)
		}
	}

	mapState.QueuedScenarios = append(mapState.QueuedScenarios, id)

	logz.Println("Scenario Queued", "queued", id, "in map", mapID)
}

func (g *Game) UnlockMapLock(mapID defs.MapID, lockID string) {
	g.EnsureMapStateExists(mapID)

	mapState := g.Dataman.GetMapState(mapID)
	lockState := mapState.MapLocks[lockID]
	lockState.Unlocked = true
	mapState.MapLocks[lockID] = lockState
}
