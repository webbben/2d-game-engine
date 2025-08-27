package game

import (
	"slices"

	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/debug"
	"github.com/webbben/2d-game-engine/npc"
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
}

func (mi *MapInfo) updateMap() {
	mi.Map.Update()

	// sort NPCs by Y position so that they render in the right order
	slices.SortFunc(mi.NPCs, func(a *npc.NPC, b *npc.NPC) int {
		return a.Entity.TilePos.Y - b.Entity.TilePos.Y
	})

	for _, n := range mi.NPCs {
		n.Update()
	}
}
