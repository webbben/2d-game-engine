package cmd

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity/npc"
)

const ScheduleIdle defs.ScheduleID = "idle"

func GetAllSchedules() []defs.ScheduleDef {
	schedules := []defs.ScheduleDef{
		defs.BuildSchedule(ScheduleIdle, map[int]defs.TaskDef{
			0: {TaskID: npc.TaskIdle},
		}),
	}

	return schedules
}
