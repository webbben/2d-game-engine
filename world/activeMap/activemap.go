// Package activemap represents the map that is currently active in the game
package activemap

import (
	"fmt"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/book"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/dialogv2"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/entity"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/internal/camera"
	"github.com/webbben/2d-game-engine/internal/debug"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/quest"
	"github.com/webbben/2d-game-engine/screen"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/utils"
	"github.com/webbben/2d-game-engine/world/npc"
	"github.com/webbben/2d-game-engine/worldgraph"
)

type WorldContext interface {
	// TODO: is this used here?
	GenerateCharacter(chargen defs.CharacterGenerator, initialMap defs.MapID, homeMap defs.MapID, homeMapBedID int) id.CharacterStateID
	HandleMapDoor(result object.ObjectUpdateResult)
	FindWorldPath(from, to defs.MapID) (pathToGoal worldgraph.WorldPath, found bool)
	ChangeMapOccupancy(charStateID id.CharacterStateID, fromMap, toMap defs.MapID, toSpawn int)
}

// ActiveMap contains information about the current room the player is in
type ActiveMap struct {
	debugData debugData
	MapID     defs.MapID

	InScenario bool // if set, this active map is part of a scenario. changes behavior about things like hourly task schedules.

	dialogSession *dialogv2.DialogSession
	bookSession   *book.BookSession

	gameCtx  defs.GameContext
	worldCtx WorldContext

	dataman   *datamanager.DataManager
	audioman  *audio.AudioManager
	eventBus  *pubsub.EventBus
	screenman *screen.ScreenManager
	questman  *quest.QuestManager
	om        *overlay.OverlayManager

	DisplayName string // the name of the map shown to the player
	Loaded      bool   // flag indicating if this map has been loaded
	ReadyToPlay bool   // flag indicating if all loading steps are done, and this map is ready to show in the game
	MapIsActive bool   // flag indicating if the map is actively being used (e.g. for rendering, updates, etc)
	Map         *tiled.Map
	ImageMap    map[string]*ebiten.Image // the map of images (tiles) used in rendering the current room
	PlayerRef   *player.Player
	Objects     []*object.Object

	hoveredObject *object.Object
	hoveredNPC    *npc.NPC

	Camera     camera.Camera
	worldScene *ebiten.Image

	sortedRenderables []sortedRenderable

	Lights       []*lights.Light  // permanent lights that are not controlled by an object
	LightObjects []*object.Object // lights controlled by an object

	daylightFader lights.LightFader

	NPCManager
}

func (m ActiveMap) GetHoverTarget() (*npc.NPC, *object.Object) {
	distThreshold := float64(config.PlayerHoverDistanceThreshold)

	if m.hoveredNPC != nil {
		if utils.EuclideanDistCenter(m.PlayerRef.Entity.CollisionRect(), m.hoveredNPC.Entity.CollisionRect()) <= distThreshold {
			return m.hoveredNPC, nil
		}
	}

	if m.hoveredObject != nil {
		if !m.hoveredObject.IsHoverable() {
			panic("non-hoverable object was selected")
		}
		if utils.EuclideanDistCenter(m.PlayerRef.Entity.CollisionRect(), m.hoveredObject.GetRect()) <= distThreshold {
			return nil, m.hoveredObject
		}
	}

	return nil, nil
}

// IsDialogActive tells you if a dialog is currently active
func (m ActiveMap) IsDialogActive() bool {
	return m.dialogSession != nil
}

func NewActiveMap(
	dataman *datamanager.DataManager,
	audioman *audio.AudioManager,
	eventbus *pubsub.EventBus,
	screenman *screen.ScreenManager,
	questman *quest.QuestManager,
	om *overlay.OverlayManager,
	gameCtx defs.GameContext,
	worldCtx WorldContext,
	mapID defs.MapID,
	regenImages bool,
) *ActiveMap {
	debug.StartTimer("NewActiveMap")

	logz.Println("ActiveMap", "Creating new active map:", mapID)

	mapInfo, mapDef, _ := dataman.GetAllMapData(mapID)

	// load and setup the map
	tiledMap := tiled.LoadMap(mapDef.ID, regenImages)

	m := &ActiveMap{
		MapID:       mapID,
		DisplayName: mapInfo.DisplayName,
		dataman:     dataman,
		audioman:    audioman,
		eventBus:    eventbus,
		screenman:   screenman,
		questman:    questman,
		om:          om,
		gameCtx:     gameCtx,
		worldCtx:    worldCtx,
		Map:         tiledMap,
		NPCManager: NPCManager{
			mapRef: tiledMap,
		},
	}

	m.Camera.SetMapLimits(tiledMap.Width, tiledMap.Height)

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

	debug.StopTimer("NewActiveMap")

	return m
}

