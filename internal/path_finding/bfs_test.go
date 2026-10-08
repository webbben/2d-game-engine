package path_finding

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	m "github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/utils"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// refHopDistanceMap computes true hop distances from `from` using an independent BFS, so the
// optimised implementation can be checked against it.
//
// Note this deliberately measures STEPS, not cost. BuildDistanceMap treats every step as cost 1
// and ignores terrain cost entirely, whereas aStar charges 1 + costMap[tile]. Keeping the reference
// on the same semantics is the whole point of the comparison -- conflating the two would hide real
// disagreements rather than expose them.
func refHopDistanceMap(from m.Coords, costMap [][]int) []int {
	height, width := len(costMap), len(costMap[0])
	dist := make([]int, width*height)
	for i := range dist {
		dist[i] = -1
	}
	dist[idx(from, width)] = 0

	queue := []m.Coords{from}
	for head := 0; head < len(queue); head++ {
		cur := queue[head]
		for _, n := range getNeighbors(cur, costMap) {
			i := idx(n, width)
			if dist[i] != -1 {
				continue
			}
			dist[i] = dist[idx(cur, width)] + 1
			queue = append(queue, n)
		}
	}
	return dist
}

// refBestMonotonicEndpoint independently finds the furthest tile reachable from `npc` by stepping
// only to neighbours whose threat-distance is exactly one greater. Used to check FleeFromPosition
// returns the BEST escape rather than merely a valid one.
func refBestMonotonicEndpoint(threat, npc m.Coords, costMap [][]int) (best m.Coords, bestDist int) {
	width := len(costMap[0])
	dist := refHopDistanceMap(threat, costMap)

	best, bestDist = npc, dist[idx(npc, width)]
	if bestDist == -1 {
		return npc, -1
	}

	visited := map[m.Coords]bool{npc: true}
	queue := []m.Coords{npc}
	for head := 0; head < len(queue); head++ {
		cur := queue[head]
		curDist := dist[idx(cur, width)]
		if curDist > bestDist {
			best, bestDist = cur, curDist
		}
		for _, n := range getNeighbors(cur, costMap) {
			if visited[n] || dist[idx(n, width)] != curDist+1 {
				continue
			}
			visited[n] = true
			queue = append(queue, n)
		}
	}
	return best, bestDist
}

// countReached returns how many entries of a distance map were actually populated.
func countReached(dist []int) int {
	n := 0
	for _, d := range dist {
		if d >= 0 {
			n++
		}
	}
	return n
}

// assertPanicsWith asserts fn panics, and that the panic message contains want. Note logz panics
// route through crashreport.WriteCrashReport, which is a no-op in tests because the reports dir is
// never configured -- so these subtests do not write any files.
func assertPanicsWith(t *testing.T, label, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("%s: expected a panic containing %q, but the call returned normally", label, want)
			return
		}
		if msg := fmt.Sprint(r); !strings.Contains(msg, want) {
			t.Errorf("%s: panicked with %q, want it to contain %q", label, msg, want)
		}
	}()
	fn()
}

// assertFleeingPathIsMonotonic checks the contract that matters most for FleeFromPosition: every
// step must move exactly one tile AND strictly increase the distance from the threat. A step that
// fails to increase the distance means the NPC is walking toward what it is fleeing from.
func assertFleeingPathIsMonotonic(t *testing.T, threat, npc m.Coords, path []m.Coords, costMap [][]int, label string) {
	t.Helper()
	assertValidSteps(t, npc, path, costMap, label)

	dist := refHopDistanceMap(threat, costMap)
	prev, prevDist := npc, dist[idx(npc, len(costMap[0]))]
	for i, c := range path {
		d := dist[idx(c, len(costMap[0]))]
		if d != prevDist+1 {
			t.Errorf("%s: step %d (%v -> %v) has threat-distance %d, want %d (must increase by exactly 1)",
				label, i, prev, c, d, prevDist+1)
			return
		}
		prev, prevDist = c, d
	}
}

