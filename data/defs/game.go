package defs

import (
	"time"

	"github.com/webbben/2d-game-engine/clock"
)

/*

The purpose of GameContext is to group together all the different context interfaces that different parts of the
game engine use. Ultimately, all of these contexts end up pointing to the Game struct, and some of them may share overlaps in
functions.

TODO: These contexts have gotten really messy, and on top of that it seems like a lot of things just simply use GameContext instead
of a more specific context. That's not necessarily a big problem, but one structural issue it does pose is everything therefore has to route through the Game
struct (since the GameContext is effectively the Game struct). This means that even if ALL of the logic for a given context (Like WorldEffectContext)
lies in a more specific place (like the World struct), we have to still make all the functions at a "top level" in the Game struct, and then just have them
pass directly into World.
.
I think in the future it would be good to continue to refine the groupings of these contexts, but also start passing those
specific contexts to the places that would specifically use them. For example, Screens should take the GameScreenContext, and if needed, maybe some other
context too. If we need more than one type of "Screen" that has different capabilities (like MainScreen vs InGameScreen) we can consider that too.
.
Anytime a new function is needed for a context somewhere, let's take care to analyze what that function is for in a more general purpose.

*/

type GameStage string

// GameContext is a context that contains all other contexts. Basically, the Game struct.
type GameContext interface {
	SaveFileContext
	EventContext

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
	WorldInfoContext
	WorldEffectContext
}

type GameQuestContext interface {
	WorldEffectContext
}

type WorldInfoContext interface {
	GetCurrentGameTime() clock.GameTime
	GetMapID() MapID
	GetActiveMapDef() MapDef
	GetPlayerInfo() PlayerInfo
}

// WorldEffect is an effect you can apply to the game world. They are generally intended to be used
// by dialogs, quests, or other things that might give the player items, change his gold, or manipulate the in-game world in some way.
type WorldEffect interface {
	Apply(ctx WorldEffectContext)
}

// WorldEffectContext represents specific "effects" that can be applied to the world.
// These are for specifically changing something, like adding items or gold to the player or altering the state of a map, etc.
type WorldEffectContext interface {
	// Map and world

	// moves the player to a new map, with the option of time passage too.
	// Note: if hours is set, then this function calls a TimeLapse which is async; you will need to listen for the SysTimeLapse event to know
	// when the time lapse is done.
	TravelToMap(mapID MapID, spawnIndex int, hours int)
	QueueScenario(id ScenarioID)
	UnlockMapLock(mapID MapID, lockID string)
	GetCurrentGameTime() clock.GameTime // needed for scheduling future effects
	EventContext

	// Player specific

	AddGold(amount int)
	RemoveGold(amount int)
	AddItem(itemID ItemID, quantity int)
	AddRole(roleID RoleID)
	RemoveRole(roleID RoleID)

	// NPCs

	AssignTaskToNPC(id CharacterDefID, taskDef TaskDef, requireListener bool)
}

type EventContext interface {
	BroadcastEvent(event Event)
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

	// starts a time lapse. could take an extra update loop or two if the simulation loop is actively doing something; time lapse doesn't occur
	// until the simulation loop is successfully paused.
	StartTimeLapse(newTime clock.GameTime)

	GetLoadingStatus() (complete bool, progress float64)
	GetGameStage() GameStage
	SetGameStage(stage GameStage)

	TransitionContext

	SaveFileContext

	WorldInfoContext
}

type ActiveMapContext interface {
	StartDialogSession(dialogProfileID DialogProfileID, npcID string)
	TogglePlayerMenu()

	ShowMiscScreen(scrID ScreenID, params any)

	GetHoverTargetInfo() (*NPCInfo, *ObjectInfo)
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
