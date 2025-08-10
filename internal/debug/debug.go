package debug

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

var peakAlloc uint64 = 0
var lastLogAlloc uint64 = 0

var alloCount uint64 = 0
var allocPerSec uint64 = 0
var peakAllocPerSec uint64 = 0
var lastLogAllocPerSec uint64 = 0

var startTime time.Time = time.Now()
var warmupDone bool = false // wait for the game to finish starting up before recording performance stats
var lastLog time.Time = time.Now()

var peakGoRout int = 0
var lastLogGoRout int = 0

var currentFPS float64 = 0
var minFPS float64 = 9999
var lastLogFPS float64 = 9999
var currentTPS float64 = 0
var minTPS float64 = 9999
var lastLogTPS float64 = 9999

var currentLog string = "No data yet"

func GetLog() string {
	return currentLog
}

func UpdatePerformanceMetrics() {
	if !warmupDone {
		if time.Since(startTime).Seconds() < 5 {
			return
		}
		warmupDone = true
		startTime = time.Now()
	}
	var m runtime.MemStats

	// track FPS and TPS
	currentFPS = ebiten.ActualFPS()
	if currentFPS < minFPS {
		minFPS = currentFPS
	}
	currentTPS = ebiten.ActualTPS()
	if currentTPS < minTPS {
		minTPS = currentTPS
	}

	runtime.ReadMemStats(&m)
	if m.Alloc > peakAlloc {
		peakAlloc = m.Alloc
	}
	numGoRout := runtime.NumGoroutine()
	if numGoRout > peakGoRout {
		peakGoRout = numGoRout
	}

	// measure number of allocations each second
	alloCount += m.Alloc
	if time.Since(startTime).Seconds() >= 1 {
		allocPerSec = alloCount
		alloCount = 0
		startTime = time.Now()
	}
	if allocPerSec > peakAllocPerSec {
		peakAllocPerSec = allocPerSec
	}
	currentLog = fmt.Sprintf("Alloc: %s, Peak Alloc: %s, Alloc/s: %s, Peak Alloc/s: %s\nSys: %s, NumGC: %v\nGoRout: %v, Peak GoRout: %v\nFPS: %.2f, minFPS: %.2f, TPS: %.2f, minTPS: %.2f",
		formatBytes(m.Alloc), formatBytes(peakAlloc), formatBytes(allocPerSec), formatBytes(peakAllocPerSec),
		formatBytes(m.Sys), m.NumGC,
		numGoRout, peakGoRout, currentFPS, minFPS, currentTPS, minTPS)

	// log if there is a new peak
	if time.Since(lastLog).Seconds() >= 1 &&
		(peakAlloc > lastLogAlloc || peakAllocPerSec > lastLogAllocPerSec || peakGoRout > lastLogGoRout || minFPS < lastLogFPS || minTPS < lastLogTPS) {
		peaks := ""
		if peakAlloc > lastLogAlloc {
			peaks += "Alloc "
		}
		if peakAllocPerSec > lastLogAllocPerSec {
			peaks += "Alloc/s "
		}
		if peakGoRout > lastLogGoRout {
			peaks += "GoRout "
		}
		if minFPS < lastLogFPS {
			peaks += "FPS "
		}
		if minTPS < lastLogTPS {
			peaks += "TPS "
		}
		lastLog = time.Now()
		lastLogAlloc = peakAlloc
		lastLogAllocPerSec = peakAllocPerSec
		lastLogGoRout = peakGoRout
		lastLogFPS = minFPS
		lastLogTPS = minTPS
		log.Print("New peak: "+peaks+"\n", currentLog, "\n ")
	}
}

func formatBytes(b uint64) string {
	kb := float64(b) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.2f KB", kb)
	}
	mb := kb / 1024
	return fmt.Sprintf("%.2f KB (%.1f MB)", kb, mb)
}
