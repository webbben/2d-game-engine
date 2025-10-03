package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/image/font"
)

var (
	GameScale           float64 = 3.5 // how much the game view is scaled up
	UIScale             float64 = 1   // how much the UI (not in game world) is scaled up
	DrawGridLines               = false
	ShowEntityPositions         = false // show the logical positions and collision boxes of entities
	ShowCollisions              = false // show the areas that are collisions on the map
	ShowPlayerCoords            = false
	ShowNPCPaths                = false // highlight the paths that NPCs are following
	TrackMemoryUsage            = false // show a report in the console of memory usage every few seconds
	ShowGameDebugInfo           = false // show a report of various debugging info (like F12 in minecraft)

	HourSpeed time.Duration = time.Minute // how long it takes for an hour to pass in game

	DefaultFont font.Face // must be set by game

	MapPathOverride string = "" // set this if you have a custom directory where maps are stored
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

// given a map ID, returns the full path to the map's TMJ file for loading purposes.
func ResolveMapPath(mapID string) string {
	absPath := ""
	if MapPathOverride != "" {
		absPath = filepath.Join(MapPathOverride, fmt.Sprintf("%s.tmj", mapID))
	} else {
		absPath = filepath.Join(GameAssetsPath(), "maps", fmt.Sprintf("%s.tmj", mapID))
	}

	return absPath
}
