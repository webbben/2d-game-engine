package npc

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/utils"
)

const (
	SightDist float64 = 8
)

type debug struct {
	lastDebugPrint time.Time
}

func (n *NPC) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if n.Entity == nil {
		panic("tried to draw NPC that doesn't have an entity!")
	}
	n.Entity.Draw(screen, n.ActiveMapCtx.GetOverlayManager(), offsetX, offsetY)
}

func (n *NPC) Update() {
	// if time.Since(n.debug.lastDebugPrint) > 10*time.Second {
	// 	n.debug.lastDebugPrint = time.Now()
	// 	logz.Println(n.ID(), "== DEBUG PRINT ==")
	// 	logz.Println(n.ID(), "IsActive:", n.IsActive())
	// 	if n.CurrentTask != nil {
	// 		logz.Println(n.ID(), "Current Task:", n.CurrentTask.GetName())
	// 		logz.Println(n.ID(), "Status:", n.CurrentTask.GetStatus())
	// 	}
	// }

	// check if player is nearby/in sight range
	n.playerInSightRange = utils.EuclideanDistCoords(n.WorldCtx.GetPlayerPosition(), n.Entity.TilePos()) < SightDist
	if n.playerInSightRange {
		n.lastPlayerSightingTime = time.Now()
	}

	n.npcUpdates()
	n.Entity.Update()
}

func (mgmt *TaskMGMT) Update(n *NPC) {
	// NOTE: as of now, this function doesn't really do anything besides check for next tasks and/or run the next scheduled task.

	current := mgmt.getCurrentTask()
	if current == nil || !current.IsDone() {
		return
	}

	owner := current.GetOwner()

	// 1. an explicitly chained task (e.g. from a quest) runs before anything else.
	if nextTask := current.GetNextTaskDef(); nextTask != nil {
		mgmt.RunTask(*nextTask, owner)
		return
	}

	// 2. otherwise, decide if the NPC should return to a scheduled task for the current hour, or resume a
	//    task that was preempted earlier. The hourly schedule wins; a preempted task is only resumed for
	//    NPCs that have no scheduled task for this hour.
	hour := n.WorldCtx.GetCurrentGameTime().Hour
	if taskDef, ok := mgmt.Schedule.Hourly[hour]; ok && taskDef.TaskID != "" {
		mgmt.clearInterruptedTask()
		mgmt.RunTask(taskDef, owner)
		return
	}
	if taskDef := mgmt.takeInterruptedTask(); taskDef != nil {
		mgmt.RunTask(*taskDef, owner)
		return
	}

	// 3. nothing to chain, schedule, or resume: just clear the task (do nothing).
	mgmt.clearTask()
}

// OnHourChange handles NPC updates that should occur on hour change. mainly consideration about if scheduled tasks should run.
func (n *NPC) OnHourChange(hour int) {
	nextHourTask := n.Schedule.Hourly[hour]
	if n.CurrentTask == nil || !n.CurrentTask.GetDef().Equals(nextHourTask) {
		logz.Println("OnHourChange", "NPC is changing scheduled task.", n.WhoAmI())
		n.RunScheduleTask(hour, n)
	}
}

func (mgmt *TaskMGMT) RunScheduleTask(hour int, n *NPC) {
	mgmt.clearTask()

	taskDef := mgmt.Schedule.Hourly[hour]

	mgmt.RunTask(taskDef, n)
}

// RunTask is where a task is run for an NPC. These are "top level" tasks that define a fully fleshed chain of logic.
//
// Still considering if we should allow "sub tasks" to be run here.
// On one hand, it's good to only do tasks that are designed to be flexible and have fallback behavior.
// On the other hand, some quests or scenarios might make use of assigning smaller tasks one at a time to get a sequence of behaviors.
// Like how the prison ship scenario goes, where a guard is assigned the Goto task, startdialog task, and goto tasks as a series of chained tasks.
func (mgmt *TaskMGMT) RunTask(taskDef defs.TaskDef, n *NPC) {
	logz.Println(n.ID(), "attempting to run task:", taskDef.TaskID)

	if taskDef.TaskID == TaskDoNothing {
		// do nothing tasks are just a way for the schedule to tell an NPC to... do nothing. be frozen in one spot.
		// Decision: keep the nil shortcut rather than a hollow "do nothing" task. A nil current task naturally
		// falls through to the decision loop (schedule/resume/interrupted) each tick, which is exactly the
		// "frozen" behavior the schedule wants, with no extra task type to maintain.
		mgmt.clearTask()
		return
	}

	mgmt.switchTask(mgmt.buildTask(taskDef, n))
}

