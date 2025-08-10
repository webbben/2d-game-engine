package config

import (
	"log"
	"os"
	"path/filepath"
)

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	TileSize     = 16
	WindowTitle  = "Ancient Rome!"

	GameScale        = 3 // how much the game view is scaled up
	DrawGridLines    = false
	ShowPlayerCoords = false
	TrackMemoryUsage = true // show a report in the console of memory usage every few seconds

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
