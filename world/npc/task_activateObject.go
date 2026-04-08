package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/utils"
)

// ActivateObjectTask is a smaller "sub-task" that simply directs an NPC to go to an object and activate it.
type ActivateObjectTask struct {
	TaskBase
	NoActiveState
	targetObj *object.Object

	gotoTask *GotoTask
	skipGoto bool

	Success    bool                     // if true, we successfully activated the target object
	FailReason activateObjectFailReason // if failed, this should tell you what went wrong
}

type activateObjectFailReason struct {
	failedToReachObject bool
	activationFailed    bool
}

func NewActivateObjectTask(n *NPC, obj *object.Object) *ActivateObjectTask {
	if obj == nil {
		panic("obj was nil")
	}
	if obj.GetTargetingNPC() != "" {
		panic("object is already being targeted by another NPC; you should ensure the object is not targeted before setting the activateObjectTask.")
	}
	obj.SetTargetingNPC(n.Entity.ID())
	return &ActivateObjectTask{
		targetObj: obj,
		TaskBase: TaskBase{
			Name:        "Activate Object",
			Description: "NPC goes to an object and tries to activate it",
			Owner:       n,
		},
	}
}

func (t ActivateObjectTask) ZzInterfaceCheck() {
	_ = append([]Task{}, &t)
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
	if t.gotoTask != nil && !t.gotoTask.IsDone() {
		t.gotoTask.Update()
		return
	}

	// 2. once next to the object, try to activate it
	// confirm we are next to the target object now
	objPos := t.targetObj.TilePos()
	dist := utils.EuclideanDistCoords(t.Owner.Entity.TilePos(), objPos)
	if dist > 2 {
		// it seems we didn't manage to get close enough to the object...
		// TODO: maybe we should add a field to GotoTask that says if it successfully reached the target or not
		logz.Println("ActivateObjectTask", "failed to reach object; didn't get close enough. distance to object:", dist, "objPos:", objPos, "objID:", t.targetObj.ID, "whoami:", t.Owner.WhoAmI())
		t.Status = TaskEnded
		t.targetObj.ClearTargetingNPC()
		t.Success = false
		t.FailReason.failedToReachObject = true
		return
	}

	// try to activate the object
	charState := *t.Owner.CharacterStateRef
	t.targetObj.ClearTargetingNPC()
	res := t.targetObj.Activate(t.Owner.Entity.X, t.Owner.Entity.Y, object.ObjectActivationParams{
		ActivatorID: state.CharacterStateID(t.Owner.ID()),
		LockIDs:     characterstate.GetLockIDs(charState),
	})
	if res.UpdateOccurred {
		t.Owner.handleObjectUpdate(t.targetObj, res)
		t.Success = true
	} else {
		t.FailReason.activationFailed = true
	}
	t.Status = TaskEnded
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
