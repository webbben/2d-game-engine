package model

import (
	"fmt"
	"testing"

	"github.com/webbben/2d-game-engine/config"
)

func TestConvertPxToTilePos(t *testing.T) {
	type testcase struct {
		x, y     float64
		expected Coords
	}

	testCases := []testcase{
		{
			x:        0,
			y:        0,
			expected: Coords{0, 0},
		},
		// adding a whole tilesize to 0,0 should produce the next tile over, not the same one.
		// to get the last position within a tile, you add tilesize-1 to x or y.
		{
			x:        config.TileSize,
			y:        config.TileSize,
			expected: Coords{1, 1},
		},
		{
			x:        config.TileSize - 1,
			y:        config.TileSize - 1,
			expected: Coords{0, 0},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test ConvertPxToTilePos: Case %v", i), func(t *testing.T) {
			res := ConvertPxToTilePos(testCase.x, testCase.y)
			t.Log("res:", res)
			t.Log("exp:", testCase.expected)

			if !res.Equals(testCase.expected) {
				t.Errorf("result didn't match expectation. result: %s expected: %s", res, testCase.expected)
			}
		})
	}
}

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
			exp := testCase.expected

			t.Log("res:", res)
			t.Log("exp:", exp)

			if len(res) != len(exp) {
				t.Errorf("number of results doesn't match what we expected. num results: %v num expected: %v", len(res), len(exp))
			}
			for i := range res {
				if !res[i].Equals(exp[i]) {
					t.Errorf("tile in result didn't match expected tile. result tile: %s, exp tile: %s", res[i], exp[i])
				}
			}
		})
	}
}
