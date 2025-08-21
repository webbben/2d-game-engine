package game

import (
	"fmt"
	"log/slog"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/npc"
	"github.com/webbben/2d-game-engine/player"
)

// information about the current room the player is in
type MapInfo struct {
	Map      tiled.Map
	Entities []*entity.Entity         // the entities in the map
	NPCs     []*npc.NPC               // the NPC entities in the map
	ImageMap map[string]*ebiten.Image // the map of images (tiles) used in rendering the current room
}

func (mi *MapInfo) Preprocess() {
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
	for _, ent := range mi.Entities {
		if ent.TilePos.Equals(c) {
			return true
		}
	}

	return false
}

func (mi MapInfo) FindPath(start, goal model.Coords) []model.Coords {
	// TODO factor in entity positions
	return path_finding.FindPath(start, goal, mi.Map.CostMap)
}

func (mi MapInfo) MapDimensions() (width int, height int) {
	return mi.Map.Width, mi.Map.Height
}
