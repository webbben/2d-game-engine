package tiled

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	imagePkg "github.com/webbben/2d-game-engine/imgutil/image"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
)

func TilesetExists(tilesetName string) bool {
	return config.FileExists(config.ResolveTilePath(tilesetName))
}

// LoadJSONData loads the tileset data from it's JSON file. The JSON file path should be defined in the Tileset already, at t.Source (this is how the data is saved in Tiled maps).
// If loading the tileset from a map file, pass the absolute path of the map file.  This is because t.Source specifies a relative path to the source image.
// So, we need to be able to construct the absolute path to that file.
func (t *Tileset) LoadJSONData(mapAbsPath string) error {
	// initially, when loading from a map file, we only have these two values:
	// 1. FirstGID
	// 2. Source

	if mapAbsPath != "" {
		if filepath.Ext(mapAbsPath) != ".tmj" {
			panic("invalid absolute map path; file doesn't end with .tmj: " + mapAbsPath)
		}
	}

	originalSrc := t.Source

	// if the path is not absolute, follow the relative path starting from the map file's absolute path
	if !filepath.IsAbs(t.Source) {
		if mapAbsPath == "" {
			// if no map path is passed, then we are not loading a relative path from a tiled map file
			// resolve the full path as usual for a tileset
			t.Source = config.ResolveTilesetPath(t.Source)
		} else {
			mapAbsPath = filepath.Dir(mapAbsPath)
			t.Source = filepath.Clean(filepath.Join(mapAbsPath, t.Source))
		}
	}

	if filepath.Ext(t.Source) != ".tsj" {
		logz.Panicf("tileset source does not have a .tsj extension. Is it invalid? %s (original: %s)", t.Source, originalSrc)
	}

	var loaded Tileset
	bytes, err := os.ReadFile(t.Source)
	if err != nil {
		return fmt.Errorf("failed to read source JSON file: %w (mapAbsPath=%s)", err, mapAbsPath)
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
	loaded.Loaded = true
	*t = loaded

	t.validate()

	return nil
}

func LoadTileset(source string) (Tileset, error) {
	if source == "" {
		return Tileset{}, errors.New("no source passed to LoadTileset")
	}
	t := Tileset{
		FirstGID: 0,
		Source:   source,
	}
	err := t.LoadJSONData("")

	if !TilesetExists(t.Name) {
		err = t.GenerateTiles()
	}

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
	defer func() { _ = f.Close() }()

	img, _, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("failed to decode tileset source image data: %w", err)
	}
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	tilesetDir := config.ResolveTilePath(tileset.Name)

	// delete the directory and its contents if it already exists
	if config.FileExists(tilesetDir) {
		_ = os.RemoveAll(tilesetDir)
	}

	err = os.MkdirAll(tilesetDir, os.ModePerm)
	if err != nil {
		return err
	}

	tileset.GeneratedImagesPath = tilesetDir

	tileIndex := 0
	for y := 0; y < imgHeight; y += tileset.TileHeight {
		for x := 0; x < imgWidth; x += tileset.TileWidth {
			if x+tileset.TileWidth > imgWidth {
				// the image of the tileset does not evenly divide by the tile size;
				// so, this last "tile" is actually a partial one that should be skipped.
				// this is mainly because Tiled does not count these and skips over them.
				logz.Println(tileset.Name, "tileset image does not evenly divide by tilesize. skipping a 'partial tile' at the end of a row.")
				break
			}
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
				_ = outFile.Close()
				return fmt.Errorf("error while encoding tile image: %w", err)
			}
			_ = outFile.Close()

			tileIndex++
		}
	}

	fmt.Printf("Created tile images for tileset %s (%v tiles created)\n", tileset.Name, tileIndex)
	return nil
}

// GetTileByGID Gets Tile from Tileset, by GID.
//
// This is for Tile properties, animations, etc, NOT for the tile image. Not all tiles will have a Tile.
// Returns the Tile if found, and a boolean indicating if the tile was successfully found
func (m Map) GetTileByGID(gid int) (Tile, Tileset, bool) {
	for _, tileset := range m.Tilesets {
		if !tileset.Loaded {
			if m.AbsSourcePath != "" {
				err := tileset.LoadJSONData(m.AbsSourcePath)
				if err != nil {
					panic(err)
				}
			} else {
				panic("tried to get tile from tileset before tileset was loaded!")
			}
		}
		localTileID := gid - tileset.FirstGID
		for _, tile := range tileset.Tiles {
			if tile.ID == localTileID {
				return tile, tileset, true
			}
		}
	}
	return Tile{}, Tileset{}, false
}