// openCostMap builds an all-walkable square map of the given size.
func openCostMap(size int) [][]int {
	cm := make([][]int, size)
	for y := range cm {
		cm[y] = make([]int, size)
	}
	return cm
}

// ---------------------------------------------------------------------------
// BuildDistanceMap
// ---------------------------------------------------------------------------

func TestBuildDistanceMap(t *testing.T) {
	// Unrestricted (center == nil) must agree exactly with the reference on every tile.
	for _, size := range []int{8, 20, 50} {
		t.Run(fmt.Sprintf("matches reference %dx%d", size, size), func(t *testing.T) {
			costMap := testCostMap(size, 4242)
			from, _ := benchEndpoints(size)

			want := refHopDistanceMap(from, costMap)
			got, _ := BuildDistanceMap([]m.Coords{from}, nil, 0, costMap)

			for y := range costMap {
				for x := range costMap[y] {
					i := y*size + x
					if got[i] != want[i] {
						t.Errorf("tile (%d,%d): distance %d, want %d", x, y, got[i], want[i])
					}
				}
			}
		})
	}

	// A radius must never populate a tile outside it, and can only ever make distances worse
	// (or leave a tile unreachable) -- never better.
	t.Run("radius restricts coverage", func(t *testing.T) {
		size := 20
		costMap := testCostMap(size, 4242)
		from, _ := benchEndpoints(size)
		const radius = 4

		unrestricted := refHopDistanceMap(from, costMap)
		got, _ := BuildDistanceMap([]m.Coords{from}, &from, radius, costMap)

		for y := range costMap {
			for x := range costMap[y] {
				i := y*size + x
				fromCenter := utils.ManhattanDistCoords(from, m.Coords{X: x, Y: y})
				if fromCenter > radius && got[i] != -1 {
					t.Errorf("tile (%d,%d) is %d from center (radius %d) but got distance %d",
						x, y, fromCenter, radius, got[i])
				}
				if got[i] >= 0 && got[i] < unrestricted[i] {
					t.Errorf("tile (%d,%d): restricted distance %d is better than unrestricted %d, impossible",
						x, y, got[i], unrestricted[i])
				}
			}
		}
		if got[idx(from, size)] != 0 {
			t.Errorf("from tile has distance %d, want 0", got[idx(from, size)])
		}
	})

	// The returned index must point at a tile holding the maximum populated distance.
	t.Run("farthest index holds the maximum distance", func(t *testing.T) {
		for _, size := range []int{8, 20, 50} {
			costMap := testCostMap(size, 4242)
			from, _ := benchEndpoints(size)
			dist, farthestIdx := BuildDistanceMap([]m.Coords{from}, nil, 0, costMap)

			maxDist := 0
			for _, d := range dist {
				if d > maxDist {
					maxDist = d
				}
			}
			if farthestIdx < 0 || farthestIdx >= len(dist) {
				t.Fatalf("%dx%d: farthest index %d out of range (len %d)", size, size, farthestIdx, len(dist))
			}
			if dist[farthestIdx] != maxDist {
				t.Errorf("%dx%d: tile at farthest index has distance %d, want the maximum %d",
					size, size, dist[farthestIdx], maxDist)
			}
		}
	})

	t.Run("single tile map", func(t *testing.T) {
		dist, farthestIdx := BuildDistanceMap([]m.Coords{{X: 0, Y: 0}}, nil, 0, [][]int{{0}})
		if len(dist) != 1 || dist[0] != 0 {
			t.Errorf("distances = %v, want [0]", dist)
		}
		if farthestIdx != 0 {
			t.Errorf("farthestIdx = %d, want 0", farthestIdx)
		}
	})

	// A full wall must leave the far side unreachable rather than leaking a distance through it.
	t.Run("wall splits reachability", func(t *testing.T) {
		costMap := openCostMap(5)
		for y := range costMap {
			costMap[y][2] = BlockThreshold
		}
		dist, _ := BuildDistanceMap([]m.Coords{{X: 0, Y: 0}}, nil, 0, costMap)
		for y := 0; y < 5; y++ {
			if dist[idx(m.Coords{X: 3, Y: y}, 5)] != -1 {
				t.Errorf("tile (3,%d) is behind a wall but got distance %d, want -1", y, dist[idx(m.Coords{X: 3, Y: y}, 5)])
			}
		}
	})

	// Distances must increase by at most 1 between adjacent reachable tiles (a BFS invariant).
	t.Run("adjacent reachable tiles differ by at most one", func(t *testing.T) {
		size := 20
		costMap := testCostMap(size, 4242)
		from, _ := benchEndpoints(size)
		dist, _ := BuildDistanceMap([]m.Coords{from}, nil, 0, costMap)

		for y := range costMap {
			for x := range costMap[y] {
				here := dist[idx(m.Coords{X: x, Y: y}, size)]
				if here < 0 {
					continue
				}
				for _, n := range getNeighbors(m.Coords{X: x, Y: y}, costMap) {
					there := dist[idx(n, size)]
					if there < 0 {
						continue
					}
					if d := there - here; d > 1 || d < -1 {
						t.Errorf("tiles (%d,%d)=%d and %v=%d differ by %d", x, y, here, n, there, d)
					}
				}
			}
		}
	})

	t.Run("panics", func(t *testing.T) {
		valid := openCostMap(4)
		blocked := openCostMap(4)
		blocked[1][1] = BlockThreshold
		from := m.Coords{X: 1, Y: 1}
		center := m.Coords{X: 2, Y: 2}

		assertPanicsWith(t, "empty costMap", "costmap was empty!", func() {
			BuildDistanceMap([]m.Coords{from}, nil, 0, [][]int{})
		})
		assertPanicsWith(t, "zero width", "costmap has no width!", func() {
			BuildDistanceMap([]m.Coords{from}, nil, 0, [][]int{{}})
		})
		assertPanicsWith(t, "blocked from", "from position was invalid!", func() {
			BuildDistanceMap([]m.Coords{{X: 1, Y: 1}}, nil, 0, blocked)
		})
		assertPanicsWith(t, "blocked center", "center was invalid!", func() {
			// `from` must be valid, otherwise the from check trips first
			BuildDistanceMap([]m.Coords{{X: 0, Y: 0}}, &m.Coords{X: 1, Y: 1}, 2, blocked)
		})
		assertPanicsWith(t, "zero searchRadius with center", "search radius must be > 0", func() {
			BuildDistanceMap([]m.Coords{from}, &center, 0, valid)
		})
		// from sits 2 tiles from center, but the radius is 1. Documented as a caller error.
		assertPanicsWith(t, "from outside radius", "from is outside search radius range of center", func() {
			BuildDistanceMap([]m.Coords{from}, &center, 1, valid)
		})
	})
}

