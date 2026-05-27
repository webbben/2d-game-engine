// Package object defines logic for using objects from tiled maps
package object

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/tiled"
)

const (
	TypeDoor       defs.ObjectType = "DOOR" // a door/portal to another map
	TypeGate       defs.ObjectType = "GATE" // a "gate" is a openable/closable barrier (e.g. a physical door, or gate) in a map
	TypeSpawnPoint defs.ObjectType = "SPAWN_POINT"
	TypeLight      defs.ObjectType = "LIGHT"     // lights can be embedded in objects too
	TypeContainer  defs.ObjectType = "CONTAINER" // a container is essentially an inventory that you can open and move items to/from
	TypeMisc       defs.ObjectType = "MISC"      // general purpose; just takes up space

	TypeSign defs.ObjectType = "SIGN" // a sign that you can read (opens a BookDef)

	TypeItem  defs.ObjectType = "ITEM"
	TypeBed   defs.ObjectType = "BED"
	TypeChair defs.ObjectType = "CHAIR"

	TypeTaskArea defs.ObjectType = "TASK_AREA"

	// Types that aren't actually supported here (specific cases)

	// TODO: is this even used?
	TypeEntity defs.ObjectType = "ENTITY" // This shouldn't be used for actual objects - just for static entities in maps that are defined by objects.
)

const (
	PropOnObjID     = "on_obj_id" // if set, this object should render on top of another object
	PropOwnerCharID = "owner_id"  // if set, a unique character of the given ID owns this object. note that this doesn't work for non-unique characters.
	PropRoleID      = "role_id"   // if set, characters with this role can use this object or effective assume "ownership" (in the absense of a specific owner character.)

	PropContainerDefID = "container_def_id"
	PropContainerGenID = "container_gen_id"
)

// TODO: *Sigh* this probably could use some refactoring. I've been avoiding admitting it, but it would most likely work cleaner as an interface.
// As it stands right now, we are basically mashing a bunch of different types of objects into this one object struct. This works, but it feels kind of messy.
// I think it would be smarter to make an interface that has all the methods needed for common things, but then the inner logic can be more cleanly separated.
// But, not gonna tackle this right now, because I don't want to get diverted on yet another big refactor lol.

type Object struct {
	eventBus *pubsub.EventBus
	subIDs   []string

	Name string // TODO: I don't think most objects actually have Names; the name property in Tiled is usually left empty. should we just delete this?

	DisplayName string // Not implemented; create a property in Tiled for this (Name won't work, since you can't give names to tiles in tilesets)

	targetedByNPC id.CharacterStateID // if set, this means an NPC is currently trying to come and activate this object.

	mapID defs.MapID // used for knowing which map state to check

	ID         int             // the ID property from Tiled; just a counter I believe.
	Type       defs.ObjectType // NOT from Tiled; set by our code in Load
	xPos, yPos float64         // logical position in the map
	zOffset    int             // for influencing render order

	// Ownership and roles - determines who can use objects such as beds and chairs.

	OwnerID id.CharacterStateID // ID of character who owns this object
	RoleID  defs.RoleID         // ID of role that owns or is assigned to this object

	// some confusing object render order stuff

	CustomY      float64 // for setting a custom Y value (used when sorting renderables to determine what shows on top of what)
	OnTopOfObjID int     // if set (not -1) then this object is on top of another object, like on a table for example.

	DrawX, DrawY    float64 // the actual position on the screen where this was last drawn - for things like click detection
	Width, Height   int
	rect            model.Rect // rect used for step detection (covers entire object)
	collisionRect   model.Rect // rect used for collision (e.g. for gates, only covers bottom tiles)
	collidable      bool       // if set, game will check for collisions with this object
	collisionHeight int        // number of tiles in height the collision should be. must not be 0 or bigger than object.

	tileData tiled.TileData // data of a tile embedded in this object

	lockID string // if this object has a lock, this will be set. lock info can be accessed from map state.

	// frames for a "nextTile" animation when changing state. can go forward and backwards
	imgFrames      []*ebiten.Image
	imgFrameIndex  int
	animSpeedMs    int
	animLastUpdate time.Time

	// object-specific data

	TaskArea   TaskArea
	Door       Door
	Gate       Gate
	Container  Container
	Sign       Sign
	Light      Light
	Bed        Bed
	Chair      Chair
	SpawnPoint SpawnPoint

	World WorldContext

	PlayerHovering bool

	AudioMgr *audio.AudioManager
	dataman  *datamanager.DataManager
}

