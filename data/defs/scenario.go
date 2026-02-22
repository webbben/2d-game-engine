package defs

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
type ScenarioDef struct {
	ID    ScenarioID
	MapID MapID

	Characters []ScenarioCharDef
}

type ScenarioCharDef struct {
	// we load a character def and instantiate it into a new, temporary state. we don't load existing character states for scenarios.
	// this is because, scenarios are supposed to be completely cut off from the rest of the "game world", so we don't want anything that happens in it
	// to be influenced by or influence the outside world.
	CharDefID CharacterDefID

	// defines what the character is doing, by default, in the scenario. of course, dialog or quest scripting can override this.
	DefaultSchedule ScheduleID

	// the dialog profile this character should use in this scenario. if this is empty, then whatever is set in the character def will be used.
	DialogProfileID DialogProfileID

	SpawnCoordX, SpawnCoordY int // the spawn point (in tile coords, not abs pixels) for this character
}