// ---------------------------------------------------------------------------
// FleeFromPosition
// ---------------------------------------------------------------------------

func TestFleeFromPositionScenarios(t *testing.T) {
	// A 1-tall corridor bounded by walls, with an NPC at one end and a threat at the other.
	corridor := openCostMap(5)
	for x := 1; x < 5; x++ {
		corridor[1][x] = BlockThreshold
		corridor[3][x] = BlockThreshold
	}

	// An NPC walled in on three sides with the only opening facing the threat: reachable from the
	// threat, but with no strictly-increasing neighbour to step onto.
	closet := openCostMap(5)
	closet[2][1] = BlockThreshold
	closet[2][3] = BlockThreshold
	closet[1][2] = BlockThreshold

	// A sealed room the threat cannot enter at all.
	sealed := openCostMap(5)
	sealed[1][2], sealed[1][3], sealed[1][4] = BlockThreshold, BlockThreshold, BlockThreshold
	sealed[2][2] = BlockThreshold
	sealed[3][2] = BlockThreshold
	sealed[4][3], sealed[4][4] = BlockThreshold, BlockThreshold

	// An NPC whose escape must first pass the threat: impossible to express with a strictly
	// increasing route, so it reports cannotFlee even though a route exists. This pins the known
	// limitation -- if an A* fallback is ever added, this expectation should change deliberately.
	doorway := openCostMap(5)
	doorway[2][1] = BlockThreshold
	doorway[2][3] = BlockThreshold
	doorway[1][2] = BlockThreshold

	// A room whose far door is away from the threat.
	room := openCostMap(7)
	for x := 1; x <= 5; x++ {
		room[1][x] = BlockThreshold
		room[3][x] = BlockThreshold
	}

	// An NPC with a wall immediately behind it and the threat in front. This is the case that
	// broke the original "aim at the opposite point" approach: the escape point was inside the
	// wall, and the nearest-open-tile snap could pick a tile back toward the threat.
	wallBehind := openCostMap(5)
	for x := 0; x < 5; x++ {
		wallBehind[3][x] = BlockThreshold
	}

	tests := []struct {
		name          string
		costMap       [][]int
		threat        m.Coords
		npc           m.Coords
		radius        int
		wantReachable bool
		wantFlee      bool
		extraCheck    func(t *testing.T, path []m.Coords)
	}{
		{
			name: "open map, threat to the west", costMap: openCostMap(5),
			threat: m.Coords{X: 1, Y: 2}, npc: m.Coords{X: 2, Y: 2}, radius: 4,
			wantReachable: true, wantFlee: true,
		},
		{
			name: "npc adjacent to threat", costMap: openCostMap(5),
			threat: m.Coords{X: 2, Y: 2}, npc: m.Coords{X: 3, Y: 2}, radius: 4,
			wantReachable: true, wantFlee: true,
		},
		{
			name: "wall immediately behind npc", costMap: wallBehind,
			threat: m.Coords{X: 2, Y: 1}, npc: m.Coords{X: 2, Y: 2}, radius: 4,
			wantReachable: true, wantFlee: true,
			extraCheck: func(t *testing.T, path []m.Coords) {
				for i, c := range path {
					if c.Y >= 3 {
						t.Errorf("step %d reached %v, which is inside the wall row", i, c)
					}
				}
			},
		},
		{
			name: "room, far door away from threat", costMap: room,
			threat: m.Coords{X: 4, Y: 2}, npc: m.Coords{X: 2, Y: 2}, radius: 6,
			wantReachable: true, wantFlee: true,
		},
		{
			name: "dead-end corridor", costMap: corridor,
			threat: m.Coords{X: 0, Y: 2}, npc: m.Coords{X: 4, Y: 2}, radius: 5,
			wantReachable: true, wantFlee: false,
		},
		{
			name: "only exit passes the threat (known limitation)", costMap: doorway,
			threat: m.Coords{X: 2, Y: 4}, npc: m.Coords{X: 2, Y: 2}, radius: 4,
			wantReachable: true, wantFlee: false,
		},
		{
			name: "npc sealed away from threat", costMap: sealed,
			threat: m.Coords{X: 0, Y: 0}, npc: m.Coords{X: 3, Y: 2}, radius: 6,
			wantReachable: false, wantFlee: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path, reachable, cannotFlee := FleeFromPositions([]m.Coords{tc.threat}, tc.npc, tc.radius, tc.costMap)

			if reachable != tc.wantReachable {
				t.Fatalf("reachable = %v, want %v (cannotFlee=%v, path=%v)", reachable, tc.wantReachable, cannotFlee, path)
			}
			if !tc.wantFlee {
				// Two distinct "no escape" outcomes, both returning a nil path:
				//   reachable=false -- the threat cannot reach the npc at all, so cannotFlee is
				//                    also false and neither flag is set
				//   reachable=true  -- the threat can reach it, but no route moves strictly away,
				//                    so cannotFlee is true
				if path != nil {
					t.Errorf("no escape expected, but path was %v", path)
				}
				if want := tc.wantReachable; cannotFlee != want {
					t.Errorf("cannotFlee = %v, want %v (reachable = %v)", cannotFlee, want, reachable)
				}
				return
			}

			if cannotFlee {
				t.Fatalf("unexpected cannotFlee; path was %v", path)
			}
			if len(path) == 0 {
				t.Fatal("reported an escape but returned an empty path")
			}
			assertFleeingPathIsMonotonic(t, tc.threat, tc.npc, path, tc.costMap, tc.name)
			if tc.extraCheck != nil {
				tc.extraCheck(t, path)
			}

			// every step is +1, so the path length is exactly the distance gained from the threat
			dist := refHopDistanceMap(tc.threat, tc.costMap)
			startDist := dist[idx(tc.npc, len(tc.costMap[0]))]
			endDist := dist[idx(path[len(path)-1], len(tc.costMap[0]))]
			if want := endDist - startDist; len(path) != want {
				t.Errorf("path length %d, but threat-distance went %d -> %d, so expected %d steps",
					len(path), startDist, endDist, want)
			}
		})
	}
}

