package path_finding

import (
	"fmt"
	"math/rand"
	"testing"

	m "github.com/webbben/2d-game-engine/model"
)

// FindPathTestCase is a hand-built fixture asserting a specific routing outcome.
//
// These assert on path COST rather than an exact route. When several routes share the minimum cost
// -- which is common on these maps, since every passable tile costs 0 -- which one comes back depends
// on the open set's pop order, which is an implementation detail rather than a correctness property.
// Asserting an exact path made these tests fail whenever the search was made faster or the heuristic
// changed, even though every returned route stayed optimal. Optimality itself is covered by
// TestFindPathIsOptimal; these fixtures exist to pin down specific wall-detour and unreachable cases.
type FindPathTestCase struct {
	Start, Goal m.Coords
	CostMap     [][]int
	// ExpectedCost is the minimum cost of a route from Start to Goal, which is the path length in
	// these fixtures because every passable tile costs 0. It is unused (and set to -1) when
	// FoundRoute is false.
	ExpectedCost int
	FoundRoute   bool
}

func pathCost(path []m.Coords, costMap [][]int) int {
	total := 0
	for _, c := range path {
		total += 1 + costMap[c.Y][c.X]
	}
	return total
}

func intAbs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func assertValidSteps(t *testing.T, start m.Coords, path []m.Coords, costMap [][]int, label string) {
	t.Helper()
	prev := start
	for i, c := range path {
		if intAbs(c.X-prev.X)+intAbs(c.Y-prev.Y) != 1 {
			t.Fatalf("%s: step %d goes %v -> %v, which is not a single cardinal move", label, i, prev, c)
		}
		if costMap[c.Y][c.X] >= BlockThreshold {
			t.Fatalf("%s: step %d enters blocked tile %v", label, i, c)
		}
		prev = c
	}
}

// refShortestCost computes the true minimum cost from start to goal using Bellman-Ford with a
// worklist -- a deliberately different algorithm from A*, so a bug in the open-set handling cannot
// hide by being shared with the reference. Reports false when the goal is unreachable.
//
// Neighbours are enumerated inline instead of calling getNeighbors, for the same reason: the
// reference should not share code with the thing it is checking.
func refShortestCost(start, goal m.Coords, costMap [][]int) (int, bool) {
	neighborsOf := func(c m.Coords) []m.Coords {
		out := make([]m.Coords, 0, 4)
		for _, d := range [4]m.Coords{{X: 0, Y: -1}, {X: -1, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 0}} {
			n := m.Coords{X: c.X + d.X, Y: c.Y + d.Y}
			if n.X < 0 || n.Y < 0 || n.X >= len(costMap[0]) || n.Y >= len(costMap) {
				continue
			}
			if costMap[n.Y][n.X] >= BlockThreshold {
				continue
			}
			out = append(out, n)
		}
		return out
	}

	dist := map[m.Coords]int{start: 0}
	worklist := []m.Coords{start}
	for len(worklist) > 0 {
		cur := worklist[0]
		worklist = worklist[1:]
		for _, n := range neighborsOf(cur) {
			nd := dist[cur] + 1 + costMap[n.Y][n.X]
			if prev, ok := dist[n]; !ok || nd < prev {
				dist[n] = nd
				worklist = append(worklist, n)
			}
		}
	}

	cost, ok := dist[goal]
	return cost, ok
}

// testCostMap builds a deterministic map using the same value set buildCostMap produces: 0 for
// open floor, BlockThreshold for blocked, EntityCollision where entities stand (overlapping ones
// reach 4). Laid out as rooms and corridors rather than open field, so walls actually constrain
// the search the way they do in a real map.
func testCostMap(size int, seed int64) [][]int {
	rng := rand.New(rand.NewSource(seed))
	costMap := make([][]int, size)
	for y := range costMap {
		costMap[y] = make([]int, size)
	}
	for i := 0; i < size; i++ {
		costMap[0][i] = BlockThreshold
		costMap[size-1][i] = BlockThreshold
		costMap[i][0] = BlockThreshold
		costMap[i][size-1] = BlockThreshold
	}
	for y := 2; y < size-2; y += 4 {
		for x := 2; x < size-2; x += 4 {
			costMap[y][x] = BlockThreshold
		}
	}
	for i := 0; i < size/4; i++ {
		costMap[rng.Intn(size)][rng.Intn(size)] += EntityCollision
	}
	return costMap
}

