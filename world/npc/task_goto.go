package npc

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
)

type GotoTask struct {
	TaskBase

	goalPos          model.Coords
	isGoingTo        bool
	unknownCollision int // counter for repeated unknown collisions; used for triggering failure.
}

type GotoTaskParams struct {
	TileX, TileY int
}

func NewGotoTask(params GotoTaskParams, owner *NPC, def defs.TaskDef) *GotoTask {
	if def.TaskID != TaskGoto {
		panic("task def has wrong ID")
	}

	return &GotoTask{
		TaskBase: NewTaskBase(def, "Goto", "Goto a position", owner),
		goalPos:  model.Coords{X: params.TileX, Y: params.TileY},
	}
}

func init() {
	registerTask(TaskGoto, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			params, ok := def.Params.(GotoTaskParams)
			if !ok {
				logz.Println("GotoTask", def.Params)
				logz.Panicln("GotoTask", "tried to run a goto task, but the params could not be converted into GotoTaskParams. make sure you are using the right struct")
			}
			return NewGotoTask(params, owner, def)
		},
		validateParams: func(def defs.TaskDef) error {
			_, ok := def.Params.(GotoTaskParams)
			if !ok {
				return fmt.Errorf("GotoTask params must be GotoTaskParams, got %T", def.Params)
			}
			return nil
		},
	})
}

// Start hook for a GoTo task. Just sets the entity on course for the goal position.
func (t *GotoTask) Start() {
	t.Status = TaskNotStarted
	if t.goalPos.Equals(t.Owner.Entity.TilePos()) {
		logz.Println("GotoTask", t.goalPos, "npcPos:", t.Owner.Entity.TilePos(), t.Owner.WhoAmI())
		logz.Panicln("GotoTask", "tried to go to the position the NPC is already at")
	}
	// if sitting or sleeping, go back to standing
	if t.Owner.Entity.IsSleeping {
		t.Owner.Entity.LeaveBed()
	}
	if t.Owner.Entity.IsSitting {
		t.Owner.Entity.LeaveChair()
	}
	actualGoal, me := t.Owner.Entity.GoToPos(t.goalPos, true)
	if !me.Success {
		logz.Println(t.Owner.DisplayName(), "goto task: failed to call GoToPos:", me)
		t.Owner.Wait(time.Second)
		return
	}
	// since the goal position could've been changed (due to path being blocked), update it here
	t.goalPos = actualGoal
	t.Status = TaskInProg
	if !t.Owner.Entity.HasPath() {
		panic("started goto, but target path is empty")
	}
	t.unknownCollision = 0
	logz.Println(t.Owner.ID(), "GotoTask started")
}

func (t *GotoTask) Update() {
	switch t.GetStatus() {
	case TaskNotStarted:
		t.Start()
		return
	case TaskInProg:
		result := t.HandleNPCCollision()
		if result.Wait {
			return
		}
		if result.ReRoute {
			t.Start()
			return
		}
		if result.UnknownCollision {
			t.unknownCollision++
			if t.unknownCollision > 5 {
				t.FinishFail("gave up after repeated unknown collisions")
				return
			}
		} else {
			t.unknownCollision = 0
		}
		if t.isComplete() {
			t.FinishSuccess()
			return
		}
		if !t.Owner.Entity.IsMoving() {
			if !t.Owner.Entity.HasPath() {
				// not moving and has no target; shouldn't we have reached our goal then?
				logz.Println(t.Owner.ID(), "supposed to be going towards a goal, but entity has no target path")
				logz.Panicln("GotoTask", "supposed to be going towards a goal, but entity has no target path")
			}
			// entity is not moving, but also is not being blocked (and has a path to follow still).
			// jump start its path again.
			t.Owner.Entity.ResumePath()
		}
	case TaskEnded:
		return
	}
}

func (t GotoTask) isComplete() bool {
	return t.Owner.Entity.TilePos().Equals(t.goalPos)
}

func (t *GotoTask) SetupActiveState() {
	// TODO(#99): this is actually reachable for a top-level GotoTask (quest/schedule assigns TaskGoto, and
	// map-loading calls SetupActiveState on the current task). Decide whether to implement a real setup or
	// confirm the path can't happen; tracked before touching this.
	panic("not yet implemented! could this ever be called anyway? i think goto tasks are only created when NPC is in same map as player.")
}

func (t GotoTask) DisableDefaultSpeechBubbles() bool {
	// NPC shouldn't be talking to you while heading somewhere
	return true
}
