package defs

import "github.com/webbben/2d-game-engine/data/id"

// SneakXPGainContext contains the contextual information used to determine how much XP a character gains
// while sneaking during a given XP tick
type SneakXPGainContext struct {
	CharacterStateID id.CharacterStateID
	Hidden           bool    // true if character was not detected by any nearby NPC this tick
	NearestNPCDist   float64 // distance (in tiles) to the nearest NPC within sight range
}

// StealthSystemCalc includes all necessary functions to handle stealth-related calculations.
type StealthSystemCalc interface {
	// SneakVisibilityMultiplier returns a multiplier [0, 1] applied to how visible this character is to NPCs,
	// based on their sneak skill. The engine's base factors (view cone of NPC, distance, etc) are also applied.
	SneakVisibilityMultiplier(charStateID id.CharacterStateID, attrs map[AttributeID]int, skills map[SkillID]int) float64
	// SneakXPGain returns how much XP (>= 0) and which skill gains it, for one sneaking tick.
	SneakXPGain(ctx SneakXPGainContext) (skillID SkillID, xp int)
}
