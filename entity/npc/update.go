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
	n.Entity.Draw(screen, offsetX, offsetY)
}

func (n *NPC) Update() {
	if time.Since(n.debug.lastDebugPrint) > 10*time.Second {
		n.debug.lastDebugPrint = time.Now()
		logz.Println(n.ID, "== DEBUG PRINT ==")
		logz.Println(n.ID, "IsActive:", n.IsActive())
		if n.CurrentTask != nil {
			logz.Println(n.ID, "Current Task:", n.CurrentTask.GetName())
			logz.Println(n.ID, "Status:", n.CurrentTask.GetStatus())
		}
	}

	n.npcUpdates()
	n.Entity.Update()
}

func (mgmt *TaskMGMT) Update() {
	// check if current task is active, and if we should assign a task based on schedule

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

	if mgmt.CurrentTask == nil {
		// no task assigned; default to schedule
	}
}

func (mgmt *TaskMGMT) RunScheduleTask(hour int, n *NPC) {
	mgmt.CurrentTask = nil

	taskDef := mgmt.Schedule.Hourly[hour]
	if taskDef.TaskID == "" {
		panic("task ID was empty")
	}

	mgmt.RunTask(taskDef, n)
}

func (mgmt *TaskMGMT) RunTask(taskDef defs.TaskDef, n *NPC) {
	if taskDef.TaskID == "" {
		panic("task ID was empty")
	}

	logz.Println(n.ID, "attempting to run task:", taskDef.TaskID)

	var t Task

	switch taskDef.TaskID {
	case TaskGoto:
		gotoParams, ok := taskDef.Params.(GotoTaskParams)
		if !ok {
			logz.Panicln("RunTask", "tried to run a task, but the params could not be converted into the right type. make sure you are using the right struct.")
		}
		t = NewGotoTask(gotoParams, n, taskDef.Priority, taskDef.NextTask)
	case TaskStartDialog:
		params, ok := taskDef.Params.(StartDialogTaskParams)
		if !ok {
			logz.Panicln("RunTask", "tried to run a task, but the params could not be converted into the right type. make sure you are using the right struct.")
		}
		t = NewStartDialogTask(params, n, taskDef.Priority, taskDef.NextTask)
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

	logz.Println(n.ID, "setting task:", t.GetID())
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
