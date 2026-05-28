package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/worldgraph"
)

// RouteTask is a task that sends an NPC to a new map. For simply moving to a new position on the same map, use the Goto task.
// This task works in the following way:
//
// 1) If the NPC is in the same map as the player, and needs to go to a new map, they will go to the next door (via Goto task)
//
// 2) Once the NPC is no longer in the same map as the player, background simulation will give this task progress updates:
//
//   - each map has a certain distance between doors, so based on that distance, a certain number of updates/time will need to elapse until it goes all the way through
//
//   - this is repeated for each map in the path.
//
//     3. If the player enters the map that the NPC is moving through, then the NPC will appear at an approximately correct position in the map, based on current progress,
//     and the NPC will keep walking towards the next door (again, using Goto task)
//
// 4) Eventually, the NPC will be in the correct destination map, and will then walk to the specific position it should be at.
type RouteTask struct {
	TaskBase
	DestinationMapID defs.MapID
	pathCalculated   bool
	worldPath        *worldgraph.WorldPath
	worldPathIndex   int // index of the current node we are at in the world path

	// the path that the NPC would travel on to get to the next map. used for tracking progress with simulation updates, and approximating travel time and current position.
	currentSimPath      []model.Coords
	currentPathProgress float64 // used during simulation to determine progress in moving to the next map

	lastMapID defs.MapID // the last map that the NPC was in - to confirm if changing maps works

	activateDoorTask *ActivateObjectTask

	awaitingMapChange bool       // set when we send a change map occupancy event; causes code to wait until we are confirmed in the next map.
	awaitMapChangeTo  defs.MapID // map we are waiting to change to
}

type RouteTaskParams struct {
	DestinationMapID defs.MapID
}

func NewRouteTask(params RouteTaskParams, owner *NPC, p defs.TaskPriority) *RouteTask {
	if params.DestinationMapID == "" {
		panic("destination map ID was empty")
	}
	if owner.CharacterStateRef.CurrentMap == params.DestinationMapID {
		logz.Panicln("RouteTask", "NPC is already in destination map!", owner.WhoAmI())
	}
	def := defs.TaskDef{
		TaskID:   TaskRoute,
		Priority: p,
	}

	logz.Println("RouteTask", "created new routing task. dest:", params.DestinationMapID, owner.WhoAmI())

	return &RouteTask{
		TaskBase:         NewTaskBase(def, "Route", "Find and follow a route to a new map", owner),
		DestinationMapID: params.DestinationMapID,
	}
}

func (t *RouteTask) BackgroundAssist() {
	t.RecordBgAssist()
	if t.pathCalculated {
		return
	}
	if t.Owner.CharacterStateRef.CurrentMap != t.Owner.WorldCtx.GetActiveMapID() {
		logz.Warnln("RouteTask", "BackgroundAssist was called when NPC is not in the active map. Did the NPC not get properly removed from the active map?")
		return
	}
	// we do the route calculation during background assist, to avoid slowing down the main update loop.
	t.findWorldPath()
}

func (t *RouteTask) findWorldPath() {
	from := t.Owner.CharacterStateRef.CurrentMap
	to := t.DestinationMapID
	if from == to {
		logz.Panicln("RouteTask", "NPC is already at destination map...", t.Owner.WhoAmI(), "dest:", t.DestinationMapID)
	}
	wp, found := t.Owner.WorldCtx.FindWorldPath(from, to)
	if !found {
		logz.Panicln("RouteTask", "failed to find world path to target map. target map:", to, t.Owner.WhoAmI())
	}
	if len(wp.Path) == 0 {
		logz.Panicln("RouteTask", "world path has no length...", t.Owner.WhoAmI(), "dest:", t.DestinationMapID)
	}

	t.worldPath = &wp

	t.pathCalculated = true
}

func (t RouteTask) inDestinationMap() bool {
	return t.Owner.CharacterStateRef.CurrentMap == t.DestinationMapID
}

