package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/world"
)

// RouteTask is a task that sends an NPC to a new map. For simply moving to a new position on the same map, use the Goto task.
type RouteTask struct {
	TaskBase
	DestinationMapID defs.MapID
	pathCalculated   bool
}

type RouteTaskParams struct {
	DestinationMapID defs.MapID
}

func NewRouteTask(params RouteTaskParams, owner *NPC, p defs.TaskPriority, nextTask *defs.TaskDef) *RouteTask {
	if params.DestinationMapID == "" {
		panic("destination map ID was empty")
	}
	def := defs.TaskDef{
		TaskID:   TaskRoute,
		NextTask: nextTask,
		Priority: p,
	}

	return &RouteTask{
		TaskBase:         NewTaskBase(def, "Route", "Find and follow a route to a new map", owner),
		DestinationMapID: params.DestinationMapID,
	}
}

func (t *RouteTask) BackgroundAssist(wg *world.WorldGraph) {
	if t.pathCalculated {
		return
	}
	// we do the route calculation during background assist, to avoid slowing down the main update loop.
	logz.TODO("Route task", "call world graph to calculate path to destination map")

	t.pathCalculated = true
}

func (t *RouteTask) SimulationUpdate(wg *world.WorldGraph) {
	// TODO: do this
}

func (t *RouteTask) Update() {
	if !t.pathCalculated {
		return
	}
	logz.TODO("Route task", "need to implement actual logic for this still. for now, leaving it empty.")
}
