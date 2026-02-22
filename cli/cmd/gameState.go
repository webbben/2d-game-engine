package cmd

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/game"
)

type gameParams struct {
	startHour   int
	startMapID  defs.MapID
	startEvents []defs.Event
}

func SetupGameState(params gameParams) *game.Game {
	g := game.NewGame(params.startHour)

	LoadDefMgr(g.DefinitionManager)
	LoadAudioManager(g.AudioManager)
	questDefs := GetAllQuestDefs()
	g.QuestManager.LoadAllQuestData(questDefs, []state.QuestState{})

	// run the list of events, if set. these events may be for things like triggering quest starts, which can have effects like queueing scenarios, etc.
	if len(params.startEvents) > 0 {
		for _, e := range params.startEvents {
			g.EventBus.Publish(e)
		}
	}

	err := g.SetupMap(params.startMapID, &game.OpenMapOptions{
		RunNPCManager:    true,
		RegenerateImages: true,
	})
	if err != nil {
		panic(err)
	}

	// make the player

	charStateID := entity.CreateNewCharacterState("player", entity.NewCharacterStateParams{IsPlayer: true}, g.DefinitionManager)
	playerEnt := entity.LoadCharacterStateIntoEntity(charStateID, g.DefinitionManager, g.AudioManager)

	p := player.NewPlayer(g.DefinitionManager, playerEnt)
	_ = g.PlacePlayerAtSpawnPoint(&p, 0)
	g.Player = &p

	// add my test key bindings
	addCustomKeyBindings(g)

	// TODO: make these interfaces similar to how we handle HUDs
	g.PlayerMenu = GetPlayerMenu(g.Player, g.DefinitionManager)
	g.TradeScreen = GetTradeScreen(g.Player, g.DefinitionManager)

	g.PlayerMenu.InventoryPage.LoadPlayerItemsIn()

	hud := NewWorldHUD(WorldHUDParams{
		ClockTilesetSrc:     "ui/clock.tsj",
		ClockDayIndex:       1,
		ClockEveningIndex:   2,
		ClockNightIndex:     3,
		ClockLateNightIndex: 4,
	})
	g.SetHUD(&hud)

	return g
}
