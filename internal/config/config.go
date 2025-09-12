package config

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/image/font"
)

var (
	GameScale         float64 = 2.5 // how much the game view is scaled up
	DrawGridLines             = false
	ShowPlayerCoords          = false
	ShowNPCPaths              = false // highlight the paths that NPCs are following
	TrackMemoryUsage          = false // show a report in the console of memory usage every few seconds
	ShowGameDebugInfo         = false // show a report of various debugging info (like F12 in minecraft)

	DefaultFont font.Face // must be set by game
)

const (
	TileSize = 16

	game_dir = "2d_game_engine"
)

func GameDataRootPath() string {
	homePath, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(homePath, game_dir)
}

func GameAssetsPath() string {
	return filepath.Join(GameDataRootPath(), "assets")
}

func GameDefsPath() string {
	return filepath.Join(GameDataRootPath(), "defs")
}
