package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/utils"
)

// LoungeTask is a task that causes an NPC to find an area nearby to "lounge" and hang out.
// Similar to Idle, but the NPC actually does something a little more "interesting", like sits in a chair.
// If no interesting objects (like chairs) are nearby, then the NPC just idles.
type LoungeTask struct {
	TaskBase
	isLounging         bool // if set, then NPC is considered to be actively lounging, and no updates are needed.
	activateObjectTask *ActivateObjectTask
	idleTask           *IdleTask
}

func NewLoungeTask(n *NPC, def defs.TaskDef) *LoungeTask {
	if def.TaskID != TaskLounge {
		panic("def had wrong task ID")
	}
	return &LoungeTask{
		TaskBase: NewTaskBase(
			def,
			"Lounge",
			"NPC finds a place in the map to lounge, and just chill, dawg.",
			n,
		),
	}
}

func (t *LoungeTask) Update() {
	if t.isLounging {
		// chilling in a chair somewhere; no need to do anything
		return
	}
	if !t.RouteToStartMap(false) {
		// we are routing to a different map to start task
		return
	}
	if t.activateObjectTask != nil {
		if t.Status != TaskInProg {
			panic("task status should be in progress. did we forget to update it somewhere?")
		}
		// working on activating a chair object
		t.activateObjectTask.Update()
		if t.activateObjectTask.IsDone() {
			if t.activateObjectTask.Success {
				// successfully got in the chair; now we chill
				t.isLounging = true
			} else {
				// failed to activate chair. should we just idle then?
				t.startIdleTask()
			}
		}
		return
	}
	if t.idleTask != nil {
		if t.Status != TaskInProg {
			panic("task status should be in progress. did we forget to update it somewhere?")
		}
		// no chair was found, so just idling the day away
		t.idleTask.Update()
		if t.idleTask.IsDone() {
			// idle task is done? not sure why that would be!
			panic("idle task is done...")
		}
		return
	}

	t.Status = TaskInProg

	// no lounging target yet; find a target
	// see if there are any chairs nearby
	closestChair := t.findChair()
	if closestChair != nil {
		// found a chair! let's sit in it
		t.activateObjectTask = NewActivateObjectTask(t.Owner, closestChair)
	} else {
		// no chairs found; just idle then
		t.startIdleTask()
	}
}

func (t LoungeTask) findChair() *object.Object {
	var closestChair *object.Object
	var closestDist float64
	for _, obj := range t.Owner.ActiveMapCtx.GetAllObjects() {
		if obj.Type == object.TypeChair && !obj.Chair.InUse && obj.GetTargetingNPC() == "" {
			if !t.Owner.SatisfiesObjectOwnership(*obj) {
				// NPC doesn't have right roles or is not the owner of this object
				continue
			}
			x, y := obj.Pos()
			objPos := model.ConvertPxToTilePos(x, y)
			dist := utils.EuclideanDistCoords(t.Owner.Entity.TilePos(), objPos)

			if closestChair == nil {
				closestChair = obj
				closestDist = dist
			} else if dist < closestDist {
				closestChair = obj
				closestDist = dist
			}
		}
	}
	return closestChair
}

func (t *LoungeTask) startIdleTask() {
	t.idleTask = NewIdleTask(t.Owner, defs.TaskDef{
		TaskID:   TaskIdle,
		Priority: t.GetPriority(),
	})
}

func (t *LoungeTask) SetupActiveState() {
	if !t.InStartMap() {
		// if we aren't in start map at this function call, then we should already be routing there; setupActiveState for underlying routing task.
		t.RouteToStartMapSetupActiveState()
		return
	}
	// basically, we need to do what we do in the main update, but without the "gotos".
	// 1. find a chair if one is free, and immediately activate/sit in it.
	t.Status = TaskInProg // TODO: is taskinProg checked anywhere?
	if t.Owner.Entity.IsSitting {
		logz.Panicln("LoungeTask", "SetupActiveState was called, but for some reason the NPC was already sitting in a chair... did entity not get reset from a previous map?", t.Owner.WhoAmI())
	}
	closestChair := t.findChair()
	if closestChair != nil {
		// set the NPC in an open spot right next to the chair first, so that they have a valid "position before sitting" set in entity
		// this is to ensure that when the NPC leaves the chair, they will appear next to it as one would expect.
		c := closestChair.TilePos()
		nearest, found := t.Owner.getNearestOpenTile(c, 2, true)
		if !found {
			logz.Panicln("LoungeTask", "failed to find open tile near chair. chairID:", closestChair.ID)
		}
		t.Owner.Entity.SetPosition(nearest)

		// now, actually activate the chair and put the NPC in it.
		params := object.ObjectActivationParams{
			ActivatorID: t.Owner.Entity.ID(),
			LockIDs:     characterstate.GetLockIDs(*t.Owner.CharacterStateRef),
		}
		x, y := closestChair.Pos()
		res := closestChair.Activate(x, y, params)
		if res.UpdateOccurred {
			logz.Println("LoungeTask", "sitting in chair:", closestChair.ID, "npcID:", t.Owner.ID())
			t.isLounging = true
			t.Owner.handleObjectUpdate(closestChair, res)
			return
		}
	}
	logz.Println("LoungeTask", "failed to find a chair to sit in:", t.Owner.ID())
	// 2. if no chair exists, then just idle in a random place.
	t.startIdleTask()
	t.idleTask.SetupActiveState()
}

func (t *LoungeTask) SimulationUpdate() {
	t.RouteToStartMap(true)
}

func (t *LoungeTask) BackgroundAssist() {
	t.RouteToStartMapBgAssist()
}
