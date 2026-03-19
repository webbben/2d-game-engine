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
	NoBackgroundWork
	isLounging         bool // if set, then NPC is considered to be actively lounging, and no updates are needed.
	activateObjectTask *TaskActivateObject
	idleTask           *IdleTask
}

func NewLoungeTask(n *NPC, p defs.TaskPriority) *LoungeTask {
	return &LoungeTask{
		TaskBase: NewTaskBase(
			defs.TaskDef{
				TaskID:   TaskLounge,
				Priority: p,
			},
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
	for _, obj := range t.Owner.World.GetAllObjects() {
		if obj.Type == object.TypeChair && !obj.Chair.InUse && obj.GetTargetingNPC() == "" {
			// TODO: need to also check if chair is suitable for the specific NPC (i.e. has character restrictions), but skipping for now
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
	t.idleTask = NewIdleTask(t.Owner, t.GetPriority())
}

func (t *LoungeTask) SetupActiveState() {
	// basically, we need to do what we do in the main update, but without the "gotos".
	// 1. find a chair if one is free, and immediately activate/sit in it.
	t.Status = TaskInProg
	closestChair := t.findChair()
	if closestChair != nil {
		// activate it right now, instead of walking to it first
		x, y := closestChair.Pos()
		params := object.ObjectActivationParams{
			ActivatorID: t.Owner.Entity.ID(),
			LockIDs:     characterstate.GetLockIDs(*t.Owner.CharacterStateRef),
		}
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
