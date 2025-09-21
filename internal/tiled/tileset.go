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

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/model"
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

// load the tileset data from it's JSON file. The JSON file should be defined in the Tileset already, at t.Source (this is how the data is saved in Tiled maps).
// If loading the tileset from a map file, pass the absolute path of the map file.  This is because t.Source specifies a relative path to the source image.
// So, we need to be able to construct the absolute path to that file.
func (t *Tileset) LoadJSONData(mapAbsPath string) error {
	// initially, when loading from a map file, we only have these two values:
	// 1. FirstGID
	// 2. Source

	// if the path is not absolute, follow the relative path starting from the map file's absolute path
	if !filepath.IsAbs(t.Source) {
		if mapAbsPath == "" {
			// if no map path is passed, then we don't need to do any back-tracking. just use t.Source.
			fmt.Println("absPath is empty")
			p, err := filepath.Abs(t.Source)
			if err != nil {
				return err
			}
			t.Source = p
		} else {
			mapAbsPath = filepath.Dir(mapAbsPath)
			t.Source = filepath.Clean(filepath.Join(mapAbsPath, t.Source))
		}
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

	if !filepath.IsAbs(loaded.Image) {
		// if the source image for the tileset has a relative path, compute absolute path based on tileset absolute path
		loaded.Image = filepath.Clean(filepath.Join(filepath.Dir(t.Source), loaded.Image))
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
	err := t.LoadJSONData("")
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

// Gets Tile from Tileset, by GID.
//
// This is for Tile properties, animations, etc, NOT for the tile image. Not all tiles will have a Tile.
// Returns the Tile if found, and a boolean indicating if the tile was successfully found
func (m Map) GetTileByGID(gid int) (Tile, bool) {
	for _, tileset := range m.Tilesets {
		localTileId := gid - tileset.FirstGID
		for _, tile := range tileset.Tiles {
			if tile.ID == localTileId {
				return tile, true
			}
		}
	}
	return Tile{}, false
}

// given a gid for a tile, returns the coordinates for all places that this tile is placed in a map (in any tile layer).
// used for positioning things like lights which are embedded in certain tiles
func (m Map) GetAllTilePositions(gid int) []model.Coords {
	coords := []model.Coords{}
	for _, layer := range m.Layers {
		if layer.Type != LAYER_TYPE_TILE {
			continue
		}
		for i, dataGID := range layer.Data {
			if dataGID == gid {
				y := i / layer.Width
				x := i % layer.Width
				coords = append(coords, model.Coords{
					X: x,
					Y: y,
				})
			}
		}
	}

	return coords
}

func (t Tileset) GetTileImage(id int) (*ebiten.Image, error) {
	tileImg, _, err := ebitenutil.NewImageFromFile(filepath.Join(tilePath, t.Name, fmt.Sprintf("%v.png", id)))
	if err != nil {
		return nil, fmt.Errorf("failed to load tile image: %w", err)
	}
	return tileImg, nil
}

type LightProps struct {
	TileID            int
	ColorPreset       string
	GlowFactor        float64
	InnerRadiusFactor float64
	OffsetY           int
	Radius            int
}

func GetTileType(tile Tile) string {
	for _, prop := range tile.Properties {
		if prop.Name == "TYPE" {
			return prop.GetStringValue()
		}
	}

	return ""
}

func GetLightPropsFromTile(tile Tile) LightProps {
	props := LightProps{
		TileID: tile.ID,
	}

	for _, prop := range tile.Properties {
		switch prop.Name {
		case "light_color_preset":
			props.ColorPreset = prop.GetStringValue()
		case "light_glow_factor":
			props.GlowFactor = prop.GetFloatValue()
		case "light_offset_y":
			props.OffsetY = prop.GetIntValue()
		case "light_radius":
			props.Radius = prop.GetIntValue()
		case "light_inner_radius_factor":
			props.InnerRadiusFactor = prop.GetFloatValue()
		}
	}

	return props
}

func GetBoolProperty(propName string, props []Property) (found, value bool) {
	for _, prop := range props {
		if prop.Name == propName {
			return true, prop.GetBoolValue()
		}
	}

	return false, false
}
