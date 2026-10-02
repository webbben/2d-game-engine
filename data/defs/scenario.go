package defs

import (
	"github.com/webbben/2d-game-engine/data/id"
)

type ScenarioID string

// A ScenarioDef defines what to put in a map for a specific scenario.
// A "scenario" is a special, controlled situation that happens in a map. It basically controls things like:
//
// - Which characters/NPCs appear, and where
//
// - What tasks or schedules the NPCs have
//
// - What dialog profiles the NPCs are using
//
// So, it pretty much defines a set of NPCs to show in the map, and what their behavior is.
// One reason you might have a scenario is that a quest requires it; you need to have specific characters (and only those characters)
// in a map at a specific time, and you don't want any of the "outside world" to possibly interfere.
//
// Ordinarily, a city map will not have a scenario running in it, and so the various NPCs that inhabit that city would be going about
// their schedules, wandering here and there, etc. But a schedule allows you to use an existing map but for a special, temporary purpose.
//
// It also gives you the ability to play cutscenes. This is probably going to end up being one of the main reasons to run a scenario,
// since in the long run I don't anticipate needing so many custom scenarios in the game, but surely various quests will benefit from triggering cutscenes.
type ScenarioDef struct {
	ID    ScenarioID
	MapID MapID

	// if set to true, instead of playing in this scenario, the map will load in the "regular world".
	// this is used when you just want a specific cutscene to play, but the player shouldn't be able to enter that cutscene's "world" and play.
	ResumeRegularWorld bool

	Characters []ScenarioCharDef
}

type ScenarioCharDef struct {
	// we load a character def and instantiate it into a new, temporary state. we don't load existing character states for scenarios.
	// this is because, scenarios are supposed to be completely cut off from the rest of the "game world", so we don't want anything that happens in it
	// to be influenced by or influence the outside world.
	CharDefID id.CharacterDefID

	// defines what the character is doing. if not set, the character will just stand there, doing nothing.
	DefaultSchedule ScheduleID

	// the dialog profile this character should use in this scenario. if this is empty, then whatever is set in the character def will be used.
	DialogProfileID DialogProfileID

	SpawnCoordX, SpawnCoordY int  // the spawn point (in tile coords, not abs pixels) for this character
	SpawnDirection           byte // allows you to set which direction the NPC is facing; defaults to down. Mainly for characters that have no task and just stand in one place.
}
