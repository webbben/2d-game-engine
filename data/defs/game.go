package defs

/*

The purpose of GameContext is to group together all the different context interfaces that different parts of the
game engine use. Ultimately, all of these contexts end up pointing to the Game struct, and some of them may share overlaps in
functions.

*/

type GameContext interface {
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
	AddPlayerToMap(mapID MapID, spawnIndex int)
}

// PlayerInfo is information about the player that dialogs might use
type PlayerInfo struct {
	PlayerName    string
	PlayerCulture CultureID
}
