// Package object defines logic for using objects from tiled maps
package object

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity/npc"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

const (
	TypeDoor       = "DOOR" // a door/portal to another map
	TypeGate       = "GATE" // a "gate" is a openable/closable barrier (e.g. a physical door, or gate) in a map
	TypeSpawnPoint = "SPAWN_POINT"
	TypeLight      = "LIGHT"     // lights can be embedded in objects too
	TypeContainer  = "CONTAINER" // TODO
	TypeMisc       = "MISC"      // general purpose; just takes up space

	// Types that aren't actually supported here (specific cases)

	TypeEntity = "ENTITY" // This shouldn't be used for actual objects - just for static entities in maps that are defined by objects.
)

type Object struct {
	Name          string
	Type          string  // NOT from Tiled; set by our code in Load
	xPos, yPos    float64 // logical position in the map
	zOffset       int     // for influencing render order
	DrawX, DrawY  float64 // the actual position on the screen where this was last drawn - for things like click detection
	Width, Height int
	Rect          model.Rect // rect used for step detection (covers entire object)
	CollisionRect model.Rect // rect used for collision (e.g. for gates, only covers bottom tiles)
	collidable    bool       // if set, game will check for collisions with this object

	tileData tiled.TileData // data of a tile embedded in this object

	// frames for a "nextTile" animation when changing state. can go forward and backwards
	imgFrames      []*ebiten.Image
	imgFrameIndex  int
	animSpeedMs    int
	animLastUpdate time.Time

	MouseBehavior mouse.MouseBehavior

	Door Door

	Gate Gate

	Light Light

	SpawnPoint SpawnPoint

	World WorldContext

	PlayerHovering bool

	AudioMgr *audio.AudioManager
}

// GetRect is general purpose function to get the rect that this object occupies in the map. does not scale the values.
func (obj Object) GetRect() model.Rect {
	if obj.collidable {
		return obj.CollisionRect
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
	default:
		return false
	}
}

func (obj Object) Collides(other model.Rect) model.IntersectionResult {
	if !obj.IsCollidable() {
		return model.IntersectionResult{}
	}
	return obj.CollisionRect.IntersectionArea(other)
}

func (obj Object) Y() float64 {
	return obj.yPos + float64(obj.zOffset)
}

func (obj Object) Pos() (x, y float64) {
	return obj.xPos, obj.yPos
}

type Light struct {
	Light *lights.Light
	On    bool
}

type WorldContext interface {
	GetPlayerRect() model.Rect
	GetPlayer() *player.Player
	GetNearbyNPCs(posX, posY, radius float64) []*npc.NPC
}

type Door struct {
	targetMapID      string
	targetSpawnIndex int
	openSound        *audio.Sound
	activateType     string // "click", "step"
}

type Gate struct {
	open          bool
	changingState bool
	openSFXID     defs.SoundID
}

func (g Gate) IsOpen() bool {
	return g.open && !g.changingState
}

type SpawnPoint struct {
	SpawnIndex int
}

