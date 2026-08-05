package npc

import (
	"fmt"
	"sync/atomic"
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
	TaskActivateObj defs.TaskID = "ACTIVATE_OBJECT"
	TaskStartDialog defs.TaskID = "START_DIALOG"
	TaskFaceDir     defs.TaskID = "FACE_DIR" // TODO
	TaskBartender   defs.TaskID = "BARTENDER"
	TaskShopkeeper  defs.TaskID = "SHOPKEEPER"
	TaskGoToTavern  defs.TaskID = "GO_TO_TAVERN"
)

const (
	Schedule defs.TaskPriority = iota
	Assign
	Emergency
)

// taskMeta describes how a task type is constructed and validated.
//
// build constructs the task on the main loop, given its def and owner. For tasks that take params, the
// build closure performs the def.Params type assertion (panicking with a clear message on a mismatch) and
// passes the extracted params to the task's existing constructor.
//
// validateParams performs data-level validation of a task def without needing an NPC at runtime. It is used
// by ValidateTaskDef so that malformed TaskDefs (wrong param struct, impossible param combos) are caught at
// data-validation/CI time rather than mid-playtest. nil means the task takes no params.
type taskMeta struct {
	build          func(def defs.TaskDef, owner *NPC) Task
	validateParams func(def defs.TaskDef) error
}

// taskRegistry is the single source of truth for which task types exist. Each task_*.go file registers its
// own entry in init(), so adding a task means editing one file instead of a shared switch.
var taskRegistry = map[defs.TaskID]taskMeta{}

func registerTask(id defs.TaskID, meta taskMeta) {
	if _, exists := taskRegistry[id]; exists {
		panic("registerTask: task already registered: " + string(id))
	}
	if meta.build == nil {
		panic("registerTask: build fn is nil for task: " + string(id))
	}
	taskRegistry[id] = meta
}

// ValidateTaskDef checks a TaskDef for data correctness: the task ID is known, the params have the right
// concrete type (and pass any cheap data-level rules the task defines), and the start location is
// structurally sound. It recurses through NextTask chains. It returns an error rather than panicking so
// that data validation tooling can collect and report all issues at once.
func ValidateTaskDef(def defs.TaskDef) error {
	if def.TaskID == "" {
		return fmt.Errorf("task ID is empty")
	}
	if def.TaskID == TaskDoNothing {
		// do-nothing tasks have no logic of their own; nothing further to validate.
		return nil
	}
	meta, ok := taskRegistry[def.TaskID]
	if !ok {
		return fmt.Errorf("unknown task ID %q (not registered; child-only tasks like ROUTE/FOLLOW/ACTIVATE_OBJECT can't be top-level)", def.TaskID)
	}
	if meta.validateParams != nil {
		if err := meta.validateParams(def); err != nil {
			return err
		}
	}
	if err := validateTaskStartLocation(def.StartLocation); err != nil {
		return err
	}
	if def.NextTask != nil {
		return ValidateTaskDef(*def.NextTask)
	}
	return nil
}

// validateTaskStartLocation checks the structural sanity of a task's start location. Resolving whether the
// referenced map actually exists is out of scope here (see issue #11), so this only catches conflicts that
// require no runtime context.
func validateTaskStartLocation(sl *defs.TaskStartLocation) error {
	if sl == nil {
		return nil
	}
	if sl.UseHomeMap && sl.MapID != "" {
		return fmt.Errorf("start location sets both UseHomeMap and an explicit MapID; use only one")
	}
	if (sl.TileX == nil) != (sl.TileY == nil) {
		return fmt.Errorf("start location must set both TileX and TileY, or neither")
	}
	return nil
}

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

// TaskResultStatus describes how a task ended. Every task is finished with a result (see TaskBase.Finish).
type TaskResultStatus int

const (
	// ResultUndefined means no result has been recorded yet (the task hasn't finished).
	// It's the zero value so reading Result before Finish is visible as "not finished" rather than looking successful.
	ResultUndefined TaskResultStatus = iota
	// ResultSuccess means the task achieved its goal (Goto reached the tile, dialog ended, bed slept in).
	ResultSuccess
	// ResultFailed means the task genuinely couldn't do it and should not auto-retry.
	ResultFailed
	// ResultAborted means the task was interrupted or invalidated rather than failing (preempted, target died, condition changed).
	ResultAborted
)