// Sweep generated maps checking only the invariants: a contiguous, passable path that moves
// strictly away from the threat every step and ends on the best monotonic escape available.
func TestFleeFromPositionGeneratedMaps(t *testing.T) {
	rng := rand.New(rand.NewSource(20261003))

	checked := 0
	for iter := 0; iter < 600; iter++ {
		size := rng.Intn(10) + 8
		costMap := testCostMap(size, int64(iter))

		from, to := benchEndpoints(size)
		if !isValidCoords(from, costMap) || !isValidCoords(to, costMap) {
			continue
		}
		radius := utils.ManhattanDistCoords(from, to)
		if radius <= 0 {
			continue
		}
		checked++

		label := fmt.Sprintf("iter %d (%dx%d) threat=%v npc=%v radius=%d", iter, size, size, from, to, radius)
		path, reachable, cannotFlee := FleeFromPositions([]m.Coords{from}, to, radius, costMap)

		switch {
		case !reachable:
			if path != nil {
				t.Errorf("%s: reachable=false but path was %v", label, path)
			}
			continue
		case cannotFlee:
			if path != nil {
				t.Errorf("%s: cannotFlee=true but path was %v", label, path)
			}
			continue
		}

		if len(path) == 0 {
			t.Errorf("%s: reported an escape but returned an empty path", label)
			continue
		}
		assertFleeingPathIsMonotonic(t, from, to, path, costMap, label)

		dist := refHopDistanceMap(from, costMap)
		_, wantDist := refBestMonotonicEndpoint(from, to, costMap)
		endDist := dist[idx(path[len(path)-1], size)]
		if endDist != wantDist {
			t.Errorf("%s: ended at threat-distance %d (%v), best monotonic escape is %d",
				label, endDist, path[len(path)-1], wantDist)
		}
		if want := endDist - dist[idx(to, size)]; len(path) != want {
			t.Errorf("%s: path length %d, expected %d steps", label, len(path), want)
		}
	}
	if checked == 0 {
		t.Fatal("no usable generated cases were exercised")
	}
	t.Logf("checked %d generated flee cases", checked)
}

