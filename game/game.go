package game

import (
	"fmt"
	"sort"

	"github.com/webbben/2d-game-engine/camera"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/debug"
	"github.com/webbben/2d-game-engine/dialog"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/player"
	"github.com/webbben/2d-game-engine/room"

	"github.com/hajimehoshi/ebiten/v2"
)

// information about the current room the player is in
type RoomInfo struct {
	Room     room.Room                // the room the player is currently in
	Entities []*entity.Entity         // the entities in the current room
	Objects  []object.Object          // the objects in the current room
	ImageMap map[string]*ebiten.Image // the map of images (tiles) used in rendering the current room
}

// pre-processes data in the room for various purposes, but mostly to prepare for rendering
// and performance improvements
func (ri *RoomInfo) Preprocess() {
	sort.Slice(ri.Objects, func(i, j int) bool {
		return ri.Objects[i].Y < ri.Objects[j].Y
	})
}

// game state
type Game struct {
	RoomInfo
	Player                player.Player                // the player
	Camera                camera.Camera                // the camera/viewport
	Conversation          *dialog.Conversation         // if set, the player is in a conversation or being shown general text to read.
	GlobalKeyBindings     map[ebiten.Key]func(g *Game) // global keybindings. mainly for testing purposes.
	activeGlobalKeyBindFn map[ebiten.Key]bool          // maps which keybinding functions are actively executing, to prevent repeated calls from long key presses.
}

// generates a cost map for the contents of the game state
//
// currently includes:
//
// * entities in the room
func (g Game) GenerateCostMap() [][]int {
	costMap := make([][]int, g.Room.Height)
	for i := 0; i < len(costMap); i++ {
		costMap[i] = make([]int, g.Room.Width)
	}
	for i := 0; i < len(g.Entities); i++ {
		costMap[int(g.Entities[i].Y)][int(g.Entities[i].X)] = 10
	}
	return costMap
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

func (g *Game) Update() error {
	if g.GlobalKeyBindings != nil {
		g.handleGlobalKeyBindings()
	}

	// update dialog if currently in a dialog session
	if g.Conversation != nil {
		if g.Conversation.End {
			// if dialog has ended, remove it from game state
			g.Conversation = nil
		} else {
			g.Conversation.UpdateConversation()
		}
	} else {
		// handle player updates if no conversation is active
		g.Player.Update(g.Room.BarrierLayout)
	}

	// move camera as needed
	g.Camera.MoveCamera(g.Player.X, g.Player.Y)

	// sort entities by Y position for rendering
	if len(g.Entities) > 1 {
		sort.Slice(g.Entities, func(i, j int) bool {
			return g.Entities[i].Y < g.Entities[j].Y
		})
	}

	if config.TrackMemoryUsage {
		debug.UpdatePerformanceMetrics()
	}

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return config.ScreenWidth, config.ScreenHeight
}
