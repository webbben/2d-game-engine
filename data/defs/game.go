package defs

import "github.com/webbben/2d-game-engine/clock"

/*

The purpose of GameContext is to group together all the different context interfaces that different parts of the
game engine use. Ultimately, all of these contexts end up pointing to the Game struct, and some of them may share overlaps in
functions.

*/

type GameContext interface {
	SaveGame() (saveFileName string)

	GameDialogContext
	GameQuestContext
	GameScreenContext
}

type GameDialogContext interface {
	GetPlayerInfo() PlayerInfo
	SetPlayerName(name string)
	DialogCtxAddGold(amount int)
}

type GameQuestContext interface {
	AssignTaskToNPC(id CharacterDefID, taskDef TaskDef)
	QueueScenario(id ScenarioID)
	UnlockMapLock(mapID MapID, lockID string)
}

type GameScreenContext interface {
	EnterMap(mapID MapID, spawnIndex int)

	// NOTE: not great that we ref clock in here, but clock doesn't import anything so it works.
	// we could consider moving types like GameTime into defs, but just gonna leave things as they are for now.
	InitializeGameWorld(initTime clock.GameTime)
}

// PlayerInfo is information about the player that dialogs might use
type PlayerInfo struct {
	PlayerName    string
	PlayerCulture CultureID
}
