package clock

import (
	"testing"
	"time"
)

// newTestClock returns a Clock sitting at midnight on the first day of the first season, using
// the package defaults.
//
// Note that NewClock mutates the DaysInSeason and HourSpeed package globals. Passing a
// seasonDays of 0 keeps them at their defaults, which keeps these tests independent of each
// other, but it also means clock tests must not be marked parallel.
func newTestClock() *Clock {
	return NewClock(time.Minute, 0, 0, 0, 0, 0, 0)
}

func TestAddTime(t *testing.T) {
	type testcase struct {
		name     string
		start    GameTime
		hours    int
		expected GameTime
	}

	testCases := []testcase{
		{
			name:     "adding zero hours is a no-op",
			start:    GameTime{Hour: 10, Minute: 0},
			hours:    0,
			expected: GameTime{Hour: 10, Minute: 0},
		},
		{
			name:     "a single hour stays inside the day",
			start:    GameTime{Hour: 0, Minute: 0},
			hours:    1,
			expected: GameTime{Hour: 1, Minute: 0},
		},
		{
			// a day is HoursPerDay hours long, so the last hour of the day must be reachable.
			name:     "the last hour of the day is reachable",
			start:    GameTime{Hour: 0, Minute: 0},
			hours:    HoursPerDay - 1,
			expected: GameTime{Hour: HoursPerDay - 1, Minute: 0},
		},
		{
			name:     "crossing the end of the day carries into the next one",
			start:    GameTime{Hour: 22, Minute: 0},
			hours:    3,
			expected: GameTime{DayOfSeason: 1, Hour: 1, Minute: 0},
		},
		{
			name:     "24 hours lands on the same hour of the next day",
			start:    GameTime{Hour: 10, Minute: 0},
			hours:    24,
			expected: GameTime{DayOfSeason: 1, Hour: 10, Minute: 0},
		},
		{
			// this is the case that was wrong before: 24 hours from the last-but-two hour
			// used to skip a whole extra day.
			name:     "24 hours from late in the day does not skip two days",
			start:    GameTime{Hour: 22, Minute: 0},
			hours:    24,
			expected: GameTime{DayOfSeason: 1, Hour: 22, Minute: 0},
		},
		{
			name:     "minutes are carried, not reset",
			start:    GameTime{Hour: 23, Minute: 45},
			hours:    1,
			expected: GameTime{DayOfSeason: 1, Hour: 0, Minute: 45},
		},
		{
			// the Q003 "he holds a grudge for a week" case, which used to land 7 hours late
			// because (14 + 24*7) % 23 == 21.
			name:     "a full week lands on the same hour exactly 7 days later",
			start:    GameTime{DayOfSeason: 3, Hour: 14, Minute: 0},
			hours:    24 * 7,
			expected: GameTime{DayOfSeason: 10, Hour: 14, Minute: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gt := tc.start
			gt.AddTime(tc.hours)
			if gt != tc.expected {
				t.Errorf("AddTime(%d) from %+v:\n got      %+v\n expected %+v", tc.hours, tc.start, gt, tc.expected)
			}
		})
	}
}

func TestAddTimeExactPeriods(t *testing.T) {
	// derived rather than hardcoded, so these stay correct if DaysInSeason or the season
	// count ever change.
	seasonLength := HoursPerDay * DaysInSeason
	yearLength := seasonLength * len(Seasons)

	type testcase struct {
		name     string
		hours    int
		expected GameTime
	}

	testCases := []testcase{
		{
			name:     "exactly one season",
			hours:    seasonLength,
			expected: GameTime{Season: 1, DayOfSeason: 0, Hour: 0, Minute: 0},
		},
		{
			name:     "exactly one year, including all four seasons",
			hours:    yearLength,
			expected: GameTime{Year: 1, Season: 0, DayOfSeason: 0, Hour: 0, Minute: 0},
		},
		{
			name:     "exactly four years",
			hours:    yearLength * 4,
			expected: GameTime{Year: 4, Season: 0, DayOfSeason: 0, Hour: 0, Minute: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gt := GameTime{}
			gt.AddTime(tc.hours)
			if gt != tc.expected {
				t.Errorf("AddTime(%d):\n got      %+v\n expected %+v", tc.hours, gt, tc.expected)
			}
		})
	}
}

// TestAddTimeReachesEverySeason guards the seasons-per-year count. It used to be hardcoded to 3
// while Seasons has four entries, which made Winter unreachable.
func TestAddTimeReachesEverySeason(t *testing.T) {
	gt := GameTime{}
	visited := make([]Season, 0, len(Seasons))
	for i := 0; i < len(Seasons); i++ {
		gt.AddTime(HoursPerDay * DaysInSeason)
		// AddTime already validates, but be explicit: this test is about Season staying a
		// usable index into Seasons.
		gt.Validate()
		visited = append(visited, Seasons[gt.Season])
	}

	expected := make([]Season, 0, len(Seasons))
	for i := 1; i <= len(Seasons); i++ {
		expected = append(expected, Seasons[i%len(Seasons)])
	}

	if len(visited) != len(expected) {
		t.Fatalf("visited %d seasons, expected %d", len(visited), len(expected))
	}
	for i := range expected {
		if visited[i] != expected[i] {
			t.Errorf("season %d was %q, expected %q", i, visited[i], expected[i])
		}
	}
	if gt.Year != 1 {
		t.Errorf("after advancing through %d seasons, Year was %d, expected 1", len(Seasons), gt.Year)
	}
}

func TestAddTimeNegativeHoursPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected AddTime to panic when given a negative number of hours")
		}
	}()

	gt := GameTime{Hour: 0, Minute: 0}
	gt.AddTime(-1)
}

func TestValidateRejectsOutOfRange(t *testing.T) {
	type testcase struct {
		name string
		gt   GameTime
	}

	testCases := []testcase{
		{"minute too high", GameTime{Minute: 60}},
		{"minute negative", GameTime{Minute: -1}},
		{"hour too high", GameTime{Hour: HoursPerDay}},
		{"hour negative", GameTime{Hour: -1}},
		{"dayOfSeason equal to the season length", GameTime{DayOfSeason: DaysInSeason}},
		{"dayOfSeason negative", GameTime{DayOfSeason: -1}},
		{"season equal to the number of seasons", GameTime{Season: len(Seasons)}},
		{"season negative", GameTime{Season: -1}},
		{"year negative", GameTime{Year: -1}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("expected Validate to panic for %+v", tc.gt)
				}
			}()
			tc.gt.Validate()
		})
	}
}

func TestValidateAcceptsLatestValues(t *testing.T) {
	// the newest legal instant before each field rolls over.
	gt := GameTime{
		Hour:        HoursPerDay - 1,
		Minute:      59,
		DayOfSeason: DaysInSeason - 1,
		Season:      len(Seasons) - 1,
		Year:        1,
	}
	gt.Validate()
}

// TestTickTockMatchesAddTime is the regression test for the actual defect. Clock.TickTock and
// GameTime.AddTime are two independent implementations of the same hour/day/season/year
// rollover, and they had silently diverged: TickTock was correct, so the running clock was
// fine, while AddTime used a 23-hour day and 3 seasons per year. Anything that advances time
// through AddTime -- GetFutureGameTime, opinion modifier expiries, timed dialog memory, quest
// wait days -- was wrong as a result.
//
// Advancing a live clock by every single minute of a full in-game year (518,400 ticks, so every
// day, season and year boundary is crossed) must land on exactly the same GameTime as adding
// the equivalent number of hours in one call.
func TestTickTockMatchesAddTime(t *testing.T) {
	daysInYear := DaysInSeason * len(Seasons)
	minutes := HoursPerDay * daysInYear * 60
	hours := minutes / 60
	if hours%60 != 0 {
		t.Fatalf("test setup is wrong: %d minutes is not a whole number of hours", minutes)
	}

	c := newTestClock()
	// dayOfWeek is seeded by SetDateAndTime from the dow basis year, so its starting value is
	// whatever that formula produces rather than 0. Measure the advance instead of the value.
	startDow := c.dayOfWeek

	for i := 0; i < minutes; i++ {
		c.TickTock()
	}

	ticked := c.GetCurrentGameTime()

	expected := GameTime{}
	expected.AddTime(hours)

	if ticked != expected {
		t.Errorf("TickTock and AddTime disagree after %d minutes (%d hours):\n TickTock: %+v\n AddTime:  %+v",
			minutes, hours, ticked, expected)
	}

	if ticked != (GameTime{Year: 1, Season: 0, DayOfSeason: 0, Hour: 0, Minute: 0}) {
		t.Errorf("expected exactly one in-game year to have elapsed, got %+v", ticked)
	}

	// TickTock also tracks day of week, which AddTime has no notion of. Check that it advanced
	// by exactly one day per day that elapsed. Accessing the field directly is fine here; we're
	// in the same package.
	if want := (startDow + daysInYear%len(DaysOfWeek)) % len(DaysOfWeek); c.dayOfWeek != want {
		t.Errorf("dayOfWeek was %d, expected %d (started at %d, advanced %d days)", c.dayOfWeek, want, startDow, daysInYear)
	}
}

// TestPassTime covers the one path that advances the live clock through AddTime. It has no
// callers today, but it is the only way AddTime's result can reach c.currentTime, so it is the
// path most worth pinning down.
func TestPassTime(t *testing.T) {
	c := NewClock(time.Minute, 10, 30, 0, 0, 0, 0)
	c.PassTime(24)

	expected := GameTime{Hour: 10, Minute: 30, Year: 0, Season: 0, DayOfSeason: 1}
	if got := c.GetCurrentGameTime(); got != expected {
		t.Errorf("PassTime(24) from 10:30:\n got      %+v\n expected %+v", got, expected)
	}
}
