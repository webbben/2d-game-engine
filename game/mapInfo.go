package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/internal/pubsub"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/npc"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/player"
)

// information about the current room the player is in
type MapInfo struct {
	ID          string
	DisplayName string // the name of the map shown to the player
	Loaded      bool   // flag indicating if this map has been loaded
	ReadyToPlay bool   // flag indicating if all loading steps are done, and this map is ready to show in the game
	MapIsActive bool   // flag indicating if the map is actively being used (e.g. for rendering, updates, etc)
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

type OpenMapOptions struct {
	// set to true if this map should run a NPC manager background process.
	// this is not mandatory for using NPCs, just helps improve their behavior, especially when there are a lot of them in a map.
	RunNPCManager bool
}

// sets up a map and puts the player in it at the given position. meant for use once player already exists in game state
func (g *Game) EnterMap(mapID string, op *OpenMapOptions, playerSpawnIndex int) error {
	if g.MapInfo != nil {
		g.MapInfo.CloseMap()
	}

	err := g.SetupMap(mapID, op)
	if err != nil {
		return err
	}

	return g.PlacePlayerAtSpawnPoint(&g.Player, playerSpawnIndex)
}

// prepare the MapInfo for in-game play
func (g *Game) SetupMap(mapID string, op *OpenMapOptions) error {
	if op == nil {
		op = &OpenMapOptions{}
	}

	if g.MapInfo != nil {
		g.MapInfo.CloseMap()
	}

	mapSource := config.ResolveMapPath(mapID)
	fmt.Println("map source:", mapSource)

	// load and setup the map
	m, err := tiled.OpenMap(mapSource)
	if err != nil {
		return fmt.Errorf("error while opening Tiled map: %w", err)
	}

	err = m.Load()
	if err != nil {
		return fmt.Errorf("error while loading map: %w", err)
	}

	g.MapInfo = &MapInfo{
		ID:          m.ID,
		DisplayName: m.DisplayName,
	}
	g.MapInfo.Map = m
	g.MapInfo.NPCManager.mapRef = g.MapInfo.Map

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
			g.MapInfo.Lights = append(g.MapInfo.Lights, &l)
			fmt.Println("light")
			fmt.Println("x:", l.X, "y:", l.Y)
			fmt.Println("radius:", l.MinRadius, l.MaxRadius)
		}
	}

	// find all objects in the map
	for _, layer := range m.Layers {
		if layer.Type == tiled.LAYER_TYPE_OBJECT {
			for _, obj := range layer.Objects {
				g.MapInfo.AddObjectToMap(obj)
			}
		}
	}

	// start up background jobs loop
	if g.MapInfo.NPCManager.backgroundJobsRunning {
		panic("backgroundJobsRunning flag is already true while initializing map")
	}
	if op.RunNPCManager {
		g.MapInfo.RunBackgroundJobs = true
		g.MapInfo.NPCManager.startBackgroundNPCManager()
	}

	g.MapInfo.Loaded = true

	g.EventBus.Publish(pubsub.Event{
		Type: pubsub.Event_VisitMap,
		Data: map[string]any{
			"MapID":          g.MapInfo.ID,
			"MapDisplayName": g.MapInfo.DisplayName,
		},
	})

	return nil
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
func (mi *MapInfo) AddPlayerToMap(p *player.Player, x, y float64) {
	if !mi.Loaded {
		panic("map not loaded yet. use SetupMap before using this.")
	}
	if p == nil {
		panic("player is nil when adding to map")
	}
	r := model.Rect{
		X: x,
		Y: y,
		W: config.TileSize,
		H: config.TileSize,
	}
	if mi.Collides(r, "", true) {
		// this also handles placement outside of map bounds
		panic("player added to map on colliding tile")
	}
	mi.PlayerRef = p
	p.Entity.World = mi
	p.Entity.SetPositionPx(x, y)
}

