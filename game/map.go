package game

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/tiled"
)

// EnsureMapStateExists checks if a map state exists, and instantiates it if it doesn't.
// So, anytime you're dealing with a map, just run this function to make sure its state exists and/or has been instantiated correctly.
func (g *Game) EnsureMapStateExists(mapID defs.MapID) {
	// make sure def exists for this mapID
	_ = g.Dataman.GetMapDef(mapID)

	if g.Dataman.MapStateExists(mapID) {
		return
	}

	g.CreateNewMapState(mapID)
}

func (g *Game) CreateNewMapState(mapID defs.MapID) {
	if g.Dataman.MapStateExists(mapID) {
		logz.Panicln("CreateNewMapState", "tried to create a new map state, but one already exists:", mapID)
	}

	mapState := state.MapState{
		ID:       mapID,
		MapLocks: make(map[string]state.LockState),
		MapBeds:  make(map[int]state.BedState),
	}

	// look through Tiled map data to initialize state for things like pre-defined item placements, door locks, etc.
	m := tiled.LoadMap(mapID, false)

	objLayers := getAllObjectLayers(m.Layers)

	// Get initial items
	for _, layer := range objLayers {
		for _, obj := range layer.Objects {
			objectInfo := m.GetObjectPropsAndTile(obj)
			objType, found := tiled.GetStringProperty("TYPE", objectInfo.AllProps)
			if !found {
				logz.Panicln("CreateNewMapState", "object didn't have a TYPE property:", obj.Name, obj.ID, "mapID:", mapID)
			}

			// check if there is a lock on this object
			var lockLevel int
			var lockID string
			lockLevel, found = tiled.GetIntProperty("lock_level", objectInfo.AllProps)
			if found {
				if lockLevel <= 0 {
					logz.Panicln("CreateNewMapState", "found lock level property on object, but it had a level of <= 0.", obj.Name, obj.ID, "mapID:", mapID)
				}
				lockID, found = tiled.GetStringProperty("lock_id", objectInfo.AllProps)
				if !found {
					logz.Panicln("CreateNewMapState", "found lock level property on object, but no lock ID.", obj.Name, obj.ID, "mapID:", mapID)
				}
				if lockID == "" {
					logz.Panicln("CreateNewMapState", "lock ID property was empty:", obj.Name, obj.ID, "mapID:", mapID)
				}
				// add it to the lock map
				mapState.MapLocks[lockID] = state.LockState{
					LockID:    lockID,
					LockLevel: lockLevel,
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
				_ = g.Dataman.GetItemDef(defID)
				mapState.MapItems = append(mapState.MapItems, state.MapItemState{
					ItemInstance: defs.ItemInstance{DefID: defID},
					Quantity:     1,
					X:            obj.X,
					Y:            obj.Y,
				})
			case object.TypeBed:
				if lockID != "" {
					logz.Panicln("CreateNewMapState", "a lock was put on a bed, which doesn't make any sense.", obj.Name, obj.ID, "mapID:", mapID)
				}
				// instantiate the NPC that is associated with this bed
				var charStateID state.CharacterStateID
				charGenID, found := tiled.GetStringProperty("characterGeneratorID", objectInfo.AllProps)
				if found {
					charGen := g.Dataman.GetCharacterGenerator(charGenID)
					charStateID = g.World.GenerateCharacter(charGen, mapID, mapID, obj.ID)
				} else {
					charDefID, found := tiled.GetStringProperty("characterDefID", objectInfo.AllProps)
					if found {
						// this bed has a specific character def ID set
						params := entity.NewCharacterStateParams{
							InitialMapID: mapID,
							HomeMapID:    mapID,
							HomeMapBedID: obj.ID,
						}
						charStateID = entity.CreateNewCharacterState(defs.CharacterDefID(charDefID), params, g.Dataman)
					}
				}

				// if a character was created, register this bed in the character state
				if charStateID != "" {
					charState := g.Dataman.GetCharacterState(charStateID)
					charState.HomeMapID = mapID
					charState.HomeMapBedID = obj.ID
				}

				// add this bed to the MapBeds
				mapState.MapBeds[obj.ID] = state.BedState{
					MapObjID:            obj.ID,
					ResidentCharStateID: charStateID,
				}
			}
		}
	}

	g.Dataman.LoadMapState(mapState)
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
