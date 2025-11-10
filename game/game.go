package game

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/dialog"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/internal/camera"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/pubsub"
	playermenu "github.com/webbben/2d-game-engine/playerMenu"
	"github.com/webbben/2d-game-engine/trade"

	"github.com/hajimehoshi/ebiten/v2"
)

// game state
type Game struct {
	MapInfo *MapInfo
	Player  *player.Player // the player
	Camera  camera.Camera  // the camera/viewport

	Dialog          *dialog.Dialog // if set, a dialog is shown
	PlayerMenu      playermenu.PlayerMenu
	ShowPlayerMenu  bool
	TradeScreen     trade.TradeScreen // screen for handling trades
	ShowTradeScreen bool

	GlobalKeyBindings map[ebiten.Key]func(g *Game) // global keybindings. mainly for testing purposes.
	TestDataMap       map[string]any               // a general purpose map; used for testing in the update hook functions

	activeGlobalKeyBindFn map[ebiten.Key]bool // maps which keybinding functions are actively executing, to prevent repeated calls from long key presses.
	GamePaused            bool                // if true, the game is paused

	Hour           int
	lastHourChange time.Time
	daylightFader  lights.LightFader

	outsideWidth, outsideHeight int

	worldScene *ebiten.Image

	EventBus *pubsub.EventBus

	OverlayManager *overlay.OverlayManager

	UpdateHooks

	DefinitionManager *definitions.DefinitionManager

	debugData debugData // just used for the debug drawing
}

func (g *Game) SetupTradeSession(shopkeeperID string) {
	shopkeeper := g.DefinitionManager.GetShopkeeper(shopkeeperID)
	g.TradeScreen.SetupTradeSession(shopkeeper)
	g.ShowTradeScreen = true
}

func (g *Game) StartDialog(dialogID string) {
	d := g.DefinitionManager.GetDialog(dialogID)
	g.Dialog = &d
}

type UpdateHooks struct {
	UpdateMapHook func(*Game)
}

// run startup functions to prepare the game to play.
// only need to call once, at the beginning.
func InitialStartUp() error {
	err := config.InitFileStructure()
	if err != nil {
		return err
	}
	err = lights.LoadShaders()
	return err
}

func (g *Game) RunGame() error {
	return ebiten.RunGame(g)
}

func NewGame(hour int) *Game {
	g := Game{
		worldScene:        ebiten.NewImage(display.SCREEN_WIDTH, display.SCREEN_HEIGHT),
		lastHourChange:    time.Now(),
		daylightFader:     lights.NewLightFader(lights.LightColor{1, 1, 1}, 0, 0.1, config.HourSpeed/20),
		EventBus:          pubsub.NewEventBus(),
		OverlayManager:    &overlay.OverlayManager{},
		DefinitionManager: definitions.NewDefinitionManager(),
	}

	g.SetHour(hour, true)

	return &g
}

func (g *Game) SetHour(hour int, skipFade bool) {
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

	g.Hour = hour

	g.EventBus.Publish(pubsub.Event{
		Type: pubsub.Event_TimePass,
		Data: map[string]any{
			"hour": hour,
		},
	})
}

// Binds a key to a given function for global keybindings.
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

// this is called whenever the screen/window resizes.
// we keep an internal fixed screen size, and then scale up or down to meet the real size of the window.
// but, we record the real screen size here in case its useful
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	g.outsideWidth = outsideWidth
	g.outsideHeight = outsideHeight
	return display.SCREEN_WIDTH, display.SCREEN_HEIGHT
}
