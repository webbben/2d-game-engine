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
	isLounging bool // if set, then NPC is considered to be actively lounging, and no updates are needed.
	// the child slot holds whichever sub-task is active: an activate-object task (going to sit in a chair) or an
	// idle task (no chair available). Both run as this task's single child.
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

func init() {
	registerTask(TaskLounge, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			return NewLoungeTask(owner, def)
		},
	})
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
	if !t.ChildDone() {
		if t.Status != TaskInProg {
			panic("task status should be in progress. did we forget to update it somewhere?")
		}
		// a child is active (going to a chair, or idling); advance it
		t.TaskBase.Update()
		return
	}

	if t.HasChild() {
		// a child finished. The idle child should never finish; only the activate-object child can complete.
		if t.loadChild().GetID() == TaskIdle {
			panic("idle task is done...")
		}
		childResult := t.ChildResult()
		t.EndChild()
		if childResult.Status == ResultSuccess {
			// successfully got in the chair; now we chill
			t.isLounging = true
		} else {
			// failed to activate chair. should we just idle then?
			t.startIdleTask()
		}
		return
	}

	t.Status = TaskInProg

	// no lounging target yet; find a target
	// see if there are any chairs nearby
	closestChair := t.findChair()
	if closestChair != nil {
		// found a chair! let's sit in it
		t.RunChild(NewActivateObjectTask(t.Owner, closestChair))
	} else {
		// no chairs found; just idle then
		t.startIdleTask()
	}
}

func (t LoungeTask) findChair() *object.Object {
	// find the closest chair, or a chair that has the right owner_id
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
				if closestChair.OwnerID != t.Owner.CharacterStateRef.ID || obj.OwnerID == t.Owner.CharacterStateRef.ID {
					// only overwrite if we haven't chosen a chair by ownerID, or if this one also is owned by the NPC
					closestChair = obj
					closestDist = dist
				}
			} else if obj.OwnerID == t.Owner.CharacterStateRef.ID {
				closestChair = obj
				closestDist = dist
			}
		}
	}
	return closestChair
}

func (t *LoungeTask) startIdleTask() {
	t.RunChild(NewIdleTask(t.Owner, defs.TaskDef{
		TaskID:   TaskIdle,
		Priority: t.GetPriority(),
	}))
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
		logz.Println("LoungeTask", t.Owner.WhoAmI())
		logz.Panicln("LoungeTask", "SetupActiveState was called, but for some reason the NPC was already sitting in a chair... did entity not get reset from a previous map?")
	}
	closestChair := t.findChair()
	if closestChair != nil {
		// set the NPC in an open spot right next to the chair first, so that they have a valid "position before sitting" set in entity
		// this is to ensure that when the NPC leaves the chair, they will appear next to it as one would expect.
		c := closestChair.TilePos()
		nearest, found := t.Owner.getNearestOpenTile(c, 2, true)
		if !found {
			logz.Println("LoungeTask", closestChair.ID)
			logz.Panicln("LoungeTask", "failed to find open tile near chair")
		}
		t.Owner.Entity.SetPosition(nearest)

		// now, actually activate the chair and put the NPC in it.
		params := object.ObjectActivationParams{
			ActivatorID: t.Owner.Entity.ID(),
			LockIDs:     characterstate.GetLockIDs(*t.Owner.CharacterStateRef, t.Owner.dataman),
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
	t.TaskBase.SetupActiveState()
}

func (t *LoungeTask) SimulationUpdate() {
	t.RouteToStartMap(true)
}

func (t *LoungeTask) BackgroundAssist() {
	t.TaskBase.BackgroundAssist()
}
