package model

import (
	"fmt"
	"testing"

	"github.com/webbben/2d-game-engine/config"
)

func TestGetOverlappingTiles(t *testing.T) {
	type getOverlapTilesTestcase struct {
		r        Rect
		expected []Coords
	}

	testCases := []getOverlapTilesTestcase{
		{
			r: Rect{X: 0, Y: 0, W: config.TileSize, H: config.TileSize},
			expected: []Coords{
				{X: 0, Y: 0},
			},
		},
		{
			r: Rect{X: 0, Y: 0, W: config.TileSize + 1, H: config.TileSize + 1},
			expected: []Coords{
				{X: 0, Y: 0},
				{X: 0, Y: 1},
				{X: 1, Y: 0},
				{X: 1, Y: 1},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test GetOverlappingTiles: Case %v", i), func(t *testing.T) {
			res := testCase.r.GetOverlappingTiles()
			t.Log("res:", res)
			t.Log("exp:", testCase.expected)

			if len(res) != len(testCase.expected) {
				t.Errorf("number of results doesn't match what we expected. num results: %v num expected: %v", len(res), len(testCase.expected))
			}
			for i := range res {
				if !res[i].Equals(testCase.expected[i]) {
					t.Errorf("tile in result didn't match expected tile. result tile: %s, exp tile: %s", res[i], testCase.expected[i])
				}
			}
		})
	}
}
