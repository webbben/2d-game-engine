package main

import (
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/debug"
	"github.com/webbben/2d-game-engine/entity"
	g "github.com/webbben/2d-game-engine/game"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/player"
	"github.com/webbben/2d-game-engine/room"
	"github.com/webbben/2d-game-engine/tileset"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// track memory usage
	if config.TrackMemoryUsage {
		go debug.DisplayResourceUsage(60)
	}

	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle(config.WindowTitle)

	playerSprites, err := tileset.LoadTileset(tileset.Ent_Player)
	if err != nil {
		panic(err)
	}
	player := player.CreatePlayer(50, 50, playerSprites)
	testEnt := entity.CreateEntity(entity.Old_Man_01, "test_npc", "Pepe", "")
	if testEnt == nil {
		panic("failed to create entity")
	}
	houseObj, houseImg := object.CreateObject(object.Latin_house_1, 10, 10)
	imageMap := make(map[string]*ebiten.Image)
	imageMap[houseObj.Name] = houseImg

	room.GenerateRandomRoom("test_room", 100, 100)
	currentRoom := room.CreateRoom("test_room")

	game := &g.Game{
		RoomInfo: g.RoomInfo{
			Room:     currentRoom,
			Entities: []entity.Entity{*testEnt},
			Objects:  []object.Object{*houseObj},
			ImageMap: imageMap,
		},
		Player: player,
	}
	game.RoomInfo.Preprocess()

	// go game.entities[0].TravelToPosition(model.Coords{X: 99, Y: 50}, currentRoom.BarrierLayout)
	go game.Entities[0].FollowPlayer(&game.Player, currentRoom.BarrierLayout)

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