func (rs TaskResultStatus) String() string {
	switch rs {
	case ResultUndefined:
		return "UNDEFINED (0)"
	case ResultSuccess:
		return "SUCCESS (1)"
	case ResultFailed:
		return "FAILED (2)"
	case ResultAborted:
		return "ABORTED (3)"
	default:
		return "(Error: task result status has no string representation yet)"
	}
}

// TaskResult is the official outcome of a finished task. Set once, via TaskBase.Finish.
type TaskResult struct {
	Status TaskResultStatus
	Reason string // optional, for logs/telemetry
}

// Task defines an interface that can be used to implement a task. The functions defined in here are only used
// by the "outside logic", so we only need to define functions that the general task management logic would need to access.
// Anything that is task-specific and not needing exposure to the task management system should be left out of this interface.
type Task interface {
	GetID() defs.TaskID
	GetDef() defs.TaskDef

	// GetDeclaredStartLocation returns the start location declared in this task's def (map + optional tile),
	// or nil if the task declares no static start (e.g. dynamic tasks like GO_TO_TAVERN). Sleep tasks with no
	// location resolve to the character's home map. This is used at runtime by navigation (routing to the start
	// map, checking if the NPC is in the start map, standing on a start tile). It has no world knowledge and no
	// notion of the schedule.
	GetDeclaredStartLocation() *defs.TaskStartLocation

	// ResolveStartMap returns the mapID this task places the NPC in for a scheduled hour. It's the scheduling/placement
	// counterpart to GetDeclaredStartLocation: dynamic tasks (no declared location) override it to compute their start
	// map from context, using the anchor (the map the schedule most recently placed the NPC in). Static tasks inherit the
	// TaskBase default, which returns the declared start location (or the anchor if none is declared).
	ResolveStartMap(anchor defs.MapID) defs.MapID

	GetNextTaskDef() *defs.TaskDef

	// the NPC who "owns" this task (i.e. the NPC who is currently running this task)
	GetOwner() *NPC

	GetDescription() string

	GetPriority() defs.TaskPriority

	GetName() string

	GetStatus() TaskStatus // current status of function

	IsDone() bool   // flag that indicates this task is finished or ended. causes no further updates to process.
	IsActive() bool // indicates that the task is currently underway (already started, and hasn't stopped yet)

	// Start is called by the task manager when a task becomes current. It should set up any state the task needs
	// before its first Update, and mark the task in-progress. Tasks with real start logic override this.
	Start()

	// Finish is the single way a task ends (natural completion or preemption). It records the result and runs cleanup.
	// Tasks with resources to release (subscriptions, object targeting, child tasks) override this and call
	// TaskBase.Finish when done. The task manager calls Finish with an Aborted result when it preempts a task.
	Finish(result TaskResult)

	// GetResult returns the outcome of a finished task. Only meaningful once the task is done.
	GetResult() TaskResult

	// logic to execute on each update tick
	Update()

	// provides access to asynchronous work for this task; this is called in the background for tasks that an NPC runs in a map.
	// E.g. calculating routes for an NPC that is chasing someone; it might be bad to hold up the update loop with that kind of work, so we can offload it
	// to another goroutine.
	//
	// Runs on the background jobs goroutine (see switchTask's concurrency contract). It must only read
	// atomic slots and never read/write task Status/Result or NPC/entity/character state.
	BackgroundAssist()

	// allows a task to update while an NPC is not in the current map. Not meant for most tasks, only ones that do things like move an NPC across their path
	// to new maps.
	SimulationUpdate()

	SetupActiveState()

	// modify default NPC behavior while this task is active

	// If true, the default speech bubble activations that run for NPCs will be disabled.
	// Set this to true if the task needs customized speech bubble behavior, or if there should be no speech bubbles at all.
	DisableDefaultSpeechBubbles() bool
}

