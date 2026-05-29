package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
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
	if t.IsDone() {
		// task ended for some reason
		return
	}

	if !t.RouteToStartMap(false) {
		// we are routing to a different map
		return
	}
	// we are in the start map (already - it should be the active map, and the character must've entered the map now or was already in it.)
	if !t.InStartMap() {
		panic("not in start map?")
	}

	if t.bedObj == nil {
		// find the bed we are to sleep in
		t.findBed()
		if t.bedObj == nil {
			panic("bed object was nil")
		}
		// go to the bed and activate it
		t.activateObjectTask = NewActivateObjectTask(t.Owner, t.bedObj)
		return
	}

	if t.activateObjectTask != nil {
		if t.inBed {
			panic("NPC is in bed, but activateObjectTask wasn't cleared")
		}
		t.handleActivateObj()
		return
	}

	if t.inBed {
		// in the bed; mission accomplished
		return
	}

	logz.Panicln("SleepTask", "somehow, we are neither in bed or trying to activate the bed... npcID:", t.Owner.ID(), "bedID:", t.bedObj.ID)
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
			t.activateObjectTask = nil
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
	for _, obj := range t.Owner.ActiveMapCtx.GetAllObjects() {
		if obj.ID == bedID {
			if !t.Owner.SatisfiesObjectOwnership(*obj) {
				// don't sleep in someone else's bed
				continue
			}
			if obj.Bed.InUse {
				if obj.Bed.SleeperID != id.CharacterStateID(defs.PlayerID) {
					logz.Panicln("SleepTask", "Another (non-player) character appears to be sleeping in NPC's bed! npcID:", t.Owner.ID(), "bedID:", bedID, "sleeperID:", obj.Bed.SleeperID)
				}
				// the player is using this NPC's bed... we should probably cause something to happen, but let's just cancel the task for now.
				t.Status = TaskEnded
				return
			}
			targetingNPC := obj.GetTargetingNPC()
			if targetingNPC != "" {
				if targetingNPC == id.CharacterStateID(defs.PlayerID) {
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
	if !t.InStartMap() {
		// if we aren't in start map at this function call, then we should already be routing there; setupActiveState for underlying routing task.
		t.RouteToStartMapSetupActiveState()
		return
	}

	t.findBed()
	if t.bedObj == nil {
		panic("bed was nil")
	}
	// directly activate the bed, so that the NPC is already in it
	// set the NPC in an open spot right next to the bed first, so that they have a valid "position before sleeping" set in entity
	// this is to ensure that when the NPC leaves the bed, they will appear next to it as one would expect.

	if t.Owner.Entity.IsSleeping {
		logz.Panicln("SleepTask", "SetupActiveState was called, but for some reason the NPC was already sleeping... did entity not get reset from a previous map?", t.Owner.WhoAmI())
	}

	c := t.bedObj.TilePos()
	nearest, found := t.Owner.getNearestOpenTile(c, 2, true)
	if !found {
		logz.Panicln("SleepTask", "failed to find open tile near bed. bedID:", t.bedObj.ID, t.Owner.WhoAmI())
	}
	t.Owner.Entity.SetPosition(nearest)

	params := object.ObjectActivationParams{
		ActivatorID: t.Owner.Entity.ID(),
		LockIDs:     characterstate.GetLockIDs(*t.Owner.CharacterStateRef, t.Owner.dataman),
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
	if !t.InStartMap() {
		// sleep needs to route to a new map, so send bg assist to route so it can calculate the path
		// NOTE: it seems there's a slim chance of background assist calling after the task starts, but before the task has had its routing set up.
		// this is probably a race condition technically, but I think it won't be an issue if I handle it with this condition.
		// However, if more race-condition stuff keeps happening in the future, maybe it'll be work adding some synchronization to the different goroutines,
		// like a Mutex or something.
		t.RouteToStartMapBgAssist()
		return
	}
	if !t.InActiveMap() {
		// NOTE: considered doing a panic here, but figured there is a bg assist calling at least one tick even after the task exited the active map,
		// due to the separate goroutines/race case possibility.
		// if we switch to mutexes, then this should be made into a panic.
		return
	}
	if t.activateObjectTask != nil && !t.activateObjectTask.IsDone() {
		t.activateObjectTask.BackgroundAssist()
	}
}

func (t *SleepTask) SimulationUpdate() {
	// the only thing simulation needs to handle is getting the NPC to the right map
	t.RouteToStartMap(true)
}
