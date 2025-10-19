package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/webbben/2d-game-engine/internal/logz"
)

/*

Here is the file structure for game data:

├── assets
│   ├── audio
│   │   ├── music
│   │   └── sfx
│   └── fonts
├── generated
│   └── tiles
└── tiled
    ├── maps
    └── tilesets

*/

const (
	game_dir = "2d_game_engine"
)

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func GameDataRootPath() string {
	if GameDataPathOverride != "" {
		return GameDataPathOverride
	}
	homePath, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	return filepath.Join(homePath, game_dir)
}

func InitFileStructure() error {
	if !FileExists(gameAssetsPath()) {
		err := os.MkdirAll(gameAssetsPath(), os.ModePerm)
		if err != nil {
			return fmt.Errorf("error while creating game assets path: %w", err)
		}
	}

	// tiles
	// remove all existing tiles on startup, to ensure that all tiles are up to date with source data
	if FileExists(tilePath()) {
		logz.Println("SYSTEM", "deleting all existing generated tiles from previous runs")
		os.RemoveAll(tilePath())
	}
	err := os.MkdirAll(tilePath(), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error while creating image tile path: %w", err)
	}

	return nil
}

// GENERATED

func gameGeneratedPath() string {
	return filepath.Join(GameDataRootPath(), "generated")
}

func tilePath() string {
	return filepath.Join(gameGeneratedPath(), "tiles")
}

func ResolveTilePath(relTilePath string) string {
	return filepath.Join(tilePath(), relTilePath)
}

// ASSETS

// root/assets
func gameAssetsPath() string {
	return filepath.Join(GameDataRootPath(), "assets")
}

// root/assets/audio
func gameAudioPath() string {
	return filepath.Join(gameAssetsPath(), "audio")
}

func ResolveAudioPath(audioPath string) string {
	if filepath.IsAbs(audioPath) {
		panic("given path is absolute; you should give a relative path based on the audio directory structure.")
	}
	ext := filepath.Ext(audioPath)
	if ext != ".mp3" {
		panic("given audio path has a non-mp3 extension; audio files must be mp3. ext: " + ext)
	}
	fullPath := filepath.Join(gameAudioPath(), audioPath)
	return fullPath
}

// TILED

func gameMapsPath() string {
	return filepath.Join(GameDataRootPath(), "tiled", "maps")
}

func gameTilesetsPath() string {
	return filepath.Join(GameDataRootPath(), "tiled", "tilesets")
}

func ResolveTilesetPath(relTilesetPath string) string {
	return filepath.Join(gameTilesetsPath(), relTilesetPath)
}

// given a map ID, returns the full path to the map's TMJ file for loading purposes.
func ResolveMapPath(mapID string) string {
	return filepath.Join(gameMapsPath(), fmt.Sprintf("%s.tmj", mapID))
}

func gameFontsPath() string {
	return filepath.Join(gameAssetsPath(), "fonts")
}

func ResolveFontPath(fontRelPath string) string {
	return filepath.Join(gameFontsPath(), fontRelPath)
}
