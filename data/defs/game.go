package defs

import (
	"time"

	"github.com/webbben/2d-game-engine/clock"
)

/*

The purpose of GameContext is to group together all the different context interfaces that different parts of the
game engine use. Ultimately, all of these contexts end up pointing to the Game struct, and some of them may share overlaps in
functions.

*/

type GameStage string

type GameContext interface {
	PublishEvent(event Event)

	SaveFileContext

	GameDialogContext
	GameQuestContext
	GameScreenContext

	ActiveMapContext
}

type SaveFileContext interface {
	SaveGame() (saveFileName string)
	GetAllExistingCharacters() []ExistingCharacterInfo
	LoadGame(saveFilePath string)
}

type GameDialogContext interface {
	GetCurrentGameTime() clock.GameTime
	GetMapID() MapID
	GetActiveMapDef() MapDef
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
	GetPlayerInventoryRef() *StandardInventory
	EnterMap(mapID MapID, spawnIndex int, doTransition bool)
	PlacePlayerInMap(mapID MapID, x, y float64, doTransition bool)

	// NOTE: not ideal that we ref clock in here, but clock doesn't import anything so it works.
	// we could consider moving types like GameTime into defs, but just gonna leave things as they are for now.
	InitializeGameWorld(initTime clock.GameTime)
	GetCurrentGameTime() clock.GameTime
	SetGameTime(clock.GameTime)

	// starts a time lapse. could take an extra update loop or two if the simulation loop is actively doing something; time lapse doesn't occur
	// until the simulation loop is successfully paused.
	StartTimeLapse(newTime clock.GameTime)

	GetLoadingStatus() (complete bool, progress float64)
	GetGameStage() GameStage
	SetGameStage(stage GameStage)

	TransitionContext
}

type ActiveMapContext interface {
	StartDialogSession(dialogProfileID DialogProfileID, npcID string)
	StartTradeSession(shopkeeperID ShopID)
	TogglePlayerMenu()

	ShowMiscScreen(scrID ScreenID)
}

type TransitionContext interface {
	StartLoadScreen(loadFunction func(GameContext))
	StartCustomLoadScreen(scrID ScreenID, open, close Transition, loadFunction func(ctx GameContext))
	StartSyncTransition(open, close Transition, lightWeightSetup func(ctx GameContext))
}

// PlayerInfo is information about the player that dialogs might use
type PlayerInfo struct {
	PlayerName    string
	PlayerCulture CultureID
}

// SaveInfo is just an overview of a single save file - NOT the actual save data.
// This is used for showing a preview of a save file, without actually loading all the data.
type SaveInfo struct {
	UniquePlayerID  UniquePlayerID
	CharacterName   string
	LastPlay        time.Time
	CurrentMapID    MapID
	CurrentGameTime clock.GameTime
	SaveFilePath    string
}

// ExistingCharacterInfo gives info about an existing character that has save files.
// A single character can have multiple save files, so this just gives you an overview of that character's info and save data.
type ExistingCharacterInfo struct {
	UniquePlayerID UniquePlayerID
	DisplayName    string
	RecentSave     SaveInfo
	SaveFilePaths  []string
}
