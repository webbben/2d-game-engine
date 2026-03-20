package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/object"
)

type SleepTask struct {
	TaskBase
	activateObjectTask *ActivateObjectTask
	bedObj             *object.Object
	inBed              bool
}

func NewSleepTask(n *NPC, p defs.TaskPriority) *SleepTask {
	t := &SleepTask{
		TaskBase: NewTaskBase(
			defs.TaskDef{
				TaskID:   TaskSleep,
				Priority: p,
			},
			"Sleep",
			"NPC sleeps in his bed",
			n,
		),
	}

	return t
}

func (t *SleepTask) Update() {
	switch t.Status {
	case TaskNotStarted:
		// find bed and start going to it
		homeMap := t.Owner.CharacterStateRef.HomeMapID
		if homeMap != t.Owner.CharacterStateRef.CurrentMap {
			// if we aren't in the home map right now, then this NPC should route to his home map.
			// TODO: add routing task
			t.Status = TaskEnded
			return
		}
		t.Status = TaskInProg
		t.findBed()
		if t.bedObj == nil {
			panic("bed object was nil")
		}
		t.activateObjectTask = NewActivateObjectTask(t.Owner, t.bedObj)
	case TaskInProg:
		// either going to the bed, or already in the bed
		if t.inBed {
			// in bed; nothing to do
			return
		}
		if t.activateObjectTask != nil {
			t.handleActivateObj()
			return
		}
		// not in bed, and not activating bed object... what's going on?
		logz.Panicln("SleepTask", "somehow, we are neither in bed or trying to activate the bed... npcID:", t.Owner.ID(), "bedID:", t.bedObj.ID)
	case TaskEnded:
		// task ended for some reason
		return
	}
}

func (t *SleepTask) handleActivateObj() {
	if t.activateObjectTask.IsDone() {
		// sanity check: make sure targetingNPC was unset
		if t.bedObj.GetTargetingNPC() == t.Owner.CharacterStateRef.ID {
			logz.Panicln("SleepTask", "activate object finished, but this NPC is still set as the targeting NPC...", "npcID:", t.Owner.ID(), "bedID:", t.bedObj.ID)
		}
		// bed has been activated
		if t.activateObjectTask.Success {
			// we should be currently in bed right now; nothing to do.
			if !t.bedObj.Bed.InUse {
				logz.Panicln("SleepTask", "activate object task was reportedly a success, but bed is not in use? bedID:", t.bedObj.ID, "npcID:", t.Owner.ID())
			}
			if t.bedObj.Bed.SleeperID != t.Owner.CharacterStateRef.ID {
				logz.Panicln("SleepTask", "activate object task was reportedly a success, but bed object sleeper ID doesn't match NPC ID. sleeperID:", t.bedObj.Bed.SleeperID, "npcID:", t.Owner.ID())
			}
			if !t.Owner.Entity.IsSleeping {
				logz.Panicln("SleepTask", "we are supposed to be sleeping, but entity is not marked as sleeping...")
			}

			t.inBed = true
			return
		}
		// failed to activate bed...
		if t.bedObj.Bed.InUse {
			logz.Panicln("SleepTask", "failed to activate bed; bed is already in use. sleeperID:", t.bedObj.Bed.SleeperID, "objID:", t.bedObj.ID, "npc:", t.Owner.WhoAmI())
		}
		if t.activateObjectTask.FailReason.failedToReachObject {
			logz.Panicln("SleepTask", "failed to activate bed; failed to reach target bed.", t.Owner.WhoAmI(), "bedID:", t.bedObj.ID)
		}
		logz.Panicln("SleepTask", "failed to activate bed, due to some unknown reason... ", t.Owner.WhoAmI(), "bedID:", t.bedObj.ID)
	} else {
		t.activateObjectTask.Update()
	}
}

func (t *SleepTask) findBed() {
	bedID := t.Owner.CharacterStateRef.HomeMapBedID
	homeMap := t.Owner.CharacterStateRef.HomeMapID
	for _, obj := range t.Owner.World.GetAllObjects() {
		if obj.ID == bedID {
			if obj.Bed.InUse {
				if obj.Bed.SleeperID != state.CharacterStateID(defs.PlayerID) {
					logz.Panicln("SleepTask", "Another (non-player) character appears to be sleeping in NPC's bed! npcID:", t.Owner.ID(), "bedID:", bedID, "sleeperID:", obj.Bed.SleeperID)
				}
				// the player is using this NPC's bed... we should probably cause something to happen, but let's just cancel the task for now.
				t.Status = TaskEnded
				return
			}
			targetingNPC := obj.GetTargetingNPC()
			if targetingNPC != "" {
				if targetingNPC == state.CharacterStateID(defs.PlayerID) {
					// player is... targeting the bed?  doesn't seem possible, since player doesn't use tasks right?
					logz.Panicln("SleepTask", "player is apparently targeting the bed... this shouldnt be happening, right?")
				}
				logz.Panicln("SleepTask", "a different NPC is targeting my bed! that shouldn't be happening... npcID:", t.Owner.ID(), "targetingNPC:", targetingNPC, "bedID:", obj.ID)
			}
			t.bedObj = obj
			return
		}
	}
	logz.Panicln("SleepTask", "couldn't find NPC's bed! it's supposed to be in this map... homeMapID:", homeMap, "bedID:", bedID, "npcID:", t.Owner.ID())
}

func (t *SleepTask) SetupActiveState() {
	homeMap := t.Owner.CharacterStateRef.HomeMapID
	if t.Owner.CharacterStateRef.CurrentMap != homeMap {
		// TODO: route to new map?
		t.Status = TaskEnded
		return
	}

	t.findBed()
	if t.bedObj == nil {
		panic("bed was nil")
	}
	// directly activate the bed, so that the NPC is already in it
	// set the NPC in an open spot right next to the bed first, so that they have a valid "position before sleeping" set in entity
	// this is to ensure that when the NPC leaves the bed, they will appear next to it as one would expect.

	c := t.bedObj.TilePos()
	nearest, found := t.Owner.getNearestOpenTile(c, 2, true)
	if !found {
		logz.Panicln("SleepTask", "failed to find open tile near bed. bedID:", t.bedObj.ID, t.Owner.WhoAmI())
	}
	t.Owner.Entity.SetPosition(nearest)

	params := object.ObjectActivationParams{
		ActivatorID: t.Owner.Entity.ID(),
		LockIDs:     characterstate.GetLockIDs(*t.Owner.CharacterStateRef),
	}
	x, y := t.bedObj.Pos()
	res := t.bedObj.Activate(x, y, params)
	if res.UpdateOccurred {
		t.inBed = true
		t.Owner.handleObjectUpdate(t.bedObj, res)
		t.Status = TaskInProg
		return
	}
	logz.Panicln("SleepTask", "failed to activate bed...", t.Owner.WhoAmI(), "bedID:", t.bedObj.ID)
}

func (t *SleepTask) BackgroundAssist() {
	if t.activateObjectTask != nil && !t.activateObjectTask.IsDone() {
		t.activateObjectTask.BackgroundAssist()
	}
}

func (t *SleepTask) SimulationUpdate() {}
