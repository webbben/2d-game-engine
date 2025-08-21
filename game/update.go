package game

import (
	"sort"

	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/debug"
)

func (g *Game) Update() error {
	if g.GlobalKeyBindings != nil {
		g.handleGlobalKeyBindings()
	}

	if g.CurrentScreen != nil {
		g.CurrentScreen.UpdateScreen()
		return nil
	}

	if !g.GamePaused {
		g.worldUpdates()
	}

	if config.TrackMemoryUsage {
		debug.UpdatePerformanceMetrics()
	}

	return nil
}

// All "In World" updates happen here - basically anything happening when the player is walking in a room
func (g *Game) worldUpdates() {
	// update dialog if currently in a dialog session
	if g.Conversation != nil {
		if g.Conversation.End {
			// if dialog has ended, remove it from game state
			g.Conversation = nil
		} else {
			g.Conversation.UpdateConversation()
		}
	} else {
		// handle player and npc updates
		g.Player.Update()
		g.MapInfo.updateMap()
	}

	// move camera as needed
	g.Camera.MoveCamera(g.Player.Entity.X, g.Player.Entity.Y)

	// sort entities by Y position for rendering
	// TODO - does this have an impact on performance? If entities got long enough, it seems like it could
	// Maybe we can make a separate goroutine that runs "performance intensive" tasks so the update loop isn't affected
	if len(g.Entities) > 1 {
		sort.Slice(g.Entities, func(i, j int) bool {
			return g.Entities[i].Y < g.Entities[j].Y
		})
	}
}

func (mi *MapInfo) updateMap() {
	for _, n := range mi.NPCs {
		n.Update()
	}
	for _, e := range mi.Entities {
		e.Update()
	}
}
