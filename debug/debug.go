package debug

import (
	"fmt"
	"log"
	"runtime"
	"time"
)

var peakAlloc uint64 = 0
var lastLogAlloc uint64 = 0

var alloCount uint64 = 0
var allocPerSec uint64 = 0
var peakAllocPerSec uint64 = 0
var lastLogAllocPerSec uint64 = 0

var startTime time.Time = time.Now()
var lastLog time.Time = time.Now()

var peakGoRout int = 0
var lastLogGoRout int = 0

func GetMemoryUsageStats() string {
	var m runtime.MemStats

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
	out := fmt.Sprintf("Alloc: %s, Peak Alloc: %s, Alloc/s: %s, Peak Alloc/s: %s\nSys: %s, NumGC: %v\nGoRout: %v, Peak GoRout: %v",
		formatBytes(m.Alloc), formatBytes(peakAlloc), formatBytes(allocPerSec), formatBytes(peakAllocPerSec),
		formatBytes(m.Sys), m.NumGC,
		numGoRout, peakGoRout)

	// log if there is a new peak
	if time.Since(lastLog).Seconds() >= 1 && (peakAlloc > lastLogAlloc || peakAllocPerSec > lastLogAllocPerSec || peakGoRout > lastLogGoRout) {
		peaks := ""
		if peakAlloc > lastLogAlloc {
			peaks += "Alloc, "
		}
		if peakAllocPerSec > lastLogAllocPerSec {
			peaks += "Alloc/s, "
		}
		if peakGoRout > lastLogGoRout {
			peaks += "GoRout"
		}
		lastLog = time.Now()
		lastLogAlloc = peakAlloc
		lastLogAllocPerSec = peakAllocPerSec
		lastLogGoRout = peakGoRout
		log.Print("New peak: "+peaks+"\n", out, "\n ")
	}
	return out
}

func formatBytes(b uint64) string {
	kb := float64(b) / 1024
	if kb < 1024 {
		return fmt.Sprintf("%.2f KB", kb)
	}
	mb := kb / 1024
	return fmt.Sprintf("%.2f KB (%.1f MB)", kb, mb)
}
