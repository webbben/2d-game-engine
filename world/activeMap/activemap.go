// Package activemap represents the map that is currently active in the game
package activemap

import (
	"fmt"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/entity"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/internal/camera"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/utils"
	"github.com/webbben/2d-game-engine/world/npc"
)

type GameContext interface {
	StartDialogSession(dialogProfileID defs.DialogProfileID, npcID string)
	StartTradeSession(shopkeeperID defs.ShopID)
}

type WorldContext interface {
	// TODO: is this used here?
	GenerateCharacter(chargen defs.CharacterGenerator, initialMap defs.MapID, homeMap defs.MapID, homeMapBedID int) state.CharacterStateID
	HandleMapDoor(result object.ObjectUpdateResult)
}

// ActiveMap contains information about the current room the player is in
type ActiveMap struct {
	debugData debugData
	MapID     defs.MapID

	gameCtx  GameContext
	worldCtx WorldContext

	dataman  *datamanager.DataManager
	audioman *audio.AudioManager
	eventBus *pubsub.EventBus

	DisplayName string // the name of the map shown to the player
	Loaded      bool   // flag indicating if this map has been loaded
	ReadyToPlay bool   // flag indicating if all loading steps are done, and this map is ready to show in the game
	MapIsActive bool   // flag indicating if the map is actively being used (e.g. for rendering, updates, etc)
	Map         *tiled.Map
	ImageMap    map[string]*ebiten.Image // the map of images (tiles) used in rendering the current room
	PlayerRef   *player.Player
	Objects     []*object.Object

	Camera     camera.Camera
	worldScene *ebiten.Image

	sortedRenderables []sortedRenderable

	Lights       []*lights.Light  // permanent lights that are not controlled by an object
	LightObjects []*object.Object // lights controlled by an object

	daylightFader lights.LightFader

	NPCManager
}

func NewActiveMap(
	dataman *datamanager.DataManager,
	audioman *audio.AudioManager,
	eventbus *pubsub.EventBus,
	gameCtx GameContext,
	worldCtx WorldContext,
	mapID defs.MapID,
	regenImages bool,
) *ActiveMap {
	mapDef := dataman.GetMapDef(mapID)
	// load and setup the map
	tiledMap := tiled.LoadMap(mapID, regenImages)

	m := &ActiveMap{
		MapID:       mapID,
		DisplayName: mapDef.DisplayName,
		dataman:     dataman,
		audioman:    audioman,
		eventBus:    eventbus,
		gameCtx:     gameCtx,
		worldCtx:    worldCtx,
		Map:         tiledMap,
		NPCManager: NPCManager{
			mapRef: tiledMap,
		},
	}

	m.daylightFader = lights.NewLightFader(lights.LightColor{1, 1, 1}, 0, 0.1, config.HourSpeed/20)
	m.worldScene = ebiten.NewImage(display.SCREEN_WIDTH, display.SCREEN_HEIGHT)

	// find all lights embedded in tiles
	for _, tileset := range m.Map.Tilesets {
		for _, tile := range tileset.Tiles {
			// determine if this is a light tile
			tileType := tiled.GetTileType(tile)
			if tileType == "LIGHT" {
				lightProps := tiled.GetLightProps(tile.Properties)
				gid := tile.ID + tileset.FirstGID

				lightPositions := m.Map.GetAllTilePositions(gid)
				for _, pos := range lightPositions {
					// center on the tile so the light doesn't show in the tile's top-left corner
					x := (pos.X * config.TileSize) + (config.TileSize / 2)
					y := (pos.Y * config.TileSize) + (config.TileSize / 2)
					l := lights.NewLight(x, y, lightProps, nil)
					m.Lights = append(m.Lights, &l)
					fmt.Printf("light found at x: %v y: %v\n", x, y)
				}
			}
		}
	}

	// find all objects in the map
	for _, layer := range m.Map.Layers {
		m.addAllObjectsToMap(layer)
	}

	// start up background jobs loop
	if m.backgroundJobsRunning {
		panic("backgroundJobsRunning flag is already true while initializing map")
	}
	m.RunBackgroundJobs = true
	m.startBackgroundNPCManager()

	m.Loaded = true

	m.eventBus.Publish(defs.Event{
		Type: pubsub.EventVisitMap,
		Data: map[string]any{
			"MapID":          mapID,
			"MapDisplayName": m.DisplayName,
		},
	})

	return m
}

func (m *ActiveMap) OnHourChange(hour int, skipFade bool) {
	newDaylight, darknessFactor := lights.CalculateDaylight(hour)
	if skipFade {
		m.daylightFader.SetCurrentColor(newDaylight)
		m.daylightFader.TargetColor = newDaylight
		m.daylightFader.SetCurrentDarknessFactor(darknessFactor)
		m.daylightFader.TargetDarknessFactor = darknessFactor
	} else {
		m.daylightFader.TargetColor = newDaylight
		m.daylightFader.TargetDarknessFactor = darknessFactor
	}
}

