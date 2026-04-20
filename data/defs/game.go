package defs

import (
	"github.com/webbben/2d-game-engine/clock"
)

/*

The purpose of GameContext is to group together all the different context interfaces that different parts of the
game engine use. Ultimately, all of these contexts end up pointing to the Game struct, and some of them may share overlaps in
functions.

*/

type GameStage string

type GameContext interface {
	SaveGame() (saveFileName string)
	PublishEvent(event Event)

	GameDialogContext
	GameQuestContext
	GameScreenContext

	ActiveMapContext
}

type GameDialogContext interface {
	GetCurrentGameTime() clock.GameTime
	GetMapID() MapID
	GetPlayerInfo() PlayerInfo
	SetPlayerName(name string)
	DialogCtxAddGold(amount int)
	RemoveGold(amount int)

	TransitionContext
}

type GameQuestContext interface {
	AssignTaskToNPC(id CharacterDefID, taskDef TaskDef)
	QueueScenario(id ScenarioID)
	UnlockMapLock(mapID MapID, lockID string)
}

// GameScreenContext isn't actually directly used anywhere; we just have it here to keep these functions organized by intended use,
// and to prevent the other contexts from using them. Screens just have direct access to GameContext.
type GameScreenContext interface {
	SetPlayerMenu(scrID ScreenID)
	GetPlayerInventoryRef() *StandardInventory
	EnterMap(mapID MapID, spawnIndex int, doTransition bool)
	PlacePlayerInMap(mapID MapID, x, y float64, doTransition bool)

	// NOTE: not great that we ref clock in here, but clock doesn't import anything so it works.
	// we could consider moving types like GameTime into defs, but just gonna leave things as they are for now.
	InitializeGameWorld(initTime clock.GameTime)
	GetCurrentGameTime() clock.GameTime
	SetGameTime(clock.GameTime)

	GetLoadingStatus() (complete bool, progress float64)
	GetGameStage() GameStage
	SetGameStage(stage GameStage)

	TransitionContext
}

type ActiveMapContext interface {
	StartDialogSession(dialogProfileID DialogProfileID, npcID string)
	StartTradeSession(shopkeeperID ShopID)
	TogglePlayerMenu()
}

type TransitionContext interface {
	StartLoadScreen(loadFunction func(GameContext))
	StartCustomLoadScreen(scrID ScreenID, open, close Transition, loadFunction func(ctx GameContext))
	StartBasicTransition(open, close Transition, lightWeightSetup func(ctx GameContext))
}

// PlayerInfo is information about the player that dialogs might use
type PlayerInfo struct {
	PlayerName    string
	PlayerCulture CultureID
}
