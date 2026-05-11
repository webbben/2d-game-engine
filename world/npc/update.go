package npc

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
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
	if time.Since(n.debug.lastDebugPrint) > 10*time.Second {
		n.debug.lastDebugPrint = time.Now()
		logz.Println(n.ID(), "== DEBUG PRINT ==")
		logz.Println(n.ID(), "IsActive:", n.IsActive())
		if n.CurrentTask != nil {
			logz.Println(n.ID(), "Current Task:", n.CurrentTask.GetName())
			logz.Println(n.ID(), "Status:", n.CurrentTask.GetStatus())
		}
	}

	n.npcUpdates()
	n.Entity.Update()
}

func (mgmt *TaskMGMT) Update() {
	// NOTE: as of now, this function doesn't really do anything besides check for next tasks and/or remove ended tasks.

	if mgmt.CurrentTask != nil {
		if mgmt.CurrentTask.IsDone() {
			// check if there's a chained task
			nextTask := mgmt.CurrentTask.GetNextTaskDef()
			if nextTask != nil {
				mgmt.RunTask(*nextTask, mgmt.CurrentTask.GetOwner())
				return
			}
			// no next task, so disconnect this one.
			mgmt.CurrentTask = nil
			return
		}
	}
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
	mgmt.CurrentTask = nil

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
	if taskDef.TaskID == "" {
		panic("taskID was empty. if this is a 'do nothing' task, use the task ID for that.")
	}

	logz.Println(n.ID(), "attempting to run task:", taskDef.TaskID)

	var t Task

	switch taskDef.TaskID {
	case TaskDoNothing:
		// do nothing tasks are just a way for the schedule to tell an NPC to... do nothing. be frozen in one spot.
		// TODO: should we still make a task for this? it can just be an empty task with no logic in its update functions, of course.
		n.CurrentTask = nil
		return
	case TaskIdle:
		t = NewIdleTask(n, taskDef)
	case TaskLounge:
		t = NewLoungeTask(n, taskDef)
	case TaskSleep:
		// Note: not passing whole task def because sleep task doesn't have things (as of now) like start location and such. that's just assumed to be home map.
		t = NewSleepTask(n, taskDef.Priority)
	case TaskGoto:
		gotoParams, ok := taskDef.Params.(GotoTaskParams)
		if !ok {
			logz.Panicln("RunTask", "tried to run a task, but the params could not be converted into the right type. make sure you are using the right struct.")
		}
		t = NewGotoTask(gotoParams, n, taskDef)
	case TaskStartDialog:
		params, ok := taskDef.Params.(StartDialogTaskParams)
		if !ok {
			logz.Panicln("RunTask", "tried to run a task, but the params could not be converted into the right type. make sure you are using the right struct.")
		}
		t = NewStartDialogTask(params, n, taskDef)
	case TaskBartender:
		t = NewBartenderTask(n, taskDef)
	case TaskShopkeeper:
		t = NewShopkeeperTask(n, taskDef)
	case TaskRoute:
		// we don't plan to allow this as a "top level" task (it's considered a "sub-task" that should be used inside other tasks' logic)
		// TODO: if we decide for sure that a task (like routing) should never be "top level", maybe we should make it private (lowercase) so that schedules can't add it.
		logz.Panicln("TaskMGMT", "This task is not intended to use as a top-level task:", taskDef.TaskID, "If this is a mistake, we can always change that of course.")
	default:
		logz.Panicln("TaskMGMT", "unknown task ID:", taskDef.TaskID)
	}

	// TODO: just add this to a validation function for Task?
	if t.GetOwner() == nil {
		logz.Panicln("SetTask", "task owner was empty; it should've been set in the task creation function")
	}

	if n.CurrentTask != nil {
		// compare priorities
		currentTaskPriority := n.CurrentTask.GetPriority()
		if currentTaskPriority > t.GetPriority() {
			// can't override current task
			logz.Warnln("TaskMGMT", "unable to run task; existing task has higher priority. existing task:", mgmt.CurrentTask.GetID(), "task to run:", t.GetID())
		}
	}

	logz.Println(n.ID(), "setting task:", t.GetID())
	n.CurrentTask = t
}

// Updates related to NPC behavior or tasks
func (n *NPC) npcUpdates() {
	if time.Until(n.waitUntil) > 0 {
		return
	}
	if n.waitUntilDoneMoving {
		if n.Entity.Movement.IsMoving {
			return
		}
		n.waitUntilDoneMoving = false
	}

	n.TaskMGMT.Update()

	if n.IsActive() {
		if n.CurrentTask == nil {
			panic("NPC is marked as active, but there is no current task set")
		}
		n.HandleTaskUpdate()
	}
}
