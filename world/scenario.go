package world

import (
	"github.com/webbben/2d-game-engine/config"
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
		logz.Println("SetupMap", w.ActiveMap.MapID, "found in scenario def:", scenarioDef.MapID)
		logz.Panicln("SetupMap", "found queued scenario for map, but mapID in scenario def doesn't match")
	}

	for _, charDef := range scenarioDef.Characters {
		params := entity.NewCharacterStateParams{
			Temp:                    true,
			OverrideDialogProfileID: charDef.DialogProfileID,
			OverrideScheduleID:      charDef.DefaultSchedule,
			InitialMapID:            scenarioDef.MapID,
			LevelSysParams:          w.LevelSysParams,
		}
		charStateID := entity.CreateNewCharacterState(
			charDef.CharDefID,
			params,
			w.Dataman)

		npcParams := npc.NPCParams{
			CharStateID:             charStateID,
			SpeechBubbleTileset:     config.SpeechBubbleBox.TilesetSrc,
			SpeechBubbleOriginIndex: config.SpeechBubbleBox.OriginIndex,
			SpeechBubbleFont:        config.SpeechBubbleFont,
		}
		// TODO: should scenario NPCs be able to use world context?
		n := npc.NewNPC(npcParams, w.Dataman, w.Audioman, w.EventBus, w)

		startPos := model.Coords{X: charDef.SpawnCoordX, Y: charDef.SpawnCoordY}
		w.ActiveMap.AddNPCToMap(n, startPos)
		// set the start position, to make sure a task doesn't put the NPC in a random place (such as what Idle will do, if no position is set)
		startLocation := defs.TaskStartLocation{TileX: &startPos.X, TileY: &startPos.Y, MapID: scenarioDef.MapID}
		if n.CurrentTask == nil {
			logz.Warnln("loadScenario", "just wondering - should an NPC be able to not have a current task at this point?")
			n.SetupTaskState(w.Clock.GetCurrentGameTime(), &startLocation)
		}
		if n.CurrentTask == nil {
			// if current task is still nil, NPC must have the "do nothing" task, so no task was assigned.
			continue
		}
		n.CurrentTask.SetupActiveState()
	}
}
