package defs

import "github.com/webbben/2d-game-engine/data/id"

type ObjectType string

// ObjectInfo provides basic information about an object and its current state, mainly intended for things like UI to display.
// Defining it here since we need to be able to reference it inside some of the game contexts, and we can't reference outside packages here in defs.
type ObjectInfo struct {
	ID           int    // ID in tiled map
	DisplayName  string // what the user sees
	Type         ObjectType
	Activatable  bool
	ActivateText string // text that tells you what the activation does (e.g. "Open", "Close", "Pick up", etc)
}

type NPCInfo struct {
	CharID       id.CharacterStateID
	DisplayName  string
	ActivateText string // text that tells you what the activation does (e.g. "Talk", "Pickpocket", etc)
}
