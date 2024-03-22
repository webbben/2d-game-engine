package main

import (
	"fmt"
	"image/color"

	"github.com/webbben/2d-game-engine/camera"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/debug"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/player"
	"github.com/webbben/2d-game-engine/room"
	"github.com/webbben/2d-game-engine/tileset"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// game state
type Game struct {
	room     room.Room
	player   player.Player
	camera   camera.Camera
	entities []entity.Entity
}

func (g *Game) Update() error {
	// Your game logic goes here

	// handle player updates
	g.player.Update(g.room.BarrierLayout)

	// move camera as needed
	g.camera.MoveCamera(g.player.X, g.player.Y)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	offsetX, offsetY := g.camera.GetAbsPos()

	// draw the terrain of the room
	g.room.DrawFloor(screen, offsetX, offsetY)
	g.room.DrawCliffs(screen, offsetX, offsetY)
	if config.DrawGridLines {
		g.drawGridLines(screen, offsetX, offsetY)
	}

	// draw the player
	g.player.Draw(screen, offsetX, offsetY)

	// draw entities
	for _, e := range g.entities {
		e.Draw(screen, offsetX, offsetY)
	}

	// draw objects
	g.room.DrawObjects(screen, offsetX, offsetY)

	if config.ShowPlayerCoords {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("Player pos: [%v, %v]", g.player.X, g.player.Y))
	}
}

func (g *Game) drawGridLines(screen *ebiten.Image, offsetX float64, offsetY float64) {
	offsetX = offsetX * config.GameScale
	offsetY = offsetY * config.GameScale
	lineColor := color.RGBA{255, 0, 0, 255}
	maxWidth := float64(len(g.room.TileLayout[0]) * config.TileSize * config.GameScale)
	maxHeight := float64(len(g.room.TileLayout) * config.TileSize * config.GameScale)

	// vertical lines
	for x := 0; x < len(g.room.TileLayout[0]); x++ {
		drawX := float64(x*config.TileSize*config.GameScale) - offsetX
		drawY := -offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(drawX), float32(maxHeight-offsetY), 1, lineColor, true)
	}
	// horizontal lines
	for y := 0; y < len(g.room.TileLayout); y++ {
		drawX := -offsetX
		drawY := float64(y*config.TileSize*config.GameScale) - offsetY
		vector.StrokeLine(screen, float32(drawX), float32(drawY), float32(maxWidth-offsetX), float32(drawY), 1, lineColor, true)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return config.ScreenWidth, config.ScreenHeight
}

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

	room.GenerateRandomRoom("test_room", 100, 100)
	currentRoom := room.CreateRoom("test_room")

	game := &Game{
		room:     currentRoom,
		player:   player,
		entities: []entity.Entity{*testEnt},
	}

	go game.entities[0].TravelToPosition(model.Coords{X: 99, Y: 50}, currentRoom.BarrierLayout)

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
