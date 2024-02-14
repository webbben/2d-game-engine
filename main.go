package main

import (
	"ancient-rome/player"
	"ancient-rome/tileset"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 640
	screenHeight = 480
	tileSize     = 16
	playerSpeed  = 0.25
)

type Game struct {
	roomLayout   [][]string
	player       player.Player
	tilesetFloor map[string]*ebiten.Image
}

func (g *Game) Update() error {
	// Your game logic goes here

	// handle player updates
	g.player.Update()

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	// tile to use if tile fails to be found for some reason
	defaultTile, _ := g.tilesetFloor["grass"]

	for y, row := range g.roomLayout {
		for x, tileKey := range row {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x*tileSize), float64(y*tileSize))

			tileImage, ok := g.tilesetFloor[tileKey]
			if !ok {
				tileImage = defaultTile
			}

			screen.DrawImage(tileImage, op)
		}
	}

	g.player.Draw(screen, tileSize)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Ancient Rome!")

	roomLayout := [][]string{
		{"grass", "grass", "grass", "grass", "grass"},
		{"grass", "grass", "grass", "grass", "grass"},
		{"grass", "grass", "road-dirt", "grass", "grass"},
		{"grass", "grass", "grass", "grass", "grass"},
		{"grass", "grass", "grass", "grass", "grass"},
	}

	floorTileset, err := tileset.LoadTileset(tileset.Txt_Outdoor_Grass_01)
	if err != nil {
		panic(err)
	}

	playerSprites, err := tileset.LoadTileset(tileset.Ent_Player)
	if err != nil {
		panic(err)
	}

	player := player.CreatePlayer(1, 1, playerSprites)

	game := &Game{
		roomLayout:   roomLayout,
		tilesetFloor: floorTileset,
		player:       player,
	}

	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
