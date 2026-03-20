package world

import (
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/object"
)

func (w *World) Update() {
	// update time
	if w.Clock.Update() {
		// hour just changed
		_, h, _, _, _, _ := w.Clock.GetCurrentDateAndTime()
		w.OnHourChange(h, false, true)
	}

	if w.ActiveMap == nil {
		// TODO: are we loading a map or something? why are we here, but no map exists?
	} else {
		w.Player.Update(w.BlockPlayerChanges)
		w.ActiveMap.Update(w.BlockPlayerChanges)
	}
}

func (w *World) HandleMapDoor(result object.ObjectUpdateResult) {
	if result.ChangeMapID == "" {
		panic("tried to do map door change, but no map ID is set in object update result")
	}
	logz.Println("WORLD", "Handling map door:", result.ChangeMapID, result.ChangeMapSpawnIndex)
	w.EnterMap(result.ChangeMapID, result.ChangeMapSpawnIndex, true)
}
