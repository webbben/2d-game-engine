// Package worldgraph contains a graph representation of the world and its maps, and handles finding routes between maps
package worldgraph

import (
	"fmt"
	"slices"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/internal/debug"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/utils"
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
	ID          defs.MapID
	Edges       []MapEdge
	SpawnPoints map[int]model.Coords
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

func BuildWorldGraph(dataman *datamanager.DataManager) *WorldGraph {
	debug.StartTimer("BuildWorldGraph")
	logz.Println("WORLD", "Building World Graph...")

	wg := WorldGraph{
		Nodes:        make(map[defs.MapID]*MapNode),
		MapDataCache: make(map[defs.MapID]*tiled.Map),
	}

	for mapID := range dataman.MapDefs {
		node := MapNode{
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

			node.Edges = append(node.Edges, MapEdge{
				To:           defs.MapID(doorTo),
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
	return &wg
}

// A PathStep represents one map's segment of travel in a WorldPath. Each PathStep assumes you spawn in at a certain point,
// and are heading towards another edge. The first step will not have a spawn point (since the character is, of course, already in that map)
// and the last step will not have an edge (since the character will have arrived at the destination map).
type PathStep struct {
	SpawnCoords model.Coords // the coords of the spawn point where the character arrives in this map/path step
	MapID       defs.MapID
	NextEdge    *MapEdge // the edge this step heads to, if going to a new map
	MapPath     []model.Coords
}

func (ps PathStep) String() string {
	s := fmt.Sprintf("mapID: %s\n", ps.MapID)
	s += fmt.Sprintf("spawnCoords: %s\n", ps.SpawnCoords)
	s += fmt.Sprintf("nextEdge: %v\n", ps.NextEdge)
	s += fmt.Sprintf("mapPath: %s", ps.MapPath)
	return s
}

type WorldPath struct {
	FromMapID defs.MapID
	ToMapID   defs.MapID
	Path      []PathStep
}

func (wp WorldPath) Validate() {
	if wp.FromMapID == "" {
		panic("from map is empty")
	}
	if wp.ToMapID == "" {
		panic("to map is empty")
	}
	if len(wp.Path) == 0 {
		panic("Path is empty")
	}
	if wp.Path[0].MapID != wp.FromMapID {
		panic("first path mapID doesn't match FromMapID")
	}
	if wp.Path[len(wp.Path)-1].NextEdge.To != wp.ToMapID {
		panic("last path segment's edge to map ID doesn't match ToMapID")
	}

	for i, ps := range wp.Path {
		if ps.MapID == "" {
			panic("mapID was empty")
		}
		if ps.NextEdge == nil {
			panic("next edge was nil")
		}
		if i == 0 {
			// first step
			if !ps.SpawnCoords.Equals(model.Coords{X: 0, Y: 0}) {
				panic("first path step had spawn coords defined, but it's not supposed to")
			}
			if len(ps.MapPath) > 0 {
				panic("first step had a map path defined, but that shouldn't happen since findPath doesn't know the starting coordinates")
			}
		} else {
			if ps.SpawnCoords.Equals(model.Coords{X: 0, Y: 0}) {
				panic("middle step didn't have a spawn defined")
			}
			if len(ps.MapPath) == 0 {
				panic("map path was empty")
			}
		}
	}
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
				path := wg.reconstructPath(prev, prevEdge, from, to)
				pathToGoal = WorldPath{
					FromMapID: from,
					ToMapID:   to,
					Path:      path,
				}
				pathToGoal.Validate()
				return pathToGoal, true
			}

			queue = append(queue, next)
		}
	}

	return WorldPath{}, false
}

// prev: gets you the mapID that came before this mapID
// prevEdge: gets you the mapEdge that sent you to this mapID
func (wg *WorldGraph) reconstructPath(prev map[defs.MapID]defs.MapID, prevEdge map[defs.MapID]MapEdge, start, goal defs.MapID) []PathStep {
	path := []PathStep{}

	for current := goal; current != start; {
		// the step here is the step that leads to the CURRENT map.
		// So, therefore this step has the MapID of parent, and next edge of parentEdge
		// we are starting this loop from the goal mapID, and on each loop we are asking:
		// "what map did we come from, and what edge did that map use to get here?"
		// This means that goal (mapID) doesn't have a step.
		parent := prev[current]
		parentEdge := prevEdge[current]
		step := PathStep{
			MapID:    parent,
			NextEdge: &parentEdge,
		}

		// we can look back at the previous reconstructed step in path and fill in the spawn point for that step.
		// this allows us to calculate the actual in-map route that the NPC should take in this path step.
		if len(path) > 0 {
			lastStep := path[len(path)-1]
			lastSpawnPoint := parentEdge.ToSpawn
			lastSpawnCoords := wg.Nodes[lastStep.MapID].SpawnPoints[lastSpawnPoint]
			path[len(path)-1].SpawnCoords = lastSpawnCoords
			mapData := wg.MapDataCache[lastStep.MapID]
			if mapData == nil {
				panic("map data was nil!")
			}
			// Note: this cost map does NOT factor in things like objects or other entities.
			// doesn't really matter for paths that are used as estimates of in-map routes and RouteTask progress, but
			// once loading into the actual map, we will probably need to recalculate the route rather than using this same one.
			// that's because, it seems very possible that objects will be obstacles that influence a route, so we wouldn't want to assume this route is
			// guaranteed to be free of obstacles.
			costmap := mapData.CostMap
			if len(costmap) == 0 {
				panic("costmap appears to not have been calculated yet. that should've happened in the Map.load function")
			}
			inMapPath, found := path_finding.FindPath(lastSpawnCoords, lastStep.NextEdge.EdgeCoords, costmap)
			if !found {
				if inMapPath == nil {
					logz.Panicln("WorldGraph", "failed to find path between spawn and edge of path step; complete failure! maybe the spawn point is blocked?")
				}
				// for some reason, path finding failed...
				// TODO: for now, we won't panic since it seems possible some doors may not be "reachable", in that they are directly on top of objects
				// that are collidable. I don't know if the pathfinding function tries to walk directly onto the goal tile, but just in case we won't panic for now.
				// If it turns out that this shouldn't be a problem, then we can go ahead and make this a panic case.
				// Until then, I'm going to add a check to make sure that the path at least got "close enough".
				lastPathPos := inMapPath[len(inMapPath)-1]
				dist := utils.EuclideanDistCoords(lastPathPos, lastStep.NextEdge.EdgeCoords)
				if dist > 2 {
					logz.Println("WorldGraph", "start:", lastSpawnCoords, "goal:", lastStep.NextEdge.EdgeCoords)
					logz.Println("WorldGraph", "last path pos:", lastPathPos, "dist from goal:", dist)
					logz.Panicln("WorldGraph", "failed to find path between spawn point and edge of path step; last step of path didn't get close enough to goal (dist > 2)")
				}
				// if the incomplete path is close enough, then let's just use it
			}
			path[len(path)-1].MapPath = inMapPath
		}

		path = append(path, step)
		current = parent
	}

	slices.Reverse(path)

	return path
}
