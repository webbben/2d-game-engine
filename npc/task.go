package npc

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

const (
	// TODO
	// NPC travels to a position.
	TYPE_GOTO = "goto"
	// TODO
	// NPC follows another entity.
	TYPE_FOLLOW = "follow"
	// TODO
	// NPC fights another entity.
	TYPE_FIGHT = "fight"
	// TODO
	// NPC runs away from another entity.
	TYPE_EVADE = "evade"
	// TODO
	// NPC does some miscellaneous activity, mostly consisting of an animation.
	// Can have side-effects too.
	TYPE_ACTIVITY = "activity"
)

type Task struct {
	Owner       *NPC
	Description string
	Type        string

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

	done bool // flag that signals this task is done. stops execution of hook functions.

	// specific task types
	GotoTask
	FollowTask
}

type GotoTask struct {
	GoalPos model.Coords
}

type FollowTask struct {
	TargetEntity *entity.Entity
	Distance     int // number of tiles behind the target entity to stand. default, 0, means the tile directly behind.
}

func (t Task) Copy() Task {
	task := Task{
		Owner:        nil,
		Type:         t.Type,
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

func (t *Task) Start() {
	if t.Owner == nil {
		panic("tried to start task that has no owner")
	}
	t.StartTime = time.Now()
	if t.StartFn != nil {
		t.StartFn(t)
	}
	t.started = true
	t.done = false

	switch t.Type {
	case TYPE_GOTO:
		t.startGoto()
	case TYPE_FOLLOW:
		t.startFollow()
	}
}

func (t Task) IsComplete() bool {
	if !t.started {
		return false
	}
	if t.IsCompleteFn != nil {
		return t.IsCompleteFn(t)
	}
	switch t.Type {
	case TYPE_GOTO:
		return t.isCompleteGoto()
	}
	return false
}

func (t Task) IsFailure() bool {
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

func (t *Task) OnUpdate() {
	if t.done {
		return
	}
	if !t.started {
		t.Start()
		return
	}
	if t.Restart {
		t.Restart = false
		logz.Println(t.Owner.DisplayName, "restarting task")
		t.Start()
		return
	}
	if t.IsComplete() {
		t.EndTask()
		return
	}
	if t.IsFailure() {
		t.EndTask()
		return
	}
	if t.OnUpdateFn != nil {
		t.OnUpdateFn(t)
	}
	switch t.Type {
	case TYPE_GOTO:
		t.updateGoto()
	case TYPE_FOLLOW:
		t.updateFollow()
	}
}

func (t *Task) EndTask() {
	if t.Owner == nil {
		panic("tried to end a task that has no owner assigned")
	}
	if t.EndFn != nil {
		t.EndFn(t)
	}
	t.Owner.Active = false
	t.done = true
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
