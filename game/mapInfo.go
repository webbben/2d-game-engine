package game

import (
	"fmt"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/npc"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/internal/pubsub"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/object"
)

// MapInfo contains information about the current room the player is in
type MapInfo struct {
	gameRef *Game

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

	Lights       []*lights.Light  // permanent lights that are not controlled by an object
	LightObjects []*object.Object // lights controlled by an object

	NPCManager
}

type sortedRenderable interface {
	Y() float64
	Draw(screen *ebiten.Image, offsetX, offsetY float64)
}

type OpenMapOptions struct {
	// set to true if this map should run a NPC manager background process.
	// this is not mandatory for using NPCs, just helps improve their behavior, especially when there are a lot of them in a map.
	RunNPCManager    bool
	RegenerateImages bool // set to true if tile images should be regenerated
}

// EnterMap sets up a map and puts the player in it at the given position. meant for use once player already exists in game state
func (g *Game) EnterMap(mapID string, op *OpenMapOptions, playerSpawnIndex int) error {
	if g.MapInfo != nil {
		g.MapInfo.CloseMap()
	}

	err := g.SetupMap(mapID, op)
	if err != nil {
		return err
	}

	return g.PlacePlayerAtSpawnPoint(g.Player, playerSpawnIndex)
}

// SetupMap prepares the MapInfo for in-game play
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

	err = m.Load(op.RegenerateImages)
	if err != nil {
		return fmt.Errorf("error while loading map: %w", err)
	}

	g.MapInfo = &MapInfo{
		ID:          m.ID,
		DisplayName: m.DisplayName,
	}
	g.MapInfo.Map = m
	g.MapInfo.mapRef = g.MapInfo.Map
	g.MapInfo.gameRef = g

	// find all lights embedded in tiles
	for _, tileset := range m.Tilesets {
		for _, tile := range tileset.Tiles {
			// determine if this is a light tile
			tileType := tiled.GetTileType(tile)
			if tileType == "LIGHT" {
				lightProps := tiled.GetLightProps(tile.Properties)
				gid := tile.ID + tileset.FirstGID

				lightPositions := m.GetAllTilePositions(gid)
				for _, pos := range lightPositions {
					// center on the tile so the light doesn't show in the tile's top-left corner
					x := (pos.X * config.TileSize) + (config.TileSize / 2)
					y := (pos.Y * config.TileSize) + (config.TileSize / 2)
					l := lights.NewLight(x, y, lightProps, nil)
					g.MapInfo.Lights = append(g.MapInfo.Lights, &l)
					fmt.Printf("light found at x: %v y: %v\n", x, y)
				}
			}
		}
	}

	// find all objects in the map
	entObjs := []tiled.Object{}
	for _, layer := range m.Layers {
		if layer.Type == tiled.LayerTypeObject {
			for _, obj := range layer.Objects {
				// look for objects that are actually static entities
				// Note: if the entity type is stored deeper than the object itself, I don't think it'll be found here.
				objType, found := tiled.GetStringProperty("TYPE", obj.Properties)
				if found && objType == object.TypeEntity {
					entObjs = append(entObjs, obj)
					continue
				}
				g.MapInfo.AddObjectToMap(obj, m)
			}
		}
	}

	// start up background jobs loop
	if g.MapInfo.backgroundJobsRunning {
		panic("backgroundJobsRunning flag is already true while initializing map")
	}
	if op.RunNPCManager {
		g.MapInfo.RunBackgroundJobs = true
		g.MapInfo.startBackgroundNPCManager()
	}

	g.MapInfo.Loaded = true

	// add static ents after Loaded = true, since that is checked in AddNPCToMap
	// StaticEntities: These are basically one-off characters that appear in a specific room for a specific reason.
	// They aren't actually "existing in the overworld" as other characters will. So, we instantiate them with their set character def,
	// and show them in the world.
	// TODO: things we will need to decide/handle:
	// - do static entities have their state saved long term? or do we need to "clean it up" after its done?
	for _, obj := range entObjs {
		charDefID, found := tiled.GetStringProperty("CHARACTER_DEF_ID", obj.Properties)
		if !found {
			logz.Panicln("NewStaticEntity", "no character def ID property found on object")
		}

		charStateID := entity.CreateNewCharacterState(defs.CharacterDefID(charDefID), entity.NewCharacterStateParams{}, g.DefinitionManager)
		n := npc.NewNPC(npc.NPCParams{
			CharStateID: charStateID,
		}, g.DefinitionManager, g.AudioManager)

		_ = n.SetTask(&npc.IdleTask{}, true)
		r := model.NewRect(obj.X, obj.Y, obj.Width, obj.Height)
		tilePos := model.GetTilePosOfRectCenter(r)
		logz.Println("static NPC pos", tilePos, "original obj pos:", "x:", obj.X, "y:", obj.Y)
		g.MapInfo.AddNPCToMap(&n, tilePos)
	}

	g.EventBus.Publish(defs.Event{
		Type: pubsub.EventVisitMap,
		Data: map[string]any{
			"MapID":          g.MapInfo.ID,
			"MapDisplayName": g.MapInfo.DisplayName,
		},
	})

	return nil
}

