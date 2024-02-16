package main

import (
	"ancient-rome/camera"
	"ancient-rome/config"
	"ancient-rome/player"
	"ancient-rome/room"
	"ancient-rome/tileset"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	g.room.DrawFloor(screen, offsetX, offsetY)
	g.drawGridLines(screen, offsetX, offsetY)
	g.player.Draw(screen, getDefaultDrawOptions(), offsetX, offsetY)
	g.room.DrawObjects(screen, offsetX, offsetY)
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

	player := player.CreatePlayer(0, 0, playerSprites)
	room := room.CreateRoom("hello_world")

	game := &Game{
		room:   room,
		player: player,
	}

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
