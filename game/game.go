package game

import (
	"fmt"

	"github.com/webbben/2d-game-engine/internal/camera"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/dialog"
	"github.com/webbben/2d-game-engine/player"
	"github.com/webbben/2d-game-engine/screen"

	"github.com/hajimehoshi/ebiten/v2"
)

// game state
type Game struct {
	MapInfo               *MapInfo
	Player                player.Player                // the player
	Camera                camera.Camera                // the camera/viewport
	Conversation          *dialog.Conversation         // if set, the player is in a conversation or being shown general text to read.
	GlobalKeyBindings     map[ebiten.Key]func(g *Game) // global keybindings. mainly for testing purposes.
	activeGlobalKeyBindFn map[ebiten.Key]bool          // maps which keybinding functions are actively executing, to prevent repeated calls from long key presses.
	GamePaused            bool                         // if true, the game is paused

	CurrentScreen *screen.Screen // if set, a screen is being displayed and we are not in the game world
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

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return config.ScreenWidth, config.ScreenHeight
}
