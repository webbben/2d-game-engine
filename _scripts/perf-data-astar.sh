#!/bin/bash

# capture profile data for cpu and memory for a specific benchmark testing function
go test -run=NONE -bench=BenchmarkFindPathScaling/150x150 -cpuprofile=/tmp/cpu.out -memprofile=/tmp/mem.out ./internal/path_finding

# cpu sorted by self time
go tool pprof -top -nodecount=25 /tmp/cpu.out

# memory allocations
go tool pprof -top -alloc_space -nodecount=25 /tmp/mem.out

# prints the source of a function line by line with time and allocations per line
go tool pprof -list='aStar' /tmp/cpu.out

# clean up (built from go test)
rm path_finding.test
