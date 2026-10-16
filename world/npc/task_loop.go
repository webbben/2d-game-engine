package npc

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

// LoopTaskParams configures a LOOP task: one task def to run, repeatedly.
type LoopTaskParams struct {
	// Inner is the task to run. Its NextTask chain is followed in order, and once the chain runs out
	// the loop returns to Inner and starts again. So a back-and-forth walk is a two-step chain
	// (GOTO here, GOTO there) rather than bespoke task logic.
	Inner defs.TaskDef
}

// loopMaxConsecutiveFailures caps how many times in a row the inner task may fail before the loop
// gives up and finishes.
//
// Without this, a loop whose inner task can't succeed (an unreachable GOAL, say) would re-run it every
// single frame forever, which looks identical to a movement bug when you're reading logs. Five
// consecutive failures is well clear of a goal being briefly unreachable -- a door opening, an NPC
// briefly in the way -- while still being finite.
//
// Success resets the count, so an inner task that fails occasionally over a long walk is fine.
const loopMaxConsecutiveFailures = 5

// LoopTask runs another task over and over. It exists because "do this again afterwards" is a common
// thing to want -- patrol a route, walk somewhere and back, shuttle between two spots -- and answering
// it with a dedicated task each time means a new task type that is the same shape as the last one.
//
// Why a wrapping task rather than a flag on every task: a task that restarts itself would have to
// remember what it originally was, and that memory is exactly what a NextTask chain destroys. A task
// with a NextTask finishes into something else, so "restart when done" is ambiguous -- restart the task
// itself, or the one it chained to? LOOP keeps that decision outside the inner task, which is the whole
// point of the arrangement.
//
// It deliberately follows the chain itself rather than relying on the task manager. TaskMGMT.Update only
// consults the NextTask of the *top-level* task, so a chained task run as a child would silently drop
// its NextTask on the floor. Flattening the chain up front and driving it from here is what makes the
// whole sequence repeat, rather than just its first step.
type LoopTask struct {
	TaskBase

	// chain is Inner's NextTask chain, flattened into execution order. Built once in Start.
	chain []defs.TaskDef

	// chainIndex is the position in chain of the task currently running as a child.
	chainIndex int

	// consecutiveFailures counts inner-task failures since the last success.
	consecutiveFailures int
}

var _ Task = (*LoopTask)(nil)

func NewLoopTask(params LoopTaskParams, owner *NPC, def defs.TaskDef) *LoopTask {
	if def.TaskID != TaskLoop {
		panic("task def has wrong ID")
	}

	return &LoopTask{
		TaskBase: NewTaskBase(def, "Loop", "Run a task on repeat", owner),
		chain:    flattenTaskChain(params.Inner),
	}
}

func init() {
	registerTask(TaskLoop, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			params, ok := def.Params.(LoopTaskParams)
			if !ok {
				logz.Println("LoopTask", def.Params)
				logz.Panicln("LoopTask", "tried to run a loop task, but the params could not be converted into LoopTaskParams. make sure you are using the right struct")
			}
			return NewLoopTask(params, owner, def)
		},
		validateParams: validateLoopParams,
	})
}

// flattenTaskChain returns def followed by its NextTask chain, in execution order.
//
// A cycle in the chain can't happen through normal data (a chain is built by nesting values, so it ends),
// but the inner def is runtime input here rather than something only data tooling produces, so the loop
// guard is cheap insurance against a nil-chasing infinite loop.
func flattenTaskChain(def defs.TaskDef) []defs.TaskDef {
	chain := []defs.TaskDef{def}
	for cur := def.NextTask; cur != nil; cur = cur.NextTask {
		if len(chain) > maxLoopChainLength {
			logz.Panicln("LoopTask", "inner task chain is longer than", maxLoopChainLength, "steps; a cycle is probably present")
		}
		chain = append(chain, *cur)
	}
	return chain
}

// maxLoopChainLength bounds the chain flatten. Far longer than any legitimate sequence of tasks, and
// only exists so flattenTaskChain terminates on malformed input.
const maxLoopChainLength = 64

func validateLoopParams(def defs.TaskDef) error {
	inner, ok := def.Params.(LoopTaskParams)
	if !ok {
		return fmt.Errorf("LOOP params must be LoopTaskParams, got %T", def.Params)
	}
	if inner.Inner.TaskID == "" {
		return fmt.Errorf("LOOP has an empty inner task ID")
	}
	if inner.Inner.TaskID == TaskLoop {
		// A LOOP wrapping a LOOP never terminates, because the outer loop would restart the inner one
		// forever without ever advancing past it.
		return fmt.Errorf("LOOP cannot wrap another LOOP")
	}
	if _, registered := taskRegistry[inner.Inner.TaskID]; !registered {
		return fmt.Errorf("LOOP inner task %q is not a registered top-level task (child-only tasks like ROUTE/FOLLOW/ACTIVATE_OBJECT can't be run directly)", inner.Inner.TaskID)
	}
	if inner.Inner.StartLocation != nil {
		// StartLocation routes the NPC to another map before the task starts. Repeating that every
		// iteration would bounce the NPC between maps forever, so it's rejected rather than ignored.
		return fmt.Errorf("LOOP inner task %q has a StartLocation; a looping task cannot route to a start map on every iteration", inner.Inner.TaskID)
	}

	// Validates the inner task and recurses through the rest of the chain.
	return ValidateTaskDef(inner.Inner)
}

func (t *LoopTask) Start() {
	t.Status = TaskInProg
	t.chainIndex = 0
	t.consecutiveFailures = 0
	t.startCurrent()
}

func (t *LoopTask) Update() {
	if t.IsDone() {
		return
	}

	if !t.ChildDone() {
		// A child is mid-run; let it advance. TaskBase.Update just forwards Update to it.
		t.TaskBase.Update()
		return
	}

	// The child finished. Decide whether that's a reason to keep going or to stop.
	result := t.ChildResult()
	if result.Status == ResultFailed {
		t.consecutiveFailures++
		if t.consecutiveFailures >= loopMaxConsecutiveFailures {
			t.Finish(TaskResult{
				Status: ResultFailed,
				Reason: fmt.Sprintf("inner task failed %d times in a row (last reason: %q)", t.consecutiveFailures, result.Reason),
			})
			return
		}
	} else if result.Status == ResultSuccess {
		t.consecutiveFailures = 0
	}
	// ResultAborted and ResultUndefined both mean "didn't really run"; they neither count as progress
	// nor count as failure. The loop just advances, which keeps a looping task from being brought down
	// by a single aborted iteration.

	t.chainIndex = (t.chainIndex + 1) % len(t.chain)
	t.startCurrent()
}

// startCurrent builds the task at chainIndex and runs it as a child. RunChild aborts any child already
// in the slot, so calling this with a live child is safe, though Update only reaches it once the child
// is done.
func (t *LoopTask) startCurrent() {
	inner := t.chain[t.chainIndex]
	t.RunChild(t.Owner.TaskMGMT.buildTask(inner, t.Owner))
}
