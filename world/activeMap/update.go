package activemap

import (
	"math"
	"sort"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
)

func (m *ActiveMap) Update(blockPlayerChanges bool) {
	m.blockPlayerChanges = blockPlayerChanges

	m.daylightFader.Update()

	m.refreshPathfindingSnapshot()

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

	// if we have a book session open while in inventory, pause inventory updates
	pausePlayerMenuUpdates := false
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
		pausePlayerMenuUpdates = true
	}

	if m.showPlayerMenu {
		// set last player update to now, so that the time hud doesn't immediately display
		m.PlayerRef.LastUserInput = time.Now()

		if !pausePlayerMenuUpdates {
			m.playerMenuViewer.Update()
			if m.playerMenuViewer.IsDone() {
				logz.Println("", "player menu done")
				m.showPlayerMenu = false
			}
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

	m.blockMapUpdates = blockMapUpdates
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
		// 2026-10-02: verified - keep the inline handling. Object.Update only ever returns
		// UpdateOccurred for step-activated doors: updateGate() and updateContainer() both return an
		// empty ObjectUpdateResult, and activateDoor() (the only Update-reachable path that sets the
		// flag) always sets ChangeMapID at the same time. So for doors HandleObjectUpdate behaves
		// identically to this branch, and swapping gains nothing. The reason to keep it inline is that
		// the else-panic below is a deliberate invariant check: HandleObjectUpdate also handles
		// beds/chairs/signs/lights/items, so routing through it would silently accept one of those
		// types wrongly reporting an update from Update() instead of Activate(), rather than
		// catching it. It also unconditionally logs "Activate Area", which is wrong for a passive
		// per-frame update.
		if result.UpdateOccurred && !blockPlayerChanges {
			if result.ChangeMapID != "" {
				m.worldCtx.HandleMapDoor(result)
				return
			} else {
				logz.Println("Update Map", result)
				logz.Panicln("Update Map", "an object update apparently occurred, but we didn't handle it.")
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

	m.entitySeparation()

	// sort all sortable renderable things on the map
	// do this after all entities have updated, so that in case they've moved, we've properly sorted them for drawing next
	m.updateSortedRenderables()

	m.checkPlayerVisibility()
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

func (m *ActiveMap) entitySeparation() {
	if !config.EntitySeparationEnabled {
		return
	}

	entities := make([]*entity.Entity, 0, len(m.NPCs)+1)
	if m.PlayerRef != nil {
		if !m.PlayerRef.Entity.DisableCollisions && !m.PlayerRef.Entity.IsDead() {
			entities = append(entities, m.PlayerRef.Entity)
		}
	}
	for _, n := range m.NPCs {
		if !n.Entity.DisableCollisions && !n.Entity.IsDead() {
			entities = append(entities, n.Entity)
		}
	}
	if len(entities) < 2 {
		return
	}

	deltas := make([]model.Vec2, len(entities))

	// pre-compute each entity's heading
	headings := make([]model.Vec2, len(entities))
	for i, e := range entities {
		headings[i] = separationHeading(e)
	}

	// compute deltas for how much each entity should be nudged
	for i := range entities {
		entA := entities[i]
		centerA := entA.CollisionRect().GetCenter()
		for j := i + 1; j < len(entities); j++ {
			// if the two entities are walking directly into each other, don't apply any separation push
			// this is to allow them to move through each other and not get stuck, e.g. in a narrow doorway
			// or a position in a path they both need to walk through.
			if headings[i].Dot(headings[j]) < 0 {
				continue
			}

			entB := entities[j]
			centerB := entB.CollisionRect().GetCenter()

			if centerA.Equals(centerB) {
				// centers are exactly the same, so we need to adjust one of them so that a push direction can be resolved.
				// let's just arbitrarily change one of them for now. realistically, this should almost never happen since
				// entities should be getting pushed away from each other when starting to get close anyway
				centerB.X += 1
				centerB.Y += 1
			}

			// use radius rather than collision rect overlap
			// this way, it's actually based on distance and not based on if two square corners happen to be overlapping
			away := centerA.Sub(centerB)
			dist := away.Len()
			if dist >= config.EntitySeparationRadius {
				continue
			}
			// sanity check: make sure dist never ends up as 0
			if dist == 0 {
				logz.PanicCtx(
					"entitySeparation",
					"dist was 0, which can only happen if two entities are in the exact same position. did we fail to handle this edge case?",
					entA.ID(), entB.ID(), centerA, centerB)
			}
			dir := away.Scale(1 / dist)

			// strength scales with how far inside the radius they are
			penetration := config.EntitySeparationRadius - dist
			push := math.Min(penetration*config.EntitySeparationGain, config.EntitySeparationMaxPush)

			// equal and opposite, and accumulated
			step := dir.Scale(push)
			deltas[i] = deltas[i].Add(step)
			deltas[j] = deltas[j].Sub(step)
		}
	}

	// apply deltas to each entity
	for i, e := range entities {
		d := deltas[i]
		if d.X == 0 && d.Y == 0 {
			continue
		}

		// some entities could be getting pushed by multiple other entities, so make sure their accumulated
		// push is still clamped to max push
		d = d.ClampLength(config.EntitySeparationMaxPush)
		e.NudgeBy(d.X, d.Y)
	}
}

// returns the unit vector an entity is currently trying to move along, or a zero vector if not moving.
// i.e. the direction the entity is "heading" towards.
func separationHeading(e *entity.Entity) model.Vec2 {
	if !e.Movement.IsMoving {
		return model.Vec2{}
	}
	toTarget := model.Vec2{X: e.TargetX - e.X, Y: e.TargetY - e.Y}
	l := toTarget.Len()
	if l > 0 {
		return toTarget.Scale(1 / l)
	}
	return model.Vec2{}
}
