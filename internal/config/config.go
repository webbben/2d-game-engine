// Package config defines global configuration variables that are used in the game engine
package config

import (
	"time"

	"golang.org/x/image/font"
)

type DefaultBox struct {
	TilesetSrc  string
	OriginIndex int
}

var (
	// how much the game view is scaled up.
	// seems better to keep this a whole number.
	// Using .5's sometimes seems to cause slight distortion in rendered pixels.
	GameScale float64 = 4

	// how much the UI (not in game world) is scaled up.
	// does not affect actual size of resulting GUI, but scales the tiles and images that are used.
	// so, tiles or images used in UI should be created with a consistent tile size (e.g 16px)
	UIScale float64 = 3

	// how much the HUD is scaled up.
	// the HUD refers to things like the clock or player's health bar which are shown on top of the game world (but not part of it)
	HUDScale float64 = 4

	// debug options

	DrawGridLines       = false
	ShowEntityPositions = false // show the logical positions and collision boxes of entities
	ShowCollisions      = false // show the areas that are collisions on the map
	ShowPlayerCoords    = false
	ShowNPCPaths        = false // highlight the paths that NPCs are following
	TrackMemoryUsage    = false // show a report in the console of memory usage every few seconds
	ShowGameDebugInfo   = false // show a report of various debugging info (like F12 in minecraft)

	// misc

	HourSpeed time.Duration = time.Minute // how long it takes for an hour to pass in game

	DefaultFont      font.Face // default font for most body text (e.g. item info tooltips); must be set by game
	DefaultTitleFont font.Face // default font for titles of text areas (e.g. item info tooltips); must be set by game

	// default box used for simple tooltips (e.g. tooltips for tabs); must be set by game
	// (this one is actually required as of now)
	DefaultTooltipBox DefaultBox
	DefaultUIBox      DefaultBox // default box used for UI menus (e.g. the player's inventory menu, etc). required.

	GameDataPathOverride  string = ""   // set this to customize game data root directory location
	GameDataDirectoryName string = "2d" // set this to customize name of game engine data directory

	// custom paths for other data

	EntityDefsDirectory string = "" // this is required to be set in order to load entity data.
)

const (
	TileSize = 16
)

func GetScaledTilesize() float64 {
	return TileSize * UIScale
}
