package debug

import (
	"fmt"
	"runtime"
	"time"
)

var peakAlloc uint64 = 0
var alloCount uint64 = 0
var allocPerSec uint64 = 0
var peakAllocPerSec uint64 = 0
var startTime time.Time = time.Now()
var peakGoRout int = 0

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

	return fmt.Sprintf("Alloc: %v KB, Peak Alloc: %v KB, Alloc/s: %v KB, Peak Alloc/s: %v\nSys: %v KB, NumGC: %v\nGoRout: %v, Peak GoRout: %v",
		bToKb(m.Alloc), bToKb(peakAlloc), bToKb(allocPerSec), bToKb(peakAllocPerSec),
		bToKb(m.Sys), m.NumGC,
		numGoRout, peakGoRout)
}

func bToKb(b uint64) uint64 {
	return b / 1024
}
