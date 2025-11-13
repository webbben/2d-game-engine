package npc

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/internal/logz"
)

type TaskStatus string

const (
	// the task has not started yet
	TASK_STATUS_NOTSTARTED TaskStatus = "not_yet_started"
	// task has only started, and no update has occurred yet
	TASK_STATUS_START TaskStatus = "start"
	// task has started processing updates
	TASK_STATUS_INPROG TaskStatus = "in_progress"
	// task has ended and is no longer active
	TASK_STATUS_END TaskStatus = "end"
)

type Task interface {
	// the NPC who "owns" this task (i.e. the NPC who is currently running this task)
	GetOwner() *NPC
	SetOwner(ownerNPC *NPC)

	GetDescription() string
	SetDescription(desc string)

	GetID() string
	SetID(id string)

	GetName() string
	SetName(name string)

	GetStatus() TaskStatus // current status of function
	SetStatus(status TaskStatus)

	IsDone() bool // flag that indicates this task is finished or ended. causes no further updates to process.
	SetDone()

	// Optional custom logic for task completion
	IsComplete() bool
	// Optional custom logic for task failure
	IsFailure() bool

	// logic to execute in order to start the task
	Start()
	// logic to execute on each update tick
	Update()
	// logic to execute when task is done - in case some kind of cleanup should occur
	Cleanup()

	// background task assistance
	BackgroundAssist()
	//lastBackgroundAssist time.Time // last time a the bg task loop assisted this task
}

type TaskBase struct {
	Owner       *NPC
	Description string
	ID          string
	Name        string
	Status      TaskStatus
}

func NewTaskBase(id, name, desc string) TaskBase {
	return TaskBase{
		ID:          id,
		Name:        name,
		Description: desc,
		Status:      TASK_STATUS_NOTSTARTED,
	}
}

func (tb TaskBase) GetOwner() *NPC {
	return tb.Owner
}
func (tb *TaskBase) SetOwner(n *NPC) {
	tb.Owner = n
}
func (tb TaskBase) GetDescription() string {
	return tb.Description
}
func (tb *TaskBase) SetDescription(desc string) {
	tb.Description = desc
}
func (tb TaskBase) GetID() string {
	return tb.ID
}
func (tb *TaskBase) SetID(id string) {
	tb.ID = id
}
func (tb TaskBase) GetName() string {
	return tb.Name
}
func (tb *TaskBase) SetName(name string) {
	tb.Name = name
}
func (tb TaskBase) GetStatus() TaskStatus {
	return tb.Status
}
func (tb *TaskBase) SetStatus(status TaskStatus) {
	tb.Status = status
}
func (tb TaskBase) IsDone() bool {
	return tb.Status == TASK_STATUS_END
}
func (tb *TaskBase) SetDone() {
	tb.Status = TASK_STATUS_END
}

// does various checks to ensure task status is consistent with other state variables.
func validateStatus(t Task) {
	status := t.GetStatus()
	if !(status == TASK_STATUS_START ||
		status == TASK_STATUS_INPROG ||
		status == TASK_STATUS_END ||
		status == TASK_STATUS_NOTSTARTED) {
		panic("invalid task status:" + status)
	}
	if status == TASK_STATUS_END {
		if t.GetOwner() != nil {
			panic("task.validateStatus: task has ended but still has an owner")
		}
		if !t.IsDone() {
			panic("task.validateStatus: task has ended but 'done' bool is still false")
		}
	} else {
		if t.IsDone() {
			panic("task.validateStatus: task marked as 'done' but hasn't ended yet")
		}
	}
}

func (n *NPC) HandleTaskUpdate() {
	if n.CurrentTask.GetOwner() == nil {
		panic("tried to run task that has no owner set")
	}
	if n.CurrentTask.IsDone() {
		return
	}

	switch n.CurrentTask.GetStatus() {
	case TASK_STATUS_NOTSTARTED:
		n.CurrentTask.Start()
	default:
		n.CurrentTask.Update()
	}

	validateStatus(n.CurrentTask)
}

func (t *TaskBase) HandleNPCCollision(continueFunc func()) {
	if !t.Owner.Entity.Movement.Interrupted {
		return
	}
	logz.Println(t.Owner.DisplayName, "NPC interrupted; handling collision")
	// path entity was moving on has been interrupted.
	// if interrupted by NPC, try to negotiate resolution to collision.
	if !t.Owner.Entity.Movement.TargetTile.Equals(t.Owner.Entity.TilePos) {
		panic("Goto task: since NPC movement was interrupted, we expect its target position to be the same as its current position")
	}
	if len(t.Owner.Entity.Movement.TargetPath) == 0 {
		panic("Goto task: npc movement was interrupted, but there is no next step in target path")
	}

	// get NPCs that are at the next target tile - i.e. the next position in the target path
	nextTarget := t.Owner.Entity.Movement.TargetPath[0]
	collisionNpc, found := t.Owner.World.FindNPCAtPosition(nextTarget)
	if collisionNpc.ID == t.Owner.ID {
		panic("Goto task: npc collided with itself?! that shouldn't be happening")
	}
	if !found {
		// TODO check if NPC collided with player. NPC doesn't collide with player yet.

		// if NPC didn't collide with player, and no collision NPC found, then it seems the blocking NPC is now gone, so proceed
		continueFunc()
		return
	} else {
		logz.Println(t.Owner.DisplayName, "collided with NPC", collisionNpc.DisplayName)
	}
	if collisionNpc.Priority > t.Owner.Priority {
		// other NPC is higher priority; let him go first.
		logz.Println(t.Owner.DisplayName, "waiting for other NPC to go first")
		t.Owner.Wait(time.Second) // wait a second before we check logic again for this NPC
		return
	} else {
		if collisionNpc.Priority == t.Owner.Priority {
			fmt.Println(collisionNpc.DisplayName, collisionNpc.Priority, " / ", t.Owner.DisplayName, t.Owner.Priority)
			panic("updateGoto: other NPC has same priority as this one! that's not suppposed to happen.")
		}
		// this NPC has higher priority, so it can re-route first
		continueFunc()
		return
	}
}