func (m *ActiveMap) addAllObjectsToMap(layer tiled.Layer) {
	allObjs := tiled.GetAllObjectsFromLayer(layer)
	for _, obj := range allObjs {
		m.AddObjectToMap(obj, *m.mapRef)
	}
}

type NPCManager struct {
	NPCs []*npc.NPC // the NPC entities in the map

	mapRef       *tiled.Map // map info so we can get map size, tile adjacency, etc
	nextPriority int        // the next priority value to assign to an NPC

	// if true, the background jobs goroutine will run.
	// if false, the background jobs goroutine will stop.
	RunBackgroundJobs     bool
	backgroundJobsRunning bool // flag that indicates if background jobs loop already running.
}

type sortedRenderable interface {
	Y() float64
	Draw(screen *ebiten.Image, offsetX, offsetY float64)
}

// AddPlayerToMap is the official way to add the player to a map
func (mi *ActiveMap) AddPlayerToMap(p *player.Player, x, y float64) {
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
func (mi *ActiveMap) AddNPCToMap(n *npc.NPC, startPos model.Coords) {
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
		logz.Panicln("AddNPCToMap", "NPC added to map on colliding tile. r:", r, "mapID:", mi.MapID, "npcID:", n.ID())
		panic("npc added to map on colliding tile")
	}
	n.Entity.World = mi
	n.World = mi // NPC has its own world context it needs, that isn't relevant to entity
	n.Entity.SetPosition(startPos)
	n.Priority = mi.getNextNPCPriority()
	mi.NPCs = append(mi.NPCs, n)
}

func (mi *ActiveMap) AddObjectToMap(obj tiled.Object, m tiled.Map) {
	o := object.LoadObject(obj, m, mi.audioman, mi.dataman, mi.MapID, mi)
	mi.Objects = append(mi.Objects, o)
	if o.Light.On {
		mi.LightObjects = append(mi.LightObjects, o)
	}
}