func (mi *MapInfo) CloseMap() {
	mi.RunBackgroundJobs = false
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
	npcJams := mi.findNPCJams()
	if len(npcJams) == 0 {
		// no jams found
		return
	}

	// resolve each jam, one at a time
	for _, npcJam := range npcJams {
		mi.resolveJam(npcJam)
	}
}

// AddPlayerToMap is the official way to add the player to a map
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
	if res := mi.Collides(r, ""); res.Collides() {
		// this also handles placement outside of map bounds
		panic("player added to map on colliding tile")
	}
	mi.PlayerRef = p
	p.Entity.World = mi
	p.Entity.SetPositionPx(x, y)

	p.World = mi
}

// AddNPCToMap is the official way to add an NPC to a map.
// Note: uses Tile coordinates - not absolute coordinates!
func (mi *MapInfo) AddNPCToMap(n *npc.NPC, startPos model.Coords) {
	if !mi.Loaded {
		panic("map not loaded yet. use SetupMap before using this.")
	}

	numCols, numRows := mi.MapDimensions()
	if startPos.X > numCols || startPos.Y > numRows {
		panic("given position appears to be off the map. are you using absolute coordinates on accident? this expects tile based coordinates.")
	}

	w := n.Entity.CollisionRect().W
	h := n.Entity.CollisionRect().H
	r := model.Rect{
		X: float64(startPos.X * config.TileSize),
		Y: float64(startPos.Y * config.TileSize),
		W: w,
		H: h,
	}
	if res := mi.Collides(r, ""); res.Collides() {
		panic("npc added to map on colliding tile")
	}
	n.Entity.World = mi
	n.World = mi // NPC has its own world context it needs, that isn't relevant to entity
	n.Entity.SetPosition(startPos)
	n.Priority = mi.getNextNPCPriority()
	mi.NPCs = append(mi.NPCs, n)
}

func (mi *MapInfo) AddObjectToMap(obj tiled.Object, m tiled.Map) {
	o := object.LoadObject(obj, m, mi.gameRef.AudioManager)
	o.World = mi
	mi.Objects = append(mi.Objects, o)
	if o.Light.On {
		mi.LightObjects = append(mi.LightObjects, o)
	}
}

