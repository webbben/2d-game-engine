package state

import "github.com/webbben/2d-game-engine/data/defs"

type MapState struct {
	ID defs.MapID

	// scenarios that have been queued to run in this map. when this map loads, it will run the scenario at the top of this slice,
	// if any scenario exists. otherwise, it would run it's default behavior.
	QueuedScenarios []defs.ScenarioID
}
