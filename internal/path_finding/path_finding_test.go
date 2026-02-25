package path_finding

import (
	"fmt"
	"math/rand"
	"testing"

	m "github.com/webbben/2d-game-engine/model"
)

type FindPathTestCase struct {
	Start, Goal  m.Coords
	ExpectedPath []m.Coords
	CostMap      [][]int
	FoundRoute   bool
}

func pathIsSame(pathA []m.Coords, pathB []m.Coords) bool {
	if len(pathA) != len(pathB) {
		return false
	}
	for i := 0; i < len(pathA); i++ {
		if pathA[i] != pathB[i] {
			return false
		}
	}
	return true
}

func TestFindPath(t *testing.T) {
	costMap := [][]int{
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0},
		{0, 0, 0, 10, 10},
		{0, 0, 0, 0, 0},
		{0, 10, 0, 0, 0},
	}

	testCases := []FindPathTestCase{
		{
			Start: m.Coords{X: 0, Y: 0},
			Goal:  m.Coords{X: 4, Y: 0},
			ExpectedPath: []m.Coords{
				{X: 1, Y: 0},
				{X: 2, Y: 0},
				{X: 3, Y: 0},
				{X: 4, Y: 0},
			},
			CostMap:    costMap,
			FoundRoute: true,
		},
		{
			Start: m.Coords{X: 0, Y: 0},
			Goal:  m.Coords{X: 0, Y: 4},
			ExpectedPath: []m.Coords{
				{X: 0, Y: 1},
				{X: 0, Y: 2},
				{X: 0, Y: 3},
				{X: 0, Y: 4},
			},
			CostMap:    costMap,
			FoundRoute: true,
		},
		{
			Start: m.Coords{X: 0, Y: 4},
			Goal:  m.Coords{X: 4, Y: 4},
			ExpectedPath: []m.Coords{
				{X: 0, Y: 3},
				{X: 1, Y: 3},
				{X: 2, Y: 3},
				{X: 3, Y: 3},
				{X: 3, Y: 4},
				{X: 4, Y: 4},
			},
			CostMap:    costMap,
			FoundRoute: true,
		},
		{
			Start: m.Coords{X: 4, Y: 0},
			Goal:  m.Coords{X: 4, Y: 4},
			ExpectedPath: []m.Coords{
				{X: 4, Y: 1},
				{X: 3, Y: 1},
				{X: 2, Y: 1},
				{X: 2, Y: 2},
				{X: 2, Y: 3},
				{X: 3, Y: 3},
				{X: 3, Y: 4},
				{X: 4, Y: 4},
			},
			CostMap:    costMap,
			FoundRoute: true,
		},
		{
			Start: m.Coords{X: 0, Y: 0},
			Goal:  m.Coords{X: 4, Y: 0},
			ExpectedPath: []m.Coords{
				{X: 0, Y: 1},
				{X: 1, Y: 1},
				{X: 2, Y: 1},
				{X: 3, Y: 1},
				{X: 4, Y: 1},
				{X: 4, Y: 0},
			},
			CostMap: [][]int{
				{0, 0, 0, 10, 0},
				{0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0},
			},
			FoundRoute: true,
		},
		{
			Start: m.Coords{X: 0, Y: 0},
			Goal:  m.Coords{X: 0, Y: 4},
			ExpectedPath: []m.Coords{
				{X: 0, Y: 1},
				{X: 1, Y: 1},
				{X: 1, Y: 2},
				{X: 1, Y: 3},
				{X: 0, Y: 3},
				{X: 0, Y: 4},
			},
			CostMap: [][]int{
				{0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0},
				{10, 0, 0, 0, 0},
				{0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0},
			},
			FoundRoute: true,
		},
		{
			Start: m.Coords{X: 0, Y: 0},
			Goal:  m.Coords{X: 0, Y: 4},
			ExpectedPath: []m.Coords{
				{X: 0, Y: 1},
				{X: 0, Y: 2},
			},
			CostMap: [][]int{
				{0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0},
				{0, 0, 0, 0, 0},
				{10, 10, 10, 0, 0},
				{0, 0, 0, 10, 0},
			},
			FoundRoute: false,
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test FindPath: Case %v", i), func(t *testing.T) {
			resultPath, foundPath := FindPath(testCase.Start, testCase.Goal, testCase.CostMap)
			if !pathIsSame(resultPath, testCase.ExpectedPath) {
				t.Error("Result path doesn't match expected path. Result:", resultPath, "Expected:", testCase.ExpectedPath)
			}
			if foundPath != testCase.FoundRoute {
				t.Error("found path result doesn't match expected found path result. result:", foundPath, "expected:", testCase.FoundRoute)
			}
		})
	}
}

func BenchmarkFindPathSM(b *testing.B) {
	costMap := [][]int{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	width := len(costMap[0])
	height := len(costMap)

	for i := 0; i < b.N; i++ {
		start := m.Coords{X: rand.Intn(width), Y: rand.Intn(height)}
		goal := m.Coords{X: rand.Intn(width), Y: rand.Intn(height)}
		result, _ := FindPath(start, goal, costMap)
		_ = result
	}
}

func BenchmarkFindPathMD(b *testing.B) {
	barrierMap := [][]int{
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}
	width := len(barrierMap[0])
	height := len(barrierMap)

	for i := 0; i < b.N; i++ {
		start := m.Coords{X: rand.Intn(width), Y: rand.Intn(height)}
		goal := m.Coords{X: rand.Intn(width), Y: rand.Intn(height)}
		result, _ := FindPath(start, goal, barrierMap)
		_ = result
	}
}
