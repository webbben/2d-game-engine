package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/npc"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/player"
)

// information about the current room the player is in
type MapInfo struct {
	MapIsActive bool // flag indicating if the map is actively being used (e.g. for rendering, updates, etc)
	Map         tiled.Map
	ImageMap    map[string]*ebiten.Image // the map of images (tiles) used in rendering the current room
	PlayerRef   *player.Player
	Objects     []*object.Object

	sortedRenderables []sortedRenderable

	Lights []*lights.Light

	NPCManager
}

type sortedRenderable interface {
	Y() float64
	Draw(screen *ebiten.Image, offsetX, offsetY float64)
}

// Do all required setup for creating MapInfo and preparing it for use.
// Note: if the runBackgroundJobs flag is set to true, this is where the background jobs loop is started.
func SetupMap(mi MapInfo, mapSource string) (*MapInfo, error) {
	// load and setup the map
	m, err := tiled.OpenMap(mapSource)
	if err != nil {
		return nil, fmt.Errorf("error while opening Tiled map: %w", err)
	}

	err = m.Load()
	if err != nil {
		return nil, fmt.Errorf("error while loading map: %w", err)
	}
	mi.Map = m
	mi.NPCManager.mapRef = mi.Map

	lightProps := []tiled.LightProps{}

	// find all lights embedded in tiles
	for _, tileset := range m.Tilesets {
		for _, tile := range tileset.Tiles {
			// determine if this is a light tile
			tileType := tiled.GetTileType(tile)
			if tileType == "LIGHT" {
				tileProps := tiled.GetLightPropsFromTile(tile)
				tileProps.TileID += tileset.FirstGID
				lightProps = append(lightProps, tileProps)
			}
		}
	}

	// find the positions of the tiles where the lights are set
	for _, lightProp := range lightProps {
		lightPositions := m.GetAllTilePositions(lightProp.TileID)
		for _, pos := range lightPositions {
			l := lights.NewLight(pos.X*config.TileSize+(config.TileSize/2), (pos.Y*config.TileSize)+lightProp.OffsetY+(config.TileSize/2), lightProp, nil)
			mi.Lights = append(mi.Lights, &l)
			fmt.Println("light")
			fmt.Println("x:", l.X, "y:", l.Y)
			fmt.Println("radius:", l.MinRadius, l.MaxRadius)
		}
	}

	// find all objects in the map
	for _, layer := range m.Layers {
		if layer.Type == tiled.LAYER_TYPE_OBJECT {
			for _, obj := range layer.Objects {
				mi.Objects = append(mi.Objects, object.LoadObject(obj))
			}
		}
	}

	// start up background jobs loop
	if mi.NPCManager.backgroundJobsRunning {
		panic("backgroundJobsRunning flag is already true while initializing map")
	}
	if mi.NPCManager.RunBackgroundJobs {
		mi.NPCManager.startBackgroundNPCManager()
	}

	return &mi, nil
}

func (mi *MapInfo) CloseMap() {
	mi.NPCManager.RunBackgroundJobs = false
}

type NPCManager struct {
	NPCs []*npc.NPC // the NPC entities in the map

	mapRef       tiled.Map // map info so we can get map size, tile adjacency, etc
	nextPriority int       // the next priority value to assign to an NPC

	StuckNPCs []string // IDs of NPCs who are currently stuck (while trying to execute a task)

	// if true, the background jobs goroutine will run.
	// if false, the background jobs goroutine will stop.
	RunBackgroundJobs     bool
	backgroundJobsRunning bool // flag that indicates if background jobs loop already running.
}

func (mi *MapInfo) ResolveNPCJams() {
	// get all existing NPC jams
	npcJams := mi.NPCManager.findNPCJams()
	if len(npcJams) == 0 {
		// no jams found
		return
	}

	// resolve each jam, one at a time
	for _, npcJam := range npcJams {
		mi.NPCManager.resolveJam(npcJam)
	}
}

// The official way to add the player to a map
func (mi *MapInfo) AddPlayerToMap(p *player.Player, startPos model.Coords) {
	r := model.Rect{
		X: float64(startPos.X * config.TileSize),
		Y: float64(startPos.Y * config.TileSize),
		W: config.TileSize,
		H: config.TileSize,
	}
	if mi.Collides(r, "", false) {
		// this also handles placement outside of map bounds
		panic("player added to map on colliding tile")
	}
	mi.PlayerRef = p
	p.Entity.World = mi
	p.Entity.SetPosition(startPos)
}