// Collides detects if the given rect collides in the map.
// rectBased param determines if collisions check for collision rects (e.g. for buildings with nuanced collision rects)
// or if it just uses tile-based collisions (if a tile contains a collision rect, the entire tile is marked as a collision).
// generally, the player should use rect-based, and NPCs should use tile-based (since NPCs usually can't do partial tile/px based movement)
func (mi ActiveMap) Collides(r model.Rect, excludeEntID string) model.CollisionResult {
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
func (mi ActiveMap) FindPath(start, goal model.Coords) ([]model.Coords, bool) {
	return path_finding.FindPath(start, goal, mi.CostMap())
}

// MapDimensions gives the TILE dimensions of a map (columns = width, rows = height).
// This is not a pixels/absolute dimensions function.
func (mi ActiveMap) MapDimensions() (width int, height int) {
	return mi.Map.Width, mi.Map.Height
}

// CostMap gets a cost map for a map.
//
// Currently includes:
//
// - tiledMap costmap (mainly just collision rects embedded in tiles, I believe)
//
// - NPC positions
//
// Not included:
//
// - Object collisions; These are shown in the debug "showCollisions", but not actually in the cost map here.
func (mi ActiveMap) CostMap() [][]int {
	if mi.Map.CostMap == nil {
		panic("tried to get ActiveMap cost map before Map costmap was created")
	}
	// make deep copy so that original cost map isn't altered
	costMap := make([][]int, len(mi.Map.CostMap))
	for i := range mi.Map.CostMap {
		costMap[i] = append([]int{}, mi.Map.CostMap[i]...)
	}

	for _, n := range mi.NPCs {
		tilePos := n.Entity.TilePos()
		costMap[tilePos.Y][tilePos.X] += 10
		// if the entity is currently moving, mark its destination tile as a collision too
		targetTile := n.Entity.TargetTilePos()
		if !targetTile.Equals(tilePos) {
			costMap[targetTile.Y][targetTile.X] += 10
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

func (mi *ActiveMap) GetLights() []*lights.Light {
	return mi.Lights
}

func (mi *ActiveMap) GetSpawnPosition(index int) (x, y float64, found bool) {
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

func (mi ActiveMap) GetPlayerRect() model.Rect {
	if mi.PlayerRef == nil {
		panic("player ref is nil")
	}
	return mi.PlayerRef.Entity.CollisionRect()
}

func (mi *ActiveMap) StartTradeSession(shopkeeperID defs.ShopID) {
	mi.gameCtx.StartTradeSession(shopkeeperID)
}

func (mi *ActiveMap) StartDialog(dialogProfileID defs.DialogProfileID, npcID string) {
	mi.gameCtx.StartDialogSession(dialogProfileID, npcID)
}

func (mi *ActiveMap) GetNearbyNPCs(posX, posY, radius float64) []*npc.NPC {
	npcs := []*npc.NPC{}

	for _, n := range mi.NPCs {
		dist := utils.EuclideanDist(
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

func (mi *ActiveMap) GetPlayer() *player.Player {
	return mi.PlayerRef
}

func (mi ActiveMap) GetDistToPlayer(x, y float64) float64 {
	return utils.EuclideanDist(x, y, mi.PlayerRef.Entity.X, mi.PlayerRef.Entity.Y)
}

func (mi *ActiveMap) GetGroundMaterial(tileX, tileY int) string {
	if tileY < 0 || tileY >= len(mi.Map.GroundMaterial) {
		panic("tileY outside of bounds of ground material matrix")
	}
	if tileX < 0 || tileX >= len(mi.Map.GroundMaterial[0]) {
		panic("tileX outside of bounds of ground material matrix")
	}

	return mi.Map.GroundMaterial[tileY][tileX]
}

func (mi *ActiveMap) AttackArea(attackInfo entity.AttackInfo) {
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
// NOTE: for now, this is only used by the player. if this becomes a general purpose "activate an area" function,
// then we need to pass info in about who is activating - so that we know which locks can be opened, for example.
func (mi *ActiveMap) ActivateArea(r model.Rect, originX, originY float64) bool {
	// check for activated objects
	// try to get the object that is the "best match" (i.e. closest to the center of the activated area)
	var closestObject *object.Object = nil
	closestObjectDist := float64(config.TileSize * 1000)
	for _, obj := range mi.Objects {
		if !obj.IsActivatable() {
			continue
		}
		if r.Intersects(obj.GetRect()) {
			dist := utils.EuclideanDistCenter(r, obj.GetRect())
			if closestObject == nil || dist < closestObjectDist {
				closestObject = obj
				closestObjectDist = dist
			}
		}
	}
	if closestObject != nil {
		activateParams := object.ObjectActivationParams{
			LockIDs: characterstate.GetLockIDs(*mi.PlayerRef.CharacterStateRef),
		}
		result := closestObject.Activate(originX, originY, activateParams)
		logz.Println("Activate Area", closestObject.Type, "Activating...")

		if result.UpdateOccurred {
			logz.Println("Activate Area", closestObject.Type, "Activation occurred")
			if result.ChangeMapID != "" {
				mi.worldCtx.HandleMapDoor(result)
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
			dist := utils.EuclideanDistCenter(r, n.Entity.CollisionRect())
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
func (mi *ActiveMap) HandleMouseClick(mouseX, mouseY int) bool {
	distThreshold := float64(config.TileSize * 2)
	// check for object clicks
	for _, obj := range mi.Objects {
		if !obj.IsActivatable() {
			continue
		}
		if obj.GetDrawRect().Within(mouseX, mouseY) {
			if utils.EuclideanDistCenter(mi.GetPlayerRect(), obj.GetRect()) <= distThreshold {
				fmt.Println("object clicked")
				x, y := mi.PlayerRef.Entity.X, mi.PlayerRef.Entity.Y
				activateParams := object.ObjectActivationParams{
					LockIDs: characterstate.GetLockIDs(*mi.PlayerRef.CharacterStateRef),
				}
				result := obj.Activate(x, y, activateParams)
				if result.UpdateOccurred {
					if result.ChangeMapID != "" {
						mi.worldCtx.HandleMapDoor(result)
					}
				}
				return true
			}
		}
	}

	// check for NPC clicks
	for _, n := range mi.NPCs {
		if n.Entity.GetDrawRect().Within(mouseX, mouseY) {
			if utils.EuclideanDistCenter(mi.GetPlayerRect(), n.Entity.CollisionRect()) <= distThreshold {
				n.Activate()
				return true
			}
		}
	}

	return false
}

// FindObjectsAtPosition finds all objects that intersect with a given tile position.
// This includes collidable and non-collidable objects, as long as they have a draw rect.
func (mi *ActiveMap) FindObjectsAtPosition(c model.Coords) []*object.Object {
	posRect := model.NewRect(float64(c.X)*config.TileSize, float64(c.Y)*config.TileSize, config.TileSize, config.TileSize)
	objs := []*object.Object{}
	for _, obj := range mi.Objects {
		if obj.GetRect().Intersects(posRect) {
			objs = append(objs, obj)
		}
	}
	return objs
}

// RectCollidesWithOthers is a general purpose function to see if a rect in a world map collides with anything.
// Can pass exclusion IDs so that the caller can ignore itself.
func (mi *ActiveMap) RectCollidesWithOthers(r model.Rect, excludeEntID string, excludeObjID int) bool {
	for _, n := range mi.NPCs {
		if string(n.Entity.ID()) == excludeEntID {
			continue
		}
		if n.Entity.CollisionRect().Intersects(r) {
			return true
		}
	}

	for _, obj := range mi.Objects {
		if obj.ID == excludeObjID {
			continue
		}
		if obj.GetRect().Intersects(r) {
			return true
		}
	}

	return false
}

func (m *ActiveMap) PlacePlayerAtSpawnPoint(p *player.Player, spawnIndex int) {
	x, y, found := m.GetSpawnPosition(spawnIndex)
	if !found {
		logz.Panicf("given spawn point index not found in map: %v", spawnIndex)
	}
	m.PlacePlayerAtPosition(p, x, y)
}

func (m *ActiveMap) PlacePlayerAtPosition(p *player.Player, x, y float64) {
	m.AddPlayerToMap(p, x, y)
	m.Camera.SetCameraPosition(x, y)
}
