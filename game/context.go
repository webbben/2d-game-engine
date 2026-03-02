package game

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/logz"
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

// TODO: should this function take a map ID too? for now, map is setup elsewhere and we just use this so that the MainMenu (appearance designer)
// can make the player spawn.
func (g *Game) AddPlayerToMap(mapID defs.MapID, spawnIndex int) {
	err := g.SetupMap(mapID, &OpenMapOptions{
		RunNPCManager:    true,
		RegenerateImages: true,
	})
	if err != nil {
		panic(err)
	}

	// TODO: should the "player" ID somehow be passed in here, or determined by a property on Game or something?
	// Originally I was having it passed in to this function as a parameter, but I ran into an inconvenient problem where I couldn't put state.CharacterStateID
	// into the context interface in defs, since it would cause a circular dependency. So, for now, just leaving it hard coded here..
	playerEnt := entity.LoadCharacterStateIntoEntity("player", g.Dataman, g.AudioManager)

	p := player.NewPlayer(g.Dataman, playerEnt)
	_ = g.PlacePlayerAtSpawnPoint(&p, spawnIndex)
	g.Player = &p

	// TODO: make these interfaces similar to how we handle HUDs
	//
	// TODO: We need to convert these to Screens, so I'm gonna comment this out until that's been done.
	//
	// g.PlayerMenu = setup.GetPlayerMenu(g.Player, g.Dataman)
	// g.TradeScreen = setup.GetTradeScreen(g.Player, g.Dataman)
	//
	// g.PlayerMenu.InventoryPage.LoadPlayerItemsIn()
}
