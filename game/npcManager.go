package game

import (
	"fmt"
	"slices"
	"time"

	"github.com/webbben/2d-game-engine/entity/npc"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
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
		if n.Entity.TilePos.Equals(c) {
			return *n, true
		}
		if n.Entity.Movement.TargetTile.Equals(c) {
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

// Report an NPC that is stuck and cannot recover on its own
func (nm *NPCManager) ReportStuckNPC(npcID string) {
	if slices.Contains(nm.StuckNPCs, npcID) {
		return
	}
	nm.StuckNPCs = append(nm.StuckNPCs, npcID)
}

// Report that a previously stuck NPC is now recovered
func (nm *NPCManager) RecoverStuckNPC(npcID string) {
	i := slices.Index(nm.StuckNPCs, npcID)
	if i == -1 {
		logz.Warnln("", "tried to recover stuck NPC ID that was not present in stuck NPC list:", npcID)
		return
	}
	nm.StuckNPCs = general_util.RemoveIndexUnordered(nm.StuckNPCs, i)
}

type StuckNPC struct {
	npcID     string
	direction byte
	position  model.Coords
}

func (nm *NPCManager) resolveJam(npcJam []StuckNPC) {
	// strategy to resolve Jam:
	//
	// 1. identify majority direction
}

func (nm *NPCManager) findNPCJams() [][]StuckNPC {
	// map out geographic locations of stuck NPCs
	jamMap := make([][]StuckNPC, nm.mapRef.Height)
	for i := 0; i < nm.mapRef.Height; i++ {
		jamMap[i] = make([]StuckNPC, nm.mapRef.Width)
	}
	for _, npcID := range nm.StuckNPCs {
		n := nm.getNPCByID(npcID)
		if n == nil {
			logz.Errorln("Stuck NPCs", "stuck NPC ID not found in MapInfo NPC list:", npcID, "(should this have been cleaned up elsewhere?)")
			continue
		}
		x := n.Entity.TilePos.X
		y := n.Entity.TilePos.Y
		stuckNPC := StuckNPC{
			npcID:     npcID,
			direction: n.Entity.Movement.Direction,
			position:  model.Coords{X: x, Y: y},
		}
		jamMap[y][x] = stuckNPC
	}

	// identify clusters of adjacent stuck NPCs
	// for each stuck NPC:
	// - find all adjacent stuck NPCs
	//   - for each adjacent stuck NPC:
	//     - if NPC is already in a cluster, add to same cluster
	// - if no existing cluster is found (among adjacent NPCs), create a new one and put this NPC in it

	// maps "cluster ID" ("x/y" of first member) to array of StuckNPC
	clusters := make(map[string][]StuckNPC)
	// maps NPC ID to "cluster ID", which is "x/y" of first member
	clusterMembership := make(map[string]string)
	for y, row := range jamMap {
		for x, stuckNPC := range row {
			if stuckNPC.npcID == "" {
				// empty
				continue
			}

			// check for adjacent stuck NPCs
			clusterFound := false
			for _, c := range nm.mapRef.GetAdjTiles(stuckNPC.position) {
				n := jamMap[c.Y][c.X]
				if n.npcID != "" {
					clusterName, exists := clusterMembership[n.npcID]
					if !exists {
						continue
					}
					// existing cluster found - add to this existing cluster
					clusterMembership[stuckNPC.npcID] = clusterName
					clusters[clusterName] = append(clusters[clusterName], stuckNPC)
					clusterFound = true
					break
				}
			}

			if !clusterFound {
				// create a new cluster, since no existing adjacent one was found
				clusterName := fmt.Sprintf("%v/%v", x, y)
				clusterMembership[stuckNPC.npcID] = clusterName
				clusters[clusterName] = []StuckNPC{stuckNPC}
			}
		}
	}

	// extract all clusters into an array of clusters
	jams := make([][]StuckNPC, 0)
	for _, stuckNPCs := range clusters {
		jams = append(jams, stuckNPCs)
	}
	return jams
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
