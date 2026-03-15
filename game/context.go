package game

import (
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/dialogv2"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
)

func (g *Game) DialogCtxAddGold(amount int) {
	characterstate.EarnMoney(&g.World.Player.CharacterStateRef.StandardInventory, amount, g.Dataman)
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
	lockState, exists := mapState.MapLocks[lockID]
	if !exists {
		logz.Panicln("UnlockMapLock", "given lock ID was not found in map. mapID:", mapID, "lockID:", lockID)
	}
	lockState.Unlocked = true
	mapState.MapLocks[lockID] = lockState
}

// EnterMap adds the player to a map. creates the active map too, in the process.
// Used in the NewGame flow to actually put the player in a map once his character has been created.
func (g *Game) EnterMap(mapID defs.MapID, spawnIndex int) {
	g.World.EnterMap(mapID, spawnIndex)

	// TODO: We need to convert these to Screens, so I'm gonna comment this out until that's been done.
	// we can probably create some kind of "in-game screen" that these can be set into to cause them to appear.
	//
	// g.PlayerMenu = setup.GetPlayerMenu(g.Player, g.Dataman)
	// g.TradeScreen = setup.GetTradeScreen(g.Player, g.Dataman)
	//
	// g.PlayerMenu.InventoryPage.LoadPlayerItemsIn()
}

// PlacePlayerInMap is the same as EnterMap, but for putting the player at a specific position (instead of just at a spawn point).
// Used by the LoadGame flow, since you will be appearing not at a spawn point, but at the position you were in last when the game was saved.
func (g *Game) PlacePlayerInMap(mapID defs.MapID, x, y float64) {
	g.World.EnterMapAtPosition(mapID, x, y)
}

func (g *Game) SetPlayerName(name string) {
	g.World.SetPlayerName(name)
}

func (g *Game) GetPlayerInfo() defs.PlayerInfo {
	charDef := g.Dataman.GetCharacterDef(g.World.Player.CharacterStateRef.DefID)
	return defs.PlayerInfo{
		PlayerName:    g.World.Player.CharacterStateRef.DisplayName,
		PlayerCulture: charDef.CultureID,
	}
}

func (g *Game) StartTradeSession(shopkeeperID defs.ShopID) {
	shopkeeperDef := g.Dataman.GetShopkeeperDef(shopkeeperID)
	shopkeeperState := g.Dataman.GetShopkeeperState(shopkeeperID)
	g.TradeScreen.SetupTradeSession(*shopkeeperDef, shopkeeperState)
	g.ShowTradeScreen = true
}

// StartDialogSession starts a dialog session with the given dialog profile ID
func (g *Game) StartDialogSession(dialogProfileID defs.DialogProfileID, npcID string) {
	if npcID == "" {
		panic("npcID was empty")
	}
	params := dialogv2.DialogSessionParams{
		NPCID:         npcID,
		ProfileID:     dialogProfileID,
		BoxTilesetSrc: "boxes/boxes.tsj",
		BoxOriginID:   16,
		TextFont:      config.DefaultFont,
	}
	ds := dialogv2.NewDialogSession(params, g.EventBus, g.Dataman, g.ScreenManager, g)

	g.dialogSession = &ds
}