// Collides detects if the given rect collides in the map.
// rectBased param determines if collisions check for collision rects (e.g. for buildings with nuanced collision rects)
// or if it just uses tile-based collisions (if a tile contains a collision rect, the entire tile is marked as a collision).
// generally, the player should use rect-based, and NPCs should use tile-based (since NPCs usually can't do partial tile/px based movement)
func (mi MapInfo) Collides(r model.Rect, excludeEntID string) model.CollisionResult {
	tl := model.ConvertPxToTilePos(int(r.X), int(r.Y))
	tr := model.ConvertPxToTilePos(int(r.X+r.W), int(r.Y))
	bl := model.ConvertPxToTilePos(int(r.X), int(r.Y+r.H))
	br := model.ConvertPxToTilePos(int(r.X+r.W), int(r.Y+r.H))
	// check if any part of the target is outside the map
	maxTileX := len(mi.Map.CostMap[0]) - 1
	maxTileY := len(mi.Map.CostMap) - 1

	cr := model.CollisionResult{}

	// first, check for corners of the rect that are off the map
	if !tl.WithinBounds(0, maxTileX, 0, maxTileY) {
		cr.TopLeft.Intersects = true
	}
	if !tr.WithinBounds(0, maxTileX, 0, maxTileY) {
		cr.TopRight.Intersects = true
	}
	if !bl.WithinBounds(0, maxTileX, 0, maxTileY) {
		cr.BottomLeft.Intersects = true
	}
	if !br.WithinBounds(0, maxTileX, 0, maxTileY) {
		cr.BottomRight.Intersects = true
	}

	// next, check for regular collisions on the map
	if !cr.TopLeft.Intersects {
		// rects in CollisionRects don't have an X or Y value set, so we "offset" them to put them in their actual correct place
		// then, we check if that collision rect intersects with r.
		r1 := mi.Map.CollisionRects[tl.Y][tl.X]
		if r1.IsCollision {
			cr.TopLeft = r1.OffsetRect(float64(tl.X*config.TileSize), float64(tl.Y*config.TileSize)).IntersectionArea(r)
		}
	}
	if !cr.TopRight.Intersects {
		r2 := mi.Map.CollisionRects[tr.Y][tr.X]
		if r2.IsCollision {
			cr.TopRight = r2.OffsetRect(float64(tr.X*config.TileSize), float64(tr.Y*config.TileSize)).IntersectionArea(r)
		}
	}
	if !cr.BottomLeft.Intersects {
		r3 := mi.Map.CollisionRects[bl.Y][bl.X]
		if r3.IsCollision {
			cr.BottomLeft = r3.OffsetRect(float64(bl.X*config.TileSize), float64(bl.Y*config.TileSize)).IntersectionArea(r)
		}
	}
	if !cr.BottomRight.Intersects {
		r4 := mi.Map.CollisionRects[br.Y][br.X]
		if r4.IsCollision {
			cr.BottomRight = r4.OffsetRect(float64(br.X*config.TileSize), float64(br.Y*config.TileSize)).IntersectionArea(r)
		}
	}

	// if any static collisions are found, report those
	if cr.Collides() {
		cr.Assert() // can probably remove these asserts
		return cr
	}

	// if no static collisions are found, find any entity or object collisions
	// check the corner points of r and see if they are inside any of the entity/object rects.
	// if a corner point is inside a rect, then that corresponding corner gets the collision.
	// if no corner point is inside the rect but there is still a collision (e.g. the two rects aren't the same size) then the collision is set to Other.
	if mi.PlayerRef != nil {
		if string(mi.PlayerRef.Entity.ID()) != excludeEntID {
			newCr := checkCornerCollision(r, mi.PlayerRef.Entity.CollisionRect())
			if newCr.Collides() {
				newCr.Assert()
			}
			cr.MergeOtherCollisionResult(newCr)
		}
	}
	for _, n := range mi.NPCs {
		if string(n.Entity.ID()) == excludeEntID {
			continue
		}
		newCr := checkCornerCollision(r, n.Entity.CollisionRect())
		if newCr.Collides() {
			newCr.Assert()
		}
		cr.MergeOtherCollisionResult(newCr)
	}

	// check for collidable objects (gates, etc)
	for _, obj := range mi.Objects {
		if !obj.IsCollidable() {
			continue
		}
		newCr := checkCornerCollision(r, obj.GetRect())
		if newCr.Collides() {
			newCr.Assert()
		}
		cr.MergeOtherCollisionResult(newCr)
	}

	cr.Assert()
	return cr
}

func checkCornerCollision(r, targetRect model.Rect) model.CollisionResult {
	cr := model.CollisionResult{}

	res := r.IntersectionArea(targetRect)
	if res.Intersects {
		corners := 0
		if res.FromTL {
			cr.TopLeft = res
			corners++
		}
		if res.FromTR {
			cr.TopRight = res
			corners++
		}
		if res.FromBL {
			cr.BottomLeft = res
			corners++
		}
		if res.FromBR {
			cr.BottomRight = res
			corners++
		}
		if corners > 2 {
			panic("how can a rect intersect with another rect on more than 2 corners? is it inside the other rect?")
		}
	}

	return cr
}

// FindPath returns a path to the goal, or if it cannot be reached, a path to the closest reachable position.
// The boolean indicates if the goal was successfully reached.
func (mi MapInfo) FindPath(start, goal model.Coords) ([]model.Coords, bool) {
	return path_finding.FindPath(start, goal, mi.CostMap())
}

// MapDimensions gives the TILE dimensions of a map (columns = width, rows = height).
// This is not a pixels/absolute dimensions function.
func (mi MapInfo) MapDimensions() (width int, height int) {
	return mi.Map.Width, mi.Map.Height
}

