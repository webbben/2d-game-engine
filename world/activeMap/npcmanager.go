package activemap

import (
	"time"

	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/world/npc"
)

// FindNPCAtPosition finds an NPC at a given position, if one is to be found there. Second return value indicates if NPC successfully found.
func (nm *NPCManager) FindNPCAtPosition(c model.Coords) (npc.NPC, bool) {
	for _, n := range nm.NPCs {
		if n.Entity.TilePos().Equals(c) {
			return *n, true
		}
		if n.Entity.TargetTilePos().Equals(c) {
			return *n, true
		}
	}
	return npc.NPC{}, false
}

func (nm *NPCManager) getNextNPCPriority() int {
	nextPriority := nm.nextPriority
	nm.nextPriority++
	return nextPriority
}

func (nm *NPCManager) startBackgroundNPCManager() {
	if !nm.RunBackgroundJobs {
		panic("NPC Manager: tried to start background jobs loop even though flag is set to false.")
	}
	if nm.backgroundJobsRunning {
		panic("NPC Manager: tried to start more than one background jobs loop!")
	}
	nm.backgroundJobsRunning = true
	go nm._asyncJobs()
}

// async jobs that the NPC Manager runs in a separate go-routine.
//
// DO NOT call this directly! Call StartBackgroundNPCManager instead!
// maxBgLoopSpeed bounds how fast the background jobs loop can repeat. Deliberately aligned with the
// pathfinding snapshot refresh throttle (see ActiveMap.refreshPathfindingSnapshot, 100ms), since bg
// assist pathfinding runs against that snapshot — faster passes would just recompute against an
// identical cost map. Without throttling, the loop would busy-spin and peg a core.
const maxBgLoopSpeed = time.Millisecond * 100

func (nm *NPCManager) _asyncJobs() {
	defer func() {
		nm.backgroundJobsRunning = false
		logz.Println("NPC Manager", "stopping background jobs loop")
	}()

	logz.Println("NPC Manager", "starting background jobs loop")

	for {
		start := time.Now()
		if !nm.RunBackgroundJobs {
			return
		}

		// doing this to avoid data race stuff
		nm.npcMu.RLock()
		npcs := make([]*npc.NPC, len(nm.NPCs))
		copy(npcs, nm.NPCs)
		nm.npcMu.RUnlock()

		for _, n := range npcs {
			task := n.GetCurrentTaskForBgAssist()
			if task == nil {
				continue
			}
			task.BackgroundAssist()
		}

		if time.Since(start) < maxBgLoopSpeed {
			time.Sleep(maxBgLoopSpeed - time.Since(start))
		}
	}
}
