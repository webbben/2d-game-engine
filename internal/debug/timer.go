package debug

import (
	"time"

	"github.com/webbben/2d-game-engine/logz"
)

var Timer timer = timer{
	records: make(map[string]*timerRecord),
}

// A Timer is just a convenience utility for timing various things. Probably will be used for performance monitoring type stuff.
type timer struct {
	records map[string]*timerRecord
}

func (t *timer) startTimer(key string) {
	if _, exists := t.records[key]; !exists {
		t.records[key] = &timerRecord{
			key: key,
		}
	}
	record := t.records[key]
	record.lastStartTime = time.Now()
}

func (t timer) stopTimer(key string) {
	record, exists := t.records[key]
	if !exists {
		logz.Warnln("TIMER", "timer key doesn't exist:", key)
		return
	}

	elapsed := time.Since(record.lastStartTime)
	record.callCount++
	record.elapsedTimeSum += elapsed
	logz.Println("TIMER", key, "time elapsed:", elapsed)

	if elapsed < record.lowestElapsed || record.lowestElapsed == 0 {
		record.lowestElapsed = elapsed
	}
	if elapsed > record.highestElapsed {
		record.highestElapsed = elapsed
	}
}

// just prints the current elapsed time, without recording anything
func (t timer) reportElapsedTime(key string) {
	record, exists := t.records[key]
	if !exists {
		logz.Warnln("TIMER", "timer key doesn't exist:", key)
		return
	}
	elapsed := time.Since(record.lastStartTime)
	logz.Println("TIMER", key, "time elapsed:", elapsed)
}

type timerRecord struct {
	key                           string
	lastStartTime                 time.Time
	callCount                     int           // number of times the function being recorded was called (i.e. number of reports called)
	elapsedTimeSum                time.Duration // sum of all elapsed times; used for calculating an average, or just seeing total amount of time spent.
	highestElapsed, lowestElapsed time.Duration // min and max durations
}

func (tr timerRecord) report() {
	logz.Println("TIMER", "== Record Report ==")
	logz.Println(tr.key, "Total calls:", tr.callCount, "Ave:", tr.elapsedTimeSum/time.Duration(tr.callCount), "Total:", tr.elapsedTimeSum)
	logz.Println(tr.key, "Min:", tr.lowestElapsed, "Max:", tr.highestElapsed)
	logz.Println("TIMER", "== End of Report ==")
}

func StartTimer(key string) {
	Timer.startTimer(key)
}

func StopTimer(key string) {
	Timer.stopTimer(key)
}

func ReportTimeElapsed(key string) {
	Timer.reportElapsedTime(key)
}

func ShowFullReport(key string) {
	record, exists := Timer.records[key]
	if !exists {
		logz.Warnln("TIMER", "timer key doesn't exist:", key)
		return
	}
	record.report()
}

func ShowAllReports() {
	for _, record := range Timer.records {
		record.report()
	}
}
