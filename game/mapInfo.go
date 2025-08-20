package game

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/internal/tiled"
	"github.com/webbben/2d-game-engine/object"
)

// information about the current room the player is in
type MapInfo struct {
	Map      tiled.Map
	Entities []*entity.Entity         // the entities in the current room
	Objects  []object.Object          // the objects in the current room
	ImageMap map[string]*ebiten.Image // the map of images (tiles) used in rendering the current room
}

func (mi *MapInfo) Preprocess() {
}

func (mi MapInfo) Collides(c model.Coords) bool {
	// check map's CostMap
	log.Println(c)
	maxY := len(mi.Map.CostMap)
	maxX := len(mi.Map.CostMap[0])
	if c.Y == maxY || c.X == maxX || c.X == -1 || c.Y == -1 {
		slog.Info(fmt.Sprintf("map boundaries: X = [%v, %v], Y = [%v, %v]", 0, maxX, 0, maxY))
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
