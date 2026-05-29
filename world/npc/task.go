package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/object"
)

type TaskStatus int

const (
	// the task has not started yet
	TaskNotStarted TaskStatus = iota
	// task has started processing updates
	TaskInProg
	// task has ended and is no longer active
	TaskEnded
)

const (
	TaskDoNothing   defs.TaskID = "DO_NOTHING"
	TaskIdle        defs.TaskID = "IDLE"
	TaskLounge      defs.TaskID = "LOUNGE"
	TaskSleep       defs.TaskID = "SLEEP"
	TaskGoto        defs.TaskID = "GOTO"
	TaskRoute       defs.TaskID = "ROUTE"
	TaskFollow      defs.TaskID = "FOLLOW"
	TaskFight       defs.TaskID = "FIGHT"
	TaskStartDialog defs.TaskID = "START_DIALOG"
	TaskFaceDir     defs.TaskID = "FACE_DIR" // TODO
	TaskBartender   defs.TaskID = "BARTENDER"
	TaskShopkeeper  defs.TaskID = "SHOPKEEPER"
)

const (
	Schedule defs.TaskPriority = iota
	Assign
	Emergency
)

func (ts TaskStatus) String() string {
	switch ts {
	case TaskNotStarted:
		return "NOTSTARTED (0)"
	case TaskInProg:
		return "INPROG (1)"
	case TaskEnded:
		return "END (2)"
	default:
		return "(Error: task status not given a string representation yet)"
	}
}

// Task defines an interface that can be used to implement a task. The functions defined in here are only used
// by the "outside logic", so we only need to define functions that the general task management logic would need to access.
// Anything that is task-specific and not needing exposure to the task management system should be left out of this interface.
type Task interface {
	GetID() defs.TaskID
	GetDef() defs.TaskDef

	GetStartLocation() *defs.TaskStartLocation

	GetNextTaskDef() *defs.TaskDef

	// the NPC who "owns" this task (i.e. the NPC who is currently running this task)
	GetOwner() *NPC

	GetDescription() string

	GetPriority() defs.TaskPriority

	GetName() string

	GetStatus() TaskStatus // current status of function

	IsDone() bool   // flag that indicates this task is finished or ended. causes no further updates to process.
	IsActive() bool // indicates that the task is currently underway (already started, and hasn't stopped yet)

	// logic to execute on each update tick
	Update()

	// provides access to asynchronous work for this task; this is called in the background for tasks that an NPC runs in a map.
	// E.g. calculating routes for an NPC that is chasing someone; it might be bad to hold up the update loop with that kind of work, so we can offload it
	// to another goroutine.
	BackgroundAssist()

	// allows a task to update while an NPC is not in the current map. Not meant for most tasks, only ones that do things like move an NPC across their path
	// to new maps.
	SimulationUpdate()

	SetupActiveState()
}

type TaskBase struct {
	Def         defs.TaskDef
	Owner       *NPC
	Description string
	Name        string
	Status      TaskStatus

	// routing task controlled by TaskBase; other tasks that embed TaskBase should NOT touch this.
	_baseRouting    *RouteTask
	_startTime      *time.Time // records the instant this task received its first update of any kind. used for detecting "unplugged" background assist.
	_bgAssistCalled bool       // records if bg assist has ever been called. used for detecting "unplugged" bg assist.
}

func (tb TaskBase) GetDef() defs.TaskDef {
	return tb.Def
}

func (tb *TaskBase) RecordBgAssist() {
	tb._bgAssistCalled = true
}

// BgAssistUnplugged is used for detecting if background assist is "unplugged" (i.e. not being called for a specific task).
// All you need to do is check this function in your task's Update function and make sure to call RecordBgAssist in your task's BackgroundAssist function.
//
// Note: Do NOT use this in SimulationUpdate, because of course, NPCs in the simulation loop should not be receiving Background Assist anyway.
func (tb *TaskBase) BgAssistUnplugged() bool {
	if tb._startTime == nil {
		now := time.Now()
		tb._startTime = &now
	}

	if tb._bgAssistCalled {
		// if bg assist is ever called, then it should be "plugged in"
		return false
	}

	return time.Since(*tb._startTime) > (time.Second * 10)
}

// NewTaskBase defines a task base that covers all the bases of the Task interface.
//
// nextTask: OPT (only set if you want another task to start right after this one finishes)
func NewTaskBase(def defs.TaskDef, name, desc string, owner *NPC) TaskBase {
	if owner == nil {
		panic("owner was nil")
	}
	return TaskBase{
		Def:         def,
		Name:        name,
		Description: desc,
		Status:      TaskNotStarted,
		Owner:       owner,
	}
}

