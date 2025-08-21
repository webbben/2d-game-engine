package npc

import (
	"time"
)

type Task struct {
	Owner       *NPC
	Description string

	StartFn      func(t *Task)
	IsCompleteFn func(t Task) bool
	IsFailureFn  func(t Task) bool
	OnUpdateFn   func(t *Task)
	EndFn        func(t *Task)

	StartTime time.Time
	Started   bool
	Timeout   time.Duration

	Context map[string]interface{}

	Done bool
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
}

func (t Task) IsComplete() bool {
	if !t.Started {
		return false
	}
	if t.IsCompleteFn == nil {
		return false
	}
	return t.IsCompleteFn(t)
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
