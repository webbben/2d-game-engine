// Package clock defines the internal game clock
package clock

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/internal/logz"
)

type (
	DayOfWeek string
	Season    string
)

const (
	Sunday    DayOfWeek = "Solis"
	Monday    DayOfWeek = "Lunae"
	Tuesday   DayOfWeek = "Martis"
	Wednesday DayOfWeek = "Mercurii"
	Thursday  DayOfWeek = "Jovis"
	Friday    DayOfWeek = "Veneris"
	Saturday  DayOfWeek = "Saturni"

	// shortened these since it was hard to fit in HUD clock.
	// should I make the HUD clock wider and bring these back to full length?

	Spring Season = "Spr"
	Summer Season = "Sum"
	Fall   Season = "Fall"
	Winter Season = "Wint"
)

var (
	DaysOfWeek = []DayOfWeek{Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday}
	Seasons    = []Season{Spring, Summer, Fall, Winter}
)

type Clock struct {
	hour, minute int
	year         int
	season       int
	dayOfSeason  int
	daysInSeason int // number of days that a single season lasts

	dowBasisYear int // used to calculate day of week. day 0 season 0 of this year is defined as sunday/first day of week
	dayOfWeek    int

	hourSpeed      time.Duration
	lastMinuteTick time.Time
}

func (c Clock) String() string {
	season := Seasons[c.season]
	dow := DaysOfWeek[c.dayOfWeek]
	return fmt.Sprintf("%02d:%02d Y %v S %v (%s) DoS %v/%v DoW %v (%s)", c.hour, c.minute, c.year, c.season, season, c.dayOfSeason, c.daysInSeason-1, c.dayOfWeek, dow)
}

func (c Clock) GetCurrentDateAndTime() (m, h, y, season, seasonDay int, dow DayOfWeek) {
	return c.minute, c.hour, c.year, c.season, c.dayOfSeason, DaysOfWeek[c.dayOfWeek]
}

func (c Clock) GetTimeString(formatAmPm bool) string {
	if formatAmPm {
		// do AM/PM system with 12 hour clocks
		hour := c.hour
		meridiem := "AM"
		if hour > 12 {
			meridiem = "PM"
			hour -= 12
		}
		if hour == 0 {
			// midnight is 12 AM, not 0 o'clock
			hour = 12
		}
		return fmt.Sprintf("%v:%02d %s", hour, c.minute, meridiem)
	}

	// do 24 hr clock
	return fmt.Sprintf("%v:%02d", c.hour, c.minute)
}

func (c Clock) minuteSpeed() time.Duration {
	return c.hourSpeed / 60
}

// TickTock increments minutes, hours, days, etc. basically handles all ticking time change.
func (c *Clock) TickTock() {
	c.lastMinuteTick = time.Now()

	// MINUTE
	c.minute++
	if c.minute > 59 {
		c.minute = 0

		// HOUR
		c.hour++
		if c.hour > 23 {
			c.hour = 0

			// DAY OF WEEK
			c.dayOfWeek++
			if c.dayOfWeek >= len(DaysOfWeek) {
				c.dayOfWeek = 0
			}

			// DAY OF SEASON
			c.dayOfSeason++
			if c.dayOfSeason > c.daysInSeason-1 {
				c.dayOfSeason = 0

				// SEASON
				c.season++
				if c.season >= len(Seasons) {
					c.season = 0

					// YEAR
					c.year++
				}
			}
		}
	}
}

func (c *Clock) PassTime(hours int) {
	days := hours / 23
	if days == 0 {
		c.hour += hours
		return
	}

	seasons := days / c.daysInSeason
	if seasons == 0 {
		c.dayOfSeason += days
		return
	}

	years := seasons / 3
	if years == 0 {
		c.season += seasons
	}

	// Hmmm... why are we waiting entire years? Let's panic for now, unless we have a use case in the future.
	logz.Panicln("Clock", "tried to pass a year or more of time, which seems wrong... hours:", hours)
}

// SetDateAndTime sets the exact date and time of the clock.
// For passing time in the game, you can use PassTime instead.
func (c *Clock) SetDateAndTime(hour, minute, seasonDay, season, year int) {
	if hour < 0 || hour > 23 {
		panic("hour invalid")
	}
	if minute < 0 || minute > 59 {
		panic("minute invalid")
	}
	if year < 0 || year > 10000 {
		panic("year invalid")
	}
	if seasonDay < 0 || seasonDay > c.daysInSeason-1 {
		panic("season day invalid")
	}
	if season < 0 || season >= len(Seasons) {
		panic("season invalid")
	}

	// calculate day of week
	if year <= c.dowBasisYear {
		panic("can't go to a year before or same as the dow basis year")
	}
	// "Day of Week Basis" is on the first day of the first season of the set dowBasisYear
	// We define this "basis date" as being a sunday, and therefore we can calculate the correct day of the week for any date after it.
	// Well, we could do before too, but I don't wanna do that extra math.
	// to calculate day of week, we want to calculate how many days past the dow basis date this new date is at.
	// then, we can just use modulus to get the day of week
	daysPastBasis := 0
	yearsPastBasis := year - c.dowBasisYear
	if yearsPastBasis <= 0 {
		logz.Panicln("SetDateAndTime", "sanity check: years past basis is <= 0:", yearsPastBasis)
	}
	daysInYear := (c.daysInSeason * len(Seasons))
	daysPastBasis += daysInYear * yearsPastBasis // factor in years
	daysPastBasis += season * c.daysInSeason     // factor in seasons
	daysPastBasis += seasonDay                   // factor in days

	if daysPastBasis < 0 {
		logz.Panicln("SetDateAndTime", "sanity checks: days past basis is negative:", daysPastBasis)
	}

	dow := daysPastBasis % len(DaysOfWeek)

	// math sanity check
	if dow < 0 || dow >= len(DaysOfWeek) {
		logz.Panicln("SetDateAndTime", "sanity check: calculated day of week is wrong... it's either negative or it's longer than the days of week slice:", dow)
	}
	c.dayOfWeek = dow

	c.minute = minute
	c.hour = hour
	c.season = season
	c.dayOfSeason = seasonDay
	c.year = year
}

func NewClock(hourSpeed time.Duration, initHour, initMin, initSeason, initDayOfSeason, initYear, seasonDays int) Clock {
	if seasonDays < 5 || seasonDays > 100 {
		panic("invalid season days")
	}
	if hourSpeed < time.Minute {
		panic("invalid hour speed: too short (minimum 1 minute)")
	}
	if hourSpeed > time.Hour {
		panic("invalid hour speed: too long (maximum 1 hour)")
	}

	c := Clock{
		daysInSeason: seasonDays,
		hourSpeed:    hourSpeed,
		// initialize dowBasisYear to 1000 years before init year - we'll assume that's far enough in the past to be safe.
		dowBasisYear: initYear - 1000,
	}

	c.SetDateAndTime(initHour, initMin, initDayOfSeason, initSeason, initYear)

	return c
}

func (c *Clock) Update() (hourChanged bool) {
	// update time
	beforeTickHour := c.hour
	if time.Since(c.lastMinuteTick) >= c.minuteSpeed() {
		c.TickTock()
	}
	// check if the hour changed, to pass back to caller
	if beforeTickHour != c.hour {
		hourChanged = true
	}

	return
}