type TaskBase struct {
	Def         defs.TaskDef
	Owner       *NPC
	Description string
	Name        string
	Status      TaskStatus

	// Result is the official outcome of the task, set by Finish. Only meaningful once the task is done.
	Result TaskResult

	// child holds the current child task, if any (TaskBase manages a single child slot).
	// It is written by the main loop (RunChild/EndChild) and read from the background goroutine via
	// BackgroundAssist's forwarding. It's an atomic slot so a bg-goroutine read can never race a
	// main-loop swap; this is what made the (formerly hand-rolled + mutex-guarded) FightTask follow
	// child safe to convert onto this slot — see task_fight.go.
	child atomic.Value
}

// childRef boxes the child slot's value so the atomic.Value never holds nil (which atomic.Value
// forbids). The child itself may be nil, meaning "no child".
type childRef struct {
	task Task
}

func (tb TaskBase) GetDef() defs.TaskDef {
	return tb.Def
}

func (tb TaskBase) DisableDefaultSpeechBubbles() bool {
	// NOTE: we don't forward to child tasks, because I think it would get too complicated at some
	// point; it's better to decide this at the parent task level, and if custom speech bubble behavior
	// is needed, fully control it at the parent task level rather than passing it to the child.
	// But, of course, individual tasks can decide to forward it if they like.
	return false
}

// ---- child task support ----
//
// A composite task can delegate to a single "child" task via the child slot. The parent starts the child with
// RunChild, observes completion with ChildDone()/ChildResult(), and the child's lifecycle is tied to the
// parent's: when the parent finishes (naturally or by preemption), an active child is ended with an Aborted
// result (see Finish). This replaces the hand-rolled sub-task fields/forwarding each composite used to keep.

// loadChild returns the current child task, or nil. Safe for concurrent use (atomic read); used by
// the background goroutine's forwarding and by the main loop.
func (tb *TaskBase) loadChild() Task {
	ref, _ := tb.child.Load().(*childRef)
	if ref == nil {
		return nil
	}
	return ref.task
}

// storeChild sets the current child task. Main-loop only.
func (tb *TaskBase) storeChild(t Task) {
	tb.child.Store(&childRef{task: t})
}

// RunChild makes the given task this task's single child, ending any existing child first, and starts it.
// Children inherit the parent's priority (a child can never interrupt anything). RunChild is main-loop-owned.
func (tb *TaskBase) RunChild(t Task) {
	if t == nil {
		panic("RunChild: child was nil")
	}
	if cur := tb.loadChild(); cur != nil && !cur.IsDone() {
		cur.Finish(TaskResult{Status: ResultAborted, Reason: "replaced by new child"})
	}
	tb.storeChild(t)
	t.Start()
}

// EndChild ends the current child (if any) with an Aborted result and clears the slot. Safe to call when no
// child exists. This is the generic cleanup-composition path: it runs whenever a parent finishes, so a
// preempted parent's child always gets a chance to release its resources.
func (tb *TaskBase) EndChild() {
	cur := tb.loadChild()
	if cur == nil {
		return
	}
	if !cur.IsDone() {
		cur.Finish(TaskResult{Status: ResultAborted, Reason: "parent ended"})
	}
	tb.storeChild(nil)
}

func (tb *TaskBase) HasChild() bool {
	return tb.loadChild() != nil
}

// ChildDone reports whether there is no current child, or the current child has finished.
func (tb *TaskBase) ChildDone() bool {
	cur := tb.loadChild()
	return cur == nil || cur.IsDone()
}

// ChildResult returns the result of the current child. Only meaningful once the child is done.
// Returns ResultUndefined if there is no child.
func (tb *TaskBase) ChildResult() TaskResult {
	cur := tb.loadChild()
	if cur == nil {
		return TaskResult{Status: ResultUndefined}
	}
	return cur.GetResult()
}

// Update advances the current child, if any. A task with its own per-frame logic should call t.TaskBase.Update()
// at the top of its own Update() and then drive its own state machine off ChildDone()/ChildResult().
// A task that only runs a child needs no Update() of its own.
func (tb *TaskBase) Update() {
	cur := tb.loadChild()
	if cur == nil || cur.IsDone() {
		return
	}
	cur.Update()
}

// BackgroundAssist forwards to the current child. A composite that only runs a child inherits this and needs no
// BackgroundAssist of its own; tasks with their own background work keep their own BackgroundAssist and forward
// as needed. Runs on the background goroutine; the child is read via the atomic slot, and the child's own
// BackgroundAssist must only touch atomic slots (see the concurrency contract in switchTask).
func (tb *TaskBase) BackgroundAssist() {
	cur := tb.loadChild()
	if cur == nil || cur.IsDone() {
		return
	}
	cur.BackgroundAssist()
}

