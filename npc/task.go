package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/internal/model"
)

const (
	// TODO
	// NPC travels to a position.
	TYPE_GOTO = "goto"
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
	Started   bool
	Timeout   time.Duration
	Restart   bool // flag to restart the task over again

	// key/value storage for misc memory between hooks.
	// Only use this if existing task properties and features don't handle the use case.
	Context map[string]interface{}

	Done bool

	// specific task types
	GotoTask
}

type GotoTask struct {
	GoalPos     model.Coords
	IsFailureFn func(t Task)
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
	t.Started = true

	switch t.Type {
	case TYPE_GOTO:
		// start going to the goal position
		t.startGoto()
	}
}

func (t Task) IsComplete() bool {
	if !t.Started {
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
	if !t.Started {
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
	if !t.Started {
		t.Start()
		return
	}
	if t.Restart {
		t.Restart = false
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
	t.Done = true
}
