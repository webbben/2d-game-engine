package entity

import (
	"fmt"
	"math/rand"
	"testing"
)

type FindPathTestCase struct {
	Start, Goal  Pos
	ExpectedPath []Point
}

func pathIsSame(pathA []Point, pathB []Point) bool {
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
	barrierMap := [][]bool{
		{false, false, false, false, false},
		{false, false, false, false, false},
		{false, false, false, true, true},
		{false, false, false, false, false},
		{false, true, false, false, false},
	}

	testCases := []FindPathTestCase{
		{
			Start: Pos{X: 0, Y: 0},
			Goal:  Pos{X: 4, Y: 0},
			ExpectedPath: []Point{
				{X: 1, Y: 0},
				{X: 2, Y: 0},
				{X: 3, Y: 0},
				{X: 4, Y: 0},
			},
		},
		{
			Start: Pos{X: 0, Y: 0},
			Goal:  Pos{X: 0, Y: 4},
			ExpectedPath: []Point{
				{X: 0, Y: 1},
				{X: 0, Y: 2},
				{X: 0, Y: 3},
				{X: 0, Y: 4},
			},
		},
		{
			Start: Pos{X: 0, Y: 4},
			Goal:  Pos{X: 4, Y: 4},
			ExpectedPath: []Point{
				{X: 0, Y: 3},
				{X: 1, Y: 3},
				{X: 2, Y: 3},
				{X: 3, Y: 3},
				{X: 3, Y: 4},
				{X: 4, Y: 4},
			},
		},
		{
			Start: Pos{X: 4, Y: 0},
			Goal:  Pos{X: 4, Y: 4},
			ExpectedPath: []Point{
				{X: 4, Y: 1},
				{X: 3, Y: 1},
				{X: 2, Y: 1},
				{X: 2, Y: 2},
				{X: 2, Y: 3},
				{X: 3, Y: 3},
				{X: 3, Y: 4},
				{X: 4, Y: 4},
			},
		},
	}
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("Test FindPath: Case %v", i), func(t *testing.T) {
			resultPath := FindPath(testCase.Start, testCase.Goal, barrierMap)
			if !pathIsSame(resultPath, testCase.ExpectedPath) {
				t.Error("Result path doesn't match expected path. Result:", resultPath, "Expected:", testCase.ExpectedPath)
			}
		})
	}

}

func BenchmarkFindPathSM(b *testing.B) {
	barrierMap := [][]bool{
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false},
	}
	width := len(barrierMap[0])
	height := len(barrierMap)

	for i := 0; i < b.N; i++ {
		start := Pos{X: float64(rand.Intn(width)), Y: float64(rand.Intn(height))}
		goal := Pos{X: float64(rand.Intn(width)), Y: float64(rand.Intn(height))}
		result := FindPath(start, goal, barrierMap)
		_ = result
	}
}

func BenchmarkFindPathMD(b *testing.B) {
	barrierMap := [][]bool{
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false},
	}
	width := len(barrierMap[0])
	height := len(barrierMap)

	for i := 0; i < b.N; i++ {
		start := Pos{X: float64(rand.Intn(width)), Y: float64(rand.Intn(height))}
		goal := Pos{X: float64(rand.Intn(width)), Y: float64(rand.Intn(height))}
		result := FindPath(start, goal, barrierMap)
		_ = result
	}
}