// SimulationUpdate forwards to the current child.
func (tb *TaskBase) SimulationUpdate() {
	cur := tb.loadChild()
	if cur == nil || cur.IsDone() {
		return
	}
	cur.SimulationUpdate()
}

// SetupActiveState forwards to the current child.
func (tb *TaskBase) SetupActiveState() {
	cur := tb.loadChild()
	if cur == nil || cur.IsDone() {
		return
	}
	cur.SetupActiveState()
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

// RouteToStartMap handles routing the NPC to the starting map of this task (as defined in task def) by running a
// RouteTask through the child slot. Returns true if the NPC has reached the starting map. Put this at the top of
// an Update function and return early if it returns false, to ensure the NPC makes its way to the start map
// before running the rest of the task logic.
//
// IMPORTANT: your task still needs to pass BackgroundAssist over to TaskBase (BackgroundAssist is the only way for
// the route calculation to run when the NPC is in the active map, since it's expensive enough to cause lag).
func (tb *TaskBase) RouteToStartMap(simulation bool) (reachedStartMap bool) {
	if tb.InStartMap() {
		// if the routing task made it into the start map last tick, it can be dropped now.
		if child := tb.loadChild(); child != nil && child.GetID() == TaskRoute {
			tb.EndChild()
		}
		return true
	}

	// NPC is not in the start map yet. Set up the routing task as this task's child.
	if !tb.HasChild() {
		startLoc := tb.GetDeclaredStartLocation()
		if startLoc == nil {
			panic("start location was nil!")
		}
		if startLoc.MapID == "" {
			panic("start map was empty!")
		}
		tb.RunChild(NewRouteTask(RouteTaskParams{DestinationMapID: startLoc.MapID}, tb.Owner, tb.GetPriority()))
	}

	child := tb.loadChild()
	if simulation {
		child.SimulationUpdate()
	} else {
		child.Update()
	}

	if tb.ChildDone() {
		// double check that NPC is now in correct map
		if tb.Owner.CharacterStateRef.CurrentMap != tb.GetDeclaredStartLocation().MapID {
			logz.Println("RouteToStartMap", tb.Owner.CharacterStateRef.CurrentMap, tb.GetDeclaredStartLocation().MapID)
			logz.Panicln("RouteToStartMap", "RouteTask seems to be done, but the NPC isn't in the starting map still...")
		}
		tb.EndChild()
		return true
	}

	return false
}

// RouteToStartMapBgAssist forwards a BackgroundAssist call to the routing child, when the NPC is being routed.
// Returns true if the routing task exists and its bg assist was called. This is required, or else if the NPC is in
// the active map they may never get their route calculated (pathing is expensive and deferred to background assist).
func (tb *TaskBase) RouteToStartMapBgAssist() (isRouting bool) {
	cur := tb.loadChild()
	if cur == nil || cur.IsDone() {
		return false
	}
	cur.BackgroundAssist()
	return true
}

// InStartMap tells you if the NPC is in the start map or not
func (tb TaskBase) InStartMap() bool {
	startLoc := tb.GetDeclaredStartLocation()
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

// RouteToStartMapSetupActiveState handles the SetupActiveState for the routing child. Returns true if the routing
// task is setting up an active state, to inform the calling task if it should set up its own active state or not.
func (tb *TaskBase) RouteToStartMapSetupActiveState() (isRouting bool) {
	if tb.InStartMap() {
		return false
	}
	if child := tb.loadChild(); child != nil {
		child.SetupActiveState()
		return true
	}
	// we expect a routing task to have already started if it's getting called at SetupActiveState
	logz.Println("RouteToStartMap", tb.Owner.WhoAmI())
	logz.Panicln("RouteToStartMap", "SetupActiveState called, and NPC not in start map, but routing task wasn't started yet.")
	return false
}

func (tb TaskBase) GetDeclaredStartLocation() *defs.TaskStartLocation {
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
		return nil
	}
	if startLoc.UseHomeMap {
		if startLoc.TileX != nil || startLoc.TileY != nil {
			logz.Println("GetDeclaredStartLocation", tb.Owner.WhoAmI())
			logz.Panicln("GetDeclaredStartLocation", "UseHomeMap set, but TileX or TileY was set too, which seems contradictory or invalid (you don't know which map is home).")
		}
		homeMap := tb.Owner.CharacterStateRef.HomeMapID
		if homeMap == "" {
			logz.Println("GetDeclaredStartLocation", tb.Owner.WhoAmI())
			logz.Panicln("GetDeclaredStartLocation", "no home map found.")
		}
		return &defs.TaskStartLocation{
			MapID: homeMap,
		}
	}
	return startLoc
}

// ResolveStartMap is the TaskBase default: static tasks start wherever they declare (see GetDeclaredStartLocation);
// tasks with no declared start location fall back to the anchor passed in by the scheduler.
func (tb TaskBase) ResolveStartMap(anchor defs.MapID) defs.MapID {
	if loc := tb.GetDeclaredStartLocation(); loc != nil {
		return loc.MapID
	}
	return anchor
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

// Start marks the task as in-progress. Tasks with real start logic (setup, positioning, subscribing) override this.
func (tb *TaskBase) Start() {
	tb.Status = TaskInProg
}

// Finish is the single way a task ends. It records the result, marks the task as ended, and logs the outcome.
// Tasks must always finish with a result; this is what makes "every task produces a result" structural.
func (tb *TaskBase) Finish(result TaskResult) {
	if tb.Status == TaskEnded {
		logz.Println("TaskBase.Finish", "task already ended; finishing again:", tb.Name, "new result:", result.Status)
		return
	}
	// if this task has an active child, end it now (Aborted) so its cleanup runs. This is the generic
	// composition-cleanup path that fixes e.g. the ActivateObjectTask soft-lock when a parent is preempted.
	tb.EndChild()
	tb.Result = result
	tb.Status = TaskEnded
	logz.Println(tb.Owner.DisplayName(), "task finished:", tb.Name, "result:", result.Status, "reason:", result.Reason)
}

func (tb *TaskBase) FinishSuccess() {
	tb.Finish(TaskResult{Status: ResultSuccess})
}

func (tb *TaskBase) FinishFail(reason string) {
	tb.Finish(TaskResult{Status: ResultFailed, Reason: reason})
}

func (tb *TaskBase) FinishAborted(reason string) {
	tb.Finish(TaskResult{Status: ResultAborted, Reason: reason})
}

func (tb *TaskBase) GetResult() TaskResult {
	return tb.Result
}

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

// HandleNPCCollision handles any collision that interrupted the NPC's movement.
// If the NPC ran up against an object, like a gate, it can try to open it.
// Note: entity-vs-entity contact doesn't interrupt movement; entities just slow down and incur more
// "cost" by walking into each other. Interruptions here come from static collidable objects, the
// main case being closed gates (which are excluded from the pathfinding cost map so NPCs route
// through them). Since a GotoTask always moves along a target path, an interruption here is expected
// to have a next path step; this handler only worries about resolving the obstacle.
func (t *TaskBase) HandleNPCCollision() NPCCollisionResult {
	if !t.Owner.Entity.HasStoppedUnexpectedly() {
		return NPCCollisionResult{NoneDetected: true}
	}
	logz.Println(t.Owner.DisplayName(), "NPC interrupted; handling collision")
	nextTarget, ok := t.Owner.Entity.PathAhead(0)
	if !ok {
		panic("Goto task: npc movement was interrupted, but there is no next step in target path")
	}

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

// findTaskArea finds the task-area object for the given task type that the NPC is authorized to use.
// Panics if none is found, since a Bartender/Shopkeeper/etc. task expects its area to exist in the start map.
func findTaskArea(n *NPC, taskID defs.TaskID) *object.Object {
	for _, obj := range n.ActiveMapCtx.GetAllObjects() {
		if obj.Type == object.TypeTaskArea {
			if obj.TaskArea.TaskID == string(taskID) {
				if !n.SatisfiesObjectOwnership(*obj) {
					continue
				}
				return obj
			}
		}
	}
	logz.Println("findTaskArea", taskID, n.WhoAmI())
	logz.Panicln("findTaskArea", "failed to find task area")
	return nil
}