// BuildDistanceMap panics when the threat is already farther from the npc than the search radius.
// That is a deliberate caller contract, so it is pinned here: callers (FleeTask) must stop asking
// once the npc has reached the radius, treating it as the success condition.
func TestFleeFromPositionRejectsOutOfRangeThreat(t *testing.T) {
	costMap := openCostMap(9)
	threat := []m.Coords{{X: 1, Y: 4}}
	npc := m.Coords{X: 7, Y: 4} // 6 tiles away

	assertPanicsWith(t, "threat beyond search radius", "from is outside search radius range of center", func() {
		FleeFromPositions(threat, npc, 3, costMap)
	})

	// exactly at the radius is fine
	if _, _, _ = FleeFromPositions(threat, npc, 6, costMap); false {
		t.Fatal("unreachable")
	}
}

// ---------------------------------------------------------------------------
// benchmarks
// ---------------------------------------------------------------------------

// BenchmarkBuildDistanceMap measures the distance map with and without a radius, since the radius
// is what keeps the BFS from covering an entire city-sized map on every call.
func BenchmarkBuildDistanceMap(b *testing.B) {
	for _, size := range []int{20, 50, 100, 150} {
		b.Run(fmt.Sprintf("%dx%d/no-radius", size, size), func(b *testing.B) {
			costMap := benchCostMap(size, 1, 0)
			from, _ := benchEndpoints(size)

			sample, _ := BuildDistanceMap([]m.Coords{from}, nil, 0, costMap)
			reached := countReached(sample)

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				BuildDistanceMap([]m.Coords{from}, nil, 0, costMap)
			}
			b.StopTimer()
			b.ReportMetric(float64(reached), "tiles-reached")
		})

		b.Run(fmt.Sprintf("%dx%d/radius-12", size, size), func(b *testing.B) {
			costMap := benchCostMap(size, 1, 0)
			_, npc := benchEndpoints(size)
			// threat 6 tiles away from the npc, so it sits comfortably inside the radius
			from := m.Coords{X: npc.X - 6, Y: npc.Y}

			sample, _ := BuildDistanceMap([]m.Coords{from}, &npc, 12, costMap)
			reached := countReached(sample)

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				BuildDistanceMap([]m.Coords{from}, &npc, 12, costMap)
			}
			b.StopTimer()
			b.ReportMetric(float64(reached), "tiles-reached")
		})
	}
}

