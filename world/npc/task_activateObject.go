package npc

import (
	"github.com/webbben/2d-game-engine/data/state"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/utils"
)

// TaskActivateObject is a smaller "sub-task" that simply directs an NPC to go to an object and activate it.
type TaskActivateObject struct {
	TaskBase
	NoActiveState
	targetObj *object.Object

	gotoTask *GotoTask
	skipGoto bool

	Success bool // if true, we successfully activated the target object
}

func NewActivateObjectTask(n *NPC, obj *object.Object) *TaskActivateObject {
	if obj.GetTargetingNPC() != "" {
		panic("object is already being targeted by another NPC; you should ensure the object is not targeted before setting the activateObjectTask.")
	}
	obj.SetTargetingNPC(n.Entity.ID())
	return &TaskActivateObject{
		TaskBase: TaskBase{
			Name:        "Activate Object",
			Description: "NPC goes to an object and tries to activate it",
			Owner:       n,
		},
	}
}

func (t TaskActivateObject) ZzInterfaceCheck() {
	_ = append([]Task{}, &t)
}

func (t *TaskActivateObject) Update() {
	if t.IsDone() {
		return
	}

	t.Status = TaskInProg

	// 1. go to object (if we aren't already next to it)
	if t.gotoTask == nil && !t.skipGoto {
		x, y := t.targetObj.Pos()
		objPos := model.ConvertPxToTilePos(x, y)
		dist := utils.EuclideanDistCoords(t.Owner.Entity.TilePos(), objPos)
		if dist > 2 {
			t.gotoTask = NewGotoTask(GotoTaskParams{TileX: int(x), TileY: int(y)}, t.Owner, t.GetPriority(), nil)
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
	x, y := t.targetObj.Pos()
	objPos := model.ConvertPxToTilePos(x, y)
	dist := utils.EuclideanDistCoords(t.Owner.Entity.TilePos(), objPos)
	if dist > 2 {
		// it seems we didn't manage to get close enough to the object...
		// TODO: maybe we should add a field to GotoTask that says if it successfully reached the target or not
		t.Status = TaskEnded
		t.targetObj.ClearTargetingNPC()
		t.Success = false
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
	}
	t.Status = TaskEnded
}

func (t *TaskActivateObject) BackgroundAssist() {
	t.gotoTask.BackgroundAssist()
}

func (t *TaskActivateObject) SimulationUpdate() {}

func (n *NPC) handleObjectUpdate(obj *object.Object, res object.ObjectUpdateResult) {
	switch obj.Type {
	case object.TypeChair:
		if res.UpdateOccurred {
			v := model.NewVec2(obj.Rect.X, obj.Rect.Y)
			n.Entity.SitInChair(v, obj.Chair.Direction)
		}
	}
}
