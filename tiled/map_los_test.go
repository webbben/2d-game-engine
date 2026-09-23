package tiled

import (
	"testing"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/model"
)

// newTestMap returns an in-memory Map with the given tiles marked as vision blockers.
// The map is 10x10, all other tiles are pass-through.
func newTestMap(blockers ...model.Coords) Map {
	m := Map{
		Width:          10,
		Height:         10,
		VisionBlockers: make([][]bool, 10),
	}
	for y := range m.VisionBlockers {
		m.VisionBlockers[y] = make([]bool, 10)
	}
	for _, b := range blockers {
		m.VisionBlockers[b.Y][b.X] = true
	}
	return m
}

func rectForTile(c model.Coords) model.Rect {
	return model.NewRect(float64(c.X*config.TileSize), float64(c.Y*config.TileSize), float64(config.TileSize), float64(config.TileSize))
}

func TestLineOfSightBlocked(t *testing.T) {
	diagFrom := model.Coords{X: 0, Y: 0}
	diagTo := model.Coords{X: 3, Y: 3}

	testCases := []struct {
		name           string
		from, to       model.Coords
		blockers       []model.Coords
		blockingObject []model.Rect
		expected       bool
	}{
		{
			name:     "same tile blocks nothing",
			from:     model.Coords{X: 3, Y: 3},
			to:       model.Coords{X: 3, Y: 3},
			blockers: []model.Coords{{X: 3, Y: 3}},
			expected: false,
		},
		{
			name:     "adjacent tiles never block",
			from:     model.Coords{X: 2, Y: 2},
			to:       model.Coords{X: 3, Y: 2},
			blockers: []model.Coords{{X: 2, Y: 2}, {X: 3, Y: 2}},
			expected: false,
		},
		{
			name:     "clear straight line",
			from:     model.Coords{X: 1, Y: 2},
			to:       model.Coords{X: 6, Y: 2},
			expected: false,
		},
		{
			name:     "blocker in middle of straight line",
			from:     model.Coords{X: 1, Y: 2},
			to:       model.Coords{X: 6, Y: 2},
			blockers: []model.Coords{{X: 3, Y: 2}},
			expected: true,
		},
		{
			name:     "blocker one tile in front of source",
			from:     model.Coords{X: 1, Y: 2},
			to:       model.Coords{X: 6, Y: 2},
			blockers: []model.Coords{{X: 2, Y: 2}},
			expected: true,
		},
		{
			name:     "blocker at source tile is ignored",
			from:     model.Coords{X: 1, Y: 2},
			to:       model.Coords{X: 6, Y: 2},
			blockers: []model.Coords{{X: 1, Y: 2}},
			expected: false,
		},
		{
			name:     "blocker at target tile is ignored",
			from:     model.Coords{X: 1, Y: 2},
			to:       model.Coords{X: 6, Y: 2},
			blockers: []model.Coords{{X: 6, Y: 2}},
			expected: false,
		},
		{
			name:     "blocker in middle when tracing in reverse",
			from:     model.Coords{X: 6, Y: 2},
			to:       model.Coords{X: 1, Y: 2},
			blockers: []model.Coords{{X: 3, Y: 2}},
			expected: true,
		},
		{
			name:     "blocker on the diagonal path",
			from:     diagFrom,
			to:       diagTo,
			blockers: []model.Coords{{X: 1, Y: 1}},
			expected: true,
		},
		{
			name:     "corner tiles beside the diagonal are permissive",
			from:     diagFrom,
			to:       diagTo,
			blockers: []model.Coords{{X: 1, Y: 2}, {X: 2, Y: 1}},
			expected: false,
		},
		{
			name:           "blocking object spanning the middle tile",
			from:           model.Coords{X: 1, Y: 2},
			to:             model.Coords{X: 6, Y: 2},
			blockingObject: []model.Rect{rectForTile(model.Coords{X: 3, Y: 2})},
			expected:       true,
		},
		{
			name:           "blocking object away from the line is ignored",
			from:           model.Coords{X: 1, Y: 2},
			to:             model.Coords{X: 6, Y: 2},
			blockingObject: []model.Rect{rectForTile(model.Coords{X: 8, Y: 8})},
			expected:       false,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestMap(tc.blockers...)
			result := m.LineOfSightBlocked(tc.from, tc.to, tc.blockingObject)
			if result != tc.expected {
				t.Errorf("case %d (%s): LineOfSightBlocked(%v, %v) = %v, expected %v",
					i, tc.name, tc.from, tc.to, result, tc.expected)
			}
		})
	}
}