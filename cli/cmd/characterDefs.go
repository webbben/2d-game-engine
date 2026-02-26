package cmd

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

// just creating this file to track all of the actual characters we've created.
// we can use this to validate that all known characters have a valid character def JSON too, which is good.

const (
	Player defs.CharacterDefID = "player" // this is a special one; but used to signify the player when needed.

	// Q001: Awakening

	CharJovePrisonShip     defs.CharacterDefID = "jovis_prisonship"
	CharPrisonShipGuard01  defs.CharacterDefID = "prisonship_guard_01"
	CharQ001ShipCaptain    defs.CharacterDefID = "Q001_harbor_guard"
	CharQ001MiscGuard      defs.CharacterDefID = "Q001_misc_guard"
	CharQ001CustomsOfficer defs.CharacterDefID = "Q001_customs_officer"
)

func ValidateAllCharacterDefs() {
	allDefs := []defs.CharacterDefID{
		CharJovePrisonShip, CharJovePrisonShip,
	}

	for _, charDefID := range allDefs {
		// TODO: check if each charDefID has a matching JSON file
		logz.Println("TODO", "need to add validation:", charDefID)
	}
}
