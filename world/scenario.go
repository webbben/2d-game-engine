package world

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/world/npc"
)

func (w *World) loadScenario(scenarioDef defs.ScenarioDef) {
	if w.ActiveMap == nil {
		panic("map was nil")
	}

	w.ActiveMap.InScenario = true

	logz.Println("Loading Scenario", "MapID:", scenarioDef.MapID, "ScenarioID:", scenarioDef.ID)
	if scenarioDef.MapID != w.ActiveMap.MapID {
		logz.Panicln("SetupMap", "found queued scenario for map, but mapID in scenario def doesn't match. mapID:", w.ActiveMap.MapID, "found in scenario def:", scenarioDef.MapID)
	}

	for _, charDef := range scenarioDef.Characters {
		params := entity.NewCharacterStateParams{
			Temp:                    true,
			OverrideDialogProfileID: charDef.DialogProfileID,
			OverrideScheduleID:      charDef.DefaultSchedule,
			InitialMapID:            scenarioDef.MapID,
		}
		charStateID := entity.CreateNewCharacterState(
			charDef.CharDefID,
			params,
			w.Dataman)
		n := npc.NewNPC(npc.NPCParams{
			CharStateID: charStateID,
		}, w.Dataman, w.Audioman, w.EventBus, w) // TODO: should scenario NPCs be able to use world context?

		startPos := model.Coords{X: charDef.SpawnCoordX, Y: charDef.SpawnCoordY}
		w.ActiveMap.AddNPCToMap(n, startPos)
		// set the start position, to make sure a task doesn't put the NPC in a random place (such as what Idle will do, if no position is set)
		startLocation := defs.TaskStartLocation{TileX: &startPos.X, TileY: &startPos.Y, MapID: scenarioDef.MapID}
		n.SetupTaskState(w.Clock.GetCurrentGameTime(), &startLocation)
		if n.CurrentTask == nil {
			// NPC must have the "do nothing" task, so no task was assigned.
			continue
		}
		n.CurrentTask.SetupActiveState()
	}
}
