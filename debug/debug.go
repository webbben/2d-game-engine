package debug

import (
	"fmt"
	"runtime"
	"time"
)

func DisplayResourceUsage(intervalSeconds int) {
	interval := time.Second * 10
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			printMemoryUsage()
		}
	}
}

func printMemoryUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Println("=== Memory Usage ===")
	fmt.Printf("Alloc: %v MiB\n", bToMb(m.Alloc))
	fmt.Printf("Total Alloc: %v MiB\n", bToMb(m.TotalAlloc))
	fmt.Printf("Sys: %v MiB\n", bToMb(m.Sys))
	fmt.Printf("NumGC: %v\n", m.NumGC)
	fmt.Println("====================")
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
