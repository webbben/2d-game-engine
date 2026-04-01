package activemap

import (
	"sort"
	"time"
)

func (m *ActiveMap) Update(blockPlayerChanges bool) {
	m.daylightFader.Update()

	// if m.cutsceneSession != nil {
	// 	m.updateCutscene()
	// 	blockPlayerChanges = true
	// }

	if m.dialogSession != nil {
		m.dialogSession.Update()
		if m.dialogSession.Exit {
			m.dialogSession = nil
		}
		// set last player update to now, so that the time hud doesn't immediately display
		m.PlayerRef.LastUserInput = time.Now()
		blockPlayerChanges = true
		if m.cutsceneSession == nil {
			// not in a cutscene, so don't allow mid-dialog updates to world
			return
		}
	}

	m.Camera.MoveCamera(m.PlayerRef.Entity.X, m.PlayerRef.Entity.Y)

	m.Map.Update()

	for i := range m.Objects {
		result := m.Objects[i].Update(blockPlayerChanges)
		// only handle object update reactions if player is not blocked
		// we do this to prevent accidentally handling map doors twice in a row, when walking on a map door.
		if result.UpdateOccurred && !blockPlayerChanges {
			if result.ChangeMapID != "" {
				m.worldCtx.HandleMapDoor(result)
				return
			}
		}
	}

	for _, n := range m.NPCs {
		n.Update()
	}

	// sort all sortable renderable things on the map
	// do this after all entities have updated, so that in case they've moved, we've properly sorted them for drawing next
	m.updateSortedRenderables()
}

func (m *ActiveMap) updateSortedRenderables() {
	// Always ensure data isn't nil before adding to this slice
	m.sortedRenderables = []sortedRenderable{}

	for _, n := range m.NPCs {
		m.sortedRenderables = append(m.sortedRenderables, n)
	}

	if m.PlayerRef != nil {
		m.sortedRenderables = append(m.sortedRenderables, m.PlayerRef)
	}

	for _, obj := range m.Objects {
		m.sortedRenderables = append(m.sortedRenderables, obj)
	}

	sort.Slice(m.sortedRenderables, func(i, j int) bool {
		return m.sortedRenderables[i].Y() < m.sortedRenderables[j].Y()
	})
}
