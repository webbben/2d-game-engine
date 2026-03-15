// Package world defines functions and logic used for working with the "game world", especially pertaining to generating NPCs, managing town ecosystems, etc.
package world

import (
	"math/rand"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
)

// GenerateCharacter simply generates a new instance of a character, using the given Character Generator.
// Assumes that the character SHOULD be generated, so make sure you do the necessary validation there before deciding to call this.
// (e.g. is the bed associated with this character generator open? if not, you would be instantiating more than one character for a given bed.)
func (w *World) GenerateCharacter(chargen defs.CharacterGenerator, initialMap defs.MapID, homeMap defs.MapID, homeMapBedID int) state.CharacterStateID {
	if initialMap == "" {
		panic("initial map was empty")
	}
	chargen.Validate()

	params := entity.NewCharacterStateParams{
		InitialMapID: initialMap,
		HomeMapID:    homeMap,
		HomeMapBedID: homeMapBedID,
	}

	if chargen.NameGenFn != nil {
		params.OverwriteDisplayName = chargen.NameGenFn()
	}

	// create the character state first, so it exists
	randi := 0
	if len(chargen.CharacterDefIDs) > 0 {
		randi = rand.Intn(len(chargen.CharacterDefIDs))
	}
	charDefID := chargen.CharacterDefIDs[randi]
	if len(chargen.DialogProfileIDs) > 0 {
		randi = rand.Intn(len(chargen.DialogProfileIDs))
		params.OverrideDialogProfileID = chargen.DialogProfileIDs[randi]
	}
	if chargen.ScheduleID != "" {
		params.OverrideScheduleID = chargen.ScheduleID
	}

	return entity.CreateNewCharacterState(charDefID, params, w.Dataman)
}
