package debug

import (
	"fmt"
	"runtime"
	"time"

	"github.com/webbben/2d-game-engine/general_util"
)

var peakAlloc uint64 = 0
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

	aveAlloc := general_util.RoundToDecimal(float64(bToMb(m.TotalAlloc))/time.Since(startTime).Seconds(), 1)

	return fmt.Sprintf("Alloc: %v MiB, Ave Alloc/s: %v MiB, Peak Alloc: %v MiB, Total Alloc: %v\nSys: %v MiB, NumGC: %v\nGoRout: %v, Peak GoRout: %v",
		bToMb(m.Alloc), aveAlloc, bToMb(peakAlloc), bToMb(m.TotalAlloc),
		bToMb(m.Sys), m.NumGC,
		numGoRout, peakGoRout)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