func (obj Object) OnMapClose() {
	for _, subID := range obj.subIDs {
		obj.eventBus.Unsubscribe(subID)
	}
}

func (obj Object) GetDisplayName() string {
	if obj.DisplayName != "" {
		return obj.DisplayName
	}

	switch obj.Type {
	case TypeChair:
		return "Chair"
	case TypeBed:
		return "Bed"
	case TypeGate:
		return "Gate"
	}

	return ""
}

func (obj Object) GetInfo() defs.ObjectInfo {
	return defs.ObjectInfo{
		ID:           obj.ID,
		DisplayName:  obj.GetDisplayName(),
		Type:         obj.Type,
		Activatable:  obj.IsActivatable(),
		ActivateText: obj.GetActivateText(),
	}
}

func (obj Object) GetActivateText() string {
	if !obj.IsActivatable() {
		return ""
	}

	// TODO: add custom activation text property in tiled

	switch obj.Type {
	case TypeBed:
		return "Sleep"
	case TypeChair:
		return "Sit"
	case TypeGate:
		if obj.Gate.IsOpen() {
			return "Close"
		}
		return "Open"
	case TypeDoor:
		return "Enter"
	case TypeContainer:
		return "Open"
	case TypeItem:
		return "Pick up"
	case TypeLight:
		if obj.Light.On {
			return "Put out"
		}
		return "Light"
	case TypeSign:
		return "Read"
	}

	return ""
}

func (obj Object) IsHovering(x, y int) bool {
	if !obj.IsHoverable() {
		return false
	}
	return obj.GetDrawRect().Within(x, y)
}

func (obj *Object) SetTargetingNPC(id id.CharacterStateID) {
	if obj.targetedByNPC != "" {
		logz.Panicln("Object", "tried to set targeting NPC, but there was already another ID set... make sure you clear it before trying to set a new one. existing id:", obj.targetedByNPC, "new id:", id)
	}
	obj.targetedByNPC = id
}

func (obj *Object) ClearTargetingNPC() {
	obj.targetedByNPC = ""
}

func (obj Object) GetTargetingNPC() id.CharacterStateID {
	return obj.targetedByNPC
}

func (obj Object) GetLockID() string {
	return obj.lockID
}

func (obj Object) IsCurrentlyActivating() bool {
	switch obj.Type {
	case TypeGate:
		return obj.Gate.changingState
	}

	return false
}

// GetRect is general purpose function to get the rect that this object occupies in the map. does not scale the values.
func (obj Object) GetRect() model.Rect {
	if obj.IsCollidable() {
		return obj.collisionRect
	}
	return model.Rect{
		X: obj.xPos,
		Y: obj.yPos,
		W: float64(obj.Width),
		H: float64(obj.Height),
	}
}

// GetDrawRect gets the rect that represents where the object is on the screen, as drawn. useful for detecting things like mouseclicks.
func (obj Object) GetDrawRect() model.Rect {
	return model.NewRect(obj.DrawX, obj.DrawY, float64(obj.Width)*config.GameScale, float64(obj.Height)*config.GameScale)
}

func (obj Object) IsCollidable() bool {
	if obj.Type == TypeGate {
		if obj.Gate.IsOpen() {
			return false
		}
	}
	return obj.collidable
}

func (obj Object) IsActivatable() bool {
	switch obj.Type {
	case TypeDoor:
		return true
	case TypeGate:
		return true
	case TypeLight:
		return true
	case TypeContainer:
		return true
	case TypeBed:
		return true
	case TypeChair:
		return true
	case TypeSign:
		return true
	default:
		return false
	}
}

// IsHoverable determines if an object can be "hovered" on by the mouse.
// If an object is hoverable, then it can be targeted by the player in a map (at least in terms of getting its info)
func (obj Object) IsHoverable() bool {
	switch obj.Type {
	case TypeGate, TypeLight, TypeContainer, TypeChair, TypeBed, TypeDoor, TypeItem, TypeSign:
		return true
	default:
		return false
	}
}

func (obj Object) Collides(other model.Rect) model.IntersectionResult {
	if !obj.IsCollidable() {
		return model.IntersectionResult{}
	}
	return obj.collisionRect.IntersectionArea(other)
}