// SimulationUpdate is what runs during the background NPC simulation loop, instead of Update.
// So, basically Update for the background simulation when an NPC is not in the active map.
func (t *RouteTask) SimulationUpdate() {
	if t.awaitingMapChange {
		t.handleAwaitMapChange()
		return
	}
	if t.Owner.CharacterStateRef.CurrentMap == t.Owner.WorldCtx.GetActiveMapID() {
		// SimulationUpdate shouldn't be getting called...
		logz.Panicln("RouteTask", "SimulationUpdate called when NPC should already be in active map!")
		return
	}
	if t.IsDone() {
		return
	}
	if !t.pathCalculated {
		t.findWorldPath()
		return
	}
	if t.worldPath == nil {
		panic("worldPath was nil")
	}

	// detect end state
	// I think this is where it would be caught if the NPC walks from the active map directly into the destination map.
	if t.inDestinationMap() {
		if t.worldPathIndex < len(t.worldPath.Path) {
			logz.Panicln("RouteTask", "in destination map, but worldPathIndex hasn't past the last path segment. index:", t.worldPathIndex, "len(path):", len(t.worldPath.Path))
		}
		t.Status = TaskEnded
		return
	}

	if t.worldPathIndex == 0 {
		// the first segment doesn't have a calculated map path, since we don't know a specific starting coords for the NPC
		// so, we expect it to be empty.
		if len(t.worldPath.Path[t.worldPathIndex].MapPath) > 0 {
			panic("why is map path calculated?")
		}
		// we skip this step since there isn't a good way to tell how long this segment should take
		t.currentPathProgress = 1
	} else {
		if len(t.currentSimPath) == 0 {
			// this should've been set by the logic below...
			panic("current sim path is empty!")
		}
	}

	if t.currentPathProgress >= 1 {
		// current step progress has finished
		// move the NPC to the next map
		currentMap := t.worldPath.Path[t.worldPathIndex].MapID
		nextEdge := t.worldPath.Path[t.worldPathIndex].NextEdge
		nextMap := nextEdge.To
		toSpawn := nextEdge.ToSpawn

		if nextMap == "" {
			panic("next map was empty...")
		}

		t.Owner.WorldCtx.ChangeMapOccupancyEvent(t.Owner.CharacterStateRef.ID, currentMap, nextMap, toSpawn)

		t.awaitingMapChange = true
		t.awaitMapChangeTo = nextMap
		return
	}

	// we assume a movement speed of 4 tile per tick. this is just roughly how many tiles the player walks at default speed
	// in ~ 1 second / 1 in-game minute, and the simulation loop currently runs every second.
	t.currentPathProgress += float64(4) / float64(len(t.currentSimPath))
}

func (t *RouteTask) handleAwaitMapChange() {
	if !t.awaitingMapChange {
		panic("not awaiting map change?")
	}
	// check if we are now in the next map
	if t.Owner.CharacterStateRef.CurrentMap != t.awaitMapChangeTo {
		// not there yet...
		return
	}

	// we've reached the next map (main Update loop processed the event)
	t.awaitingMapChange = false
	t.awaitMapChangeTo = ""

	// advance to the next path step
	t.worldPathIndex++
	if t.worldPathIndex >= len(t.worldPath.Path) {
		// we've reached the end
		// ensure we are at the destination map
		if !t.inDestinationMap() {
			panic("route task finished, but NPC's current map doesn't match destination map!")
		}
		t.Status = TaskEnded
		return
	}
	if t.inDestinationMap() {
		panic("huh, I didn't think we'd be at the destination at this point, since worldPathIndex is still within range of Path. Does the last goal map have a step too?")
	}
	t.currentSimPath = t.worldPath.Path[t.worldPathIndex].MapPath
	t.currentPathProgress = 0
}

// Update for RouteTask is only done when NPC is in the active map; for NPCs that are routing while in other maps, the SimulationUpdate function is used instead.
func (t *RouteTask) Update() {
	if t.BgAssistUnplugged() {
		logz.Panicln("RouteTask", "Background Assist appears to be unplugged! Whatever task is using this one should make sure to forward calls to BackgroundAssist to RouteTask's BackgroundAssist function.")
	}
	if t.awaitingMapChange {
		t.handleAwaitMapChange()
		return
	}

	if t.Owner.CharacterStateRef.CurrentMap != t.Owner.WorldCtx.GetActiveMapID() {
		logz.Panicln("RouteTask", "Update called for an NPC that isn't in the active map.", t.Owner.WhoAmI())
	}
	if t.IsDone() {
		return
	}
	if !t.pathCalculated {
		// waiting on background assist to handle the work for calculating world path
		return
	}
	if t.worldPath == nil {
		panic("no world path calculated")
	}
	if len(t.worldPath.Path) == 0 {
		panic("world path has no length")
	}

	// Update is only called when the player is in the same map. If so, the NPC would be going to his next position, i.e. a door.
	// If the NPC reaches the target map, then this task is considered completed - since we expect a different task to take over and do something afterwards.

	if t.inDestinationMap() {
		t.Status = TaskEnded
		return
	}

	if t.worldPathIndex >= len(t.worldPath.Path) {
		logz.Panicln("RouteTask", "world path seems to have been traversed (index wise) but apparently the NPC is not at the destination yet...", t.Owner.WhoAmI())
	}

	if t.activateDoorTask != nil {
		t.activateDoorTask.Update()
		if t.activateDoorTask.IsDone() {
			if t.activateDoorTask.Success {
				if t.Owner.CharacterStateRef.CurrentMap == t.lastMapID {
					logz.Panicln("RouteTask", "NPC should've successfully activated a door, but its current map didn't change...", t.Owner.WhoAmI())
				}
				if t.Owner.CharacterStateRef.CurrentMap == t.Owner.WorldCtx.GetActiveMapID() {
					logz.Panicln("RouteTask", "NPC should've successfully activated a door, but its current map is still the active map...", t.Owner.WhoAmI())
				}
				// next, we can expect the code to continue to SimulationUpdate if there are still more places for the NPC to travel
				// set the flags that SimulationUpdate will use to move to next map
				t.awaitMapChangeTo = t.worldPath.Path[t.worldPathIndex].NextEdge.To
				t.awaitingMapChange = true
			} else {
				// failed to activate door... retry?
				logz.Panicln("RouteTask", "NPC failed to activate door:", t.activateDoorTask.FailReason, t.Owner.WhoAmI())
			}
			t.activateDoorTask = nil
		}
		return
	}

	// NPC is in the same map as the player. Tell it to go to the position of the next door and activate it.
	t.goActivateDoor()
	if t.activateDoorTask == nil {
		panic("sanity check: failed to setup active door task?")
	}
}

