package npc

import (
	"math"
	"testing"

	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/model"
)

const epsilon = 1e-9

func approxEqual(a, b float64) bool {
	return math.Abs(a-b) <= epsilon
}

func testEntityInfo(pos model.Coords, dir byte, sneaking bool) entity.EntityInfo {
	return entity.EntityInfo{
		ID:        "test_entity",
		TilePos:   pos,
		Direction: dir,
		Sneaking:  sneaking,
	}
}

func pos(x, y int) model.Coords {
	return model.Coords{X: x, Y: y}
}

func TestCalculateVisibility(t *testing.T) {
	observer := testEntityInfo(pos(0, 0), 'R', false)

	testCases := []struct {
		name         string
		observer     entity.EntityInfo
		target       entity.EntityInfo
		light        float32
		sneakMult    float64
		sightBlocked bool
		expected     float64
	}{
		{
			name:     "target at max sight distance is invisible",
			observer: observer,
			target:   testEntityInfo(pos(8, 0), 'R', false),
			light:    1.0,
			expected: 0,
		},
		{
			name:     "target beyond sight distance is invisible",
			observer: observer,
			target:   testEntityInfo(pos(9, 0), 'R', false),
			light:    1.0,
			expected: 0,
		},
		{
			name:     "same tile is trivially visible",
			observer: observer,
			target:   testEntityInfo(pos(0, 0), 'R', false),
			light:    1.0,
			expected: 1,
		},
		{
			name:     "in front, in cone, full light",
			observer: observer,
			target:   testEntityInfo(pos(2, 0), 'R', false),
			light:    1.0,
			expected: 1 - 2/SightDist,
		},
		{
			name:     "distance factor scales visibility",
			observer: observer,
			target:   testEntityInfo(pos(3, 0), 'R', false),
			light:    1.0,
			expected: 1 - 3/SightDist,
		},
		{
			name:     "outside the vision cone is reduced",
			observer: observer,
			target:   testEntityInfo(pos(0, 2), 'R', false),
			light:    1.0,
			expected: (1 - 2/SightDist) * OutsideConeFactor,
		},
		{
			name:     "dim light scales visibility",
			observer: observer,
			target:   testEntityInfo(pos(2, 0), 'R', false),
			light:    0.5,
			expected: (1 - 2/SightDist) * 0.5,
		},
		{
			name:     "sneaking target uses sneak multiplier",
			observer: observer,
			target:   testEntityInfo(pos(2, 0), 'R', true),
			light:    1.0,
			sneakMult: 0.3,
			expected: (1 - 2/SightDist) * 0.3,
		},
		{
			name:         "blocked sight uses sight block factor",
			observer:     observer,
			target:       testEntityInfo(pos(2, 0), 'R', false),
			light:        1.0,
			sightBlocked: true,
			expected:     (1 - 2/SightDist) * SightBlockFactor,
		},
		{
			name:     "facing down sees targets below",
			observer: testEntityInfo(pos(0, 0), 'D', false),
			target:   testEntityInfo(pos(0, 2), 'D', false),
			light:    1.0,
			expected: 1 - 2/SightDist,
		},
		{
			name:     "facing up sees targets above",
			observer: testEntityInfo(pos(0, 0), 'U', false),
			target:   testEntityInfo(pos(0, -2), 'U', false),
			light:    1.0,
			expected: 1 - 2/SightDist,
		},
		{
			name:     "facing left sees targets to the left",
			observer: testEntityInfo(pos(0, 0), 'L', false),
			target:   testEntityInfo(pos(-2, 0), 'L', false),
			light:    1.0,
			expected: 1 - 2/SightDist,
		},
	}

	for i, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateVisibility(tc.observer, tc.target, tc.light, tc.sneakMult, tc.sightBlocked)
			if !approxEqual(result, tc.expected) {
				t.Errorf("case %d (%s): calculateVisibility = %v, expected %v", i, tc.name, result, tc.expected)
			}
		})
	}
}