// benchThreats places n threats as a loose group on ONE side of the npc, inside `radius`.
//
// Two properties matter for the benchmarks to mean anything:
//
//   - Every returned tile must be valid and inside the radius. BuildDistanceMap panics on any `from`
//     that is blocked, out of bounds, or farther from the center than the search radius, so a threat
//     that missed either would make the benchmark measure the panic path instead of the search.
//   - Tiles must be distinct. A duplicate would be seeded twice and expand twice, quietly inflating
//     the cost of an N-threat run without adding a real threat to it.
//
// Why a group on one side rather than a ring around the npc: a ring is the tidier-looking layout, and
// it works fine for BuildDistanceMap, but it makes the flee benchmark measure nothing. With threats
// adjacent on every side there is no neighbour whose distance-from-threat is one greater than the
// npc's own, so FleeFromPositions correctly reports cannotFlee and returns a zero-step path at every
// threat count. That is right behavior and a bad benchmark: it would report a flat, meaningless
// number. A group on one side is also the situation the game actually produces -- a pack of enemies
// chasing you -- and it leaves a real escape route, so path-steps stays meaningful as threats grow.
func benchThreats(costMap [][]int, npc m.Coords, n, radius int) []m.Coords {
	size := len(costMap)
	threats := make([]m.Coords, 0, n)
	seen := map[m.Coords]bool{npc: true} // never place a threat on top of the npc

	// The group sits on the far side of the npc from the map's open interior, at a standoff distance
	// rather than right on top of it. Two reasons, both about the flee benchmark measuring something:
	//
	//   - These benchmarks put the npc in the far corner of the map, so threats pressed against it
	//     leave it nowhere to go and FleeFromPositions correctly reports cannotFlee with a zero-step
	//     path at every threat count. A standoff keeps a real escape route open.
	//   - Threats all at roughly one distance means adding threats changes how many sources the BFS is
	//     seeded with, not how big the searched area is. Standoff, not a growing spread.
	const standoff = 6
	for dx := standoff; dx <= standoff+radius/2 && len(threats) < n; dx++ {
		// fan out vertically as we go, so threats form a wedge rather than a single file
		spread := dx - standoff
		for dy := -spread; dy <= spread && len(threats) < n; dy++ {
			t := m.Coords{X: npc.X - dx, Y: npc.Y + dy}
			if t.X < 1 || t.X >= size-1 || t.Y < 1 || t.Y >= size-1 || seen[t] {
				continue
			}
			if !isValidCoords(t, costMap) {
				continue
			}
			seen[t] = true
			threats = append(threats, t)
		}
	}
	return threats
}

