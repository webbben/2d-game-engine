package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/utils"
)

// ActivateObjectTask is a smaller "sub-task" that simply directs an NPC to go to an object and activate it.
type ActivateObjectTask struct {
	TaskBase
	targetObj *object.Object

	gotoTask *GotoTask
	skipGoto bool
}

func NewActivateObjectTask(n *NPC, obj *object.Object) *ActivateObjectTask {
	if obj == nil {
		panic("obj was nil")
	}
	if obj.GetTargetingNPC() != "" {
		// some objects, like doors, don't matter if they're already targeted; multiple NPCs can use the same door simultaneously to switch maps
		if obj.Type != object.TypeDoor {
			logz.Println("NewActivateObjectTask", obj.ID, obj.Type, "npc:", n.WhoAmI())
			logz.Panicln("NewActivateObjectTask", "object is already being targeted by another NPC; ensure the object is not targeted before setting activateObjectTask")
		}
	}
	if !n.SatisfiesObjectOwnership(*obj) {
		logz.Println("ActivateObjectTask", obj.ID, n.WhoAmI())
		logz.Panicln("ActivateObjectTask", "tried to create activate object task, but NPC is not authorized to use this object")
	}
	obj.SetTargetingNPC(n.Entity.ID())
	return &ActivateObjectTask{
		targetObj: obj,
		TaskBase: TaskBase{
			Def: defs.TaskDef{
				TaskID: TaskActivateObj,
			},
			Name:        "Activate Object",
			Description: "NPC goes to an object and tries to activate it",
			Owner:       n,
		},
	}
}

var _ Task = (*ActivateObjectTask)(nil)

// Finish releases the object's targeting reservation before recording the result. The object is reserved the
// moment the task is created, and only cleared on the reachable success/failure paths inside Update; a Finish
// (e.g. ResultAborted when a parent task is preempted mid-route) previously left the object reserved forever
// (bed/door/chair soft-locks). Clearing here makes the release unconditional.
func (t *ActivateObjectTask) Finish(result TaskResult) {
	if t.targetObj != nil {
		t.targetObj.ClearTargetingNPC()
	}
	t.TaskBase.Finish(result)
}

func (t *ActivateObjectTask) Update() {
	if t.IsDone() {
		return
	}

	if t.targetObj == nil {
		panic("target object was nil")
	}

	t.Status = TaskInProg

	// 1. go to object (if we aren't already next to it)
	if t.gotoTask == nil && !t.skipGoto {
		objPos := t.targetObj.TilePos()
		if objPos.Equals(model.Coords{X: 0, Y: 0}) {
			x, y := t.targetObj.Pos()
			logz.Println("ActivateObjectTask", "objID:", t.targetObj.ID, "object pos:", x, y, "rect:", t.targetObj.GetRect())
			logz.Panicln("ActivateObjectTask", "object position (tile position) came back as 0 0, which seems wrong.")
		}

		dist := utils.EuclideanDistCoords(t.Owner.Entity.TilePos(), objPos)
		if dist > 2 {
			t.gotoTask = NewGotoTask(GotoTaskParams{TileX: objPos.X, TileY: objPos.Y}, t.Owner, defs.TaskDef{
				TaskID:   TaskGoto,
				Priority: t.GetPriority(),
			})
		} else {
			// already close to the object, so no need to go to it
			t.skipGoto = true
		}
	}
	// if the goto task is not done, then keep updating it
	if t.gotoTask != nil && !t.gotoTask.IsDone() {
		t.gotoTask.Update()
		return
	}

	if t.gotoTask == nil && !t.skipGoto {
		logz.Println("ActivateObjectTask", t.Owner.WhoAmI())
		logz.Panicln("ActivateObjectTask", "goto task was unexpectedly nil, but we aren't skipping goto.")
	}

	// 2. once next to the object, try to activate it
	// confirm we are next to the target object now
	objPos := t.targetObj.TilePos()
	dist := utils.EuclideanDistCoords(t.Owner.Entity.TilePos(), objPos)
	if dist > 2 {
		if t.gotoTask != nil && t.gotoTask.Result.Status == ResultSuccess {
			logz.Println(t.Owner.WhoAmI(), "distance:", dist)
			logz.Panicln("ActivateObjectTask", "gotoTask reported success, but didn't get close enough to object.")
		} else {
			logz.Println("ActivateObjectTask", "failed to reach object; didn't get close enough. distance to object:", dist, "objPos:", objPos, "objID:", t.targetObj.ID, "whoami:", t.Owner.WhoAmI())
			t.targetObj.ClearTargetingNPC()
			t.FinishFail("failed to reach object")
			return
		}
	}

	// try to activate the object
	charState := *t.Owner.CharacterStateRef
	t.targetObj.ClearTargetingNPC()
	res := t.targetObj.Activate(t.Owner.Entity.X, t.Owner.Entity.Y, object.ObjectActivationParams{
		ActivatorID: id.CharacterStateID(t.Owner.ID()),
		LockIDs:     characterstate.GetLockIDs(charState, t.Owner.dataman),
	})
	if res.UpdateOccurred {
		t.Owner.handleObjectUpdate(t.targetObj, res)
		t.FinishSuccess()
		return
	}
	t.FinishFail("activation failed")
}

func (t *ActivateObjectTask) BackgroundAssist() {
	if t.gotoTask != nil {
		t.gotoTask.BackgroundAssist()
	}
}

func (t *ActivateObjectTask) SimulationUpdate() {}

func (n *NPC) handleObjectUpdate(obj *object.Object, res object.ObjectUpdateResult) {
	if !res.UpdateOccurred {
		panic("update didn't occur, but handleObjectUpdate was called")
	}
	switch obj.Type {
	case object.TypeChair:
		n.Entity.SitInChair(obj)
	case object.TypeBed:
		n.Entity.SleepInBed(obj)
	case object.TypeDoor:
		// remove NPC from active map, and put them into the new map's occupancy
		n.ActiveMapCtx.RemoveNPCFromActiveMap(n.CharacterStateRef.ID, res.ChangeMapID)
	}
}
