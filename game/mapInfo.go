package game

import (
	"fmt"
	"log/slog"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/npc"
	"github.com/webbben/2d-game-engine/player"
)

// information about the current room the player is in
type MapInfo struct {
	Map      tiled.Map
	ImageMap map[string]*ebiten.Image // the map of images (tiles) used in rendering the current room

	NPCManager
}

// Do all required setup for creating MapInfo
func InitializeMap(mapInfo MapInfo) MapInfo {
	mi := mapInfo
	mi.NPCManager.mapRef = mi.Map
	return mi
}

type NPCManager struct {
	NPCs []*npc.NPC // the NPC entities in the map

	mapRef       tiled.Map // map info so we can get map size, tile adjacency, etc
	nextPriority int       // the next priority value to assign to an NPC

	StuckNPCs []string // IDs of NPCs who are currently stuck (while trying to execute a task)
}

func (mi *MapInfo) ResolveNPCJams() {
	// get all existing NPC jams
	npcJams := mi.NPCManager.findNPCJams()
	if len(npcJams) == 0 {
		// no jams found
		return
	}

	// resolve each jam, one at a time
	for _, npcJam := range npcJams {
		mi.NPCManager.resolveJam(npcJam)
	}
}

// The official way to add the player to a map
func (mi *MapInfo) AddPlayerToMap(p *player.Player, startPos model.Coords) {
	if mi.Collides(startPos) {
		// this also handles placement outside of map bounds
		panic("player added to map on colliding tile")
	}
	p.Entity.World = mi
	p.Entity.SetPosition(startPos)
}

func (mi *MapInfo) AddNPCToMap(n *npc.NPC, startPos model.Coords) {
	if mi.Collides(startPos) {
		panic("npc added to map on colliding tile")
	}
	n.Entity.World = mi
	n.Entity.SetPosition(startPos)
	n.Priority = mi.NPCManager.getNextNPCPriority()
	mi.NPCs = append(mi.NPCs, n)
}

func (mi MapInfo) Collides(c model.Coords) bool {
	// check map's CostMap
	maxY := len(mi.Map.CostMap)
	maxX := len(mi.Map.CostMap[0])
	if c.Y == maxY || c.X == maxX || c.X == -1 || c.Y == -1 {
		// attempting to move past the edge of the map
		return true
	}
	if c.Y > maxY || c.X > maxX || c.X < -1 || c.Y < -1 {
		slog.Error(fmt.Sprintf("map boundaries: X = [%v, %v], Y = [%v, %v]", 0, maxX, 0, maxY))
		panic("mapInfo.Collides given a value that is far beyond map boundaries; if an entity is trying to move here, something must have gone wrong")
	}
	if mi.Map.CostMap[c.Y][c.X] >= 10 {
		return true
	}

	// check entity positions
	for _, n := range mi.NPCs {
		if n.Entity.TilePos.Equals(c) {
			return true
		}
		// include tile entity is moving into
		if n.Entity.Movement.TargetTile.Equals(c) {
			return true
		}
	}

	return false
}

func (mi MapInfo) FindPath(start, goal model.Coords) []model.Coords {
	return path_finding.FindPath(start, goal, mi.CostMap())
}

func (mi MapInfo) MapDimensions() (width int, height int) {
	return mi.Map.Width, mi.Map.Height
}

// Gets a cost map that includes entity positions
func (mi MapInfo) CostMap() [][]int {
	if mi.Map.CostMap == nil {
		panic("tried to get MapInfo cost map before Map costmap was created")
	}
	// make deep copy so that original cost map isn't altered
	costMap := make([][]int, len(mi.Map.CostMap))
	for i := range mi.Map.CostMap {
		costMap[i] = append([]int{}, mi.Map.CostMap[i]...)
	}

	for _, n := range mi.NPCs {
		costMap[n.Entity.TilePos.Y][n.Entity.TilePos.X] += 10
		// if the entity is currently moving, mark its destination tile as a collision too
		if !n.Entity.Movement.TargetTile.Equals(n.Entity.TilePos) {
			costMap[n.Entity.Movement.TargetTile.Y][n.Entity.Movement.TargetTile.X] += 10
		}
	}
	// TODO add player position?
	return costMap
}
