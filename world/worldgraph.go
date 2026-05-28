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
	"github.com/webbben/2d-game-engine/worldgraph"
)

func (w *World) BuildWorldGraph() {
	debug.StartTimer("BuildWorldGraph")
	logz.Println("WORLD", "Building World Graph...")

	wg := worldgraph.WorldGraph{
		Nodes:        make(map[defs.MapID]*worldgraph.MapNode),
		MapDataCache: make(map[defs.MapID]*tiled.Map),
	}

	for mapID, def := range w.Dataman.MapDefs {
		if def.IsMapGenTemplate {
			// template maps should not be used in the world graph; these are used for generating maps, not directly as-is.
			continue
		}
		node := worldgraph.MapNode{
			ID:          mapID,
			SpawnPoints: make(map[int]model.Coords),
		}

		// load each map and find all "edges" ("doors" as we call the objects)
		m, exists := wg.MapDataCache[mapID]
		if exists {
			// not sure why this would ever happen... we aren't doing a search, we are just going through the list of map IDs, so no two IDs should ever be the same, right?
			// I guess we can panic here to ensure there are no duplicates
			logz.Panicln("WORLD", "BuildWorldGraph found map data was already cached for a mapID, but this shouldn't happen unless there are two maps with the same ID... mapID:", mapID)
		}
		if !exists {
			m = tiled.LoadMap(mapID, false)
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
			if obj.Ellipse {
				// ellipse objects are just used for planning
				continue
			}

			objectInfo := m.GetObjectPropsAndTile(obj)
			objType, found := object.GetObjectType(objectInfo.AllProps)
			if !found {
				logz.Panicln("CreateNewMapState", "object didn't have a TYPE property:", obj.Name, obj.ID, "mapID:", mapID)
			}

			if objType == object.TypeSpawnPoint {
				// record spawn point location
				spawnID, found := tiled.GetIntProperty(object.PropSpawnIndex, obj.Properties)
				if !found {
					logz.Panicln("BuildWorldGraph", "Tried to get spawn index of spawn point, but property wasn't found. mapID:", mapID, "objID:", obj.ID)
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

			// found a door/edge; record where it goes
			var doorTo defs.MapID
			doorToProp, found := tiled.GetStringProperty(object.PropDoorTo, objectInfo.AllProps)
			if found {
				doorTo = defs.MapID(doorToProp)
			} else {
				mapGenID, found := tiled.GetStringProperty(object.PropDoorMapGeneratorID, objectInfo.AllProps)
				if !found {
					panic("door object has neither a door_to prop nor a mapGeneratorID")
				}
				logz.Println("WorldGraph", "map generator found:", mapGenID)

				returnSpawn, found := tiled.GetIntProperty("return_spawn_index", objectInfo.AllProps)
				if !found {
					logz.Panicln("WorldGraph", "found map generator, but the door didn't include the return_spawn_index prop.", mapID, obj.ID)
				}

				doorTo = w.GenerateMap(mapGenID, mapID, returnSpawn)

				// set the door overrides for this map, so this door can get to the generated map
				mapState := w.Dataman.GetMapState(mapID)
				mapState.DoorOverrides[obj.ID] = state.DoorState{
					OverrideDestinationMap: doorTo,
					// NOTE: we don't set the override spawn, since the regular spawn point prop should be correct.
				}
			}

			toSpawn, found := tiled.GetIntProperty(object.PropDoorSpawnIndex, objectInfo.AllProps)
			if !found {
				panic("door object didn't have spawn index prop!")
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
			logz.Panicln("BuildWorldGraph", "map doesn't have spawn point index 0! this is required for all maps. mapID:", mapID)
		}

		m.TileImageMap = nil // delete this stuff, since it could take up extra memory; we don't need tile images here.
		wg.MapDataCache[mapID] = m

		wg.Nodes[mapID] = &node
	}

	debug.StopTimer("BuildWorldGraph")

	w.WorldGraph = &wg
}
