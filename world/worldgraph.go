package world

import (
	"slices"

	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/tiled"
)

// MapEdge represents a door to a new map.
// "Doors" in our map system work like this:
//
// - there is a "door" object that points to the new map and spawn point in that map
// - there is a "spawn point" object right next to the "door", which represents the position where you would appear if entered the current map via that door/edge.
//
// So, every "door"/edge is supposed to have a corresponding spawn point.
type MapEdge struct {
	To           defs.MapID
	ToSpawn      int          // the spawn point this takes you to in the "to" map (where you arrive)
	EdgeCoords   model.Coords // the position where the character should walk before using the door/edge.
	EdgeObjectID int          // actual ID (from tiled) of the object that represents the door. this is for checking if the character can open the door (i.e. is it locked)
}

type MapNode struct {
	ID    defs.MapID
	Edges []MapEdge
}

// WorldGraph is a graph of all maps that comprise the entire game world. It is used for finding paths from one map to another.
//
// Future ideas:
//
// - when caching becomes important, we can start caching these routes between maps since they should be largely reusable.
// - if a route uses edges (doors) that have locks, we should track all the locks in a path, so that characters can know which paths they can or cannot take.
// - at some point, we might even need to check if the in-map path has things like locked gates, and track those locks too.
// - this also means that anytime a lock is changed (lock added or removed), these cached routes will need to be marked dirty and recalculated.
type WorldGraph struct {
	Nodes map[defs.MapID]*MapNode

	// a cache of all map data. mainly used for access to cost maps, so we can calculate local paths.
	// we purposely remove data like tile image data, since we don't use image data for path finding.
	MapDataCache map[defs.MapID]*tiled.Map
}

func BuildWorldGraph(dataman *datamanager.DataManager) WorldGraph {
	wg := WorldGraph{
		Nodes: make(map[defs.MapID]*MapNode),
	}

	for mapID := range dataman.MapDefs {
		node := MapNode{
			ID: mapID,
		}

		// load each map and find all "edges" ("doors" as we call the objects)
		m, exists := wg.MapDataCache[mapID]
		if !exists {
			m = tiled.LoadMap(mapID, false)
		}
		if m == nil {
			panic("failed to get map data... did LoadMap return nil? or did the MapDataCache?")
		}

		allObjs := []tiled.Object{}
		for _, l := range m.Layers {
			allObjs = append(allObjs, tiled.GetAllObjectsFromLayer(l)...)
		}

		// look for "edges" (door objects)
		for _, obj := range allObjs {
			objectInfo := m.GetObjectPropsAndTile(obj)
			objType, found := tiled.GetStringProperty("TYPE", objectInfo.AllProps)
			if !found {
				logz.Panicln("CreateNewMapState", "object didn't have a TYPE property:", obj.Name, obj.ID, "mapID:", mapID)
			}

			if objType != object.TypeDoor {
				continue
			}

			// found a door/edge; record where it goes
			doorTo, found := tiled.GetStringProperty(object.PropDoorTo, objectInfo.AllProps)
			if !found {
				panic("door object didn't have door_to prop!")
			}
			toSpawn, found := tiled.GetIntProperty(object.PropDoorSpawnIndex, objectInfo.AllProps)
			if !found {
				panic("door object didn't have spawn index prop!")
			}
			// TODO: these coordinates are probably a bit wrong; it'll point to the top left, but we would actually want to know the position right next to the door,
			// where a character would actually be standing when using the door.
			edgeCoords := model.ConvertPxToTilePos(int(obj.X), int(obj.Y))
			node.Edges = append(node.Edges, MapEdge{
				To:           defs.MapID(doorTo),
				ToSpawn:      toSpawn,
				EdgeCoords:   edgeCoords,
				EdgeObjectID: obj.ID,
			})
		}

		m.TileImageMap = nil // delete this stuff, since it could take up extra memory; we don't need tile images here.
		wg.MapDataCache[mapID] = m

		wg.Nodes[mapID] = &node
	}

	return wg
}

type PathStep struct {
	MapID    defs.MapID
	NextEdge *MapEdge // the edge this step heads to, if going to a new map
}

type WorldPath struct {
	FromMapID defs.MapID
	ToMapID   defs.MapID
	Path      []PathStep
}

func (wg *WorldGraph) FindPath(from, to defs.MapID) (pathToGoal WorldPath, foundPath bool) {
	if from == to {
		logz.Panicln("WorldGraph", "tried to find path to the same map:", from)
	}

	visited := map[defs.MapID]bool{}
	prev := map[defs.MapID]defs.MapID{}
	prevEdge := map[defs.MapID]MapEdge{}

	queue := []defs.MapID{from}
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		node := wg.Nodes[current]
		if node == nil {
			logz.Panicln("WorldGraph", "map node was nil!", current)
		}

		for _, edge := range node.Edges {
			next := edge.To

			if visited[next] {
				continue
			}

			visited[next] = true
			prev[next] = current
			prevEdge[next] = edge // also track which edges we are going through, for finding exact path later.

			if next == to {
				path := reconstructPath(prev, prevEdge, from, to)
				pathToGoal = WorldPath{
					FromMapID: from,
					ToMapID:   to,
					Path:      path,
				}
				return pathToGoal, true
			}

			queue = append(queue, next)
		}
	}

	return WorldPath{}, false
}

func reconstructPath(prev map[defs.MapID]defs.MapID, prevEdge map[defs.MapID]MapEdge, start, goal defs.MapID) []PathStep {
	path := []PathStep{}

	for current := goal; current != start; {
		parent := prev[current]
		parentEdge := prevEdge[current]
		step := PathStep{
			MapID:    parent,
			NextEdge: &parentEdge,
		}
		path = append(path, step)
		current = parent
	}

	slices.Reverse(path)

	return path
}
