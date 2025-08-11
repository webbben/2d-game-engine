package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/webbben/2d-game-engine/entity"
	g "github.com/webbben/2d-game-engine/game"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/dialog"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/player"
	"github.com/webbben/2d-game-engine/screen"
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

	tiled.InitFileStructure()

	// get our testrun game state
	game := setupGameState()
	game.RoomInfo.Preprocess()

	// go game.entities[0].TravelToPosition(model.Coords{X: 99, Y: 50}, currentRoom.BarrierLayout)
	// for i := range game.Entities {
	// 	go game.Entities[i].FollowPlayer(&game.Player, game.Room.BarrierLayout)
	// }
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
	player := player.CreatePlayer(15, 15, playerSprites)

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

	// // create a house object
	// houseObj, houseImg := object.CreateObject(object.Latin_house_1, 10, 10)
	// imageMap := make(map[string]*ebiten.Image)
	// imageMap[houseObj.Name] = houseImg

	// generate a room
	// TODO - make a generateRandomMap function for Tiled maps
	// room.GenerateRandomRoom("test_room", 100, 100)
	// currentRoom := room.CreateRoom("test_room")

	currentMap, err := tiled.OpenMap("assets/tiled/maps/testmap.tmj")
	if err != nil {
		log.Fatal(err)
	}

	err = currentMap.Load()
	if err != nil {
		log.Fatal(err)
	}

	// setup the game struct
	game := &g.Game{
		RoomInfo: g.RoomInfo{
			Map:      currentMap,
			Entities: ents,
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
			fmt.Println("getting dialog")
			c := GetConversation()
			g.Conversation = &c
		}()
	})
	game.SetGlobalKeyBinding(ebiten.KeyMinus, func(g *g.Game) {
		go func() {
			fmt.Println("getting title screen")
			s := GetTitleScreen()
			g.CurrentScreen = &s
		}()
	})

	game.SetGlobalKeyBinding(ebiten.KeyEscape, func(g *g.Game) {
		os.Exit(0)
	})
}

func GetConversation() dialog.Conversation {
	d := dialog.Dialog{
		Steps: []dialog.DialogStep{
			{Text: "Greetings, what can I do for you?"},
		},
	}

	c := dialog.Conversation{
		Greeting: d,
		Font: dialog.Font{
			FontName: "Planewalker",
		},
		Topics: map[string]dialog.Dialog{
			"rumors":        {Steps: []dialog.DialogStep{{Text: "I heard there are goblins in the forest."}}},
			"little advice": {Steps: []dialog.DialogStep{{Text: "Don't go into the forest alone."}}},
			"joke":          {Steps: []dialog.DialogStep{{Text: "Why did the chicken cross the road?"}, {Text: "To get to the other side!"}}},
			"the empire":    {Steps: []dialog.DialogStep{{Text: "The empire is a vast and powerful entity."}}},
		},
	}
	c.SetDialogTiles("tileset/borders/dialog_1")

	return c
}

func GetTitleScreen() screen.Screen {
	s := screen.Screen{
		Title:               "Ancient Rome!",
		TitleFontName:       "Herculanum",
		TitleFontColor:      color.White,
		BodyFontName:        "Herculanum",
		BodyFontColor:       color.White,
		BackgroundImagePath: "image/bg/dark_cistern.png",
	}

	// add a menu
	m := screen.Menu{
		Buttons: []screen.Button{
			{Text: "New Game", Callback: func() {}},
			{Text: "Load Game", Callback: func() {}},
			{Text: "Options", Callback: func() {}},
			{Text: "Quit", Callback: func() { os.Exit(0) }},
		},
		BoxTilesetPath: "tileset/borders/stone_1",
	}
	s.Menus = append(s.Menus, m)

	return s
}