// the official way to add an NPC to a map
func (mi *MapInfo) AddNPCToMap(n *npc.NPC, startPos model.Coords) {
	if !mi.Loaded {
		panic("map not loaded yet. use SetupMap before using this.")
	}
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

func (mi *MapInfo) AddObjectToMap(obj tiled.Object) {
	o := object.LoadObject(obj)
	o.WorldContext = mi
	mi.Objects = append(mi.Objects, o)
}

// detects if the given rect collides in the map.
// rectBased param determines if collisions check for collision rects (e.g. for buildings with nuanced collision rects)
// or if it just uses tile-based collisions (if a tile contains a collision rect, the entire tile is marked as a collision).
// generally, the player should use rect-based, and NPCs should use tile-based (since NPCs usually can't do partial tile/px based movement)
func (mi MapInfo) Collides(r model.Rect, excludeEntId string, rectBased bool) bool {
	tl := model.ConvertPxToTilePos(int(r.X), int(r.Y))
	tr := model.ConvertPxToTilePos(int(r.X+r.W), int(r.Y))
	bl := model.ConvertPxToTilePos(int(r.X), int(r.Y+r.H))
	br := model.ConvertPxToTilePos(int(r.X+r.W), int(r.Y+r.H))
	// check if any part of the target is outside the map
	maxTileX := len(mi.Map.CostMap[0]) - 1
	maxTileY := len(mi.Map.CostMap) - 1
	if !tl.WithinBounds(0, maxTileX, 0, maxTileY) {
		return true
	}
	if !tr.WithinBounds(0, maxTileX, 0, maxTileY) {
		return true
	}
	if !bl.WithinBounds(0, maxTileX, 0, maxTileY) {
		return true
	}
	if !br.WithinBounds(0, maxTileX, 0, maxTileY) {
		return true
	}

	if rectBased {
		r1 := mi.Map.CollisionRects[tl.Y][tl.X]
		if r1.IsCollision {
			if r1.OffsetRect(float64(tl.X*config.TileSize), float64(tl.Y*config.TileSize)).Intersects(r) {
				return true
			}
		}
		r2 := mi.Map.CollisionRects[tr.Y][tr.X]
		if r2.IsCollision {
			if r2.OffsetRect(float64(tr.X*config.TileSize), float64(tr.Y*config.TileSize)).Intersects(r) {
				return true
			}
		}
		r3 := mi.Map.CollisionRects[bl.Y][bl.X]
		if r3.IsCollision {
			if r3.OffsetRect(float64(bl.X*config.TileSize), float64(bl.Y*config.TileSize)).Intersects(r) {
				return true
			}
		}
		r4 := mi.Map.CollisionRects[br.Y][br.X]
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

	// TODO should I update the TilePos to be this position?
	if mi.PlayerRef != nil {
		playerRect := mi.PlayerRef.Entity.CollisionRect()
		playerX := playerRect.X + (playerRect.W / 2)
		playerY := playerRect.Y + (playerRect.H / 2)
		playerPos := model.ConvertPxToTilePos(int(playerX), int(playerY))
		costMap[playerPos.Y][playerPos.X] += 10
	}

	return costMap
}

func (mi *MapInfo) GetLights() []*lights.Light {
	return mi.Lights
}

func (mi *MapInfo) GetSpawnPosition(index int) (x, y float64, found bool) {
	for _, obj := range mi.Objects {
		if obj.Type == object.TYPE_SPAWN_POINT {
			if obj.SpawnPoint.SpawnIndex == index {
				return obj.X, obj.Y, true
			}
		}
	}
	return -1, -1, false
}

func (g *Game) PlacePlayerAtSpawnPoint(p *player.Player, spawnIndex int) error {
	if g.MapInfo == nil {
		panic("map info is nil")
	}
	x, y, found := g.MapInfo.GetSpawnPosition(spawnIndex)
	if !found {
		return fmt.Errorf("given spawn point index not found in map: %v", spawnIndex)
	}
	g.MapInfo.AddPlayerToMap(p, x, y)
	g.Camera.SetCameraPosition(x, y)
	return nil
}

func (mi MapInfo) GetPlayerRect() model.Rect {
	if mi.PlayerRef == nil {
		panic("player ref is nil")
	}
	return mi.PlayerRef.Entity.CollisionRect()
}
