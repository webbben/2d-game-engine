package main

import (
	"ancient-rome/camera"
	"ancient-rome/config"
	"ancient-rome/player"
	"ancient-rome/room"
	"ancient-rome/tileset"

	"github.com/hajimehoshi/ebiten/v2"
)

// game state
type Game struct {
	room   room.Room
	player player.Player
	camera camera.Camera
}

func (g *Game) Update() error {
	// Your game logic goes here

	// handle player updates
	g.player.Update()

	// move camera as needed
	g.camera.MoveCamera(g.player.X, g.player.Y)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	offsetX, offsetY := g.camera.GetAbsPos()
	g.room.Draw(screen, getDefaultDrawOptions(), offsetX, offsetY)
	g.player.Draw(screen, getDefaultDrawOptions(), offsetX, offsetY)
}

func getDefaultDrawOptions() *ebiten.DrawImageOptions {
	defaultOptions := &ebiten.DrawImageOptions{}
	return defaultOptions
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return config.ScreenWidth, config.ScreenHeight
}

func main() {
	ebiten.SetWindowSize(config.ScreenWidth, config.ScreenHeight)
	ebiten.SetWindowTitle(config.WindowTitle)

	playerSprites, err := tileset.LoadTileset(tileset.Ent_Player)
	if err != nil {
		panic(err)
	}

	player := player.CreatePlayer(1, 1, playerSprites)
	room := room.CreateRoom("hello_world")

	game := &Game{
		room:   room,
		player: player,
	}

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
