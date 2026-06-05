package world

import (
	"fmt"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/utils"
)

// EnsureMapStateExists checks if a map state exists, and instantiates it if it doesn't.
// So, anytime you're dealing with a map, just run this function to make sure its state exists and/or has been instantiated correctly.
//
// Warning: this should NOT be used on "template" map defs.
func (w *World) EnsureMapStateExists(mapID defs.MapID) {
	if w.Dataman.MapStateExists(mapID) {
		return
	}

	// make sure def exists for this mapID
	def := w.Dataman.GetMapDef(mapID)
	if def.IsMapGenTemplate {
		// shouldn't create a map state for this mapID, since it's just a template
		logz.Panicln("EnsureMapStateExists", "tried to create map state for a template map def. whoever called this should check and handle template map defs separately, and not call this.", mapID)
	}

	logz.Println("WORLD", "Map state doesn't exist yet; creating...", mapID)

	w.CreateNewMapState(mapID, "")
}

// CreateNewMapState is for creating a map state when initializing a new game. This should not be run when loading an existing game,
// since it would cause duplicate NPCs to be recreated.
//
// can pass a customMapStateID; only intended for use when generating maps using a map generator.
func (w *World) CreateNewMapState(mapID defs.MapID, customMapStateID string) {
	if w.Dataman.MapStateExists(mapID) {
		logz.Panicln("CreateNewMapState", "tried to create a new map state, but one already exists:", mapID)
	}

	mapStateID := mapID
	if customMapStateID != "" {
		mapStateID = defs.MapID(customMapStateID)
	}

	mapState := state.MapState{
		ID:            mapStateID,
		MapLocks:      make(map[string]state.LockState),
		MapBeds:       make(map[int]state.BedState),
		MapContainers: make(map[int]*state.ContainerState),
		DoorOverrides: make(map[int]state.DoorState),
	}

	// look through Tiled map data to initialize state for things like pre-defined item placements, door locks, etc.
	m := tiled.LoadMap(mapID, false)

	objLayers := getAllObjectLayers(m.Layers)

	// Get initial items
	for _, layer := range objLayers {
		for _, obj := range layer.Objects {
			if obj.Ellipse || obj.Text != nil {
				// text, ellipses, etc are ignored since they're only used for planning
				continue
			}

			objectInfo := m.GetObjectPropsAndTile(obj)
			objType, found := object.GetObjectType(objectInfo.AllProps)
			if !found {
				logz.Panicln("CreateNewMapState", "object didn't have a TYPE property:", obj.Name, obj.ID, "mapID:", mapID)
			}

			// check if there is a lock on this object
			var lockLevel int
			var lockID string
			lockLevel, found = tiled.GetIntProperty(object.PropLockLevel, objectInfo.AllProps)
			if found {
				if lockLevel <= 0 {
					logz.Panicln("CreateNewMapState", "found lock level property on object, but it had a level of <= 0.", obj.Name, obj.ID, "mapID:", mapID)
				}
				lockID, found = tiled.GetStringProperty(object.PropLockID, objectInfo.AllProps)
				if !found {
					// if no custom lock ID is set, then just generate a default one for this object
					lockID = object.GetDefaultLockID(obj.ID)
				}
				if lockID == "" {
					logz.Panicln("CreateNewMapState", "lock ID property was empty:", obj.Name, obj.ID, "mapID:", mapID)
				}
				// add it to the lock map
				mapState.MapLocks[lockID] = state.LockState{
					OriginalLockLevel: lockLevel,
					LockID:            lockID,
					LockLevel:         lockLevel,
				}
			}

			switch objType {
			case object.TypeItem:
				if lockID != "" {
					logz.Panicln("CreateNewMapState", "a lock was put on an item, which doesn't make any sense.", obj.Name, obj.ID, "mapID:", mapID)
				}
				// get item ID
				itemID, found := tiled.GetStringProperty("item_id", objectInfo.AllProps)
				if !found {
					logz.Panicln("CreateNewMapState", "found item object, but no item_id property was found:", obj.Name, obj.ID, "mapID:", mapID)
				}
				// confirm item exists
				defID := defs.ItemID(itemID)
				itemDef := w.Dataman.GetItemDef(defID)
				mapState.MapItems = append(mapState.MapItems, state.MapItemState{
					ItemState: state.ItemState{
						DefID:      defs.ItemID(itemID),
						Quantity:   1,
						Durability: itemDef.MaxDurability,
					},
					X: obj.X,
					Y: obj.Y,
				})
			case object.TypeBed:
				if lockID != "" {
					logz.Panicln("CreateNewMapState", "a lock was put on a bed, which doesn't make any sense.", obj.Name, obj.ID, "mapID:", mapID)
				}
				// instantiate the NPC that is associated with this bed
				var charStateID id.CharacterStateID
				charGenID, found := tiled.GetStringProperty("characterGeneratorID", objectInfo.AllProps)
				if found {
					// generate an random NPC for this bed
					charGen := w.Dataman.GetCharacterGenerator(charGenID)
					charStateID = w.GenerateCharacter(charGen, mapID, mapID, obj.ID)
				} else {
					// create the NPC based on its def
					charDefID, found := tiled.GetStringProperty("characterDefID", objectInfo.AllProps)
					if found {
						// this bed has a specific character def ID set
						params := entity.NewCharacterStateParams{
							InitialMapID: mapID,
							HomeMapID:    mapID,
							HomeMapBedID: obj.ID,
						}
						charStateID = entity.CreateNewCharacterState(defs.CharacterDefID(charDefID), params, w.Dataman)
					}
				}

				// if a character was created, register this bed in the character state
				if charStateID != "" {
					charState := w.Dataman.GetCharacterState(charStateID)
					charState.HomeMapID = mapID
					charState.HomeMapBedID = obj.ID
				}

				// add this bed to the MapBeds
				mapState.MapBeds[obj.ID] = state.BedState{
					MapObjID: obj.ID,
					OwnerID:  charStateID,
				}
			case object.TypeContainer:
				containerState := state.ContainerState{}
				// find out if this is a predefined container inventory, or if we should generate one
				// check for container_def_id
				if containerDefID, found := tiled.GetStringProperty(object.PropContainerDefID, objectInfo.AllProps); found {
					containerDef := w.Dataman.GetContainerDef(containerDefID)
					containerState.Inventory = item.ConvertInitialItemStateDefs(containerDef.Inventory)
				} else if containerGenID, found := tiled.GetStringProperty(object.PropContainerGenID, objectInfo.AllProps); found {
					containerGen := w.Dataman.GetContainerGenerator(containerGenID)
					containerState.Inventory = item.ConvertInitialItemStateDefs(containerGen.GenerateItems(w.GameCtx))
				} else {
					// no def or gen; use default generator
					containerGen := w.Dataman.GetContainerGenerator(config.DefaultContainerGeneratorID)

					containerState.Inventory = item.ConvertInitialItemStateDefs(containerGen.GenerateItems(w.GameCtx))
				}

				mapState.MapContainers[obj.ID] = &containerState
			}
		}
	}

	w.Dataman.LoadMapState(mapState)
}

