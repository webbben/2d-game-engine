// Package game defines the ebiten game structure and the foundation of running the game
package game

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/dialogv2"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/logz"
	playermenu "github.com/webbben/2d-game-engine/playerMenu"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/quest"
	"github.com/webbben/2d-game-engine/screen"
	"github.com/webbben/2d-game-engine/trade"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/world"

	"github.com/hajimehoshi/ebiten/v2"
)

type GameStage string

const (
	MainMenu    GameStage = "MAIN_MENU"
	InGameWorld GameStage = "IN_GAME_WORLD"
)

// Game - the root of the game state that is maintained when game is active
type Game struct {
	// the main menu screen that runs when you startup the game.
	// when this screen reaches "done" state, the game will start running in-world logic.
	MainMenu       screen.Screen
	mainMenuViewer screen.ScreenViewer
	gameStage      GameStage

	World *world.World

	dialogSession   *dialogv2.DialogSession
	PlayerMenu      playermenu.PlayerMenu
	ShowPlayerMenu  bool
	TradeScreen     trade.TradeScreen // screen for handling trades
	ShowTradeScreen bool

	GlobalKeyBindings map[ebiten.Key]func(g *Game) // global keybindings. mainly for testing purposes.

	activeGlobalKeyBindFn map[ebiten.Key]bool // maps which keybinding functions are actively executing, to prevent repeated calls from long key presses.

	outsideWidth, outsideHeight int

	EventBus *pubsub.EventBus

	OverlayManager *overlay.OverlayManager

	hud HUD // if set, this will update and be drawn on top of the in-game world

	Dataman       *datamanager.DataManager
	AudioManager  *audio.AudioManager
	QuestManager  *quest.QuestManager
	ScreenManager *screen.ScreenManager
}

func (g Game) GetGameStage() GameStage {
	return g.gameStage
}

func (g *Game) SetHUD(hud HUD) {
	g.hud = hud
}

// HUD interface provides a type that can be drawn as an HUD over the in-game world
type HUD interface {
	Draw(screen *ebiten.Image)
	Update(g *Game)
}

// InitialStartUp runs startup functions to prepare the game to play.
// only need to call once, at the beginning.
func InitialStartUp() error {
	err := config.InitFileStructure()
	if err != nil {
		return err
	}
	err = lights.LoadShaders()
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) RunGame() error {
	return ebiten.RunGame(g)
}

// NewGame gives you a newly created Game struct for use; Note that this does NOT start a new "game playthrough"
// or something like that. This just gives you a blank slate of a Game that can then be used for other things, like
// starting a new playthrough, or loading a previous save, etc.
//
// After getting a Game from here, you can:
//
// - Load all data definitions (character defs, item defs, quest defs, dialog defs, etc. all of the DEFS, not states.)
//
// - Once all defs have been loaded in, THEN you can load a save or start a new game/playthrough, etc. this is where states are created or loaded.
func NewGame() *Game {
	g := Game{
		gameStage: MainMenu,

		// managers

		EventBus:       pubsub.NewEventBus(),
		OverlayManager: &overlay.OverlayManager{},
		Dataman:        datamanager.NewDataManager(),
		AudioManager:   audio.NewAudioManager(),
		ScreenManager:  screen.NewScreenManager(),
	}

	g.QuestManager = quest.NewQuestManager(g.EventBus, &g)

	return &g
}

// InitializeGameWorld creates the World struct and builds all the NPCs and other world data that exists in the "game world".
//
// You should only call this AFTER doing the following:
//
// - Loading ALL data definitions (into Datamanager, Questmanager, etc)
//
// - Loading the player's character def and character STATE into Datamanager (using the unique ID at defs.PlayerID).
//
//	Yes, that's right: the player's character STATE should already exist. You need to create your own screen or process to instantiate this before creating the game world.
//	Other character states should not be made before calling this, however.
//
// In other words, this should only be done once the game is ready to fully launch into the "game universe" and load all characters, maps, etc.
func (g *Game) InitializeGameWorld(initTime clock.GameTime) {
	if len(g.Dataman.MapDefs) == 0 {
		logz.Panicln("InitializeGameWorld", "no map defs found. are you sure you loaded all data definitions?")
	}
	if len(g.Dataman.CharacterDefs) == 0 {
		logz.Panicln("InitializeGameWorld", "no character defs found. are you sure you loaded all data definitions?")
	}
	// process all maps and initialize their states; this is because a map serves to define what characters actually exist in the game world.
	// If a character has a bed in a map somewhere, then they officially "exist" and will have their state initialized and will have an NPC operating somewhere in the game universe.

	for id := range g.Dataman.MapDefs {
		g.EnsureMapStateExists(id)
	}

	// after the above, we should now have all NPC's entity states created.
	// Note that this doesn't mean all character defs have had a corresponding state created; it means all character defs found in reference in a bed object were created.
	// any character that doesn't have a bed object in some map will, therefore, not exist yet.

	g.World = world.NewWorld(initTime, g.Dataman, g.AudioManager, g.EventBus, g)
}

func (g *Game) SetMainMenu(scrID screen.ScreenID) {
	scr := g.ScreenManager.GetScreen(scrID)
	g.MainMenu = scr
	g.mainMenuViewer = screen.NewScreenViewer(scr, g.Dataman, g.EventBus, g)
}

func (g Game) LastPlayerUpdate() time.Time {
	if g.World.Player == nil {
		logz.Panicln("LastPlayerUpdate", "player is nil")
	}
	return g.World.Player.LastUserInput
}

// SetGlobalKeyBinding sets a key to a given function for global keybindings.
//
// Generally should only be used for testing purposes, as normally keybindings will only be applicable to certain screens, contexts, in-game scenarios, etc.
func (g *Game) SetGlobalKeyBinding(key ebiten.Key, f func(g *Game)) {
	// initialize the maps if they aren't initialized yet
	if g.GlobalKeyBindings == nil {
		g.GlobalKeyBindings = make(map[ebiten.Key]func(g *Game))
	}
	if g.activeGlobalKeyBindFn == nil {
		g.activeGlobalKeyBindFn = make(map[ebiten.Key]bool)
	}
	// bind key to function
	if _, exists := g.GlobalKeyBindings[key]; exists {
		fmt.Println("** Warning! Global key binding overwritten for key", key)
		fmt.Println("** If you are binding keys for temporary purposes during gameplay, this is probably a misuse of global key bindings.")
	}
	g.GlobalKeyBindings[key] = f
}

func (g *Game) handleGlobalKeyBindings() {
	for key, callbackFn := range g.GlobalKeyBindings {
		if ebiten.IsKeyPressed(key) && !g.activeGlobalKeyBindFn[key] {
			// do this in separate goroutine to not holdup game update thread
			go func(key ebiten.Key, callbackFn func(g *Game)) {
				g.activeGlobalKeyBindFn[key] = true
				callbackFn(g)
				// wait for key release
				for ebiten.IsKeyPressed(key) {
				}
				g.activeGlobalKeyBindFn[key] = false
			}(key, callbackFn)
		}
	}
}

// Layout is called whenever the screen/window resizes.
// we keep an internal fixed screen size, and then scale up or down to meet the real size of the window.
// but, we record the real screen size here in case its useful
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	g.outsideWidth = outsideWidth
	g.outsideHeight = outsideHeight
	return display.SCREEN_WIDTH, display.SCREEN_HEIGHT
}
