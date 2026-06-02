package activemap

import (
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/logz"
)

func (m *ActiveMap) Update(blockPlayerChanges bool) {
	m.daylightFader.Update()

	blockMapUpdates := false

	if m.dialogSession != nil {
		// clear hover targets when in dialog
		m.hoveredObject = nil
		m.hoveredNPC = nil

		m.dialogSession.Update()
		if m.dialogSession.Exit {
			m.dialogSession = nil
		}
		// set last player update to now, so that the time hud doesn't immediately display
		m.PlayerRef.LastUserInput = time.Now()
		// in dialog, so don't allow NPC updates
		blockMapUpdates = true
	}

	if m.bookSession != nil {
		// clear hover targets when in book sesh
		m.hoveredObject = nil
		m.hoveredNPC = nil

		m.bookSession.Update()
		if m.bookSession.IsDone() {
			m.bookSession = nil
		}
		// set last player update to now, so that the time hud doesn't immediately display
		m.PlayerRef.LastUserInput = time.Now()

		// in book sesh, so don't allow NPC updates
		blockMapUpdates = true
	}

	if m.showPlayerMenu {
		// set last player update to now, so that the time hud doesn't immediately display
		m.PlayerRef.LastUserInput = time.Now()
		m.playerMenuViewer.Update()
		if m.playerMenuViewer.IsDone() {
			logz.Println("", "player menu done")
			m.showPlayerMenu = false
		}
		blockMapUpdates = true
	}
	if m.showMiscScreen {
		m.PlayerRef.LastUserInput = time.Now()
		m.miscScreenViewer.Update()
		if m.miscScreenViewer.IsDone() {
			m.showMiscScreen = false
		}
		blockMapUpdates = true
	}

	m.Camera.MoveCamera(m.PlayerRef.Entity.X, m.PlayerRef.Entity.Y)

	if blockMapUpdates {
		return
	}

	m.Map.Update()

	// detect which objects and NPCs are being hovered over by the mouse
	mouseX, mouseY := ebiten.CursorPosition()
	m.hoveredObject = nil

	// get the hover target (from last update loop) and convey that to the objects on this loop
	// TODO: should NPC's have any visual effect when player is hovering? I guess not?
	_, hoverObj := m.GetHoverTarget()

	for _, obj := range m.Objects {
		hovering := hoverObj != nil && hoverObj.ID == obj.ID
		result := obj.Update(blockPlayerChanges, hovering)

		if m.hoveredObject == nil && obj.IsHovering(mouseX, mouseY) {
			m.hoveredObject = obj
		}

		// only handle object update reactions if player is not blocked
		// we do this to prevent accidentally handling map doors twice in a row, when walking on a map door.
		// TODO: why don't we use HandleObjectUpdate here?
		// I guess the only reason we check for object updates here is because door objects with step activation will detect if the player is standing on them,
		// and report a map change.
		if result.UpdateOccurred && !blockPlayerChanges {
			if result.ChangeMapID != "" {
				m.worldCtx.HandleMapDoor(result)
				return
			} else {
				logz.Panicln("Update Map", "an object update apparently occurred, but we didn't handle it.", result)
			}
		}
	}

	m.hoveredNPC = nil

	for _, n := range m.NPCs {
		n.Update()
		if m.hoveredNPC == nil && n.IsHovering(mouseX, mouseY) {
			m.hoveredNPC = n
		}
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