// RouteToStartMap handles starting and updating a RouteTask to the starting map of this task (as defined in task def).
// Returns true if NPC has reached the starting map. You can put this at the top of an Update function for a task and return if this returns false,
// to ensure that the NPC makes its way to a the start map before starting the rest of the task logic.
//
// IMPORTANT: your task still needs to pass BackgroundAssist over to TaskBase! BackgroundAssist is the only way for routing to calculate its path
// when in the active map (since it's somewhat expensive, in order to avoid lag)
func (tb *TaskBase) RouteToStartMap(simulation bool) (reachedStartMap bool) {
	if tb.InStartMap() {
		if tb._baseRouting != nil {
			// the routing task logic may require one extra update to notice its in the destination map, and so its possible for this condition to notice first.
			// so, just remove it here rather than wasting another update tick to confirm with the routing task logic.
			tb._baseRouting = nil
		}
		return true
	}

	// NPC is not in the start map yet. Setup the routing task!
	if tb._baseRouting == nil {
		startLoc := tb.GetStartLocation()
		if startLoc == nil {
			panic("start location was nil!")
		}
		if startLoc.MapID == "" {
			panic("start map was empty!")
		}
		tb._baseRouting = NewRouteTask(RouteTaskParams{DestinationMapID: startLoc.MapID}, tb.Owner, tb.GetPriority())
	}

	if simulation {
		tb._baseRouting.SimulationUpdate()
	} else {
		tb._baseRouting.Update()
	}

	if tb._baseRouting.IsDone() {
		// double check that NPC is now in correct map
		if tb.Owner.CharacterStateRef.CurrentMap != tb.GetStartLocation().MapID {
			logz.Panicln("RouteToStartMap", "RouteTask seems to be done, but the NPC isn't in the starting map still...", tb.Owner.CharacterStateRef.CurrentMap, tb.GetStartLocation().MapID)
		}
		tb._baseRouting = nil
	}

	return false
}

// RouteToStartMapBgAssist is for forwarding a BackgroundAssist call to the RouteTask that underlies the RouteToStartMap base task.
// This is required, or else if the NPC is in the active map they may never get their route calculated.
// Returns true if the base routing task exists and its BgAssist was called.
func (tb *TaskBase) RouteToStartMapBgAssist() (isRouting bool) {
	if tb._baseRouting == nil {
		return false
	}
	tb._baseRouting.BackgroundAssist()
	return true
}

// InStartMap tells you if the NPC is in the start map or not
func (tb TaskBase) InStartMap() bool {
	startLoc := tb.GetStartLocation()
	if startLoc == nil || startLoc.MapID == "" {
		// nothing set, so we assume true
		return true
	}
	if tb.Owner.CharacterStateRef.CurrentMap == "" {
		logz.Warnln("InStartMap", "NPC we are checking doesn't have a current map set")
	}
	return tb.Owner.CharacterStateRef.CurrentMap == startLoc.MapID
}

func (tb TaskBase) InActiveMap() bool {
	return tb.Owner.WorldCtx.GetActiveMapID() == tb.Owner.CharacterStateRef.CurrentMap
}

// RouteToStartMapSetupActiveState handles the SetupActiveState for any base routing task. Returns true if base routing is setting up an active state,
// to inform the calling task if it should setup its own active state or not.
func (tb *TaskBase) RouteToStartMapSetupActiveState() (isRouting bool) {
	if tb.InStartMap() {
		return false
	}
	if tb._baseRouting == nil {
		// we expect a routing task to have already started if it's getting called at SetupActiveState
		logz.Panicln("RouteToStartMap", "SetupActiveState called, and NPC not in start map, but base routing wasn't started yet.", tb.Owner.WhoAmI())
	}
	tb._baseRouting.SetupActiveState()
	return true
}

func (tb TaskBase) GetStartLocation() *defs.TaskStartLocation {
	startLoc := tb.Def.StartLocation
	if startLoc == nil {
		if tb.Def.TaskID == TaskSleep {
			// sleep tasks have special handling; if no task start location is set, it's assumed to start in the character's home map
			homeMap := tb.Owner.CharacterStateRef.HomeMapID
			if homeMap == "" {
				panic("home map was empty!")
			}
			return &defs.TaskStartLocation{
				MapID: homeMap,
			}
		}
	}
	return startLoc
}

