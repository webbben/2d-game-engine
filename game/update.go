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
	if g.Dialog != nil {
		g.Dialog.Update()
		if g.Dialog.Exit {
			// dialog has ended, so remove it
			g.Dialog = nil
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

	// sort all sortable renderable things on the map
	mi.updateSortedRenderables()

	for _, obj := range mi.Objects {
		obj.Update()
	}

	for _, n := range mi.NPCs {
		n.Update()
	}
}

func (mi *MapInfo) updateSortedRenderables() {
	mi.sortedRenderables = []sortedRenderable{}

	for _, n := range mi.NPCs {
		mi.sortedRenderables = append(mi.sortedRenderables, n)
	}

	mi.sortedRenderables = append(mi.sortedRenderables, mi.PlayerRef)

	for _, obj := range mi.Objects {
		mi.sortedRenderables = append(mi.sortedRenderables, obj)
	}

	sort.Slice(mi.sortedRenderables, func(i, j int) bool {
		return mi.sortedRenderables[i].Y() < mi.sortedRenderables[j].Y()
	})
}