func (obj Object) Y() float64 {
	if obj.OnTopOfObjID != -1 {
		return obj.CustomY
	}
	// we want to get the y position of the bottom tile; this is because the bottom tile is
	// logically where the object sits on the ground, and therefore should be used for ordering in drawing.
	bottomY := obj.yPos + float64(obj.Height) - config.TileSize
	return bottomY + float64(obj.zOffset)
}

func (obj Object) X() float64 {
	return obj.xPos
}

// Pos returns the logical position on the map; i.e. the top left corner of the object.
func (obj Object) Pos() (x, y float64) {
	return obj.xPos, obj.yPos
}

// TilePos gets the coordinates of the actual tile this object primarily lies in.
// Gets the tile position of the center of the collision rect; so this is a good representation of where the object "really is".
func (obj Object) TilePos() model.Coords {
	return model.GetTilePosOfRectCenter(obj.GetRect())
}

type WorldContext interface {
	// Used for checking how far an object is from the player, when the player tries to activate it (among other things).
	// TODO: should we have this? also, it seems possible that NPCs will need to be able to activate objects (like doors, to open them)
	GetPlayerRect() model.Rect

	// Used for checking if an object like a gate has other entities colliding with it
	RectCollidesWithOthers(r model.Rect, excludeEntID string, excludeObjID int) bool
}

type SpawnPoint struct {
	SpawnIndex int
}

