package object

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
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
	TYPE_DOOR        = "DOOR" // a door/portal to another map
	TYPE_GATE        = "GATE" // a "gate" is a openable/closable barrier (e.g. a physical door, or gate) in a map
	TYPE_SPAWN_POINT = "SPAWN_POINT"
	TYPE_LIGHT       = "LIGHT"     // lights can be embedded in objects too
	TYPE_CONTAINER   = "CONTAINER" // TODO
	TYPE_MISC        = "MISC"      // general purpose; just takes up space
)

type Object struct {
	Name          string
	Type          string
	xPos, yPos    float64 // logical position in the map
	zOffset       int     // for influencing render order
	DrawX, DrawY  float64 // the actual position on the screen where this was last drawn - for things like click detection
	Width, Height int
	Rect          model.Rect // rect used for step detection (covers entire object)
	CollisionRect model.Rect // rect used for collision (e.g. for gates, only covers bottom tiles)
	Collidable    bool       // if set, game will check for collisions with this object

	imgFrames      []*ebiten.Image
	imgFrameIndex  int
	animSpeedMs    int
	animLastUpdate time.Time

	MouseBehavior mouse.MouseBehavior

	OnLeftClick  func()
	OnRightClick func()

	Door Door

	Gate Gate

	Light Light

	SpawnPoint SpawnPoint

	World WorldContext

	PlayerHovering bool
}

func (obj Object) Collides(other model.Rect) model.IntersectionResult {
	if !obj.Collidable {
		return model.IntersectionResult{}
	}
	switch obj.Type {
	case TYPE_GATE:
		if obj.Gate.IsOpen() {
			return model.IntersectionResult{}
		}
		return obj.CollisionRect.IntersectionArea(other)
	default:
		return obj.CollisionRect.IntersectionArea(other)
	}
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
}

func (g Gate) IsOpen() bool {
	return g.open && !g.changingState
}

type SpawnPoint struct {
	SpawnIndex int
}

func LoadObject(obj tiled.Object, m tiled.Map) *Object {
	o := Object{
		Name:   obj.Name,
		xPos:   obj.X,
		yPos:   obj.Y,
		Width:  int(obj.Width),
		Height: int(obj.Height),
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
		panic("no object type property found")
	}
	o.Type = objType

	// load data for specific object type
	switch o.Type {
	case TYPE_DOOR:
		o.loadDoorObject(allProps)
	case TYPE_SPAWN_POINT:
		o.loadSpawnObject(allProps)
	case TYPE_GATE:
		o.loadGateObject(allProps)
	case TYPE_LIGHT:
		o.loadLightObject(allProps)
	case TYPE_CONTAINER:
		o.addDefaultCollision()
	case TYPE_MISC:
		o.addDefaultCollision()
	default:
		panic("object type invalid")
	}

	o.loadGlobal(allProps)

	if o.Type == TYPE_DOOR {
		o.validateDoorObject()
	}
	if o.Type == TYPE_GATE {
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

	tileImg, exists := m.TileImageMap[tileGID]
	if !exists {
		panic("tile attached to object, but tile not found in map's TileImageMap")
	}
	obj.imgFrames = append(obj.imgFrames, tileImg.CurrentFrame)

	obj.animSpeedMs = 100

	if len(tileProps) == 0 {
		return
	}

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
		logz.Panicf("door: invalid activation type set: %s. check Tiled object definition.", obj.Door.activateType)
	}

	if obj.Door.openSound == nil {
		panic("door: no openSound defined. check Tiled object definition.")
	}
}

func (obj *Object) loadGateObject(props []tiled.Property) {
	obj.addDefaultCollision()

	for _, prop := range props {
		switch prop.Name {
		case "gate_sound":
			gateSound := prop.GetStringValue()
			switch gateSound {
			case "wood":
				// TODO
			case "metal":
				// TODO
			default:
				logz.Panicf("loadGateProperty: gate sound value not recognized: %s", gateSound)
			}
		}
	}
}

func (obj *Object) addDefaultCollision() {
	obj.CollisionRect = model.Rect{
		X: obj.xPos,
		Y: obj.yPos + float64(obj.Height) - config.TileSize, // only TileSize height, from the bottom of the object
		W: float64(obj.Width),
		H: config.TileSize,
	}
	obj.Collidable = true
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
		case "door_sound":
			doorSound := prop.GetStringValue()
			switch doorSound {
			case "wood":
				// TODO work out a system for defining these default door sounds
				sound, err := audio.LoadSound("/Users/benwebb/dev/personal/ancient-rome/assets/audio/sfx/door/open_door_01.mp3", 0.5)
				if err != nil {
					panic("failed to load door sound:" + err.Error())
				}
				obj.Door.openSound = &sound
			default:
				panic("door_sound value not found:" + doorSound)
			}
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
	case TYPE_DOOR:
		return TYPE_DOOR
	case TYPE_SPAWN_POINT:
		return TYPE_SPAWN_POINT
	case TYPE_GATE:
		return TYPE_GATE
	case TYPE_LIGHT:
		return TYPE_LIGHT
	case TYPE_CONTAINER:
		return TYPE_CONTAINER
	case TYPE_MISC:
		return TYPE_MISC
	default:
		panic("object type doesn't exist: " + objType)
	}
}
