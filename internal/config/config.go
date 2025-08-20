package config

import (
	"log"
	"os"
	"path/filepath"
)

var (
	GameScale        float64 = 2 // how much the game view is scaled up
	DrawGridLines            = false
	ShowPlayerCoords         = false
	TrackMemoryUsage         = false // show a report in the console of memory usage every few seconds
)

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	TileSize     = 16
	WindowTitle  = "Ancient Rome!"

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
