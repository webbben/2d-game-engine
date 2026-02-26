package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/object"
)

type TaskStatus int

const (
	// the task has not started yet
	TaskNotStarted TaskStatus = iota
	// task has started processing updates
	TaskInProg
	// task has ended and is no longer active
	TaskEnded
)

const (
	TaskIdle        defs.TaskID = "IDLE"
	TaskGoto        defs.TaskID = "GOTO"
	TaskFollow      defs.TaskID = "FOLLOW"
	TaskFight       defs.TaskID = "FIGHT"
	TaskStartDialog defs.TaskID = "START_DIALOG"
	TaskFaceDir     defs.TaskID = "FACE_DIR"
)

const (
	Schedule defs.TaskPriority = iota
	Assign
	Emergency
)

func (ts TaskStatus) String() string {
	switch ts {
	case TaskNotStarted:
		return "NOTSTARTED (0)"
	case TaskInProg:
		return "INPROG (1)"
	case TaskEnded:
		return "END (2)"
	default:
		return "(Error: task status not given a string representation yet)"
	}
}

type Task interface {
	GetID() defs.TaskID

	GetNextTaskDef() *defs.TaskDef

	// the NPC who "owns" this task (i.e. the NPC who is currently running this task)
	GetOwner() *NPC

	GetDescription() string

	GetPriority() defs.TaskPriority

	GetName() string

	GetStatus() TaskStatus // current status of function

	IsDone() bool   // flag that indicates this task is finished or ended. causes no further updates to process.
	IsActive() bool // indicates that the task is currently underway (already started, and hasn't stopped yet)

	// logic to execute on each update tick
	Update()

	// background task assistance
	BackgroundAssist()
	// lastBackgroundAssist time.Time // last time a the bg task loop assisted this task
}

type TaskBase struct {
	Owner       *NPC
	Description string
	ID          defs.TaskID
	Priority    defs.TaskPriority
	Name        string
	Status      TaskStatus
	NextTaskDef *defs.TaskDef
}

// NewTaskBase defines a task base that covers all the bases of the Task interface.
//
// nextTask: OPT (only set if you want another task to start right after this one finishes)
func NewTaskBase(id defs.TaskID, name, desc string, owner *NPC, p defs.TaskPriority, nextTask *defs.TaskDef) TaskBase {
	if owner == nil {
		panic("owner was nil")
	}
	return TaskBase{
		ID:          id,
		Priority:    p,
		Name:        name,
		Description: desc,
		Status:      TaskNotStarted,
		Owner:       owner,
		NextTaskDef: nextTask,
	}
}

func (tb TaskBase) GetNextTaskDef() *defs.TaskDef {
	return tb.NextTaskDef
}

func (tb TaskBase) GetOwner() *NPC {
	return tb.Owner
}

func (tb TaskBase) GetDescription() string {
	return tb.Description
}

func (tb TaskBase) GetID() defs.TaskID {
	return tb.ID
}

func (tb TaskBase) GetPriority() defs.TaskPriority {
	return tb.Priority
}

func (tb TaskBase) GetName() string {
	return tb.Name
}

func (tb TaskBase) GetStatus() TaskStatus {
	return tb.Status
}

func (tb TaskBase) IsDone() bool {
	return tb.Status == TaskEnded
}

func (tb TaskBase) IsActive() bool {
	return tb.Status > TaskNotStarted && tb.Status < TaskEnded
}

func (n *NPC) HandleTaskUpdate() {
	if n.CurrentTask.GetOwner() == nil {
		panic("tried to run task that has no owner set")
	}
	if n.CurrentTask == nil {
		panic("current task is nil?")
	}
	if n.CurrentTask.IsDone() {
		return
	}

	n.CurrentTask.Update()
}

type NPCCollisionResult struct {
	NoneDetected bool
	Wait         bool
	ReRoute      bool
}

// HandleNPCCollision handles any collisions that are occurring.
// Basically, if the NPC has run up into an object, like a gate, they can try to open it.
// Note: entities don't collide, they should just move slower and incur more "cost" by walking into each other.
func (t *TaskBase) HandleNPCCollision() NPCCollisionResult {
	if !t.Owner.Entity.Movement.Interrupted {
		return NPCCollisionResult{NoneDetected: true}
	}
	logz.Println(t.Owner.DisplayName, "NPC interrupted; handling collision")
	// path entity was moving on has been interrupted.
	// if interrupted by NPC, try to negotiate resolution to collision.
	// TODO: this probably belongs as validation in the entity movement logic itself
	if !t.Owner.Entity.TargetTilePos().Equals(t.Owner.Entity.TilePos()) {
		logz.Println(t.Owner.ID, "Goto task: since NPC movement was interrupted, we expect its target position to be the same as its current position. target:", t.Owner.Entity.TargetTilePos(), t.Owner.Entity.TilePos())
		panic("Goto task: since NPC movement was interrupted, we expect its target position to be the same as its current position")
	}
	// TODO: same with this?
	if len(t.Owner.Entity.Movement.TargetPath) == 0 {
		panic("Goto task: npc movement was interrupted, but there is no next step in target path")
	}

	// get NPCs that are at the next target tile - i.e. the next position in the target path
	nextTarget := t.Owner.Entity.Movement.TargetPath[0]

	collidingObjs := t.Owner.World.FindObjectsAtPosition(nextTarget)
	if len(collidingObjs) > 0 {
		// see if any of these objects are things like gates, that can be opened.
		for _, obj := range collidingObjs {
			if !obj.IsCollidable() {
				continue
			}
			if obj.Type == object.TypeGate {
				if !obj.IsCurrentlyActivating() {
					x, y := t.Owner.Entity.X, t.Owner.Entity.Y
					obj.Activate(x, y)
				}
				// wait a little for the gate to open
				t.Owner.Wait(time.Second)
				return NPCCollisionResult{Wait: true}
			}
			// found an object that is collidable and cannot be opened or resolved;
			// tell NPC to reroute.
			return NPCCollisionResult{ReRoute: true}
		}
	}
	// no collidable objects found; should be good to continue?
	return NPCCollisionResult{NoneDetected: true}
}

// Made an Idle task for NPCs that don't do anything

type IdleTask struct {
	TaskBase
}

func (it IdleTask) ZzCompileCheck() {
	_ = append([]Task{}, &it)
}

func (it IdleTask) BackgroundAssist() {
}
func (it IdleTask) End() {}
func (it IdleTask) IsComplete() bool {
	return false
}

func (it IdleTask) IsFailure() bool {
	return false
}
func (it *IdleTask) Start()  {}
func (it *IdleTask) Update() {}
