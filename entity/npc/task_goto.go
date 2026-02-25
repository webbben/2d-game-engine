package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/model"
)

type GotoTask struct {
	TaskBase

	goalPos   model.Coords
	isGoingTo bool
}

func (t *GotoTask) BackgroundAssist() {
}

type GotoTaskParams struct {
	TileX, TileY int
}

func NewGotoTask(params GotoTaskParams, owner *NPC, p defs.TaskPriority, nextTask *defs.TaskDef) *GotoTask {
	t := GotoTask{
		TaskBase: NewTaskBase(TaskGoto, "Goto", "Goto a position", owner, p, nextTask),
		goalPos:  model.Coords{X: params.TileX, Y: params.TileY},
	}

	return &t
}

// Start hook for a GoTo task. Just sets the entity on course for the goal position.
func (t *GotoTask) start() {
	t.Status = TaskNotStarted
	if t.goalPos.Equals(t.Owner.Entity.TilePos()) {
		panic("goto task: tried to go to the position the NPC is already at.")
	}
	actualGoal, me := t.Owner.Entity.GoToPos(t.goalPos, true)
	if !me.Success {
		logz.Println(t.Owner.DisplayName, "goto task: failed to call GoToPos:", me)
		t.Owner.Wait(time.Second)
		return
	}
	// since the goal position could've been changed (due to path being blocked), update it here
	t.goalPos = actualGoal
	t.Status = TaskInProg
	if len(t.Owner.Entity.Movement.TargetPath) == 0 {
		panic("started goto, but target path is empty")
	}
	logz.Println(t.Owner.ID, "GotoTask started")
}

func (t *GotoTask) Update() {
	switch t.GetStatus() {
	case TaskNotStarted:
		t.start()
		return
	case TaskInProg:
		result := t.HandleNPCCollision()
		if result.Wait {
			return
		}
		if result.ReRoute {
			t.start()
			return
		}
		if t.isComplete() {
			t.Status = TaskEnded
			return
		}
		if !t.Owner.Entity.Movement.IsMoving {
			if len(t.Owner.Entity.Movement.TargetPath) == 0 {
				// not moving and has no target; shouldn't we have reached our goal then?
				logz.Panicln(t.Owner.ID, "supposed to be going towards a goal, but entity has no target path")
			}
			// entity is not moving, but also is not being blocked (and has a path to follow still).
			// let's just jump start it's path again.
			// TODO: is this the "proper way" to start moving again? seems a bit hacky to manually set the IsMoving flag outside of actual entity logic.
			// Maybe we need a function called "attemptResumePath" or something, which is publicly exposed?
			t.Owner.Entity.Movement.IsMoving = true
		}
	case TaskEnded:
		return
	}
}

func (t GotoTask) isComplete() bool {
	return t.Owner.Entity.TilePos().Equals(t.goalPos)
}