func (t *RouteTask) goActivateDoor() {
	if !t.pathCalculated {
		panic("path not calculated")
	}
	if t.worldPath == nil {
		panic("no world path calculated")
	}
	if t.Owner.CharacterStateRef.CurrentMap == t.DestinationMapID {
		panic("called GoActivateDoor, but NPC is already in the destination map")
	}
	if t.activateDoorTask != nil {
		panic("activateDoorTask is already set")
	}
	if t.Owner.CharacterStateRef.CurrentMap != t.Owner.WorldCtx.GetActiveMapID() {
		panic("goActivateDoor called, but npc is not in active map!")
	}

	pathNode := t.worldPath.Path[t.worldPathIndex]
	var doorObj *object.Object

	for _, obj := range t.Owner.ActiveMapCtx.GetAllObjects() {
		if obj.Type != object.TypeDoor {
			continue
		}
		if obj.Door.TargetMapID == pathNode.NextEdge.To && obj.Door.TargetSpawnIndex == pathNode.NextEdge.ToSpawn {
			// found the correct door obj
			doorObj = obj
			break
		}
	}
	if doorObj == nil {
		logz.Panicln("RouteTask", "failed to find next door in world path:", pathNode, "\n", t.Owner.WhoAmI())
	}

	t.lastMapID = t.Owner.CharacterStateRef.CurrentMap

	t.activateDoorTask = NewActivateObjectTask(t.Owner, doorObj)
}

// SetupActiveState will only be called when the Routing task is already in progress; the path was calculated,
// and the NPC is theoretically already heading somewhere. Perhaps the player entered the map that the NPC is in, while enroute.
// So, this function should just put the NPC in the right position along his path and setup a GoTo task, and let things be taken over by Update.
func (t *RouteTask) SetupActiveState() {
	if t.Owner.CharacterStateRef.CurrentMap != t.Owner.WorldCtx.GetActiveMapID() {
		logz.Panicln("RouteTask", "SetupActiveState was called, but NPC is not in the current map.", t.Owner.WhoAmI())
	}
	if !t.pathCalculated {
		// I think that if SetupActiveState is ever called for this task, then that means the task should've already been setup and its route calculated,
		// and the player is just encountering an NPC in a map that is on its way to another map.
		// If this function were called and the path hadn't been calculated yet, then that seems to me like a bug, or some dodgy race condition stuff happened with
		// the simulation loop that runs parallel.
		// The reason it should already have its path calculated is, if this task is set then that means some other task activated it to get to its start position.
		// Schedules don't have RouteTask as one of the tasks - at least they shouldn't.
		// So, for now, let's make this a panic, and if we decide its actually acceptable later then we can add some more careful conditions to panic but allow it too.
		// ...
		// On the other hand, it's not really a problem at all to calculate it here. Just thinking that for now, it's better to be strict about logic to avoid bugginess later on.
		logz.Panicln("RouteTask", "SetupActiveState was called, but path wasn't calculated yet, which seems like a bug to me.")
	}
	if t.worldPath == nil {
		panic("world path was nil")
	}
	if len(t.worldPath.Path) == 0 {
		panic("world path was empty")
	}
	if t.currentPathProgress >= 1 {
		logz.Panicln("RouteTask", "current path progress is already >= 1; it should've advanced to the next path segment.", t.Owner.WhoAmI())
	}

	// decide the position to put the NPC along the path in the map
	currentSimPath := t.worldPath.Path[t.worldPathIndex].MapPath
	progressIndex := min(int(max(0, t.currentPathProgress*float64(len(currentSimPath)))), len(currentSimPath)-1)
	pos := currentSimPath[progressIndex]

	t.Owner.Entity.SetPosition(pos)
	// Update can handle setting up the activeDoorTask
}