// OnHourChange just handles adjusting the lighting based on the current hour
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
		if obj.Ellipse {
			// we only use elipses for planning things in maps, so just skip em
			logz.TODO("Ellipse Object", "ellipse object found in map (skipped it though). Should we delete this ellipse?")
			continue
		}
		m.AddObjectToMap(obj, *m.mapRef)
	}

	// now, go back through all objects and adjust sortedRenderables customY values, in case some items need to be made on top of others
	// this can happen if an item is on a tabletop. without an adjustment, the item might appear underneath since it's actually positioned with a lower Y value.
	checkObjs := []*object.Object{}
	checkObjs = append(checkObjs, m.Objects...)
	for len(checkObjs) > 0 {
		recheckObjs := []*object.Object{}
		for _, obj := range checkObjs {
			if obj.OnTopOfObjID != -1 {
				// find the object it should be on top of
				found := false
				for _, objBelow := range m.Objects {
					if objBelow.ID != obj.OnTopOfObjID {
						continue
					}
					found = true
					if objBelow.OnTopOfObjID == -1 || objBelow.CustomY != 0 {
						// we can calculate this objects customY if the one below either isn't on top of another object, or if it's customY has already been calculated.
						obj.CustomY = objBelow.Y() + 1
						if obj.CustomY == 0 {
							panic("I thought customY could never be 0?")
						}
					} else {
						// object below is also on top of something, but hasn't gotten its customY yet. skip for now.
						recheckObjs = append(recheckObjs, obj)
					}
					break
				}
				if !found {
					logz.Panicln("addAllObjectsToMap", "trying to find object that is below another object, but couldn't find it. looking for object ID:", obj.OnTopOfObjID)
				}
			}
		}
		checkObjs = recheckObjs
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
	X() float64
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
		W: config.TileSize - 1,
		H: config.TileSize - 1,
	}
	if res := mi.Collides(r); res.Collides() {
		// this also handles placement outside of map bounds
		logz.Panicln("AddPlayerToMap", "player added to map on colliding position:", r, res.Note)
	}
	mi.PlayerRef = p
	p.Entity.World = mi
	p.Entity.SetPositionPx(x, y)

	p.World = mi
}

func (mi ActiveMap) GetObjByID(objID int) *object.Object {
	for _, obj := range mi.Objects {
		if obj.ID == objID {
			return obj
		}
	}
	logz.Panicln("GetObjByID", "no object of the given ID found:", objID)
	return nil
}

// AddNPCToMap is the official way to add an NPC to a map.
// Note: uses Tile coordinates - not absolute coordinates!
func (mi *ActiveMap) AddNPCToMap(n *npc.NPC, startPos model.Coords) {
	if !mi.Loaded {
		panic("map not loaded yet. use SetupMap before using this.")
	}
	logz.Println("AddNPCToMap", "adding NPC to map:", n.ID(), "pos:", startPos)

	numCols, numRows := mi.MapDimensions()
	if startPos.X > numCols || startPos.Y > numRows {
		panic("given position appears to be off the map. are you using absolute coordinates on accident? this expects tile based coordinates.")
	}

	if mi.IsTileCollision(startPos) {
		logz.Panicln("AddNPCToMap", "NPC added to map on colliding tile.", "mapID:", mi.MapID, "npcID:", n.ID(), "startPos:", startPos)
	}

	// subscribe speech bubble reaction functions to events
	n.SetupSpeechBubbleReactions(mi.gameCtx)

	n.Entity.World = mi
	n.ActiveMapCtx = mi // NPC has its own world context it needs, that isn't relevant to entity
	n.Entity.SetPosition(startPos)
	n.Priority = mi.getNextNPCPriority() // TODO: not really sure if priority is still used, since there is no collision between entities
	mi.NPCs = append(mi.NPCs, n)
}

func (mi ActiveMap) IsTileCollision(coords model.Coords) bool {
	r := model.Rect{
		X: float64(coords.X * config.TileSize),
		Y: float64(coords.Y * config.TileSize),
		W: config.TileSize - 1,
		H: config.TileSize - 1,
	}
	res := mi.Collides(r)
	if res.Collides() {
		logz.Println("IsTileCollision", "collision:", res)
	}
	return res.Collides()
}

func (m ActiveMap) IsTileEntityCollision(c model.Coords, excludeEntID string) bool {
	r := model.Rect{
		X: float64(c.X * config.TileSize),
		Y: float64(c.Y * config.TileSize),
		W: config.TileSize,
		H: config.TileSize,
	}
	collides, _ := m.CollidesWithEntity(r, excludeEntID)
	return collides
}

