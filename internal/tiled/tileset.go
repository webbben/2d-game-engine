package tiled

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/general_util"
)

var tilePath = filepath.Join(config.GameAssetsPath(), "tiles")

func InitFileStructure() {
	if !general_util.FileExists(config.GameAssetsPath()) {
		err := os.MkdirAll(config.GameAssetsPath(), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	// tiles
	if !general_util.FileExists(tilePath) {
		err := os.MkdirAll(tilePath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func LoadTileset(tilesetJsonPath string) error {
	bytes, err := os.ReadFile(tilesetJsonPath)
	if err != nil {
		return fmt.Errorf("failed to read tileset json file: %w", err)
	}

	var tileset Tileset
	err = json.Unmarshal(bytes, &tileset)
	if err != nil {
		return fmt.Errorf("failed to unmarshal tileset json: %w", err)
	}

	// open the source image file
	f, err := os.Open(tileset.Image)
	if err != nil {
		return fmt.Errorf("failed to open tileset source image: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("failed to decode tileset source image data: %w", err)
	}
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	tileIndex := 0
	for y := 0; y < imgHeight; y += tileset.TileHeight {
		for x := 0; x < imgWidth; x += tileset.TileWidth {
			rect := image.Rect(x, y, x+tileset.TileWidth, y+tileset.TileHeight)

			// extract the tile image
			tileImg := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(rect)

			outPath := filepath.Join(tilePath, tileset.Name, fmt.Sprintf("%v.png", tileIndex))
			outFile, err := os.Create(outPath)
			if err != nil {
				return fmt.Errorf("failed to create tile image: %w", err)
			}
			if err := png.Encode(outFile, tileImg); err != nil {
				outFile.Close()
				return fmt.Errorf("error while encoding tile image: %w", err)
			}
			outFile.Close()

			tileIndex++
		}
	}

	fmt.Printf("Created tile images for tileset %s (%v tiles created)\n", tileset.Name, tileIndex)
	return nil
}