func (w *World) GenerateMap(mapGeneratorID string, returnMapID defs.MapID, returnSpawnIndex int) defs.MapID {
	mapGen := w.Dataman.GetMapGenerator(mapGeneratorID)

	mapGen.Validate()

	// ensure the mapDefID set in the map generator is a template map
	def := w.Dataman.GetMapDef(mapGen.MapDefID)
	if !def.IsMapGenTemplate {
		logz.Panicln("GenerateMap", "map generator linked to a mapDef that wasn't a template map:", mapGen.MapDefID)
	}

	// first, make the map state according to the def, but using a unique mapStateID
	mapStateID := fmt.Sprintf("%s_%s", mapGen.MapDefID, utils.GenerateUUID()[:8])

	w.CreateNewMapState(mapGen.MapDefID, mapStateID)

	// now, get the map state and make the changes we need

	mapState := w.Dataman.GetMapState(defs.MapID(mapStateID))
	utils.PanicAssert(mapStateID == string(mapState.ID), "created map state isn't using the custom map state id... "+mapStateID+" vs "+string(mapState.ID))

	mapState.IsGenerated = true
	mapState.GeneratedMapDefID = mapGen.MapDefID

	if mapGen.OverrideDisplayName != "" {
		mapState.DisplayName = mapGen.OverrideDisplayName
	}
	if mapGen.OverrideRegion != "" {
		mapState.RegionID = mapGen.OverrideRegion
	}

	// insert the inhabitants into the game
	// ensure there are enough beds
	m := tiled.LoadMap(mapGen.MapDefID, false)
	objLayers := getAllObjectLayers(m.Layers)
	bedCount := 0
	doorFound := false

	for _, layer := range objLayers {
		for _, obj := range layer.Objects {
			if obj.Ellipse || obj.Text != nil {
				continue
			}

			objectInfo := m.GetObjectPropsAndTile(obj)
			objType, found := object.GetObjectType(objectInfo.AllProps)
			if !found {
				logz.Panicln("CreateNewMapState", "object didn't have a TYPE property:", obj.Name, obj.ID, "mapID:", mapGen.MapDefID)
			}

			// DOORS: make sure the following is true:
			// 1) only one door in the map (generated maps aren't allowed to have "nested" inner maps; things could get too complicated so let's avoid it.)
			// 2) the singular door does NOT have a door_to or to_spawn prop set; the door will have these things filled in for it; so panic if its set, to avoid any possible confusion.
			// 3) also doesn't have a map generator ID; not allowed to generate nested maps in another generated map
			if objType == object.TypeDoor {
				if doorFound {
					logz.Panicln("GenerateMap", "more than one door was found in template map; we only allow one door - the door to enter/exit.", mapGen.MapDefID)
				}
				doorFound = true

				if _, found := tiled.GetStringProperty(object.PropDoorTo, objectInfo.AllProps); found {
					logz.Panicln("GenerateMap", "door had a door_to prop defined; this should be empty, since its filled in during map generation.", obj.ID, mapGen.MapDefID)
				}
				if _, found := tiled.GetStringProperty(object.PropDoorSpawnIndex, objectInfo.AllProps); found {
					logz.Panicln("GenerateMap", "door had a to_spawn_index prop defined; this should be empty, since its filled in during map generation.", obj.ID, mapGen.MapDefID)
				}
				if _, found := tiled.GetStringProperty(object.PropDoorMapGeneratorID, objectInfo.AllProps); found {
					logz.Panicln("GenerateMap", "door had a genMapID prop defined; since this is a generated map, we don't allow 'nested generated maps'.", obj.ID, mapGen.MapDefID)
				}

				// If this door is valid, then let's set it as the "return door". This door returns to the map that originally linked here.
				mapState.DoorOverrides[obj.ID] = state.DoorState{
					OverrideDestinationMap:   returnMapID,
					OverrideDestinationSpawn: &returnSpawnIndex,
				}
				continue
			}

			if objType != object.TypeBed {
				continue
			}

			// ensure bed does not have a set owner in tiled props - this isn't allowed in mapGenerator maps, since we fill them in with the map generator's inhabitants.
			_, found = tiled.GetStringProperty("characterGeneratorID", objectInfo.AllProps)
			if found {
				logz.Panicln("GenerateMap", "bed in map had a character generator set; MapGenerator maps should not define bed owners.", mapGen.MapDefID)
			}
			_, found = tiled.GetStringProperty("characterDefID", objectInfo.AllProps)
			if found {
				logz.Panicln("GenerateMap", "bed in map had a character def set; MapGenerator maps should not define bed owners.", mapGen.MapDefID)
			}

			// valid bed found; now, instantiate an NPC for this bed and set the bed's state so it knows its owner
			var charStateID id.CharacterStateID

			if bedCount < len(mapGen.InhabitantCharacterDefs) {
				charDefID := mapGen.InhabitantCharacterDefs[bedCount]
				params := entity.NewCharacterStateParams{
					InitialMapID: mapState.ID,
					HomeMapID:    mapState.ID,
					HomeMapBedID: obj.ID,
				}
				charStateID = entity.CreateNewCharacterState(defs.CharacterDefID(charDefID), params, w.Dataman)
			} else if bedCount < len(mapGen.InhabitantCharacterGens) {
				// generate an random NPC for this bed
				charGenID := mapGen.InhabitantCharacterGens[bedCount]
				charGen := w.Dataman.GetCharacterGenerator(charGenID)
				charStateID = w.GenerateCharacter(charGen, mapState.ID, mapState.ID, obj.ID)
			}

			bedCount++

			bedState := mapState.MapBeds[obj.ID]
			bedState.OwnerID = charStateID
			mapState.MapBeds[obj.ID] = bedState
		}
	}

	if bedCount < len(mapGen.InhabitantCharacterDefs) || bedCount < len(mapGen.InhabitantCharacterGens) {
		// it looks like there weren't enough beds for the inhabitants set in the map generator...
		logz.Panicln("GenerateMap", "not enough beds in map for the inhabitants set in map generator. bed count:", bedCount, "# char defs:", len(mapGen.InhabitantCharacterDefs), "# char gens:", len(mapGen.InhabitantCharacterGens))
	}

	return mapState.ID
}

func getAllObjectLayers(layers []tiled.Layer) []tiled.Layer {
	objLayers := []tiled.Layer{}
	for _, layer := range layers {
		if layer.Type == tiled.LayerTypeGroup {
			objLayers = append(objLayers, getAllObjectLayers(layer.Layers)...)
			continue
		}
		if layer.Type == tiled.LayerTypeObject {
			objLayers = append(objLayers, layer)
		}
	}

	return objLayers
}
