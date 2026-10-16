package npc

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/utils"
)

type GotoTask struct {
	TaskBase

	goalPos          model.Coords
	isGoingTo        bool
	unknownCollision int // counter for repeated unknown collisions; used for triggering failure.

	shortArrivals int // counter for times this task ran out of path without reaching its goal.
}

// maxShortArrivals is how many consecutive "ran out of path, never arrived" events end this task.
const maxShortArrivals = 3

type GotoTaskParams struct {
	// TODO: should we just make these model.Coords?
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
		logz.Println(t.Owner.ID(), "goto task: failed to call GoToPos:", me)
		t.Owner.Wait(time.Second)
		return
	}
	// since the goal position could've been changed (due to path being blocked), update it here
	t.goalPos = actualGoal
	t.Status = TaskInProg
	minDist := utils.ManhattanDistCoords(t.Owner.Entity.TilePos(), t.goalPos)
	if minDist > 1 && !t.Owner.Entity.HasPath() {
		// if the goalPos is only one tile away, then GoToPos would end up leaving the target path empty.
		// so only panic here if we know for sure the path shouldn't be empty
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
		// Check for arrival before resolving collisions. Once we're on the goal there's nothing left to
		// route around, and ReRoute would call Start, which rejects a goto to the position we're already
		// on. Reachable: we can be interrupted with our path already consumed the moment we step onto the
		// goal (e.g. the player is standing there, so the next move collides).
		if t.isComplete() {
			t.Finish(TaskResult{Status: ResultSuccess})
			return
		}
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
				t.Finish(TaskResult{Status: ResultFailed, Reason: "gave up after repeated unknown collisions"})
				return
			}
		} else {
			t.unknownCollision = 0
		}
		if !t.Owner.Entity.IsMoving() {
			if !t.Owner.Entity.HasPath() {
				// out of path, standing still, and not at the goal - should this happen?
				//
				// trySetNextTargetPath pops the next (and last, when on the last segment of a path) position from TargetPath after setting it as the next movement target,
				// But there is a case where separation can shift the entity (and its target with it) so it now thinks it's done, but still not at the expected target.
				//
				// re-path, bounded, so a goto that genuinely can't reach its target gives up eventually
				t.shortArrivals++

				ent := t.Owner.Entity
				logz.Println(t.Owner.ID(), "goto task: out of path short of goal.",
					"cur pos:", ent.TilePos(), "goal:", t.goalPos, "lastMoveErr:", ent.Movement.LastMoveError.String(), "lastNudge:", ent.Movement.LastNudgedTick)

				if t.shortArrivals >= maxShortArrivals {
					t.Finish(TaskResult{
						Status: ResultFailed,
						Reason: fmt.Sprintf("ran out of path %d times without reaching the goal (goal %s, stopped at %s)",
							t.shortArrivals, t.goalPos.String(), ent.TilePos().String()),
					})
					return
				}
				t.Start()
				return
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
	// Not possible to get here because GotoTask should only be used by NPCs in an active map; never used in simulation.
	// NPCs routing between maps in simulation would be using the routing task instead.
	panic("never should've come here! (GotoTask only used in active map, never in background simulation)")
}

func (t GotoTask) DisableDefaultSpeechBubbles() bool {
	// NPC shouldn't be talking to you while heading somewhere
	return true
}