func LoadObject(obj tiled.Object, m tiled.Map, audioMgr *audio.AudioManager, dataman *datamanager.DataManager, eventBus *pubsub.EventBus, mapID defs.MapID, world WorldContext) *Object {
	if dataman == nil {
		panic("dataman was nil")
	}
	if audioMgr == nil {
		panic("audioMgr was nil")
	}
	if world == nil {
		panic("world context was nil")
	}
	if obj.Ellipse {
		panic("object was an ellipse; these aren't used in the actual active map, just for planning.")
	}
	o := Object{
		eventBus: eventBus,
		AudioMgr: audioMgr,
		dataman:  dataman,
		Name:     obj.Name,
		ID:       obj.ID,
		xPos:     obj.X,
		yPos:     obj.Y,
		Width:    int(obj.Width),
		Height:   int(obj.Height),
		rect: model.Rect{
			X: obj.X,
			Y: obj.Y,
			W: obj.Width,
			H: obj.Height,
		},
		imgFrames:    make([]*ebiten.Image, 0),
		World:        world,
		mapID:        mapID,
		OnTopOfObjID: -1,
	}

	// We need to load all properties for an object. an Object can come in two forms:
	//
	// 1) an object with a tile embedded; used when we want to place an object with an image in it somewhere, like placing barrels, or torches, etc.
	//
	// 2) an object without a tile; used when we want to just insert data at a certain position, like spawn points, or light sources that don't have tiles, etc.
	//
	// if GID is set, there should be a tile embedded in the object.

	objectInfo := m.GetObjectPropsAndTile(obj)
	allProps := objectInfo.AllProps
	if objectInfo.HasEmbeddedTile {
		// Weird bug/inconsistency issue that originates in Tiled:
		// https://discourse.mapeditor.org/t/objects-are-shown-at-the-wrong-position-in-tiled-map/1166
		// https://github.com/mapeditor/tiled/issues/91
		// TLDR: when an image is embedded in an object, the object's origin position is the bottom left instead of top left...
		// workaround is to subtract the height from the y position when an image is embedded
		o.yPos -= float64(o.Height)
		if o.yPos < 0 {
			logz.Warnln("Object", "object y is negative! is this related to the image object y position bug? ID:", o.ID, "yPos:", o.yPos)
		}

		var tileProps []tiled.Property
		if objectInfo.Tile != nil {
			tileProps = objectInfo.Tile.Properties
		}

		// also, since there is a tile embedded, load all tile-related info, including tile frames
		o.loadTileData(obj.GID, tileProps, *objectInfo.Tileset, m)
	}

	// get the type first - so we know what values to parse out
	objType, found := GetObjectType(allProps)
	if !found {
		// Note: I ran into this issue before when I had a tile object in a map and deleted its data for the tile in its tileset.
		// the object pretty much disappeared and I could only see it in the JSON. If this happens, add some pixel data back to the tile where the object's tile was (in the tileset),
		// then go back to the map and it should be there again; delete it from the map, then you can delete it from the tileset.
		logz.Panicln("Object", "no object type property found. ID:", obj.ID, "obj coords:", obj.X, obj.Y)
	}
	o.Type = objType

	// display name
	if displayName, found := tiled.GetStringProperty("displayName", allProps); found {
		o.DisplayName = displayName
	}

	// check if there's a lock
	lockID, found := tiled.GetStringProperty("lock_id", allProps)
	if found {
		if lockID == "" {
			panic("lockID is empty")
		}
		// confirm that this lockID is in the map state
		mapState := dataman.GetMapState(mapID)
		if _, exists := mapState.MapLocks[lockID]; !exists {
			logz.Panicln("LoadObject", "lock ID not found in map state:", lockID, "mapID:", mapID)
		}
		switch o.Type {
		case TypeContainer, TypeDoor, TypeGate:
			o.lockID = lockID
		default:
			logz.Panicln("Object", "cannot put lock on object type:", o.Type)
		}
	}

	// check for render order stuff
	onTopOfObj, found := tiled.GetIntProperty(PropOnObjID, allProps)
	if found {
		if onTopOfObj < 0 {
			panic("onTopOfObj was negative!")
		}
		o.OnTopOfObjID = onTopOfObj
		// once all objects have been loaded, we will come back and set the customY value to ensure this renders ontop of the other object
	}

	// other flags set in objects
	noCollision, _ := tiled.GetBoolProperty("no_collision", allProps)
	collisionHeight, found := tiled.GetIntProperty("collision_height", allProps)
	if found {
		if collisionHeight*config.TileSize > o.Height {
			logz.Panicln("Object", "collision height property is too tall for object. collision_height:", collisionHeight, "objID:", obj.ID)
		}
		if collisionHeight <= 1 {
			logz.Panicln("Object", "collision_height property set, but it was <= 1 (shouldn't set to 1, since that's default...):", collisionHeight, "objID:", obj.ID)
		}
		o.collisionHeight = collisionHeight
	}

	// check ownership and roles
	ownerID, found := tiled.GetStringProperty(PropOwnerCharID, allProps)
	if found {
		o.OwnerID = id.CharacterStateID(ownerID)
	}
	roleID, found := tiled.GetStringProperty(PropRoleID, allProps)
	if found {
		o.RoleID = defs.RoleID(roleID)
	}
	if o.OwnerID != "" && o.RoleID != "" {
		logz.Panicln("LoadObject", "object has both an owner and role ID; this can cause problems with NPCs correctly identifying who should use this object. only set one of these.")
	}

	// load data for specific object type
	switch o.Type {
	case TypeDoor:
		if objectInfo.HasEmbeddedTile && !noCollision {
			// if this door object is represented by a tile, then it should probably have built-in collisions.
			// other doors without tiles are probably just zones that the player can walk into to trigger a map change.
			o.addDefaultCollision()
		}
		o.loadDoorObject(allProps)
		mapDef := dataman.GetMapDef(o.Door.TargetMapID)
		if o.DisplayName == "" {
			o.DisplayName = fmt.Sprintf("Door to %s", mapDef.DisplayName)
		}
	case TypeSpawnPoint:
		o.loadSpawnObject(allProps)
	case TypeGate:
		if noCollision {
			panic("gate object has no collision property?")
		}
		o.addDefaultCollision()
		o.loadGateObject(allProps)
	case TypeLight:
		o.loadLightObject(allProps)
	case TypeContainer:
		if !noCollision {
			o.addDefaultCollision()
		}
		o.loadContainerObject(allProps)
	case TypeMisc:
		if !noCollision {
			o.addDefaultCollision()
		}
	case TypeChair:
		if !noCollision {
			o.addDefaultCollision()
		}
		o.loadChairObject(allProps)
	case TypeTaskArea:
		o.loadTaskAreaObject(allProps)
	case TypeItem:
		itemID, found := tiled.GetStringProperty("item_id", allProps)
		if !found {
			logz.Panic("item didn't have item_id")
		}
		itemDef := dataman.GetItemDef(defs.ItemID(itemID))
		o.DisplayName = itemDef.GetName()
	case TypeSign:
		o.loadSignObject(allProps)
	}

	o.loadGlobal(allProps)

	switch o.Type {
	case TypeDoor:
		o.validateDoorObject()
	case TypeGate:
		o.validateGateObject()
	}

	return &o
}

