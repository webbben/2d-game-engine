package activemap

import "sort"

func (m *ActiveMap) Update() {
	m.daylightFader.Update()

	m.Camera.MoveCamera(m.PlayerRef.Entity.X, m.PlayerRef.Entity.Y)

	m.Map.Update()

	// sort all sortable renderable things on the map
	m.updateSortedRenderables()

	for i := range m.Objects {
		result := m.Objects[i].Update()
		if result.UpdateOccurred {
			if result.ChangeMapID != "" {
				m.worldCtx.HandleMapDoor(result)
				return
			}
		}
	}

	for _, n := range m.NPCs {
		n.Update()
	}
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