func (mi *ActiveMap) AddObjectToMap(obj tiled.Object, m tiled.Map) {
	o := object.LoadObject(obj, m, mi.audioman, mi.dataman, mi.eventBus, mi.MapID, mi)

	// check if object is a door and has overrides; if so, we need to edit the door properties
	if o.Type == object.TypeDoor {
		mapState := mi.dataman.GetMapState(mi.MapID)
		doorOverride, exists := mapState.DoorOverrides[o.ID]
		if exists {
			o.SetDoorTarget(doorOverride.OverrideDestinationMap, doorOverride.OverrideDestinationSpawn)
		}
	}

	// validate here, so that door stuff has been resolved
	o.Validate()

	mi.Objects = append(mi.Objects, o)
	if o.Light.On {
		mi.LightObjects = append(mi.LightObjects, o)
	}
}

// CollidesWithEntity checks if the rect collides with an entity. This is meant for detecting if an entity
// collides with another entity, and should therefore move slower or not. For "hard" collisions, use Collides instead.
// TODO: at some point, when combat is more developed, we will want to make combatant entities not able to be walked through.
func (m ActiveMap) CollidesWithEntity(r model.Rect, excludeEntID string) (collides bool, dist float64) {
	if m.PlayerRef != nil {
		if string(m.PlayerRef.Entity.ID()) != excludeEntID && !m.PlayerRef.Entity.DisableCollisions {
			playerR := m.PlayerRef.Entity.CollisionRect()
			cr := checkCornerCollision(r, playerR)
			if cr.Collides() {
				dist := utils.EuclideanDistCenter(r, playerR)
				return true, dist
			}
		}
	}
	for _, n := range m.NPCs {
		if string(n.Entity.ID()) == excludeEntID {
			continue
		}
		if n.Entity.DisableCollisions {
			continue
		}
		cr := checkCornerCollision(r, n.Entity.CollisionRect())
		if cr.Collides() {
			dist := utils.EuclideanDistCenter(r, n.Entity.CollisionRect())
			return true, dist
		}
	}
	return false, -1
}

