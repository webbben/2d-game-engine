package tileset

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// texture tilesets
const (
	// a basic grass tileset
	Txt_Outdoor_Grass_01 string = "outdoor_grass_01"
)

const (
	Ent_Player string = "ent_player"
)

var (
	// maps a tileset key to its path
	pathDict = map[string]string{
		Txt_Outdoor_Grass_01: "./tileset/textures/floors/grass_tiles",
		Ent_Player:           "./tileset/entities/player",
	}
)

// loads a tileset for the given tileset key
func LoadTileset(key string) (map[int]*ebiten.Image, error) {
	tileset := make(map[int]*ebiten.Image)

	// attempt to get the path for the given key
	folderPath, ok := pathDict[key]
	if !ok {
		return nil, errors.New(fmt.Sprintf("No path found for the tileset key %s", key))
	}

	tilePaths, err := getTilePaths(folderPath)
	if err != nil {
		return nil, err
	}

	for i, path := range tilePaths {
		tileImage, _, err := ebitenutil.NewImageFromFile(path)
		if err != nil {
			panic(err)
		}
		tileset[i] = tileImage
	}

	return tileset, nil
}

func getTilePaths(folderPath string) ([]string, error) {
	folderContents, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to open the tileset at %s", folderPath))
	}

	var tilePaths []string

	for _, fileInfo := range folderContents {
		// skip directories
		if fileInfo.IsDir() {
			continue
		}
		// skip non-PNGs
		if !strings.HasSuffix(fileInfo.Name(), ".png") {
			continue
		}
		fullPath := filepath.Join(folderPath, fileInfo.Name())
		tilePaths = append(tilePaths, fullPath)
	}

	return tilePaths, nil
}
