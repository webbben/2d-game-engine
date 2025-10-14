package game

import (
	"sort"
	"time"

	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/debug"
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
	if g.Dialog != nil {
		g.Dialog.Update(g.EventBus)
		if g.Dialog.Exit {
			// dialog has ended, so remove it
			g.Dialog = nil
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
	if time.Since(g.lastHourChange) > config.HourSpeed {
		newHour := g.Hour + 1
		if newHour > 23 {
			newHour = 0
		}
		g.lastHourChange = time.Now()
		g.SetHour(newHour, false)
	}
	g.daylightFader.Update()

	// move camera as needed
	g.Camera.MoveCamera(g.Player.Entity.X, g.Player.Entity.Y)
}

func (g *Game) updateMap() {
	g.MapInfo.Map.Update()

	// sort all sortable renderable things on the map
	g.MapInfo.updateSortedRenderables()

	for i := range g.MapInfo.Objects {
		result := g.MapInfo.Objects[i].Update()

		if result.ChangeMapID != "" {
			err := g.EnterMap(result.ChangeMapID, nil, result.ChangeMapSpawnIndex)
			if err != nil {
				panic(err)
			}
			return
		}
	}

	for _, n := range g.MapInfo.NPCs {
		n.Update()
	}

	if g.UpdateHooks.UpdateMapHook != nil {
		g.UpdateHooks.UpdateMapHook(g)
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

	sort.Slice(mi.sortedRenderables, func(i, j int) bool {
		return mi.sortedRenderables[i].Y() < mi.sortedRenderables[j].Y()
	})
}
