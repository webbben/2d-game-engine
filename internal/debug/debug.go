package debug

import (
	"fmt"
	"log"
	"runtime"
	"strings"
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

var currentLog string = "No data yet\n"
var ticksSinceLastLog = 0

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

	// for better output in the console
	// currentLog = fmt.Sprintf("Alloc: %s, Peak Alloc: %s\nAlloc/s: %s, Peak Alloc/s: %s\nSys: %s, NumGC: %v\nGoRout: %v, Peak GoRout: %v\nFPS: %.2f, minFPS: %.2f, TPS: %.2f, minTPS: %.2f\n",
	// 	formatBytes(m.Alloc), formatBytes(peakAlloc), formatBytes(allocPerSec), formatBytes(peakAllocPerSec),
	// 	formatBytes(m.Sys), m.NumGC,
	// 	numGoRout, peakGoRout, currentFPS, minFPS, currentTPS, minTPS)

	// for better output on screen
	ticksSinceLastLog++
	if ticksSinceLastLog > 120 {
		ticksSinceLastLog = 0
		currentLog = createColumns(
			15,
			fmt.Sprintf("Alloc: %s", formatBytes(m.Alloc)),
			fmt.Sprintf("Peak: %s", formatBytes(peakAlloc)),
			fmt.Sprintf("Alloc/s: %s", formatBytes(allocPerSec)),
			fmt.Sprintf("Peak: %s", formatBytes(peakAllocPerSec)),
			fmt.Sprintf("Sys: %s", formatBytes(m.Sys)),
			fmt.Sprintf("NumGC: %v", m.NumGC),
			fmt.Sprintf("GoRout: %v", numGoRout),
			fmt.Sprintf("Peak: %v", peakGoRout),
			fmt.Sprintf("FPS: %.2f", currentFPS),
			fmt.Sprintf("minFPS: %.2f", minFPS),
			fmt.Sprintf("TPS: %.2f", currentTPS),
			fmt.Sprintf("minTPS: %.2f", minTPS),
		) + "\n"
	}

	// log if there is a new peak
	if time.Since(lastLog).Seconds() >= 1 &&
		(peakAlloc > lastLogAlloc || peakAllocPerSec > lastLogAllocPerSec || peakGoRout > lastLogGoRout || minFPS < lastLogFPS || minTPS < lastLogTPS) {
		peaks := ""
		if peakAlloc > lastLogAlloc {
			peaks += fmt.Sprintf("  Alloc: %v", formatBytes(peakAlloc))
		}
		if peakAllocPerSec > lastLogAllocPerSec {
			peaks += fmt.Sprintf("  Alloc/s: %v", formatBytes(peakAllocPerSec))
		}
		if peakGoRout > lastLogGoRout {
			peaks += fmt.Sprintf("  GoRout: %v", peakGoRout)
		}
		if minFPS < lastLogFPS {
			peaks += fmt.Sprintf("  minFPS: %.2f", minFPS)
		}
		if minTPS < lastLogTPS {
			peaks += fmt.Sprintf("  minTPS: %.2f", minTPS)
		}
		lastLog = time.Now()
		lastLogAlloc = peakAlloc
		lastLogAllocPerSec = peakAllocPerSec
		lastLogGoRout = peakGoRout
		lastLogFPS = minFPS
		lastLogTPS = minTPS
		log.Printf("New peak:%s\n", peaks)
	}
}

func formatBytes(b uint64) string {
	kb := float64(b) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.2f KB", kb)
	}
	mb := kb / 1024
	return fmt.Sprintf("%.0f MB", mb)
}

func createColumns(columnWidth int, data ...string) string {
	var sb strings.Builder

	for _, d := range data {
		sb.WriteString(d)
		extraSpace := columnWidth - len(d) - 2
		if extraSpace > 0 {
			for range extraSpace {
				sb.WriteByte(' ')
			}
		}
		sb.WriteString("  ")
	}

	return sb.String()
}
