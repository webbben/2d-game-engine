package object

import (
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/tiled"
)

type TaskArea struct {
	TaskID string
	Dir    byte
}

func (o *Object) loadTaskAreaObject(allProps []tiled.Property) {
	taskID, found := tiled.GetStringProperty("task_id", allProps)
	if !found {
		logz.Panicln("TaskArea", "no task_id property found.", o.ID, o.Name)
	}
	o.TaskArea.TaskID = taskID
	taskDir, found := tiled.GetStringProperty("task_dir", allProps)
	if found {
		if taskDir == "" {
			panic("taskdir was empty")
		}
		o.TaskArea.Dir = taskDir[0]
		switch o.TaskArea.Dir {
		case 'L', 'R', 'U', 'D':
			// looks good
		default:
			logz.Panicln("TaskArea", "task_dir was set, but not a valid task char:", o.TaskArea.Dir)
		}
	} else {
		o.TaskArea.Dir = 'D'
	}
}