func (mi *MapInfo) AddNPCToMap(n *npc.NPC, startPos model.Coords) {
	r := model.Rect{
		X: float64(startPos.X * config.TileSize),
		Y: float64(startPos.Y * config.TileSize),
		W: config.TileSize,
		H: config.TileSize,
	}
	if mi.Collides(r, "", false) {
		panic("npc added to map on colliding tile")
	}
	n.Entity.World = mi
	n.World = mi // NPC has its own world context it needs, that isn't relevant to entity
	n.Entity.SetPosition(startPos)
	n.Priority = mi.NPCManager.getNextNPCPriority()
	mi.NPCs = append(mi.NPCs, n)
}

func (mi MapInfo) Collides(r model.Rect, excludeEntId string, rectBased bool) bool {
	// check map's CostMap
	maxY := len(mi.Map.CostMap) * config.TileSize
	maxX := len(mi.Map.CostMap[0]) * config.TileSize
	if r.Y < 0 || r.Y+r.H > float64(maxY) || r.X < 0 || r.X+r.W > float64(maxX) {
		// attempting to move past the edge of the map
		return true
	}

	tl := model.ConvertPxToTilePos(int(r.X), int(r.Y))
	tr := model.ConvertPxToTilePos(int(r.X+r.W), int(r.Y))
	bl := model.ConvertPxToTilePos(int(r.X), int(r.Y+r.H))
	br := model.ConvertPxToTilePos(int(r.X+r.W), int(r.Y+r.H))

	if rectBased {
		r1 := mi.mapRef.CollisionRects[tl.Y][tl.X]
		if r1.IsCollision {
			if r1.OffsetRect(float64(tl.X*config.TileSize), float64(tl.Y*config.TileSize)).Intersects(r) {
				return true
			}
		}
		r2 := mi.mapRef.CollisionRects[tr.Y][tr.X]
		if r2.IsCollision {
			if r2.OffsetRect(float64(tr.X*config.TileSize), float64(tr.Y*config.TileSize)).Intersects(r) {
				return true
			}
		}
		r3 := mi.mapRef.CollisionRects[bl.Y][bl.X]
		if r3.IsCollision {
			if r3.OffsetRect(float64(bl.X*config.TileSize), float64(bl.Y*config.TileSize)).Intersects(r) {
				return true
			}
		}
		r4 := mi.mapRef.CollisionRects[br.Y][br.X]
		if r4.IsCollision {
			if r4.OffsetRect(float64(br.X*config.TileSize), float64(br.Y*config.TileSize)).Intersects(r) {
				return true
			}
		}
	} else {
		if mi.Map.CostMap[tl.Y][tl.X] >= 10 {
			return true
		}
		if mi.Map.CostMap[tr.Y][tr.X] >= 10 {
			return true
		}
		if mi.Map.CostMap[bl.Y][bl.X] >= 10 {
			return true
		}
		if mi.Map.CostMap[br.Y][br.X] >= 10 {
			return true
		}
	}

	// check entity positions
	if mi.PlayerRef != nil {
		if mi.PlayerRef.Entity.ID != excludeEntId {
			if r.Intersects(mi.PlayerRef.Entity.CollisionRect()) {
				return true
			}
		}
	}
	for _, n := range mi.NPCs {
		if n.Entity.ID == excludeEntId {
			continue
		}
		if r.Intersects(n.Entity.CollisionRect()) {
			return true
		}
	}

	return false
}

// Returns a path to the goal, or if it cannot be reached, a path to the closest reachable position.
// The boolean indicates if the goal was successfully reached.
func (mi MapInfo) FindPath(start, goal model.Coords) ([]model.Coords, bool) {
	return path_finding.FindPath(start, goal, mi.CostMap())
}

func (mi MapInfo) MapDimensions() (width int, height int) {
	return mi.Map.Width, mi.Map.Height
}

// Gets a cost map that includes entity positions
func (mi MapInfo) CostMap() [][]int {
	if mi.Map.CostMap == nil {
		panic("tried to get MapInfo cost map before Map costmap was created")
	}
	// make deep copy so that original cost map isn't altered
	costMap := make([][]int, len(mi.Map.CostMap))
	for i := range mi.Map.CostMap {
		costMap[i] = append([]int{}, mi.Map.CostMap[i]...)
	}

	for _, n := range mi.NPCs {
		costMap[n.Entity.TilePos.Y][n.Entity.TilePos.X] += 10
		// if the entity is currently moving, mark its destination tile as a collision too
		if !n.Entity.Movement.TargetTile.Equals(n.Entity.TilePos) {
			costMap[n.Entity.Movement.TargetTile.Y][n.Entity.Movement.TargetTile.X] += 10
		}
	}

	// TODO should this be centered under the player instead?
	playerPos := mi.PlayerRef.Entity.TilePos
	costMap[playerPos.Y][playerPos.X] += 10

	return costMap
}

func (mi *MapInfo) GetLights() []*lights.Light {
	return mi.Lights
}