// Collides detects if the given rect collides in the map.
func (mi ActiveMap) Collides(r model.Rect) model.CollisionResult {
	// adding a full tilesize would make you spill over into the next tile's space.
	// this is because an individual pixel position represents an actual pixel.
	// so, the space within a tile at tile position 0,0 is:
	// x: [0, tilesize-1] y: [0, tilesize-1]
	tl := model.ConvertPxToTilePos(r.X, r.Y)
	tr := model.ConvertPxToTilePos(r.X+r.W, r.Y)
	bl := model.ConvertPxToTilePos(r.X, r.Y+r.H)
	br := model.ConvertPxToTilePos(r.X+r.W, r.Y+r.H)
	// check if any part of the target is outside the map
	maxTileX := len(mi.Map.CostMap[0]) - 1
	maxTileY := len(mi.Map.CostMap) - 1

	cr := model.CollisionResult{}

	// first, check for corners of the rect that are off the map
	if !tl.WithinBounds(0, maxTileX, 0, maxTileY) {
		cr.TopLeft.Intersects = true
		cr.Note = "off map"
	}
	if !tr.WithinBounds(0, maxTileX, 0, maxTileY) {
		cr.TopRight.Intersects = true
		cr.Note = "off map"
	}
	if !bl.WithinBounds(0, maxTileX, 0, maxTileY) {
		cr.BottomLeft.Intersects = true
		cr.Note = "off map"
	}
	if !br.WithinBounds(0, maxTileX, 0, maxTileY) {
		cr.BottomRight.Intersects = true
		cr.Note = "off map"
	}

	// next, check for regular collisions on the map
	if !cr.TopLeft.Intersects {
		// rects in CollisionRects don't have an X or Y value set, so we "offset" them to put them in their actual correct place
		// then, we check if that collision rect intersects with r.
		r1 := mi.Map.CollisionRects[tl.Y][tl.X]
		if r1.IsCollision {
			cr.TopLeft = r1.OffsetRect(float64(tl.X*config.TileSize), float64(tl.Y*config.TileSize)).IntersectionArea(r)
			cr.Note = fmt.Sprintf("map collision rect TL (X: %v, Y: %v)", tl.X, tl.Y)
		}
	}
	if !cr.TopRight.Intersects {
		r2 := mi.Map.CollisionRects[tr.Y][tr.X]
		if r2.IsCollision {
			cr.TopRight = r2.OffsetRect(float64(tr.X*config.TileSize), float64(tr.Y*config.TileSize)).IntersectionArea(r)
			cr.Note = fmt.Sprintf("map collision rect TR (X: %v, Y: %v)", tr.X, tr.Y)
		}
	}
	if !cr.BottomLeft.Intersects {
		r3 := mi.Map.CollisionRects[bl.Y][bl.X]
		if r3.IsCollision {
			cr.BottomLeft = r3.OffsetRect(float64(bl.X*config.TileSize), float64(bl.Y*config.TileSize)).IntersectionArea(r)
			cr.Note = fmt.Sprintf("map collision rect BL (X: %v, Y: %v)", bl.X, bl.Y)
		}
	}
	if !cr.BottomRight.Intersects {
		r4 := mi.Map.CollisionRects[br.Y][br.X]
		if r4.IsCollision {
			cr.BottomRight = r4.OffsetRect(float64(br.X*config.TileSize), float64(br.Y*config.TileSize)).IntersectionArea(r)
			cr.Note = fmt.Sprintf("map collision rect BR (X: %v, Y: %v)", br.X, br.Y)
		}
	}

	// if any static collisions are found, report those
	if cr.Collides() {
		cr.Assert() // can probably remove these asserts
		return cr
	}

	// if no static collisions are found, find any object collisions
	// NOTE: NPC and player collisions are handled in CollidesWithEntity, since entity collisions will only slow you down, not stop you outright.

	// check for collidable objects (gates, etc)
	for _, obj := range mi.Objects {
		if !obj.IsCollidable() {
			continue
		}
		newCr := checkCornerCollision(r, obj.GetRect())
		if newCr.Collides() {
			newCr.Assert()
			newCr.Note = fmt.Sprintf("collided with object %v", obj.ID)
			newCr.ObjID = utils.Int(obj.ID)
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

// CostMap gets a cost map for a map. Includes all possible obstructions, from NPCs or objects too.
//
// Currently includes:
//
// - tiledMap costmap (just collision rects embedded in tiles)
//
// - NPC positions
//
// - Object rects
//   - Does not include gates, since those may be opened by an NPC (even if locked). NPC logic handles unlocking gates if possible.
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
		costMap[tilePos.Y][tilePos.X] += path_finding.EntityCollision
		// if the entity is currently moving, mark its destination tile as a collision too
		targetTile := n.Entity.TargetTilePos()
		if !targetTile.Equals(tilePos) {
			costMap[targetTile.Y][targetTile.X] += path_finding.EntityCollision
		}
	}

	if mi.PlayerRef != nil {
		playerRect := mi.PlayerRef.Entity.CollisionRect()
		playerX := playerRect.X + (playerRect.W / 2)
		playerY := playerRect.Y + (playerRect.H / 2)
		playerPos := model.ConvertPxToTilePos((playerX), (playerY))
		costMap[playerPos.Y][playerPos.X] += path_finding.EntityCollision
	}

	for _, obj := range mi.Objects {
		if !obj.IsCollidable() {
			continue
		}
		if obj.Type == object.TypeGate {
			// gates may be opened by NPCs, so don't include them in cost map.
			// if a gate is locked for a certain NPC, that can be applied to the costmap elsewhere.
			continue
		}
		// get all tile positions that this object's collision rect overlaps with.
		// ex: sometimes an object is larger than a single tile, or is not placed exactly in the center of a tile.
		// so, we need to make sure any tile it overlaps with is marked as blocked.
		r := obj.GetRect()
		overlapTiles := r.GetOverlappingTiles()
		for _, c := range overlapTiles {
			costMap[c.Y][c.X] += path_finding.BlockThreshold
		}
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
			ActivatorID: mi.PlayerRef.CharacterStateRef.ID,
			LockIDs:     characterstate.GetLockIDs(*mi.PlayerRef.CharacterStateRef),
		}
		result := closestObject.Activate(originX, originY, activateParams)
		logz.Println("Activate Area", closestObject.Type, "Activating...")

		if result.UpdateOccurred {
			mi.HandleObjectUpdate(result, closestObject)
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

func (mi *ActiveMap) HandleObjectUpdate(result object.ObjectUpdateResult, obj *object.Object) {
	if !result.UpdateOccurred {
		panic("no update occurred; no need to call this function")
	}

	logz.Println("Activate Area", obj.Type, "Activation occurred")

	switch obj.Type {
	case object.TypeDoor:
		if result.ChangeMapID != "" {
			mi.worldCtx.HandleMapDoor(result)
		} else {
			logz.Panicln("HandleObjectUpdate", "door object activation occurred, but no changeMapID was set")
		}
	case object.TypeBed:
		if mi.PlayerRef.Entity.IsSleeping {
			panic("trying to activate an object while in a bed. this should be prevented in player input logic.")
		} else {
			// player is now sleeping in bed
			// tell the entity so that it draws in the correct place
			mi.PlayerRef.Entity.SleepInBed(obj)
		}
	case object.TypeChair:
		if mi.PlayerRef.Entity.IsSitting {
			panic("trying to activate an object while sitting. this should be prevented by player input logic.")
		}
		mi.PlayerRef.Entity.SitInChair(obj)
	case object.TypeSign:
		// show bookSession for this sign's text
		if result.SignBookID == "" {
			panic("signBookID was empty!")
		}
		mi.StartBookSession(result.SignBookID, config.DefaultBookSessionParams)
	}
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