// buildTask constructs a task from its def via the registry, without running it. This lets callers that need the
// built task's start map (e.g. NPC placement in initializeNpcWorldState) build once, resolve, place, and then run
// the same task. RunTask is builtTask + switchTask.
func (mgmt *TaskMGMT) buildTask(taskDef defs.TaskDef, n *NPC) Task {
	if taskDef.TaskID == "" {
		panic("taskID was empty. if this is a 'do nothing' task, use the task ID for that.")
	}

	meta, ok := taskRegistry[taskDef.TaskID]
	if !ok {
		logz.Println("TaskMGMT", taskDef.TaskID)
		logz.Panicln("TaskMGMT", "unknown task ID (not registered, or a child-only task like ROUTE/FOLLOW/ACTIVATE_OBJECT)")
	}
	t := meta.build(taskDef, n)

	if t.GetOwner() == nil {
		logz.Panicln("SetTask", "task owner was empty; it should've been set in the task creation function")
	}
	return t
}

// switchTask makes t the NPC's current task, ending any task currently in progress.
//
// This is the single place task switching happens. It holds taskStateMu for the whole transition
// (priority comparison, finishing the old task, starting the new one, and the pointer swap), so
// the background jobs goroutine can never observe a torn switch or a half-started/finished task.
//
// Concurrency contract:
//   - The main game loop is the only writer of CurrentTask, interruptedTask, task Status/Result,
//     and all NPC/entity/character state.
//   - The background jobs goroutine's only entry points into task state are GetCurrentTaskForBgAssist
//     (an RLock-protected pointer read) and Task.BackgroundAssist. BackgroundAssist may read only
//     atomic slots (TaskBase.child, FollowTask.pathRequest/pathResult) and must never read or write
//     task Status/Result or NPC/entity state; anything else that crosses threads should use an atomic
//     mailbox in the FollowTask style.
//   - Known residual reads (pre-existing, not part of this pass): a few bg assists still read
//     main-owned fields directly, e.g. RouteTask.pathCalculated / CharacterStateRef.CurrentMap
//     (RouteTask/SleepTask bg assist) and ActivateObjectTask's set-once gotoTask pointer. These are
//     set-once or single-header reads that don't tear in practice; proper hardening of character-state
//     access belongs to a follow-up pass.
//
// Returns true if the switch happened; false if the current task has a higher priority and the
// new task was rejected.
func (mgmt *TaskMGMT) switchTask(t Task) bool {
	mgmt.taskStateMu.Lock()
	defer mgmt.taskStateMu.Unlock()

	if mgmt.CurrentTask != nil && !mgmt.CurrentTask.IsDone() {
		currentTaskPriority := mgmt.CurrentTask.GetPriority()
		if currentTaskPriority > t.GetPriority() {
			// can't override current task; it has higher priority
			logz.Warnln("TaskMGMT", "unable to run task; existing task has higher priority. existing task:", mgmt.CurrentTask.GetID(), "task to run:", t.GetID())
			return false
		}
		prevDef := mgmt.CurrentTask.GetDef()
		mgmt.interruptedTask = &prevDef
		mgmt.CurrentTask.Finish(TaskResult{Status: ResultAborted, Reason: "preempted by " + string(t.GetID())})
	}

	t.Start()
	mgmt.CurrentTask = t
	logz.Println(t.GetOwner().ID(), "setting task:", t.GetID())
	return true
}

// getCurrentTask returns the current task pointer, or nil. Safe for concurrent use; the background
// jobs goroutine uses this to peek at the task before calling BackgroundAssist.
func (mgmt *TaskMGMT) getCurrentTask() Task {
	mgmt.taskStateMu.RLock()
	defer mgmt.taskStateMu.RUnlock()
	return mgmt.CurrentTask
}

// clearTask nils out the current task. Main-loop only.
func (mgmt *TaskMGMT) clearTask() {
	mgmt.taskStateMu.Lock()
	mgmt.CurrentTask = nil
	mgmt.taskStateMu.Unlock()
}

// clearInterruptedTask clears any recorded preempted-task def. Main-loop only.
func (mgmt *TaskMGMT) clearInterruptedTask() {
	mgmt.taskStateMu.Lock()
	mgmt.interruptedTask = nil
	mgmt.taskStateMu.Unlock()
}

// takeInterruptedTask atomically returns and clears the recorded preempted-task def (if any).
// Main-loop only.
func (mgmt *TaskMGMT) takeInterruptedTask() *defs.TaskDef {
	mgmt.taskStateMu.Lock()
	defer mgmt.taskStateMu.Unlock()
	taskDef := mgmt.interruptedTask
	mgmt.interruptedTask = nil
	return taskDef
}

// Updates related to NPC behavior or tasks
func (n *NPC) npcUpdates() {
	if time.Until(n.waitUntil) > 0 {
		return
	}
	if n.waitUntilDoneMoving {
		if n.Entity.IsMoving() {
			return
		}
		n.waitUntilDoneMoving = false
	}

	n.TaskMGMT.Update(n)

	if n.IsActive() {
		if n.CurrentTask == nil {
			panic("NPC is marked as active, but there is no current task set")
		}
		n.HandleTaskUpdate()
	}
}
