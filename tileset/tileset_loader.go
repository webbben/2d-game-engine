package tileset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// texture tilesets
const (
	Tx_Grass_01 string = "grass_01"
)

// entity tilesets
const (
	Ent_Player string = "ent_player"
)

// object tilesets
const (
	Ob_Trees_01 string = "trees_01"
)

var (
	// maps a tileset key to its path
	pathDict = map[string]string{
		Tx_Grass_01: "./tileset/textures/floors/grass_01",
		Ent_Player:  "./tileset/entities/player",
		Ob_Trees_01: "./tileset/objects/nature/trees_01",
	}
)

// loads a tileset for the given tileset key
func LoadTileset(key string) (map[string]*ebiten.Image, error) {
	tileset := make(map[string]*ebiten.Image)

	// attempt to get the path for the given key
	folderPath, ok := pathDict[key]
	if !ok {
		return nil, fmt.Errorf("no path found for the tileset key %s", key)
	}

	tilePaths, err := getTilePaths(folderPath)
	if err != nil {
		return nil, err
	}

	for _, path := range tilePaths {
		tileImage, _, err := ebitenutil.NewImageFromFile(path)
		if err != nil {
			panic(err)
		}
		fileNameWithExt := filepath.Base(path)
		withoutExt := strings.TrimSuffix(fileNameWithExt, filepath.Ext(fileNameWithExt))
		tileset[withoutExt] = tileImage
	}

	return tileset, nil
}

func getTilePaths(folderPath string) ([]string, error) {
	folderContents, err := os.ReadDir(folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the tileset at %s", folderPath)
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
