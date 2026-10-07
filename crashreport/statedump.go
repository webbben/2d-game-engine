package crashreport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// State dumps are written alongside crash reports so the world state at the moment of a crash can be
// inspected afterwards: where the NPCs were, what they were doing, what was on the map. The report itself
// only carries a message, a stack, and a short log tail, which is rarely enough to explain what the game
// actually thought the world looked like.
//
// The provider is registered by the engine, which is the only layer that knows what a "world" is. This package
// stays ignorant of maps and just writes whatever it's handed, which keeps it free of engine dependencies.
//
// NOTE: this package cannot use logz for diagnostics -- logz imports crashreport, so importing it back would
// be a cycle. Diagnostics go to stderr instead.

// StateProvider returns a snapshot of live game state, to be serialized as JSON next to a crash report.
// Returning an error (rather than panicking) is how a provider reports that it couldn't gather its data, so
// the failure is visible rather than silently producing an empty dump.
type StateProvider func() (any, error)

var (
	stateProviderMu sync.RWMutex
	stateProvider   StateProvider
)

// RegisterStateProvider sets the provider used to dump state alongside crash reports. Registering nil
// disables state dumping. Safe to call more than once; the last registration wins.
func RegisterStateProvider(p StateProvider) {
	stateProviderMu.Lock()
	stateProvider = p
	stateProviderMu.Unlock()
}

// getStateProvider returns the currently registered provider, or nil.
func getStateProvider() StateProvider {
	stateProviderMu.RLock()
	defer stateProviderMu.RUnlock()
	return stateProvider
}

// stateDumpDirName is the subdirectory that state dumps are written to. A subdirectory rather than a sibling
// of the reports because LoadAllReports globs every *.json in the reports dir and would otherwise read dumps
// back as (empty) crash reports, which the submit tooling would then file as bogus issues.
const stateDumpDirName = "state"

// StateDumpDir returns the directory state dumps are written to, or "" if crash reporting is disabled.
func StateDumpDir() string {
	if reportsDir == "" {
		return ""
	}
	return filepath.Join(reportsDir, stateDumpDirName)
}

// stateDumpTimestampLayout is used to name dumps so the most recent one is obvious by sorting. Deliberately
// not RFC3339: colons aren't legal in filenames on all platforms.
const stateDumpTimestampLayout = "20060102-150405"

// writeStateDump asks the registered provider for a snapshot of game state and writes it to
// <reportsDir>/state/<hash>_<timestamp>.state.json.
//
// Best-effort by design. Every failure is reported to stderr and then swallowed, because the one thing this
// must never do is interfere with the crash report it accompanies or with the original panic.
//
// Callers must NOT hold mu: the provider gathers live game state from whatever goroutine crashed, which can
// be slow, and holding the package lock across it would serialize every crash in the process.
func writeStateDump(hash string) {
	if reportsDir == "" {
		return
	}
	provider := getStateProvider()
	if provider == nil {
		return
	}

	snapshot, err := callStateProvider(provider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "crashreport: state dump failed: %v\n", err)
		return
	}
	if snapshot == nil {
		// no active map, or nothing worth dumping at this point in startup/teardown.
		return
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "crashreport: could not marshal state dump: %v\n", err)
		return
	}

	dir := StateDumpDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "crashreport: could not create state dump dir %s: %v\n", dir, err)
		return
	}

	name := fmt.Sprintf("%s_%s.state.json", hash, time.Now().Format(stateDumpTimestampLayout))
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "crashreport: could not write state dump: %v\n", err)
	}
}

// callStateProvider invokes the provider, converting a panic into an error. The provider runs while the game
// is already unwinding from a crash, on whatever goroutine failed and possibly holding locks, so it is not
// trustworthy: anything it does that panics has to stay contained here rather than replacing the original
// crash with a second, less useful one.
func callStateProvider(p StateProvider) (snapshot any, err error) {
	defer func() {
		if r := recover(); r != nil {
			snapshot = nil
			err = fmt.Errorf("state provider panicked: %v", r)
		}
	}()
	return p()
}