func TestFindPathIsOptimal(t *testing.T) {
	for _, size := range []int{8, 20, 50} {
		t.Run(fmt.Sprintf("%dx%d", size, size), func(t *testing.T) {
			costMap := testCostMap(size, 4242)

			starts := []m.Coords{{X: 1, Y: 1}, {X: size - 2, Y: 1}, {X: size / 2, Y: size / 2}}
			goals := []m.Coords{{X: size - 2, Y: size - 2}, {X: 1, Y: size - 2}, {X: size / 2, Y: 2}}

			for _, start := range starts {
				for _, goal := range goals {
					if start.Equals(goal) {
						continue
					}
					label := fmt.Sprintf("%dx%d %v -> %v", size, size, start, goal)
					if costMap[start.Y][start.X] >= BlockThreshold || costMap[goal.Y][goal.X] >= BlockThreshold {
						continue
					}

					want, reachable := refShortestCost(start, goal, costMap)
					got, found := FindPath(start, goal, costMap)

					if found != reachable {
						t.Errorf("%s: FindPath found=%v, reference reachable=%v", label, found, reachable)
						continue
					}
					assertValidSteps(t, start, got, costMap, label)
					if !reachable {
						// unreachable: FindPath returns a partial route to the closest tile it found.
						// there is no cost to compare against
						continue
					}

					// reachable
					if last := got[len(got)-1]; !last.Equals(goal) {
						t.Errorf("%s: path ends at %v, want %v", label, last, goal)
						continue
					}
					if c := pathCost(got, costMap); c != want {
						t.Errorf("%s: path cost %d, want optimal %d\npath: %v", label, c, want, got)
					}
				}
			}
		})
	}
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
			Start:        m.Coords{X: 0, Y: 0},
			Goal:         m.Coords{X: 4, Y: 0},
			ExpectedCost: 4,
			CostMap:      costMap,
			FoundRoute:   true,
		},
		{
			Start:        m.Coords{X: 0, Y: 0},
			Goal:         m.Coords{X: 0, Y: 4},
			ExpectedCost: 4,
			CostMap:      costMap,
			FoundRoute:   true,
		},
		{
			Start:        m.Coords{X: 0, Y: 4},
			Goal:         m.Coords{X: 4, Y: 4},
			ExpectedCost: 6,
			CostMap:      costMap,
			FoundRoute:   true,
		},
		{
			Start:        m.Coords{X: 4, Y: 0},
			Goal:         m.Coords{X: 4, Y: 4},
			ExpectedCost: 8,
			CostMap:      costMap,
			FoundRoute:   true,
		},
		{
			Start:        m.Coords{X: 0, Y: 0},
			Goal:         m.Coords{X: 4, Y: 0},
			ExpectedCost: 6,
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
			Start:        m.Coords{X: 0, Y: 0},
			Goal:         m.Coords{X: 0, Y: 4},
			ExpectedCost: 6,
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
			Start:        m.Coords{X: 0, Y: 0},
			Goal:         m.Coords{X: 0, Y: 4},
			ExpectedCost: -1,
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
			label := fmt.Sprintf("Test FindPath: Case %v", i)
			resultPath, foundPath := FindPath(testCase.Start, testCase.Goal, testCase.CostMap)
			if foundPath != testCase.FoundRoute {
				t.Fatalf("%v: found path result doesn't match expected. result: %v, expected: %v", label, foundPath, testCase.FoundRoute)
			}
			// whichever optimal route was chosen, it still has to be well formed
			assertValidSteps(t, testCase.Start, resultPath, testCase.CostMap, label)

			if !testCase.FoundRoute {
				// unreachable: FindPath returns a partial route to the closest tile it found.
				// which tile that is stays a heuristic choice, so there's no cost to assert.
				if len(resultPath) == 0 {
					t.Errorf("%v: unreachable case returned an empty partial route", label)
				}
				return
			}

			if len(resultPath) == 0 {
				t.Fatalf("%v: found route is true but path is empty", label)
			}
			if last := resultPath[len(resultPath)-1]; !last.Equals(testCase.Goal) {
				t.Fatalf("%v: path ends at %v, want %v", label, last, testCase.Goal)
			}
			if got := pathCost(resultPath, testCase.CostMap); got != testCase.ExpectedCost {
				t.Errorf("%v: path cost %d, want optimal %d\npath: %v", label, got, testCase.ExpectedCost, resultPath)
			}
		})
	}
}

// benchCostMap builds a map resembling the ones the game actually loads: 0 for open floor,
// BlockThreshold for walls, and EntityCollision where entities stand (overlapping entities push a
// tile to 4 or 6, which is what happens in a crowd). Layout is rooms with doorways rather than
// open field, so walls actually funnel the search the way they do in a real map.
//
// entitiesPer100Tiles is roughly how many entities to scatter per 100 tiles. Real maps range from
// 0 in an empty district to a couple per tile in a packed tavern or the combat arena.
func benchCostMap(size int, seed int64, entitiesPer100Tiles float64) [][]int {
	rng := rand.New(rand.NewSource(seed))
	costMap := make([][]int, size)
	for y := range costMap {
		costMap[y] = make([]int, size)
	}

	// solid border
	for i := 0; i < size; i++ {
		costMap[0][i] = BlockThreshold
		costMap[size-1][i] = BlockThreshold
		costMap[i][0] = BlockThreshold
		costMap[i][size-1] = BlockThreshold
	}

	// room walls
	const room = 8
	for y := room; y < size-room; y += room {
		costMap[y][0] = BlockThreshold
		for x := room; x < size-room; x++ {
			costMap[y][x] = BlockThreshold
		}
	}
	for x := room; x < size-room; x += room {
		costMap[0][x] = BlockThreshold
		for y := room; y < size-room; y++ {
			costMap[y][x] = BlockThreshold
		}
	}

	// punch doorways through each wall so the whole map stays connected
	for y := room; y < size-room; y += room {
		for x := room; x < size-room; x += room / 2 {
			costMap[y][x] = 0
		}
	}
	for x := room; x < size-room; x += room {
		for y := room; y < size-room; y += room / 2 {
			costMap[y][x] = 0
		}
	}

	// scattered entities
	count := int(float64(size*size) * entitiesPer100Tiles / 100)
	for i := 0; i < count; i++ {
		costMap[rng.Intn(size)][rng.Intn(size)] += EntityCollision
	}
	return costMap
}

