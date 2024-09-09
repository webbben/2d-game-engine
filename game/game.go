package game

import (
	"sort"

	"github.com/webbben/2d-game-engine/camera"
	"github.com/webbben/2d-game-engine/config"
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
	Player player.Player  // the player
	Camera camera.Camera  // the camera/viewport
	Dialog *dialog.Dialog // if present, the player is currently in a dialog
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

func (g *Game) Update() error {
	// Your game logic goes here

	// handle player updates
	g.Player.Update(g.Room.BarrierLayout)

	// move camera as needed
	g.Camera.MoveCamera(g.Player.X, g.Player.Y)

	// sort entities by Y position for rendering
	if len(g.Entities) > 1 {
		sort.Slice(g.Entities, func(i, j int) bool {
			return g.Entities[i].Y < g.Entities[j].Y
		})
	}

	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return config.ScreenWidth, config.ScreenHeight
}
