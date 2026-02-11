package state

import (
	"github.com/webbben/2d-game-engine/data/defs"
)

// DialogProfileState represents a specific DialogProfileDef's state, as the game progresses and the player has interacted with (an NPC using) this dialog profile def.
// It remembers things, like which topics the player has discussed with this dialog profile, or any other conversation context that should be persisted.
//
// Remember:
//
// - DialogProfileDef DEFINES how a profile can behave.
//
// - DialogProfileState records and SAVES the ongoing state of a dialog profile, and its memory of previous conversations.
type DialogProfileState struct {
	ProfileID defs.DialogProfileID

	// "memory" of past dialogs and conversations. what topics have been read, what responses, have been shown, etc. "this happened"
	// This also effectly serves as a "source of truth" for a lot of things, like which topics have been seen, which topics have been unlocked, etc.
	// Note: avoid using directly. better to use the specific functions for setting and reading memory map.
	// This is profile specific - global knowledge of things like world lore topics should be stored in the Player instead.
	Memory map[string]bool
}