// benchEndpoints picks a deterministic start/goal pair for a generated map. Both are inset from the
// border wall, and start != goal, so the search always does real work -- unlike picking both with
// rand.Intn, which lands on the same tile often enough on small maps to skew the average.
func benchEndpoints(size int) (m.Coords, m.Coords) {
	return m.Coords{X: 1, Y: 1}, m.Coords{X: size - 2, Y: size - 2}
}

// BenchmarkFindPathScaling measures FindPath across the range of map sizes the game loads (roughly
// 20x20 up to 150x150), on maps that resemble real ones.
//
// It reports two numbers because they answer different questions:
//   - ns/op captures everything, including the map allocation done per call in aStar
//   - tiles-expanded is how many tiles the search actually touched, which isolates the algorithm
//     from allocation overhead and is deterministic, so it diffs cleanly across changes
func BenchmarkFindPathScaling(b *testing.B) {
	for _, size := range []int{20, 50, 100, 150} {
		b.Run(fmt.Sprintf("%dx%d", size, size), func(b *testing.B) {
			costMap := benchCostMap(size, 1, 0)
			start, goal := benchEndpoints(size)
			if _, found := FindPath(start, goal, costMap); !found {
				b.Fatalf("%dx%d: generated map is not connected; benchmark would measure the wrong thing", size, size)
			}

			// count expanded outside of actual benchmark loop
			_, visited, _ := aStar(start, goal, costMap)
			expanded := 0
			for _, v := range visited {
				if v {
					expanded++
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				aStar(start, goal, costMap)
			}
			b.StopTimer()
			b.ReportMetric(float64(expanded), "tiles-expanded")
		})
	}
}

// BenchmarkFindPathCrowded measures the case that actually produces cost variation: a map with many
// NPCs on it, where EntityCollision pushes tiles to 2, 4 or 6. This is the scenario the heuristic
// has to reason about, and the one where a weak estimate would cost real extra expansions.
func BenchmarkFindPathCrowded(b *testing.B) {
	for _, size := range []int{50, 150} {
		for _, per100 := range []float64{0, 2, 8} {
			b.Run(fmt.Sprintf("%dx%d/entities-%.0f-per-100", size, size, per100), func(b *testing.B) {
				costMap := benchCostMap(size, 1, per100)
				start, goal := benchEndpoints(size)
				if _, found := FindPath(start, goal, costMap); !found {
					b.Fatalf("%dx%d @ %.0f/100: generated map is not connected", size, size, per100)
				}

				_, visited, _ := aStar(start, goal, costMap)
				expanded := 0
				for _, v := range visited {
					if v {
						expanded++
					}
				}

				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					aStar(start, goal, costMap)
				}
				b.StopTimer()
				b.ReportMetric(float64(expanded), "tiles-expanded")
			})
		}
	}
}

// BenchmarkFindPathUnreachable measures the case where the goal cannot be reached. This is a
// distinct code path: aStar drains its entire open set instead of returning as soon as the goal is
// popped, which is the only situation where entries pushed before a cost improvement actually get
// popped and have to be recognised as stale. It is also the case that exercises the "closest tile
// found" partial-path fallback.
func BenchmarkFindPathUnreachable(b *testing.B) {
	for _, size := range []int{50, 150} {
		b.Run(fmt.Sprintf("%dx%d", size, size), func(b *testing.B) {
			costMap := benchCostMap(size, 1, 0)
			start, goal := benchEndpoints(size)
			// seal the goal in, leaving its own tile open
			for _, d := range []m.Coords{{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1}} {
				n := m.Coords{X: goal.X + d.X, Y: goal.Y + d.Y}
				costMap[n.Y][n.X] = BlockThreshold
			}
			if _, found := FindPath(start, goal, costMap); found {
				b.Fatalf("%dx%d: expected the sealed goal to be unreachable", size, size)
			}

			_, visited, _ := aStar(start, goal, costMap)
			expanded := 0
			for _, v := range visited {
				if v {
					expanded++
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				aStar(start, goal, costMap)
			}
			b.StopTimer()
			b.ReportMetric(float64(expanded), "tiles-expanded")
		})
	}
}