// benchThreatCounts are the numbers of threats to scale over. 1 is included as the baseline the
// multi-threat numbers should be read against, and 8 is roughly a worst-case group fight, which is
// the case that matters most for whether fleeing stays cheap enough to run every repick.
var benchThreatCounts = []int{1, 2, 4, 8}

// BenchmarkBuildDistanceMapMultipleThreats measures the distance map as the number of threats grows.
//
// The question this answers is whether multi-source seeding costs anything beyond the trivial O(n) of
// reading the extra coordinates. It shouldn't in the worst case: all sources are marked visited
// before any expansion, so the total work is one pass over the same reachable tiles a single threat
// would cover, and the only real cost is re-expanding neighbors of each additional seed, which the
// "already visited" check skips. tiles-reached is reported because it should stay flat as threats are
// added -- if it starts climbing, extra seeds are unlocking area a single threat couldn't reach, and
// the timing numbers would no longer be comparable.
func BenchmarkBuildDistanceMapMultipleThreats(b *testing.B) {
	const radius = 12

	for _, size := range []int{50, 100, 150} {
		for _, threats := range benchThreatCounts {
			b.Run(fmt.Sprintf("%dx%d/threats-%d", size, size, threats), func(b *testing.B) {
				costMap := benchCostMap(size, 1, 0)
				_, npc := benchEndpoints(size)
				from := benchThreats(costMap, npc, threats, radius)

				// the whole point of the benchmark is a real number of threats, so a layout that
				// couldn't place them all would silently turn this back into a single-threat run
				if len(from) != threats {
					b.Fatalf("placed %d threats, wanted %d", len(from), threats)
				}

				sample, _ := BuildDistanceMap(from, &npc, radius, costMap)
				reached := countReached(sample)

				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					BuildDistanceMap(from, &npc, radius, costMap)
				}
				b.StopTimer()
				b.ReportMetric(float64(reached), "tiles-reached")
			})
		}
	}
}

// BenchmarkBuildDistanceMapThreatsNoRadius is the same scaling question with the radius lifted, which
// is the case where a single threat would flood the entire map.
//
// This is the shape that could plausibly regress with extra seeds: with no radius cap, every source
// expands across the whole map, so more threats means more redundant neighbor checks over the same
// tiles. Reported separately because it's the unbounded worst case, not the in-game one.
func BenchmarkBuildDistanceMapThreatsNoRadius(b *testing.B) {
	const size = 150

	for _, threats := range benchThreatCounts {
		b.Run(fmt.Sprintf("%dx%d/threats-%d", size, size, threats), func(b *testing.B) {
			costMap := benchCostMap(size, 1, 0)
			_, npc := benchEndpoints(size)
			from := benchThreats(costMap, npc, threats, size)
			if len(from) != threats {
				b.Fatalf("placed %d threats, wanted %d", len(from), threats)
			}

			sample, _ := BuildDistanceMap(from, nil, 0, costMap)
			reached := countReached(sample)

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				BuildDistanceMap(from, nil, 0, costMap)
			}
			b.StopTimer()
			b.ReportMetric(float64(reached), "tiles-reached")
		})
	}
}

