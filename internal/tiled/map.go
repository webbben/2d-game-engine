package tiled

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

func OpenMap(mapSource string) (Map, error) {
	mapSource, err := filepath.Abs(mapSource)
	if err != nil {
		return Map{}, fmt.Errorf("failed to resolve absolute path for map source: %w", err)
	}
	bytes, err := os.ReadFile(mapSource)
	if err != nil {
		return Map{}, fmt.Errorf("error reading map source file: %w", err)
	}
	var m Map
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return Map{}, fmt.Errorf("error reading map source JSON data: %w", err)
	}
	m.AbsSourcePath = mapSource
	return m, nil
}

// Load a map and all its tilesets or other pre-processable data
func (m *Map) Load(regenerateImages bool) error {
	// load property data
	id, found := GetStringProperty("ID", m.Properties)
	if !found {
		panic("Map required property not found: ID. Be sure to set this as a custom property within Tiled.")
	}
	displayName, found := GetStringProperty("DisplayName", m.Properties)
	if !found {
		panic("Map required property not found: DisplayName. Be sure to set this as a custom property within Tiled.")
	}

	m.ID = id
	m.DisplayName = displayName

	// ensure all tilesets have been loaded and created
	for i, tileset := range m.Tilesets {
		err := tileset.LoadJSONData(m.AbsSourcePath)
		if err != nil {
			return err
		}
		// ensure tilesets are regenerated on game startup, just in case source image files were changed since last play
		if regenerateImages || !TilesetExists(tileset.Name) {
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

	m.CollisionRects = make([][]CollisionRect, m.Height)
	for i := range m.Height {
		m.CollisionRects[i] = make([]CollisionRect, m.Width)
	}

	// find all collision rects
	for _, layer := range m.Layers {
		for i, d := range layer.Data {
			tile, found := m.GetTileByGID(d)
			if !found {
				continue
			}
			collisionVal, found := GetStringProperty("COLLISION", tile.Properties)
			if !found {
				continue
			}

			x := i % layer.Width
			y := i / layer.Width
			cr := NewCollisionRect(collisionVal)
			m.CollisionRects[y][x] = cr
		}
	}

	m.CalculateCostMap()

	m.Loaded = true
	return nil
}

func (m *Map) LoadTileImageMap() error {
	m.TileImageMap = make(map[int]TileData)

	for _, tileset := range m.Tilesets {
		if !TilesetExists(tileset.Name) {
			return fmt.Errorf("tileset %s not found in tilesets data directory", tileset.Name)
		}

		for i := 0; i < tileset.TileCount; i++ {
			tileData := TileData{}
			tileData.ID = i
			tileImg, err := tileset.GetTileImage(i)
			if err != nil {
				return err
			}
			tileData.CurrentFrame = tileImg

			// check if this tile has an animation
			for _, tile := range tileset.Tiles {
				if tile.ID != i {
					continue
				}
				if len(tile.Animation) > 0 {
					tileData.Frames = []tileAnimation{}
					// found some animation frames
					for _, animationFrame := range tile.Animation {
						frameImg, err := tileset.GetTileImage(animationFrame.TileID)
						if err != nil {
							return err
						}
						tileData.Frames = append(tileData.Frames, tileAnimation{
							DurationMs: animationFrame.Duration,
							Image:      frameImg,
						})
					}
				}

			}

			m.TileImageMap[i+tileset.FirstGID] = tileData
		}
	}

	return nil
}

func (m *Map) Update() {
	for key, tileData := range m.TileImageMap {
		tileData.UpdateFrame()
		m.TileImageMap[key] = tileData
	}
}

func (m Map) DrawGroundLayers(screen *ebiten.Image, offsetX float64, offsetY float64) {
	if !m.Loaded {
		log.Fatal("map not loaded! ensure map is loaded before drawing layers")
	}
	if m.TileImageMap == nil {
		log.Fatal("tileImageMap is nil! ensure the tile images are loaded into memory before drawing layers")
	}

	// draw all tile layers except:
	// - the BUILDING_ROOF layer: parts of buildings that the player can walk behind
	for _, layer := range m.Layers {
		if layer.Type != LAYER_TYPE_TILE {
			continue
		}
		if layer.Name == "BUILDING_ROOF" {
			continue
		}

		m.drawTileLayer(screen, offsetX, offsetY, layer)
	}
}

func (m Map) DrawRooftopLayer(screen *ebiten.Image, offsetX, offsetY float64) {
	if !m.Loaded {
		log.Fatal("map not loaded! ensure map is loaded before drawing layers")
	}
	if m.TileImageMap == nil {
		log.Fatal("tileImageMap is nil! ensure the tile images are loaded into memory before drawing layers")
	}

	for _, layer := range m.Layers {
		if layer.Name == "BUILDING_ROOF" {
			m.drawTileLayer(screen, offsetX, offsetY, layer)
			break
		}
	}
}

func (m Map) drawTileLayer(screen *ebiten.Image, offsetX, offsetY float64, layer Layer) {
	if layer.Type != "tilelayer" {
		return
	}

	if len(layer.Data) != layer.Width*layer.Height {
		log.Fatalf("the layer data array is not the correct size; size=%v, expected=%v", len(layer.Data), layer.Width*layer.Height)
	}

	// index of the tile in the layer's Data array
	// it's important that this value is always updated correctly if we skip any rows or columns, or else tiles will render in the wrong places
	i := 0
	for y := 0; y < layer.Height; y++ {
		// skip this row if it's above the camera
		if rendering.RowAboveCameraView(float64(y), offsetY) {
			i += layer.Width
			continue
		}
		// skip all remaining rows if it's below the camera
		if rendering.RowBelowCameraView(float64(y), offsetY) {
			break
		}
		drawY := float64(y*config.TileSize) - offsetY

		for x := 0; x < layer.Width; x++ {
			if rendering.ColBeforeCameraView(float64(x), offsetX) {
				i++
				continue
			}
			// skip the rest of the columns if it's past the screen
			if rendering.ColAfterCameraView(float64(x), offsetX) {
				i += layer.Width - x
				break
			}

			drawX := float64(x*config.TileSize) - offsetX

			tileGID := layer.Data[i]
			if tileGID == 0 {
				// 0 means no tile is placed here
				i++
				continue
			}

			tileData, exists := m.TileImageMap[tileGID]
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
			screen.DrawImage(tileData.CurrentFrame, op)

			i++
		}
	}
}

type CollisionRect struct {
	IsCollision bool
	Rect        model.Rect
}

func (cr CollisionRect) OffsetRect(x, y float64) model.Rect {
	return model.Rect{
		X: cr.Rect.X + x,
		Y: cr.Rect.Y + y,
		W: cr.Rect.W,
		H: cr.Rect.H,
	}
}

func NewCollisionRect(collisionType string) CollisionRect {
	if collisionType == "" {
		return CollisionRect{IsCollision: false}
	}

	cr := CollisionRect{IsCollision: true}
	switch collisionType {
	case "HALF_L":
		cr.Rect = model.Rect{X: 0, Y: 0, W: config.TileSize / 2, H: config.TileSize}
	case "HALF_R":
		cr.Rect = model.Rect{X: config.TileSize / 2, Y: 0, W: config.TileSize / 2, H: config.TileSize}
	case "HALF_T":
		cr.Rect = model.Rect{X: 0, Y: 0, W: config.TileSize, H: config.TileSize / 2}
	case "HALF_B":
		cr.Rect = model.Rect{X: 0, Y: config.TileSize / 2, W: config.TileSize, H: config.TileSize / 2}
	case "WHOLE":
		cr.Rect = model.Rect{X: 0, Y: 0, W: config.TileSize, H: config.TileSize}
	case "CORNER_TL":
		cr.Rect = model.Rect{X: 0, Y: 0, W: config.TileSize / 2, H: config.TileSize / 2}
	case "CORNER_TR":
		cr.Rect = model.Rect{X: config.TileSize / 2, Y: 0, W: config.TileSize / 2, H: config.TileSize / 2}
	case "CORNER_BR":
		cr.Rect = model.Rect{X: config.TileSize / 2, Y: config.TileSize / 2, W: config.TileSize / 2, H: config.TileSize / 2}
	case "CORNER_BL":
		cr.Rect = model.Rect{X: 0, Y: config.TileSize / 2, W: config.TileSize / 2, H: config.TileSize / 2}
	case "CENTER_POLE":
		cr.Rect = model.Rect{X: config.TileSize / 4, Y: 0, W: config.TileSize / 2, H: config.TileSize}
	default:
		panic("collision rect type not found")
	}

	return cr
}

// CalculateCostMap calculates the cost for each position in the map.
// The CostMap is required for being able to calculate things such as path finding and collisions.
// Mainly used by NPCs, since the player has more dynamic movement ability (can partially move in tiles)
func (m *Map) CalculateCostMap() {
	if m.Height == 0 {
		panic("map cannot have a height (num rows) of 0")
	}
	if m.Width == 0 {
		panic("map cannot have a width (num cols) of 0")
	}

	m.CostMap = make([][]int, m.Height)
	for i := 0; i < m.Height; i++ {
		m.CostMap[i] = make([]int, m.Width)
	}

	// any tile that has a collision rect is blocked for NPCs
	for y, row := range m.CollisionRects {
		for x, r := range row {
			if r.IsCollision {
				m.CostMap[y][x] += 10
			}
		}
	}

	// Go through each layer and add any 'cost' properties up
	for _, layer := range m.Layers {
		i := 0
		for y := 0; y < layer.Height; y++ {
			for x := 0; x < layer.Width; x++ {
				tile, found := m.GetTileByGID(layer.Data[i])
				if found {
					for _, prop := range tile.Properties {
						if prop.Name == "cost" {
							m.CostMap[y][x] += prop.GetIntValue()
						}
					}
				}

				i++
			}
		}
	}
}

func (m Map) GetAdjTiles(c model.Coords) []model.Coords {
	l := c.GetAdj('L')
	r := c.GetAdj('R')
	u := c.GetAdj('U')
	d := c.GetAdj('D')
	adjTiles := []model.Coords{}
	if m.isWithinMapBounds(l) {
		adjTiles = append(adjTiles, l)
	}
	if m.isWithinMapBounds(r) {
		adjTiles = append(adjTiles, r)
	}
	if m.isWithinMapBounds(u) {
		adjTiles = append(adjTiles, u)
	}
	if m.isWithinMapBounds(d) {
		adjTiles = append(adjTiles, d)
	}
	return adjTiles
}

func (m Map) isWithinMapBounds(c model.Coords) bool {
	if c.X < 0 || c.X >= m.Width {
		return false
	}
	if c.Y < 0 || c.Y >= m.Height {
		return false
	}
	return true
}
