package npc

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

const (
	// the task has not started yet
	TASK_STATUS_NOTSTARTED = "not_yet_started"
	// task has only started, and no update has occurred yet
	TASK_STATUS_START = "start"
	// task has started processing updates
	TASK_STATUS_INPROG = "in_progress"
	// task has ended and is no longer active
	TASK_STATUS_END = "end"
)

type Task struct {
	Owner       *NPC
	Description string
	// Type        string
	ID   string
	Name string

	status string // current status of function
	stop   bool   // if set to true, causes end function to be called
	done   bool   // flag that indicates this task is finished or ended. causes no further updates to process.

	// Optional custom logic for task completion
	IsCompleteFn func(t Task) bool
	// Optional custom logic for task failure
	IsFailureFn func(t Task) bool

	// Optional custom logic for Start hook
	StartFn func(t *Task)
	// Optional custom logic for Update hook
	OnUpdateFn func(t *Task)
	// Optional custom logic for End hook
	EndFn func(t *Task)

	StartTime time.Time
	started   bool // flag that signals this task has started already
	Timeout   time.Duration
	Restart   bool // flag to restart the task over again

	// key/value storage for misc memory between hooks.
	// Only use this if existing task properties and features don't handle the use case.
	Context map[string]interface{}

	// fundamental built-in task logic
	GotoTask
	FollowTask

	// background task assistance
	lastBackgroundAssist time.Time // last time a the bg task loop assisted this task
}

// Returns the current programmatic status of the task.
func (t Task) GetStatus() string {
	t.validateStatus()
	return t.status
}

// does various checks to ensure task status is consistent with other state variables.
func (t Task) validateStatus() {
	if !(t.status == TASK_STATUS_START ||
		t.status == TASK_STATUS_INPROG ||
		t.status == TASK_STATUS_END ||
		t.status == TASK_STATUS_NOTSTARTED) {
		panic("invalid task status:" + t.status)
	}
	if t.status == TASK_STATUS_END {
		if t.Owner != nil {
			panic("task.validateStatus: task has ended but still has an owner")
		}
		if !t.done {
			panic("task.validateStatus: task has ended but 'done' bool is still false")
		}
	} else {
		if t.done {
			panic("task.validateStatus: task marked as 'done' but hasn't ended yet")
		}
	}
}

type GotoTask struct {
	goalPos   model.Coords
	isGoingTo bool
}

type FollowTask struct {
	targetEntity    *entity.Entity
	targetPosition  model.Coords // the last calculated target position
	distance        int          // number of tiles behind the target entity to stand. default, 0, means the tile directly behind.
	isFollowing     bool
	recalculatePath bool // if set, indicates NPC should recalculate the path to the target
}

func (t Task) Copy() Task {
	task := Task{
		Owner:        nil,
		Description:  t.Description,
		StartFn:      t.StartFn,
		IsCompleteFn: t.IsCompleteFn,
		IsFailureFn:  t.IsFailureFn,
		OnUpdateFn:   t.OnUpdateFn,
		EndFn:        t.EndFn,
		Timeout:      t.Timeout,
		Context:      make(map[string]interface{}),

		GotoTask: t.GotoTask,
	}
	return task
}

func (t *Task) start() {
	if t.Owner == nil {
		panic("tried to start task that has no owner")
	}
	t.StartTime = time.Now()
	t.started = true
	t.done = false
	if t.Restart {
		logz.Println(t.Owner.DisplayName, "restarting task")
		t.Restart = false // if restarting, be sure to unset this flag
	}
	t.status = TASK_STATUS_START

	// fundamental task logic
	if t.GotoTask.isGoingTo {
		t.startGoto()
		return
	}
	if t.FollowTask.isFollowing {
		t.startFollow()
		return
	}

	// custom task logic
	if t.StartFn != nil {
		t.StartFn(t)
	}

	// prevent restart loops (in case custom start function causes restarts)
	if t.Restart {
		panic("task.start: start hook is not allowed to trigger restarts! this could cause endless restart loops.")
	}
}

func (t Task) isComplete() bool {
	if !t.started {
		return false
	}

	// fundamental task logic
	if t.GotoTask.isGoingTo {
		return t.isCompleteGoto()
	}

	// custom task logic
	if t.IsCompleteFn != nil {
		return t.IsCompleteFn(t)
	}
	return false
}

func (t Task) isFailure() bool {
	if !t.started {
		return false
	}
	if t.Timeout != 0 {
		if time.Since(t.StartTime) > t.Timeout {
			return true
		}
	}
	if t.IsFailureFn == nil {
		return false
	}
	return t.IsFailureFn(t)
}

func (t *Task) update() {
	if t.done {
		return
	}
	// task state validation
	t.validateStatus()
	t.verifyFundamentalTasks()

	if !t.started || t.Restart {
		t.start()
		return
	}
	if t.isComplete() || t.stop {
		t.end()
		return
	}
	if t.isFailure() || t.stop {
		t.end()
		return
	}

	// fundamental task logic
	t.status = TASK_STATUS_INPROG
	if t.GotoTask.isGoingTo {
		t.updateGoto()
		return
	}
	if t.FollowTask.isFollowing {
		t.updateFollow()
		return
	}

	// if no fundamental task logic, use any customized onUpdate logic
	if t.OnUpdateFn != nil {
		t.OnUpdateFn(t)
		return
	}

	// no update logic found?
	panic("task.update: no update logic executed for this task. tasks are currently required to have an update function.")
}

// it is not allowed for multiple fundamental tasks to be active at once.
// ensure that no more than a single task is currently active.
func (t Task) verifyFundamentalTasks() {
	active := 0
	if t.GotoTask.isGoingTo {
		active++
	}
	if t.FollowTask.isFollowing {
		if active > 0 {
			panic("VerifyTaskGoals: followTask: another task is already active!")
		}
		active++
	}
}

// runs once a task has officially ended. may be used for "wrap up" logic.
// task will be disconnected from owner NPC and set to done, so no further updates will occur.
func (t *Task) end() {
	if t.Owner == nil {
		panic("tried to end a task that has no owner assigned")
	}
	if t.EndFn != nil {
		t.EndFn(t)
	}
	t.Owner.Active = false
	t.Owner = nil // remove owner reference
	t.done = true
	t.status = TASK_STATUS_END
}

func (t *Task) handleNPCCollision(continueFunc func()) {
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

func (t *Task) BackgroundAssist() {
	if time.Since(t.lastBackgroundAssist) < time.Millisecond*500 {
		// last assist was too recent
		return
	}
	if t.FollowTask.isFollowing {
		t.followBackgroundAssist()
	}
}
