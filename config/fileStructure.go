package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/utils/files"
)

/*

Here is the file structure for game data:

├── assets
│   ├── audio
│   │   ├── music  // songs and bgm
│   │   └── sfx    // short sound effects (weapon strikes, footsteps, etc)
│   └── fonts
├── generated      // all files that are generated during runtime
│   └── tiles      // individual tile images generated from tilesets
└── tiled
    ├── maps       // maps where the player or other entities can load into
    └── tilesets   // tilesets used by maps, entities, UI components, etc
*/

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
	p := filepath.Join(homePath, GameDataDirectoryName)
	if !FileExists(p) {
		logz.Panicln("GameDataRootPath", "path doesn't exist!", p)
	}
	return p
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
		_ = os.RemoveAll(tilePath())
	}
	err := os.MkdirAll(tilePath(), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error while creating image tile path: %w", err)
	}

	return nil
}

// SAVE GAME

func saveFilesPath() string {
	p := filepath.Join(GameDataRootPath(), "saves")
	// if !FileExists(p) {
	// 	logz.Panicln("saveFilesPath", "path doesn't exist!", p)
	// }
	return p
}

func getPlayerSaveDir(uniquePlayerID defs.UniquePlayerID) string {
	return filepath.Join(saveFilesPath(), string(uniquePlayerID))
}

func GetAllSaveDirs() []string {
	return files.GetListOfDirs(saveFilesPath(), true)
}

func ResolveSaveFilePath(uniquePlayerID defs.UniquePlayerID, filename string) string {
	saveDir := getPlayerSaveDir(uniquePlayerID)
	return filepath.Join(saveDir, filename)
}

func EnsurePlayerSaveDirExists(uniquePlayerID defs.UniquePlayerID) {
	saveDir := getPlayerSaveDir(uniquePlayerID)
	if !FileExists(saveDir) {
		err := os.MkdirAll(saveDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}
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

// ResolveMapPath : given a map ID, returns the full path to the map's TMJ file for loading purposes.
func ResolveMapPath(mapID defs.MapID) string {
	return filepath.Join(gameMapsPath(), fmt.Sprintf("%s.tmj", mapID))
}

func gameFontsPath() string {
	return filepath.Join(gameAssetsPath(), "fonts")
}

func ResolveFontPath(fontRelPath string) string {
	return filepath.Join(gameFontsPath(), fontRelPath)
}
