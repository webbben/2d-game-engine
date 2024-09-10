package main

import (
	"fmt"
	"log"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/dialog"
	"github.com/webbben/2d-game-engine/entity"
	g "github.com/webbben/2d-game-engine/game"
	"github.com/webbben/2d-game-engine/general_util"
	"github.com/webbben/2d-game-engine/image"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/player"
	"github.com/webbben/2d-game-engine/room"
	"github.com/webbben/2d-game-engine/tileset"

	"github.com/hajimehoshi/ebiten/v2"
)

// for now, this main function is for testing the APIs we are creating
// once this game engine is closer to being complete, I can make a separate game
// that calls this project's APIs instead

func main() {
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
}

func GetDialog() dialog.Dialog {
	box, err := image.LoadImage("assets/images/dialog_box3.png")
	if err != nil {
		log.Fatal("failed to load dialog box image:", err)
	}
	return dialog.Dialog{
		Steps: []dialog.DialogStep{
			{Text: "Greetings, what can I do for you?"},
		},
		SpeakerName: "Hamu",
		CurrentStep: 0,
		Box:         box,
	}
}