// BenchmarkFleeFromPositionMultipleThreats measures the full flee search -- distance map plus the
// outward walk from the npc -- as the number of threats grows.
//
// This is the end-to-end number that matters more than BuildDistanceMap on its own, because it's what
// FleeTask actually calls, and it's the case with the most interesting interaction: with more threats
// the distance map gets shorter everywhere (you're closer to *something*), so the outward walk has
// less room to run and should terminate sooner. path-steps is reported to make that visible, since
// time alone can't tell a faster search from a shorter one.
func BenchmarkFleeFromPositionMultipleThreats(b *testing.B) {
	for _, size := range []int{50, 100, 150} {
		for _, threats := range benchThreatCounts {
			b.Run(fmt.Sprintf("%dx%d/threats-%d", size, size, threats), func(b *testing.B) {
				costMap := benchCostMap(size, 1, 0)
				_, npc := benchEndpoints(size)
				const radius = 12
				from := benchThreats(costMap, npc, threats, radius)
				if len(from) != threats {
					b.Fatalf("placed %d threats, wanted %d", len(from), threats)
				}

				sample, _, _ := FleeFromPositions(from, npc, radius, costMap)

				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					FleeFromPositions(from, npc, radius, costMap)
				}
				b.StopTimer()
				b.ReportMetric(float64(len(sample)), "path-steps")
			})
		}
	}
}

// fleeBenchCase builds a threat/npc pair that exercises each of FleeFromPosition's three outcomes:
// a real escape, no monotonic escape available, and an npc the threat cannot reach at all.
func fleeBenchCase(topo string, size int) (costMap [][]int, threat []m.Coords, npc m.Coords, radius int) {
	costMap = benchCostMap(size, 1, 0)
	_, npc = benchEndpoints(size)
	threat = []m.Coords{{X: npc.X - 6, Y: npc.Y}}
	radius = 10 // comfortably contains the threat, so the radius never panics

	switch topo {
	case "open":
		// nothing to do; the room-and-corridor map already has an escape
	case "no-escape":
		// wall every side except the one facing the threat: reachable, but with no
		// strictly-increasing neighbour to step onto
		costMap[npc.Y][npc.X+1] = BlockThreshold
		costMap[npc.Y-1][npc.X] = BlockThreshold
		costMap[npc.Y+1][npc.X] = BlockThreshold
	case "sealed":
		// wall the npc in completely, so the distance map never reaches it
		costMap[npc.Y][npc.X-1] = BlockThreshold
		costMap[npc.Y][npc.X+1] = BlockThreshold
		costMap[npc.Y-1][npc.X] = BlockThreshold
		costMap[npc.Y+1][npc.X] = BlockThreshold
	default:
		panic("unknown topology: " + topo)
	}
	return costMap, threat, npc, radius
}

func BenchmarkFleeFromPosition(b *testing.B) {
	for _, size := range []int{20, 50, 100, 150} {
		for _, topo := range []string{"open", "no-escape", "sealed"} {
			b.Run(fmt.Sprintf("%dx%d/%s", size, size, topo), func(b *testing.B) {
				costMap, threat, npc, radius := fleeBenchCase(topo, size)

				sample, reachable, cannotFlee := FleeFromPositions(threat, npc, radius, costMap)
				switch topo {
				case "open":
					if !reachable || cannotFlee {
						b.Fatalf("expected an escape, got reachable=%v cannotFlee=%v", reachable, cannotFlee)
					}
				case "no-escape":
					if !reachable || !cannotFlee {
						b.Fatalf("expected cannotFlee, got reachable=%v cannotFlee=%v", reachable, cannotFlee)
					}
				case "sealed":
					if reachable {
						b.Fatalf("expected unreachable npc, got a path of %d steps", len(sample))
					}
				}

				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					FleeFromPositions(threat, npc, radius, costMap)
				}
				b.StopTimer()
				b.ReportMetric(float64(len(sample)), "path-steps")
			})
		}
	}
}

// BenchmarkGetNeighbors isolates the allocation cost both BFS loops pay per expanded tile. Every
// call allocates a slice literal plus repeated append growth. Compare against aStar's benchmark
// history: this was 42% of allocated bytes there before getNeighbors was made allocation-free.
func BenchmarkGetNeighbors(b *testing.B) {
	costMap := benchCostMap(150, 1, 0)
	center, _ := benchEndpoints(150)

	b.ReportAllocs()
	for b.Loop() {
		neighbors := getNeighbors(center, costMap)
		if len(neighbors) == 0 {
			b.Fatal("expected some neighbors")
		}
	}
}