func (tb TaskBase) GetNextTaskDef() *defs.TaskDef {
	return tb.Def.NextTask
}

func (tb TaskBase) GetOwner() *NPC {
	return tb.Owner
}

func (tb TaskBase) GetDescription() string {
	return tb.Description
}

func (tb TaskBase) GetID() defs.TaskID {
	return tb.Def.TaskID
}

func (tb TaskBase) GetPriority() defs.TaskPriority {
	return tb.Def.Priority
}

func (tb TaskBase) GetName() string {
	return tb.Name
}

func (tb TaskBase) GetStatus() TaskStatus {
	return tb.Status
}

func (tb TaskBase) IsDone() bool {
	return tb.Status == TaskEnded
}

func (tb TaskBase) IsActive() bool {
	return tb.Status > TaskNotStarted && tb.Status < TaskEnded
}

// NoBackgroundWork is just a struct that implements the extra, background work related functions (but has them do nothing).
// For simplicity, you can just put this in another Task struct so that you don't have to fill out the empty functions for all of them.
type NoBackgroundWork struct{}

func (x NoBackgroundWork) BackgroundAssist() {}

func (x NoBackgroundWork) SimulationUpdate() {}

type NoActiveState struct{}

func (x NoActiveState) SetupActiveState() {}

func (n *NPC) HandleTaskUpdate() {
	if n.CurrentTask.GetOwner() == nil {
		panic("tried to run task that has no owner set")
	}
	if n.CurrentTask == nil {
		panic("current task is nil?")
	}
	if n.CurrentTask.IsDone() {
		return
	}

	n.CurrentTask.Update()
}

type NPCCollisionResult struct {
	NoneDetected     bool
	UnknownCollision bool
	Wait             bool
	ReRoute          bool
}

// HandleNPCCollision handles any collisions that are occurring.
// Basically, if the NPC has run up into an object, like a gate, they can try to open it.
// Note: entities don't collide, they should just move slower and incur more "cost" by walking into each other.
func (t *TaskBase) HandleNPCCollision() NPCCollisionResult {
	if !t.Owner.Entity.Movement.Interrupted {
		return NPCCollisionResult{NoneDetected: true}
	}
	logz.Println(t.Owner.DisplayName(), "NPC interrupted; handling collision")
	// path entity was moving on has been interrupted.
	// if interrupted by NPC, try to negotiate resolution to collision.
	// TODO: this probably belongs as validation in the entity movement logic itself
	if !t.Owner.Entity.TargetTilePos().Equals(t.Owner.Entity.TilePos()) {
		logz.Println(t.Owner.ID(), "Goto task: since NPC movement was interrupted, we expect its target position to be the same as its current position. target:", t.Owner.Entity.TargetTilePos(), t.Owner.Entity.TilePos())
		panic("Goto task: since NPC movement was interrupted, we expect its target position to be the same as its current position")
	}
	// TODO: same with this?
	if len(t.Owner.Entity.Movement.TargetPath) == 0 {
		panic("Goto task: npc movement was interrupted, but there is no next step in target path")
	}

	nextTarget := t.Owner.Entity.Movement.TargetPath[0]

	collidingObjs := t.Owner.ActiveMapCtx.FindObjectsAtPosition(nextTarget)
	if len(collidingObjs) > 0 {
		// see if any of these objects are things like gates, that can be opened.
		for _, obj := range collidingObjs {
			if !obj.IsCollidable() {
				continue
			}
			if obj.Type == object.TypeGate {
				if !obj.IsCurrentlyActivating() {
					x, y := t.Owner.Entity.X, t.Owner.Entity.Y
					activateParams := object.ObjectActivationParams{
						LockIDs: characterstate.GetLockIDs(*t.Owner.CharacterStateRef, t.Owner.dataman),
					}
					obj.Activate(x, y, activateParams)
					// TODO: it looks like we don't check the result. if we don't have the lock for the gate, then we will need to handle that situation.
				}
				// wait a little for the gate to open
				t.Owner.Wait(time.Second)
				return NPCCollisionResult{Wait: true}
			}
			// found an object that is collidable and cannot be opened or resolved;
			// tell NPC to reroute.
			return NPCCollisionResult{ReRoute: true}
		}
	}
	// no collidable objects found; should be good to continue?
	logz.Println("HandleNPCCollision", "unknown collision. did the collision resolve itself?")
	return NPCCollisionResult{UnknownCollision: true}
}
