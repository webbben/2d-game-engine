package defs

import (
	"github.com/webbben/2d-game-engine/internal/logz"
)

type (
	TaskID     string
	ScheduleID string
)

// TaskPriority defines how important a task is, and if it can override an ongoing task or not.
// E.g.: "schedule", "assigned", and "emergency" -> assigned tasks override schedule, but emergency tasks override all others.
type TaskPriority int

// A TaskDef defines a single task; the taskID defines which actual task logic to run, and the params are the data to pass into the task logic.
// Note: these are not stored in a centralized place; TaskIDs are defined directly into the npc package, but these are just for getting the params directly with it.
// So, if you want to have a quest or dialog trigger a task, you just create a new TaskDef instance that points to the existing taskID, and passes the necessary
// params with it.
type TaskDef struct {
	TaskID   TaskID
	Priority TaskPriority
	Params   any

	NextTask *TaskDef // OPT: if set, this task will be run right when the parent one finishes
}

type ScheduleDef struct {
	ID     ScheduleID
	Hourly map[int]TaskDef
}

// BuildSchedule is a convenience function for building out an entire schedule, if there are only a few tasks that occur throughout the day.
func BuildSchedule(id ScheduleID, hourlyTasks map[int]TaskDef) ScheduleDef {
	sched := ScheduleDef{
		ID:     id,
		Hourly: make(map[int]TaskDef),
	}

	var lastTaskDef TaskDef
	fillInMorning := false

	for i := range 24 {
		taskDef, exists := hourlyTasks[i]
		if !exists {
			taskDef = lastTaskDef
		} else {
			lastTaskDef = taskDef
		}

		// a task ID wasn't found; that means midnight wasn't set, so we will have to loop back around at the  end.
		if taskDef.TaskID == "" {
			fillInMorning = true
		}

		sched.Hourly[i] = taskDef
	}

	if fillInMorning {
		if lastTaskDef.TaskID == "" {
			panic("is the whole schedule empty?")
		}
		for i := 0; sched.Hourly[i].TaskID == ""; i++ {
			sched.Hourly[i] = lastTaskDef
		}
	}

	sched.Validate()

	return sched
}

func (sched ScheduleDef) Validate() {
	if sched.Hourly == nil {
		panic("sched is nil")
	}
	if len(sched.Hourly) == 0 {
		panic("sched is empty")
	}
	if len(sched.Hourly) != 24 {
		logz.Panicln("ScheduleDef", "schedule is the wrong size (should have 24 entries). size:", len(sched.Hourly))
	}

	for i := range 24 {
		taskDef, exists := sched.Hourly[i]
		if !exists {
			logz.Panicln("ScheduleDef", "hour missing from schedule:", i)
		}
		if taskDef.TaskID == "" {
			logz.Panicln("ScheduleDef", "hour task was empty:", i)
		}
	}
}
