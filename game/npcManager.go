package game

import (
	"time"

	"github.com/webbben/2d-game-engine/entity/npc"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
)

func (nm *NPCManager) getNPCByID(id string) *npc.NPC {
	for _, n := range nm.NPCs {
		if n.ID == id {
			return n
		}
	}
	return nil
}

// Finds an NPC at a given position, if one is to be found there. Second return value indicates if NPC successfully found.
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
func (nm *NPCManager) _asyncJobs() {
	defer func() {
		nm.backgroundJobsRunning = false
		logz.Println("NPC Manager", "stopping background jobs loop")
	}()

	logz.Println("NPC Manager", "starting background jobs loop")

	// ensure the loop doesn't repeat faster than this length of time
	maxLoopSpeed := time.Millisecond * 100

	for {
		start := time.Now()
		if !nm.RunBackgroundJobs {
			return
		}

		for _, n := range nm.NPCs {
			if !n.IsActive() {
				continue
			}

			// check if NPC's current task can use background assistance
			if n.CurrentTask != nil {
				n.CurrentTask.BackgroundAssist()
			} else {
				logz.Println(n.DisplayName, "NPC Manager: NPC has no current task")
			}
		}

		if time.Since(start) < maxLoopSpeed {
			time.Sleep(maxLoopSpeed - time.Since(start))
		}
	}
}
