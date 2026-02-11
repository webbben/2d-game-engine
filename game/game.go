// Package game defines the ebiten game structure and the foundation of running the game
package game

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/dialogv2"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/internal/camera"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/pubsub"
	playermenu "github.com/webbben/2d-game-engine/playerMenu"
	"github.com/webbben/2d-game-engine/trade"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game - the root of the game state that is maintained when game is active
type Game struct {
	MapInfo *MapInfo
	Player  *player.Player // the player
	Camera  camera.Camera  // the camera/viewport

	dialogSession   *dialogv2.DialogSession
	PlayerMenu      playermenu.PlayerMenu
	ShowPlayerMenu  bool
	TradeScreen     trade.TradeScreen // screen for handling trades
	ShowTradeScreen bool

	GlobalKeyBindings map[ebiten.Key]func(g *Game) // global keybindings. mainly for testing purposes.
	TestDataMap       map[string]any               // a general purpose map; used for testing in the update hook functions

	activeGlobalKeyBindFn map[ebiten.Key]bool // maps which keybinding functions are actively executing, to prevent repeated calls from long key presses.
	GamePaused            bool                // if true, the game is paused

	Clock         clock.Clock
	daylightFader lights.LightFader

	outsideWidth, outsideHeight int

	worldScene *ebiten.Image

	EventBus *pubsub.EventBus

	OverlayManager *overlay.OverlayManager

	hud HUD // if set, this will update and be drawn on top of the in-game world

	UpdateHooks

	DefinitionManager *definitions.DefinitionManager

	debugData debugData // just used for the debug drawing
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
	ds := dialogv2.NewDialogSession(params, g.EventBus, g.DefinitionManager, g)

	g.dialogSession = &ds
}

// SetPlayerName - made for dialog action
func (g *Game) SetPlayerName(name string) {
	if g.Player == nil {
		panic("player was nil")
	}
	if g.Player.Entity == nil {
		panic("player entity was nil")
	}
	if g.Player.Entity.CharacterStateRef == nil {
		panic("player character state was nil")
	}
	g.Player.Entity.CharacterStateRef.DisplayName = name
}

func (g *Game) SetHUD(hud HUD) {
	g.hud = hud
}

// HUD interface provides a type that can be drawn as an HUD over the in-game world
type HUD interface {
	Draw(screen *ebiten.Image)
	Update(g *Game)
}

func (g *Game) SetupTradeSession(shopkeeperID defs.ShopID) {
	shopkeeperDef := g.DefinitionManager.GetShopkeeperDef(shopkeeperID)
	shopkeeperState := g.DefinitionManager.GetShopkeeperState(shopkeeperID)
	g.TradeScreen.SetupTradeSession(*shopkeeperDef, shopkeeperState)
	g.ShowTradeScreen = true
}

type UpdateHooks struct {
	UpdateMapHook func(*Game)
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

	// set logs to show microseconds in timestamps
	// log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	// Since we now log the update tick, I don't think this is that useful anymore

	return nil
}

func (g *Game) RunGame() error {
	return ebiten.RunGame(g)
}

func NewGame(hour int) *Game {
	g := Game{
		worldScene:        ebiten.NewImage(display.SCREEN_WIDTH, display.SCREEN_HEIGHT),
		daylightFader:     lights.NewLightFader(lights.LightColor{1, 1, 1}, 0, 0.1, config.HourSpeed/20),
		EventBus:          pubsub.NewEventBus(),
		OverlayManager:    &overlay.OverlayManager{},
		DefinitionManager: definitions.NewDefinitionManager(),
		Clock:             clock.NewClock(config.HourSpeed, hour, 0, 0, 0, 762, 90),
	}

	// make sure lighting is initialized
	g.OnHourChange(hour, true)

	return &g
}

func (g *Game) GetPlayerInfo() dialogv2.PlayerInfo {
	return dialogv2.PlayerInfo{
		PlayerName: g.Player.Entity.CharacterStateRef.DisplayName,
	}
}

// OnHourChange handles any hourly changes that should occur; such as lighting, event publishing, etc.
func (g *Game) OnHourChange(hour int, skipFade bool) {
	if hour < 0 || hour > 23 {
		panic("invalid hour")
	}

	newDaylight, darknessFactor := lights.CalculateDaylight(hour)
	if skipFade {
		g.daylightFader.SetCurrentColor(newDaylight)
		g.daylightFader.TargetColor = newDaylight
		g.daylightFader.SetCurrentDarknessFactor(darknessFactor)
		g.daylightFader.TargetDarknessFactor = darknessFactor
	} else {
		g.daylightFader.TargetColor = newDaylight
		g.daylightFader.TargetDarknessFactor = darknessFactor
	}

	g.EventBus.Publish(pubsub.Event{
		Type: pubsub.Event_TimePass,
		Data: map[string]any{
			"hour": hour,
		},
	})
}

func (g Game) LastPlayerUpdate() time.Time {
	if g.Player == nil {
		logz.Panicln("LastPlayerUpdate", "player is nil")
	}
	return g.Player.LastUserInput
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
