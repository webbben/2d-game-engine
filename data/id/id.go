// Package id defines all IDs used in the game engine
package id

// TODO: move over other popular ID types that are used in the game engine. this package should be the central place for all ID types to be defined.

const (
	PlayerStateID CharacterStateID = "player"
	PlayerDefID   CharacterDefID   = "player"
)

type (
	CharacterStateID string
	CharacterDefID   string
)
