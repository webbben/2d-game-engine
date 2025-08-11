package tiled

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

func OpenMap(mapSource string) (Map, error) {
	bytes, err := os.ReadFile(mapSource)
	if err != nil {
		return Map{}, fmt.Errorf("error reading map source file: %w", err)
	}
	var m Map
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return Map{}, fmt.Errorf("error reading map source JSON data: %w", err)
	}
	return m, nil
}

// Load a map and all its tilesets or other pre-processable data
func (m *Map) Load() error {
	m.TileImageMap = make(map[int]*ebiten.Image)

	// ensure all tilesets have been loaded and created
	for i, tileset := range m.Tilesets {
		err := tileset.LoadJSONData()
		if err != nil {
			return err
		}
		if !TilesetExists(tileset.Name) {
			err = tileset.GenerateTiles()
			if err != nil {
				return err
			}
		}

		m.Tilesets[i] = tileset
	}

	err := m.LoadTileImageMap()
	if err != nil {
		return err
	}

	m.Loaded = true
	return nil
}

func (m *Map) LoadTileImageMap() error {
	m.TileImageMap = make(map[int]*ebiten.Image)

	for _, tileset := range m.Tilesets {
		if !TilesetExists(tileset.Name) {
			return fmt.Errorf("tileset %s not found in tilesets data directory", tileset.Name)
		}

		for i := 0; i < tileset.TileCount; i++ {
			tileImg, _, err := ebitenutil.NewImageFromFile(filepath.Join(tilePath, tileset.Name, fmt.Sprintf("%v.png", i)))
			if err != nil {
				return fmt.Errorf("failed to load tile image: %w", err)
			}
			m.TileImageMap[i+tileset.FirstGID] = tileImg
		}
	}

	return nil
}

func (m Map) DrawLayers(screen *ebiten.Image, offsetX float64, offsetY float64) {
	if !m.Loaded {
		log.Fatal("map not loaded! ensure map is loaded before drawing layers")
	}
	if m.TileImageMap == nil {
		log.Fatal("tileImageMap is nil! ensure the tile images are loaded into memory before drawing layers")
	}

	for _, layer := range m.Layers {
		if layer.Type != "tilelayer" {
			continue
		}

		if len(layer.Data) != layer.Width*layer.Height {
			log.Fatalf("the layer data array is not the correct size; size=%v, expected=%v", len(layer.Data), layer.Width*layer.Height)
		}

		i := 0
		for y := 0; y < layer.Height; y++ {
			// skip this row if it's above the camera
			if rendering.RowAboveCameraView(float64(y), offsetY) {
				continue
			}
			// skip all remaining rows if it's below the camera
			if rendering.RowBelowCameraView(float64(y), offsetY) {
				break
			}
			drawY := float64(y*config.TileSize) - offsetY

			for x := 0; x < layer.Width; x++ {
				if rendering.ColBeforeCameraView(float64(x), offsetX) {
					continue
				}
				// skip the rest of the columns if it's past the screen
				if rendering.ColAfterCameraView(float64(x), offsetX) {
					break
				}

				drawX := float64(x*config.TileSize) - offsetX

				tileGID := layer.Data[i]
				if tileGID == 0 {
					// 0 means no tile is placed here
					continue
				}

				tileImg, exists := m.TileImageMap[tileGID]
				if !exists {
					keys := make([]int, 0, len(m.TileImageMap))
					for k := range m.TileImageMap {
						keys = append(keys, k)
					}
					fmt.Println(keys)
					log.Fatalf("tile GID (%v) not found in TileImageMap; was there an error during tileset initialization?", tileGID)
				}

				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(drawX, drawY)
				op.GeoM.Scale(config.GameScale, config.GameScale)
				screen.DrawImage(tileImg, op)

				i++
			}
		}
	}
}