// FindTilesetForGID tries to find a tileset that has the correct ID range for the given GID
func (m Map) FindTilesetForGID(gid int) (Tileset, bool) {
	for _, tileset := range m.Tilesets {
		if !tileset.Loaded {
			panic("tried to get tileset for GID before tileset was loaded!")
		}
		if tileset.FirstGID >= gid && gid < tileset.FirstGID+tileset.TileCount {
			return tileset, true
		}
	}

	return Tileset{}, false
}

// GetAllTilePositions given a gid for a tile, returns the coordinates for all places that this tile is placed in a map (in any tile layer).
// used for positioning things like lights which are embedded in certain tiles
func (m Map) GetAllTilePositions(gid int) []model.Coords {
	coords := []model.Coords{}
	for _, layer := range m.Layers {
		if layer.Type != LayerTypeTile {
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

// GetTileImage is the core function to get a tile image from a tileset.
// The image value will be returned as nil if it was fully transparent or lacked dimensions.
func (t Tileset) GetTileImage(id int, panicOnEmpty bool) (*ebiten.Image, error) {
	if id < 0 {
		return nil, fmt.Errorf("tile id (%v) is less than 0", id)
	}
	if id > t.TileCount {
		return nil, fmt.Errorf("tile id (%v) is greater than the tileset's tile count (%v). are we getting a bad tile ID?", id, t.TileCount)
	}
	tileDir := config.ResolveTilePath(t.Name)
	imgFilePath := filepath.Join(tileDir, fmt.Sprintf("%v.png", id))
	tileImg, err := imagePkg.LoadImage(imgFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tile image: %w", err)
	}
	if panicOnEmpty && tileImg == nil {
		logz.Panicln("GetTileImage", "tile image was empty (fully transparent or lacking dimensions) when we expected it to have actual data in it:", imgFilePath)
	}
	return tileImg, nil
}

func GetTileImage(tilesetSrc string, tileID int, panicOnEmpty bool) *ebiten.Image {
	if tilesetSrc == "" {
		panic("no tilesetSrc passed")
	}
	tileset, err := LoadTileset(tilesetSrc)
	if err != nil {
		logz.Panicf("failed to load tileset: %s", err)
	}
	img, err := tileset.GetTileImage(tileID, panicOnEmpty)
	if err != nil {
		panic(err)
	}
	return img
}

type LightProps struct {
	ColorPreset       string
	GlowFactor        float64
	InnerRadiusFactor float64
	OffsetY           int
	Radius            int
	FlickerInterval   int
	MaxBrightness     float64
	CoreRadiusFactor  float64
}

func GetTileType(tile Tile) string {
	for _, prop := range tile.Properties {
		if prop.Name == "TYPE" {
			return prop.GetStringValue()
		}
	}

	return ""
}

func GetLightProps(p []Property) LightProps {
	props := LightProps{}

	for _, prop := range p {
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
		case "light_flicker_interval":
			props.FlickerInterval = prop.GetIntValue()
		case "light_max_brightness":
			props.MaxBrightness = prop.GetFloatValue()
		case "light_core_radius":
			props.CoreRadiusFactor = prop.GetFloatValue()
		}
	}

	return props
}

func GetBoolProperty(propName string, props []Property) (value, found bool) {
	for _, prop := range props {
		if prop.Name == propName {
			return prop.GetBoolValue(), true
		}
	}

	return false, false
}

func GetStringProperty(propName string, props []Property) (val string, found bool) {
	for _, prop := range props {
		if prop.Name == propName {
			return prop.GetStringValue(), true
		}
	}

	return "", false
}

func GetFloatProperty(propName string, props []Property) (val float64, found bool) {
	for _, prop := range props {
		if prop.Name == propName {
			return prop.GetFloatValue(), true
		}
	}

	return 0, false
}

func GetIntProperty(propName string, props []Property) (val int, found bool) {
	for _, prop := range props {
		if prop.Name == propName {
			return prop.GetIntValue(), true
		}
	}

	return 0, false
}

func GetTileBoolProperty(tilesetSrc string, tileIndex int, propName string) (val bool, found bool) {
	tileset, err := LoadTileset(tilesetSrc)
	if err != nil {
		panic(err)
	}

	for _, tile := range tileset.Tiles {
		if tile.ID != tileIndex {
			continue
		}
		return GetBoolProperty(propName, tile.Properties)
	}

	return false, false
}
