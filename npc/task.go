package npc

import (
	"time"
)

type Task struct {
	Owner       *NPC
	Description string

	startFn      func(t *Task)
	isCompleteFn func(t Task) bool
	isFailureFn  func(t Task) bool
	onUpdateFn   func(t *Task)
	endFn        func(t *Task)

	startTime time.Time
	Started   bool
	Timeout   time.Duration

	Context map[string]interface{}

	Done bool
}

func (t *Task) Start() {
	if t.Owner == nil {
		panic("tried to start task that has no owner")
	}
	t.startTime = time.Now()
	if t.startFn != nil {
		t.startFn(t)
	}
	t.Started = true
}

func (t Task) IsComplete() bool {
	if !t.Started {
		return false
	}
	if t.isCompleteFn == nil {
		return false
	}
	return t.isCompleteFn(t)
}

func (t Task) IsFailure() bool {
	if !t.Started {
		return false
	}
	if t.Timeout != 0 {
		if time.Since(t.startTime) > t.Timeout {
			return true
		}
	}
	if t.isFailureFn == nil {
		return false
	}
	return t.isFailureFn(t)
}

func (t *Task) OnUpdate() {
	if !t.Started {
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
	if t.onUpdateFn != nil {
		t.onUpdateFn(t)
	}
}

func (t *Task) EndTask() {
	if t.Owner == nil {
		panic("tried to end a task that has no owner assigned")
	}
	if t.endFn != nil {
		t.endFn(t)
	}
	t.Owner.Active = false
	t.Done = true
}
