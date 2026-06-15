package world

import (
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/internal/debug"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/utils"
	"github.com/webbben/2d-game-engine/worldgraph"
)

func (w *World) BuildWorldGraph() {
	debug.StartTimer("BuildWorldGraph")
	logz.Println("WORLD", "Building World Graph...")

	wg := worldgraph.WorldGraph{
		Nodes:        make(map[defs.MapID]*worldgraph.MapNode),
		MapDataCache: make(map[defs.MapID]*tiled.Map),
	}

	for mapStateID := range w.Dataman.MapStates {
		genMapStateIDs := w.buildGraphNode(&wg, mapStateID)

		for _, genMapStateID := range genMapStateIDs {
			moreGenIDs := w.buildGraphNode(&wg, genMapStateID)
			if len(moreGenIDs) > 0 {
				logz.Panicln("BuildWorldGraph", "somehow, a generated map produced more generated maps within it... that's not supposed to be allowed.", mapStateID, genMapStateID, moreGenIDs)
			}
		}
	}

	debug.StopTimer("BuildWorldGraph")

	w.WorldGraph = &wg
}

func (w *World) buildGraphNode(wg *worldgraph.WorldGraph, mapStateID defs.MapID) (discoveredGenMaps []defs.MapID) {
	if _, exists := wg.Nodes[mapStateID]; exists {
		// it seems this map already exists... weird
		// My friend "Big Pickle" says that this could theoretically happen, since we are adding to map states with GenerateMap. So, let's be sure we don't
		// accidentally process the same map twice.
		logz.Warnln("buildGraphNode", "a map node for this mapStateID already exists:", mapStateID)
		return []defs.MapID{}
	}

	mapState := w.Dataman.GetMapState(mapStateID)

	// get actual map def ID, in case this is a generated map from a template
	mapDefID := mapStateID
	if mapState.IsGenerated {
		mapDefID = mapState.GeneratedMapDefID
	}
	mapDef := w.Dataman.GetMapDef(mapDefID)

	node := worldgraph.MapNode{
		ID:          mapStateID,
		SpawnPoints: make(map[int]model.Coords),
		Type:        mapDef.Type,
	}

	// load each map and find all "edges" ("doors" as we call the objects)
	// we cache by mapDefID (not a generated map state ID) since generated maps from the same template would use the same actual map data (same source .tmj file)
	m, exists := wg.MapDataCache[mapDefID]
	// Note: we used to have a condition that would panic if exists=true, because we thought "why would that happen at this stage?"
	// But, with generated maps, it seems possible that multiple maps would indeed use the same map data (same source .tmj file)
	if !exists {
		m = tiled.LoadMap(mapDefID, false)
	}
	if m == nil {
		panic("failed to get map data... did LoadMap return nil?")
	}

	allObjs := []tiled.Object{}
	for _, l := range m.Layers {
		allObjs = append(allObjs, tiled.GetAllObjectsFromLayer(l)...)
	}

	// a little validation; make sure spawn point 0 is found
	foundSpawn0 := false

	// look for "edges" (door objects)
	for _, obj := range allObjs {
		if obj.Ellipse || obj.Text != nil {
			// ellipse objects are just used for planning
			continue
		}

		objectInfo := m.GetObjectPropsAndTile(obj)
		objType, found := object.GetObjectType(objectInfo.AllProps)
		if !found {
			logz.Panicln("buildGraphNode", "object didn't have a TYPE property:", obj.Name, obj.ID, "mapID:", mapDefID)
		}

		if objType == object.TypeSpawnPoint {
			// record spawn point location
			spawnID, found := tiled.GetIntProperty(object.PropSpawnIndex, obj.Properties)
			if !found {
				logz.Panicln("BuildWorldGraph", "Tried to get spawn index of spawn point, but property wasn't found. mapID:", mapDefID, "objID:", obj.ID)
			}
			if spawnID == 0 {
				foundSpawn0 = true
			}
			spawnCoords := model.ConvertPxToTilePos((obj.X), (obj.Y))
			node.SpawnPoints[spawnID] = spawnCoords
			continue
		}

		if objType != object.TypeDoor {
			continue
		}

		var doorTo defs.MapID
		var toSpawn int

		// found a door/edge; record where it goes
		// for generated maps, this info should already be set in a door override, so check for that and skip the rest of the door logic.
		if mapState.IsGenerated {
			// any door in a generated map must have a door override, because generated maps aren't allowed to have explicit door_to properties and such.
			override, exists := mapState.DoorOverrides[obj.ID]
			if !exists {
				logz.Panicln("buildGraphNode", "door in generated map didn't have an override set; generated maps must only have a single door to the map that routes back to the original place that generated it.", mapStateID, obj.ID)
			}
			doorTo = override.OverrideDestinationMap
			if override.OverrideDestinationSpawn == nil {
				logz.Panicln("buildGraphNode", "overrideDestinationSpawn was unexpectedly nil.", mapStateID)
			}
			toSpawn = *override.OverrideDestinationSpawn
		} else {
			// for non-generated maps, we expect to find doorTo and spawn props on door objects
			doorToProp, found := tiled.GetStringProperty(object.PropDoorTo, objectInfo.AllProps)
			if found {
				// door to a non-generated map; so, we expect to find a doorTo and toSpawn in props
				doorTo = defs.MapID(doorToProp)
				toSpawn, found = tiled.GetIntProperty(object.PropDoorSpawnIndex, objectInfo.AllProps)
				if !found {
					logz.Panicln("buildGraphNode", "door object didn't have a spawn index prop.", obj.ID, mapStateID)
				}
			} else {
				mapGenID, found := tiled.GetStringProperty(object.PropDoorMapGeneratorID, objectInfo.AllProps)
				if !found {
					panic("door object has neither a door_to prop nor a mapGeneratorID")
				}
				logz.Println("WorldGraph", "map generator found:", mapGenID)

				returnSpawn, found := tiled.GetIntProperty("return_spawn_index", objectInfo.AllProps)
				if !found {
					logz.Panicln("WorldGraph", "found map generator, but the door didn't include the return_spawn_index prop.", mapDefID, obj.ID)
				}

				// first, check if an override already exists for this door in the map state.
				// this will be true when loading the game.
				if existingOverride, exists := mapState.DoorOverrides[obj.ID]; exists {
					doorTo = existingOverride.OverrideDestinationMap
					if existingOverride.OverrideDestinationSpawn == nil {
						logz.Panicln("buildGraphNode", "overrideDestinationSpawn was nil, but it never should be.", obj.ID, mapStateID)
					}
					toSpawn = *existingOverride.OverrideDestinationSpawn
				} else {
					doorTo = w.GenerateMap(mapGenID, mapStateID, returnSpawn)
					// set the door overrides for this map, so this door can get to the generated map
					mapState.DoorOverrides[obj.ID] = state.DoorState{
						OverrideDestinationMap:   doorTo,
						OverrideDestinationSpawn: utils.Int(0),
					}
					// Add this generated map state ID to the slice of discovered gen maps
					discoveredGenMaps = append(discoveredGenMaps, doorTo)
				}
			}
		}

		// TODO: these coordinates are probably a bit wrong; it'll point to the top left, but we would actually want to know the position right next to the door,
		// where a character would actually be standing when using the door.
		x := obj.X
		y := obj.Y
		if obj.Height > config.TileSize {
			// doors are usually 2 tiles tall, so make sure we have the bottom tile
			y = (y + obj.Height) - config.TileSize
		}
		edgeCoords := model.ConvertPxToTilePos(x, y)
		// TODO: should we go through the trouble here of actually finding the "right" position to go to in order to access the door?
		// lots of doors will be on buildings, and therefore their positions will be blocked.
		// I also wonder if path finding when the goal tile is blocked takes a performance hit at all, since it probably leads to excess searching (even if a little)

		node.Edges = append(node.Edges, worldgraph.MapEdge{
			To:           doorTo,
			ToSpawn:      toSpawn,
			EdgeCoords:   edgeCoords,
			EdgeObjectID: obj.ID,
		})
	}

	if !foundSpawn0 {
		logz.Panicln("BuildWorldGraph", "map doesn't have spawn point index 0! this is required for all maps. mapID:", mapDefID)
	}

	wg.MapDataCache[mapDefID] = m
	// reconstructPath would access the cache by mapStateID, so we can cache it here too. since the data is just pointers, it shouldn't be a memory issue.
	wg.MapDataCache[mapStateID] = m

	wg.Nodes[mapStateID] = &node

	return discoveredGenMaps
}