// CostMap gets a cost map that includes entity positions
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
		if obj.Type == object.TypeSpawnPoint {
			if obj.SpawnPoint.SpawnIndex == index {
				x, y := obj.Pos()
				return x, y, true
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

func (mi *MapInfo) StartTradeSession(shopkeeperID defs.ShopID) {
	mi.gameRef.SetupTradeSession(shopkeeperID)
}

func (mi *MapInfo) StartDialog(dialogProfileID defs.DialogProfileID, npcID string) {
	mi.gameRef.StartDialogSession(dialogProfileID, npcID)
}

func (mi *MapInfo) GetNearbyNPCs(posX, posY, radius float64) []*npc.NPC {
	npcs := []*npc.NPC{}

	for _, n := range mi.NPCs {
		dist := general_util.EuclideanDist(
			n.Entity.X,
			n.Entity.Y,
			posX,
			posY,
		)
		if dist <= radius {
			npcs = append(npcs, n)
		}
	}

	return npcs
}

func (mi *MapInfo) GetPlayer() *player.Player {
	return mi.PlayerRef
}

func (mi MapInfo) GetDistToPlayer(x, y float64) float64 {
	return general_util.EuclideanDist(x, y, mi.PlayerRef.Entity.X, mi.PlayerRef.Entity.Y)
}

func (mi *MapInfo) GetGroundMaterial(tileX, tileY int) string {
	if tileY < 0 || tileY >= len(mi.Map.GroundMaterial) {
		panic("tileY outside of bounds of ground material matrix")
	}
	if tileX < 0 || tileX >= len(mi.Map.GroundMaterial[0]) {
		panic("tileX outside of bounds of ground material matrix")
	}

	return mi.Map.GroundMaterial[tileY][tileX]
}

func (mi *MapInfo) AttackArea(attackInfo entity.AttackInfo) {
	// find all entities in the area of the rect

	logz.Println("Attack Area", "target rect:", attackInfo.TargetRect, "attack info:", attackInfo)

	if len(mi.NPCs) == 0 {
		fmt.Println("no NPCs?")
	}

	for _, n := range mi.NPCs {
		logz.Println("Attack Area", "entID:", n.Entity.ID())
		if slices.Contains(attackInfo.ExcludeEntIds, string(n.Entity.ID())) {
			continue
		}
		fmt.Println("npc rect:", n.Entity.CollisionRect())
		if attackInfo.TargetRect.Intersects(n.Entity.CollisionRect()) {
			n.Entity.ReceiveAttack(attackInfo)
		}
	}
	if mi.PlayerRef != nil && !slices.Contains(attackInfo.ExcludeEntIds, string(mi.PlayerRef.Entity.ID())) {
		if attackInfo.TargetRect.Intersects(mi.PlayerRef.Entity.CollisionRect()) {
			mi.PlayerRef.Entity.ReceiveAttack(attackInfo)
		}
	}
}

// ActivateArea attempts to activate an object or npc in an area. if an activation occurs, true is returned.
func (mi *MapInfo) ActivateArea(r model.Rect) bool {
	// check for activated objects
	// try to get the object that is the "best match" (i.e. closest to the center of the activated area)
	var closestObject *object.Object = nil
	closestObjectDist := float64(config.TileSize * 1000)
	for _, obj := range mi.Objects {
		if !obj.IsActivatable() {
			continue
		}
		if r.Intersects(obj.GetRect()) {
			dist := general_util.EuclideanDistCenter(r, obj.GetRect())
			if closestObject == nil || dist < closestObjectDist {
				closestObject = obj
				closestObjectDist = dist
			}
		}
	}
	if closestObject != nil {
		result := closestObject.Activate()
		logz.Println("Activate Area", closestObject.Type, "Activating...")

		if result.UpdateOccurred {
			logz.Println("Activate Area", closestObject.Type, "Activation occurred")
			if result.ChangeMapID != "" {
				mi.gameRef.handleMapDoor(result)
			}
		}

		return true
	}

	// check for activated entities
	// if multiple entites are present, activate the closest one to the center of the activation area
	var closestNPC *npc.NPC = nil
	closestNPCDist := float64(config.TileSize * 1000)
	for _, n := range mi.NPCs {
		if r.Intersects(n.Entity.CollisionRect()) {
			dist := general_util.EuclideanDistCenter(r, n.Entity.CollisionRect())
			if closestNPC == nil || dist < closestNPCDist {
				closestNPC = n
				closestNPCDist = dist
			}
		}
	}
	if closestNPC != nil {
		closestNPC.Activate()
		return true
	}

	return false
}

// HandleMouseClick handles a player's click in the game world; for non-ui clicks, such as clicking objects or entities in a map.
// if a click event occurs, true will be returned.
func (mi *MapInfo) HandleMouseClick(mouseX, mouseY int) bool {
	distThreshold := float64(config.TileSize * 2)
	// check for object clicks
	for _, obj := range mi.Objects {
		if !obj.IsActivatable() {
			continue
		}
		if obj.GetDrawRect().Within(mouseX, mouseY) {
			if general_util.EuclideanDistCenter(mi.GetPlayerRect(), obj.GetRect()) <= distThreshold {
				fmt.Println("object clicked")
				result := obj.Activate()
				if result.UpdateOccurred {
					if result.ChangeMapID != "" {
						mi.gameRef.handleMapDoor(result)
					}
				}
				return true
			}
		}
	}

	// check for NPC clicks
	for _, n := range mi.NPCs {
		if n.Entity.GetDrawRect().Within(mouseX, mouseY) {
			if general_util.EuclideanDistCenter(mi.GetPlayerRect(), n.Entity.CollisionRect()) <= distThreshold {
				n.Activate()
				return true
			}
		}
	}

	return false
}
