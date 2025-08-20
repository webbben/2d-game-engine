package tiled

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

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

func TilesetExists(tilesetName string) bool {
	return general_util.FileExists(filepath.Join(tilePath, tilesetName))
}

func (t *Tileset) LoadJSONData() error {
	// initially, when loading from a map file, we only have these two values:
	// 1. FirstGID
	// 2. Source

	// start by loading JSON data from source

	// TODO - this is a temporary hack. once we have a decided data directory, change how we get the source path.
	if strings.HasPrefix(t.Source, "../tilesets/") {
		t.Source = strings.TrimPrefix(t.Source, "../tilesets/")
		t.Source = "assets/tiled/tilesets/" + t.Source
	}

	var loaded Tileset
	bytes, err := os.ReadFile(t.Source)
	if err != nil {
		return fmt.Errorf("failed to read source JSON file: %w", err)
	}
	err = json.Unmarshal(bytes, &loaded)
	if err != nil {
		return fmt.Errorf("failed to unmarshal source JSON file: %w", err)
	}

	// TODO - replace this hack with the long term approach for storing tileset images
	if strings.HasPrefix(loaded.Image, "../../images/tilesets/") {
		loaded.Image = strings.TrimPrefix(loaded.Image, "../../")
		loaded.Image = "assets/" + loaded.Image
	}

	// put the two initial values into this loaded one, and replace the original Tileset data
	loaded.FirstGID = t.FirstGID
	loaded.Source = t.Source
	*t = loaded

	return nil
}

func LoadTileset(source string) (Tileset, error) {
	t := Tileset{
		FirstGID: 0,
		Source:   source,
	}
	err := t.LoadJSONData()
	return t, err
}

func (tileset *Tileset) GenerateTiles() error {
	if tileset.Image == "" {
		return errors.New("tileset JSON data not loaded yet")
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

	tilesetDir := filepath.Join(tilePath, tileset.Name)

	// delete the directory and its contents if it already exists
	if general_util.FileExists(tilesetDir) {
		os.RemoveAll(tilesetDir)
	}

	err = os.MkdirAll(tilesetDir, os.ModePerm)
	if err != nil {
		return err
	}

	tileset.GeneratedImagesPath = tilesetDir

	tileIndex := 0
	for y := 0; y < imgHeight; y += tileset.TileHeight {
		for x := 0; x < imgWidth; x += tileset.TileWidth {
			rect := image.Rect(x, y, x+tileset.TileWidth, y+tileset.TileHeight)

			// extract the tile image
			tileImg := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(rect)

			outPath := filepath.Join(tilesetDir, fmt.Sprintf("%v.png", tileIndex))
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