func LoadObject(obj tiled.Object, m tiled.Map, audioMgr *audio.AudioManager) *Object {
	o := Object{
		AudioMgr: audioMgr,
		Name:     obj.Name,
		xPos:     obj.X,
		yPos:     obj.Y,
		Width:    int(obj.Width),
		Height:   int(obj.Height),
		Rect: model.Rect{
			X: obj.X,
			Y: obj.Y,
			W: obj.Width,
			H: obj.Height,
		},
		imgFrames: make([]*ebiten.Image, 0),
	}

	allProps := []tiled.Property{}

	// first, get all properties that exist - either at the object level or tile level
	// this is because sometimes the properties might be set at the tile level
	allProps = append(allProps, obj.Properties...)
	if obj.GID != 0 {
		tile, tileset, found := m.GetTileByGID(obj.GID)

		if found {
			allProps = append(allProps, tile.Properties...)
		} else {
			// try to get the tileset still, since we need it for loading tile data
			tileset, found = m.FindTilesetForGID(obj.GID)
			if !found {
				panic("failed to find tileset for object's tile!")
			}
		}

		// Weird bug/inconsistency issue that originates in Tiled:
		// https://discourse.mapeditor.org/t/objects-are-shown-at-the-wrong-position-in-tiled-map/1166
		// https://github.com/mapeditor/tiled/issues/91
		// TLDR: when an image is embedded in an object, the object's origin position is the bottom left instead of top left...
		// workaround is to subtract the height from the y position when an image is embedded
		o.yPos -= float64(o.Height)
		if o.yPos < 0 {
			panic("object y is negative! is this related to the image object y position bug?")
		}

		// also, since there is a tile embedded, load all tile-related info, including tile frames
		o.loadTileData(obj.GID, tile.Properties, tileset, m)
	}

	// get the type first - so we know what values to parse out
	objType, found := tiled.GetStringProperty("TYPE", allProps)
	if !found {
		// Note: I ran into this issue before when I had a tile object in a map and deleted its data for the tile in its tileset.
		// the object pretty much disappeared and I could only see it in the JSON. If this happens, add some pixel data back to the tile where the object's tile was (in the tileset),
		// then go back to the map and it should be there again; delete it from the map, then you can delete it from the tileset.
		logz.Panicln("Object", "no object type property found. ID:", obj.ID, "obj coords:", obj.X, obj.Y)
	}
	o.Type = resolveObjectType(objType)

	// load data for specific object type
	switch o.Type {
	case TypeDoor:
		o.loadDoorObject(allProps)
	case TypeSpawnPoint:
		o.loadSpawnObject(allProps)
	case TypeGate:
		o.loadGateObject(allProps)
	case TypeLight:
		o.loadLightObject(allProps)
	case TypeContainer:
		o.addDefaultCollision()
	case TypeMisc:
		o.addDefaultCollision()
	default:
		panic("object type invalid")
	}

	o.loadGlobal(allProps)

	if o.Type == TypeDoor {
		o.validateDoorObject()
	}
	if o.Type == TypeGate {
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

func (obj *Object) loadLightObject(props []tiled.Property) {
	lightProps := tiled.GetLightProps(props)

	l := lights.NewLight(int(obj.xPos+float64(obj.Width/2)), int(obj.yPos+float64(obj.Height/2)), lightProps, nil)

	obj.Light = Light{
		Light: &l,
		On:    true,
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

func (obj Object) validateDoorObject() {
	// make sure required properties are defined
	if obj.Door.targetMapID == "" {
		panic("door: no target map ID set. check Tiled object definition.")
	}
	if obj.Door.activateType == "" {
		panic("door: no activate type set. check Tiled object definition.")
	}
	switch obj.Door.activateType {
	case "click":
	case "step":
	default:
		logz.Panicf("door [%s]: invalid activation type set: %s. check Tiled object definition.", obj.Name, obj.Door.activateType)
	}

	if obj.Door.openSound == nil {
		logz.Panicf("door [%s]: no openSound defined. check Tiled object definition.", obj.Name)
	}
}

func (obj *Object) loadGateObject(props []tiled.Property) {
	obj.addDefaultCollision()

	for _, prop := range props {
		switch prop.Name {
		case "SFX":
			gateSoundID := defs.SoundID(prop.GetStringValue())
			if gateSoundID == "" {
				panic("no gate sound ID found. TODO: should we make a default one?")
			}
			obj.Gate.openSFXID = gateSoundID
		}
	}

	if obj.Gate.openSFXID == "" {
		panic("no open SFX ID set for gate. make sure to set the 'SFX' property for this object in Tiled.")
	}
}

func (obj *Object) addDefaultCollision() {
	obj.CollisionRect = model.Rect{
		X: obj.xPos,
		Y: obj.yPos + float64(obj.Height) - config.TileSize, // only TileSize height, from the bottom of the object
		W: float64(obj.Width),
		H: config.TileSize,
	}
	obj.collidable = true
}

func (obj *Object) loadDoorObject(props []tiled.Property) {
	for _, prop := range props {
		switch prop.Name {
		case "door_to":
			obj.Door.targetMapID = prop.GetStringValue()
		case "door_spawn_index":
			obj.Door.targetSpawnIndex = prop.GetIntValue()
		case "door_activate":
			obj.Door.activateType = prop.GetStringValue()
		case "SFX":
			doorSound := prop.GetStringValue()
			sound, err := audio.NewSound(doorSound, 0.5)
			if err != nil {
				panic("failed to load door sound:" + err.Error())
			}
			obj.Door.openSound = &sound
		}
	}
}

func (obj *Object) loadSpawnObject(props []tiled.Property) {
	for _, prop := range props {
		switch prop.Name {
		case "spawn_index":
			obj.SpawnPoint.SpawnIndex = prop.GetIntValue()
		}
	}
}

func resolveObjectType(objType string) string {
	switch objType {
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
	default:
		panic("object type doesn't exist: " + objType)
	}
}
