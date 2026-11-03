package logz

import (
	"runtime"
	"strings"
)

// defaultStackDepth is the number of frames CallStack returns when the caller passes a non-positive depth.
const defaultStackDepth = 10

// CallStack formats the calling goroutine's stack as a short, single-line chain of function names,
// innermost first, joined by " <- " (e.g. "npc.(*FightTask).handleCombat <- npc.(*FightTask).TaskInProg").
// The arrow points from each frame to the one that called it, so reading left to right walks outward
// from this call site to the top of the goroutine.
//
// depth is the maximum number of frames to include; a non-positive depth uses defaultStackDepth.
// includeCaller controls whether the function that called CallStack is the first frame: true starts the
// chain at that call site, false starts one level further out (at CallStack's caller's caller).
// Frames belonging to this package, runtime, and crashreport are skipped, so CallStack never appears in
// its own output.
func CallStack(depth int, includeCaller bool) string {
	if depth <= 0 {
		depth = defaultStackDepth
	}

	// one extra frame so skipping CallStack itself doesn't eat into the requested depth.
	pcs := make([]uintptr, depth+1)
	// skip runtime.Callers and CallStack; optionally skip CallStack's caller too.
	skip := 2
	if !includeCaller {
		skip = 3
	}
	n := runtime.Callers(skip, pcs)
	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	names := make([]string, 0, depth)
	for len(names) < depth {
		frame, more := frames.Next()
		if !isInternalFrame(frame.Function) {
			names = append(names, shortFuncName(frame.Function))
		}
		if !more {
			break
		}
	}
	return strings.Join(names, " <- ")
}

// isInternalFrame reports whether a fully-qualified function name belongs to machinery we don't want
// in a call chain: this package, the runtime, and crashreport.
func isInternalFrame(function string) bool {
	return strings.HasPrefix(function, "runtime.") ||
		strings.Contains(function, "runtime/") ||
		strings.Contains(function, "github.com/webbben/2d-game-engine/logz") ||
		strings.Contains(function, "github.com/webbben/2d-game-engine/crashreport")
}

// shortFuncName trims a fully-qualified function name to its last path segment, keeping the package
// qualifier and method: "github.com/x/y/world/npc.(*FightTask).handleCombat" becomes
// "npc.(*FightTask).handleCombat".
func shortFuncName(function string) string {
	if function == "" {
		return "?"
	}
	if i := strings.LastIndex(function, "/"); i != -1 {
		return function[i+1:]
	}
	return function
}
