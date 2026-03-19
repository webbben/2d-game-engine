package activemap

import (
	"math/rand"
	"slices"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/internal/path_finding"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/world/npc"
)

// GetValidMapPosition finds a valid position in the active map for a given NPC.
// Basically, this can be used for placing an NPC in a random place, and ensures that the NPC doesn't spawn in inaccessible areas,
// behind locked gates (that they can't unlock), etc.
func (m ActiveMap) GetValidMapPosition(n npc.NPC) model.Coords {
	// get the cost map (which includes NPC and object collisions), and then factor in other stuff like locked gates.
	costmap := m.CostMap()

	lockIDs := characterstate.GetLockIDs(*n.CharacterStateRef)

	// TODO: at some point, if we have sufficiently big maps, should we make a separate slice for gate objects?
	// that way we aren't always looking through possibly hundreds of random objects, light objects, etc to find gates to check.
	for _, obj := range m.Objects {
		switch obj.Type {
		case object.TypeGate:
			// if the gate is locked and the NPC doesn't have the key, then mark area as a collision
			lockID := obj.GetLockID()
			if lockID != "" && !slices.Contains(lockIDs, lockID) {
				// gate is locked and NPC doesn't have the right key for it. add a collision where this gate stands.
				for _, c := range obj.CollisionRect.GetOverlappingTiles() {
					costmap[c.Y][c.X] += path_finding.BlockThreshold
				}
			}
		case object.TypeSpawnPoint:
			// also, block spawn points since we wouldn't want NPCs to spawn on top of those.
			x, y := obj.Pos()
			c := model.ConvertPxToTilePos(x, y)
			costmap[c.Y][c.X] += path_finding.BlockThreshold
			// TODO: should we block spots that are immediately adjacent to a spawn point too? don't think we want NPCs sitting right in front of the door.
		}
	}

	// use a starting position of the main spawn point in the map (index=0)
	x, y, found := m.GetSpawnPosition(0)
	if !found {
		logz.Panicln("GetValidMapPosition", "no spawn point (index=0) found in map. mapID:", m.MapID)
	}
	startPos := model.ConvertPxToTilePos(x, y)

	reachable := path_finding.GetAllReachablePositions(startPos, costmap)
	if len(reachable) == 0 {
		// TODO: it seems theoretically possible that too many NPCs could be in a single room, in which case we probably need some logic to try to prevent that from happening,
		// but also handle that situation if it occurred. We probably would just need to keep some NPCs from entering the room if that were the case, but hopefully we design maps
		// and NPC schedules well enough that this could never happen.
		// One theoretically possible one that comes to mind, though, is if there were a task like "go to tavern" in a big city. What if 50 NPCs all try to go to the same tavern?
		// I guess, the logic for choosing a tavern would need to figure out what the maximum capacity of a given map is, and skip maps that are already too crowded.
		logz.Panicln("GetValidMapPosition", "no reachable positions found! mapID:", m.MapID, "npcID:", n.ID())
	}

	i := rand.Intn(len(reachable))
	c := reachable[i]
	if c.Equals(startPos) {
		panic("chosen reachable position was the same as the given startPos! that's not supposed to be possible.")
	}
	return c
}

// FindObjectsAtPosition finds all objects that intersect with a given tile position.
// This includes collidable and non-collidable objects, as long as they have a draw rect.
func (mi ActiveMap) FindObjectsAtPosition(c model.Coords) []*object.Object {
	posRect := model.NewRect(float64(c.X)*config.TileSize, float64(c.Y)*config.TileSize, config.TileSize, config.TileSize)
	objs := []*object.Object{}
	for _, obj := range mi.Objects {
		if obj.GetRect().Intersects(posRect) {
			objs = append(objs, obj)
		}
	}
	return objs
}

func (mi *ActiveMap) StartTradeSession(shopkeeperID defs.ShopID) {
	mi.gameCtx.StartTradeSession(shopkeeperID)
}

func (mi *ActiveMap) StartDialog(dialogProfileID defs.DialogProfileID, npcID string) {
	mi.gameCtx.StartDialogSession(dialogProfileID, npcID)
}

func (m ActiveMap) GetAllObjects() []*object.Object {
	return m.Objects
}