func (obj *Object) loadGlobal(props []tiled.Property) {
	zOffset, found := tiled.GetIntProperty("z_offset", props)
	if found {
		obj.zOffset = zOffset
	}
}

func (obj *Object) loadTileData(tileGID int, tileProps []tiled.Property, tileset tiled.Tileset, m tiled.Map) {
	if !tileset.Loaded {
		logz.Println("Load Object", "tileset for object tile hasn't been loaded yet; loading now...")
		err := tileset.LoadJSONData(m.AbsSourcePath)
		if err != nil {
			logz.Panicf("failed to load tileset data (while loading object data): %s", err.Error())
		}
	}

	tileData, exists := m.TileImageMap[tileGID]
	if !exists {
		panic("tile attached to object, but tile not found in map's TileImageMap")
	}
	obj.imgFrames = append(obj.imgFrames, tileData.CurrentFrame)

	// also link the tileData itself, in case it contains a base animation for the tile
	obj.tileData = tileData

	if len(tileProps) == 0 {
		return
	}

	obj.animSpeedMs = 100 // animation speed for the change-state animation

	// attached tile has properties
	// if this tile has a "nextTile" property, that means there is a state change animation
	// load all the tiles in this "nextTile" chain until the "nextTile" property stops appearing
	var nextTileID int
	var propFound bool
	nextTileID, propFound = tiled.GetIntProperty("nextTile", tileProps)
	numberFound := 0
	for propFound {
		numberFound++
		gid := nextTileID + tileset.FirstGID

		// load tile image
		img, exists := m.TileImageMap[gid]
		if !exists {
			panic("tile attached to object, but tile not found in map's TileImageMap")
		}
		obj.imgFrames = append(obj.imgFrames, img.CurrentFrame)

		// load next tile
		nextTile, _, tileFound := m.GetTileByGID(gid)
		if !tileFound {
			// if tile not found, that means no properties exist for this tile; should be the last one in the chain.
			break
		}
		nextTileID, propFound = tiled.GetIntProperty("nextTile", nextTile.Properties)
	}
	if numberFound == 1 {
		// wait, only one tile was found in the "nextTile" chain? something must be wrong
		panic("LoadObject: only one tile was found in a nextTile chain; something must be wrong...")
	}
}

func (obj Object) validateGateObject() {
	if len(obj.imgFrames) < 2 {
		panic("gate: must have at least two tile frames to represent the open and closed states")
	}
}

func (obj *Object) addDefaultCollision() {
	tileHeight := 1
	if obj.collisionHeight > 0 {
		tileHeight = obj.collisionHeight
	}
	collisionHeight := float64(config.TileSize * tileHeight)

	obj.collisionRect = model.Rect{
		X: obj.xPos,
		Y: obj.yPos + float64(obj.Height) - collisionHeight, // only TileSize height, from the bottom of the object
		W: float64(obj.Width),
		H: collisionHeight,
	}
	obj.collidable = true
}

const (
	PropSpawnIndex string = "spawn_index"
)

func (obj *Object) loadSpawnObject(props []tiled.Property) {
	for _, prop := range props {
		switch prop.Name {
		case PropSpawnIndex:
			obj.SpawnPoint.SpawnIndex = prop.GetIntValue()
		}
	}
}

func GetObjectType(allObjProperties []tiled.Property) (defs.ObjectType, bool) {
	objType, found := tiled.GetStringProperty("TYPE", allObjProperties)
	if !found {
		return "", false
	}

	return resolveObjectType(objType), true
}

func resolveObjectType(objType string) defs.ObjectType {
	switch defs.ObjectType(objType) {
	case TypeDoor:
		return TypeDoor
	case TypeSpawnPoint:
		return TypeSpawnPoint
	case TypeGate:
		return TypeGate
	case TypeLight:
		return TypeLight
	case TypeContainer:
		return TypeContainer
	case TypeMisc:
		return TypeMisc
	case TypeItem:
		return TypeItem
	case TypeBed:
		return TypeBed
	case TypeChair:
		return TypeChair
	case TypeTaskArea:
		return TypeTaskArea
	case TypeSign:
		return TypeSign
	default:
		panic("object type doesn't exist: " + objType)
	}
}
