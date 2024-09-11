package main

import (
	"fmt"
	"os"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/dialog"
	"github.com/webbben/2d-game-engine/entity"
	g "github.com/webbben/2d-game-engine/game"
	"github.com/webbben/2d-game-engine/general_util"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/player"
	"github.com/webbben/2d-game-engine/room"
	"github.com/webbben/2d-game-engine/tileset"

	"github.com/hajimehoshi/ebiten/v2"
)

// Use this command to simulate using this game engine's APIs in a consuming application
// The real game will be developed in a different Go application that uses this game engine module.

// Note: only packages that are defined in this module should be used in this file; ebiten or other installed packages won't be
// available to a Go application that is using this game engine module.

func main() {
	// TODO move these to an API once we know how screen settings should be managed
	//ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle(config.WindowTitle)

	// get our testrun game state
	game := setupGameState()
	game.RoomInfo.Preprocess()

	// go game.entities[0].TravelToPosition(model.Coords{X: 99, Y: 50}, currentRoom.BarrierLayout)
	for i := range game.Entities {
		go game.Entities[i].FollowPlayer(&game.Player, game.Room.BarrierLayout)
	}
	//go game.Entities[0].FollowPlayer(&game.Player, currentRoom.BarrierLayout)

	for i := range game.Entities {
		fmt.Println("ent:", game.Entities[i].EntID)
	}

	// this command launches the specified game state
	// TODO wrap this in an API?
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}

func setupGameState() *g.Game {
	// create the player
	playerSprites, err := tileset.LoadTileset(tileset.Ent_Player)
	if err != nil {
		panic(err)
	}
	player := player.CreatePlayer(50, 50, playerSprites)

	// create some entities
	ents := make([]*entity.Entity, 0)
	for i := 0; i < 1; i++ {
		testEnt := entity.CreateEntity(entity.Old_Man_01, fmt.Sprintf("test_npc_%v", i), "Pepe", "")
		if testEnt == nil {
			panic("failed to create entity")
		}
		testEnt.Position.X = float64(general_util.RandInt(0, 20))
		testEnt.Position.Y = float64(general_util.RandInt(40, 60))
		ents = append(ents, testEnt)
	}

	// create a house object
	houseObj, houseImg := object.CreateObject(object.Latin_house_1, 10, 10)
	imageMap := make(map[string]*ebiten.Image)
	imageMap[houseObj.Name] = houseImg

	// generate a room
	room.GenerateRandomRoom("test_room", 100, 100)
	currentRoom := room.CreateRoom("test_room")

	// setup the game struct
	game := &g.Game{
		RoomInfo: g.RoomInfo{
			Room:     currentRoom,
			Entities: ents,
			Objects:  []object.Object{*houseObj},
			ImageMap: imageMap,
		},
		Player: player,
	}

	// add my test key bindings
	addCustomKeyBindings(game)

	return game
}

func addCustomKeyBindings(game *g.Game) {
	// open a test dialog
	game.SetGlobalKeyBinding(ebiten.KeyEqual, func(g *g.Game) {
		// doing this async since we are loading an image file
		go func() {
			for ebiten.IsKeyPressed(ebiten.KeyEqual) {
			}
			d := GetDialog()
			g.Dialog = &d
		}()
	})
	game.SetGlobalKeyBinding(ebiten.KeyEscape, func(g *g.Game) {
		// doing this async since we are loading an image file
		os.Exit(0)
	})
}

func GetDialog() dialog.Dialog {
	d := dialog.Dialog{
		Steps: []dialog.DialogStep{
			{Text: "Greetings, what can I do for you?"},
		},
		SpeakerName: "Hamu",
		CurrentStep: 0,
		FontName:    "Planewalker",
	}
	d.SetDialogTiles("tileset/borders/dialog_1")

	return d
}
