package game

import (
	"sort"

	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/debug"
	"github.com/webbben/2d-game-engine/object"
)

func (g *Game) Update() error {
	if g.GlobalKeyBindings != nil {
		g.handleGlobalKeyBindings()
	}

	if !g.GamePaused {
		g.worldUpdates()
	}

	if config.TrackMemoryUsage {
		debug.UpdatePerformanceMetrics()
	}

	return nil
}

// All "In World" updates happen here - basically anything happening when the player is walking in a map
func (g *Game) worldUpdates() {
	if g.MapInfo == nil {
		// no map to show yet
		return
	}

	// update dialog if currently in a dialog session
	if g.dialogSession != nil {
		g.dialogSession.Update()
		if g.dialogSession.Exit {
			// dialog has ended, so remove it
			g.dialogSession = nil
		}
	} else if g.ShowPlayerMenu {
		g.PlayerMenu.Update()
	} else if g.ShowTradeScreen {
		g.TradeScreen.Update()
		if g.TradeScreen.Exit {
			g.ShowTradeScreen = false
		}
	} else {
		// handle player and npc updates
		g.Player.Update()
		g.updateMap()
	}

	// update time
	if g.Clock.Update() {
		// hour just changed
		_, h, _, _, _, _ := g.Clock.GetCurrentDateAndTime()
		g.OnHourChange(h, false)
	}

	g.daylightFader.Update()

	// move camera as needed
	g.Camera.MoveCamera(g.Player.Entity.X, g.Player.Entity.Y)
}

func (g *Game) handleMapDoor(result object.ObjectUpdateResult) {
	if result.ChangeMapID == "" {
		panic("tried to do map door change, but no map ID is set in object update result")
	}
	err := g.EnterMap(result.ChangeMapID, nil, result.ChangeMapSpawnIndex)
	if err != nil {
		panic(err)
	}
}

func (g *Game) updateMap() {
	g.MapInfo.Map.Update()

	if g.hud != nil {
		g.hud.Update(g)
	}

	// sort all sortable renderable things on the map
	g.MapInfo.updateSortedRenderables()

	for i := range g.MapInfo.Objects {
		result := g.MapInfo.Objects[i].Update()
		if result.UpdateOccurred {
			if result.ChangeMapID != "" {
				g.handleMapDoor(result)
				return
			}
		}
	}

	for _, n := range g.MapInfo.NPCs {
		n.Update()
	}

	if g.UpdateMapHook != nil {
		g.UpdateMapHook(g)
	}
}

func (mi *MapInfo) updateSortedRenderables() {
	// Always ensure data isn't nil before adding to this slice
	mi.sortedRenderables = []sortedRenderable{}

	for _, n := range mi.NPCs {
		mi.sortedRenderables = append(mi.sortedRenderables, n)
	}

	if mi.PlayerRef != nil {
		mi.sortedRenderables = append(mi.sortedRenderables, mi.PlayerRef)
	}

	for _, obj := range mi.Objects {
		mi.sortedRenderables = append(mi.sortedRenderables, obj)
	}

	sort.Slice(mi.sortedRenderables, func(i, j int) bool {
		return mi.sortedRenderables[i].Y() < mi.sortedRenderables[j].Y()
	})
}
