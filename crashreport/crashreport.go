// Package crashreport provides functions for writing and managing crash reports.
// Crash reports are JSON files saved locally when a panic occurs, capturing
// the panic message, stack trace, and a SHA256 hash for deduplication.
// A separate CLI tool in the game project can submit these to GitHub Issues.
package crashreport

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	reportsDir string
	mu         sync.Mutex
	hexPattern = regexp.MustCompile(`0x[0-9a-f]+`)
	// lineNumberPattern strips the line number from "file.go:123" so the hash
	// stays stable across edits that shift lines (but not across moves to
	// different functions/files).
	lineNumberPattern = regexp.MustCompile(`\.go:\d+`)
)

// CrashReport represents a single crash report stored on disk.
type CrashReport struct {
	Hash          string   `json:"hash"`
	Message       string   `json:"message"`
	Stack         string   `json:"stack"`
	FilteredStack string   `json:"filtered_stack"`
	Logs          []string `json:"logs,omitempty"`
	Timestamp     string   `json:"timestamp"`
	Submitted     bool     `json:"submitted"` // set to true after successful GitHub submission
}

// SetReportsDir sets the directory where crash reports are stored.
// The directory will be created if it doesn't exist.
// If dir is empty, crash report writing is disabled.
func SetReportsDir(dir string) error {
	if dir == "" {
		reportsDir = ""
		return nil
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("failed to resolve crash reports directory: %w", err)
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return fmt.Errorf("failed to create crash reports directory: %w", err)
	}
	reportsDir = absDir
	return nil
}

// ReportsDir returns the currently configured crash reports directory, or empty string if not configured.
func ReportsDir() string {
	return reportsDir
}

// WriteCrashReport saves a crash report to disk, keyed by the SHA256 hash of
// the message plus the sanitized stack trace. If a file with the same hash
// already exists, a numbered suffix is appended (e.g. hash_01.json).
func WriteCrashReport(msg string, stack []byte, logs []string) {
	if reportsDir == "" {
		return
	}

	filteredStack := filterStack(stack)
	hash := hashString(msg + sanitizeStackForHash(filteredStack))

	report := CrashReport{
		Hash:          hash,
		Message:       msg,
		Stack:         string(stack),
		FilteredStack: filteredStack,
		Logs:          logs,
		Timestamp:     time.Now().Format(time.RFC3339),
		Submitted:     false,
	}

	mu.Lock()
	writeReportFile(hash, report)
	mu.Unlock()

	// Deliberately outside mu. The state provider gathers live game state on whatever goroutine crashed,
	// which can be slow, and holding this package's lock across it would serialize every crash. It also
	// keeps the dump from re-entering WriteCrashReport (and self-deadlocking on mu) if gathering trips a
	// logz panic, since logz.Panic* funnels back through here.
	writeStateDump(hash)
}

// writeReportFile writes report to a non-colliding path derived from hash, adding a numeric suffix if a
// report for that hash already exists. Callers must hold mu.
func writeReportFile(hash string, report CrashReport) {
	basePath := filepath.Join(reportsDir, hash)
	filePath := basePath + ".json"
	if _, err := os.Stat(filePath); err == nil {
		for i := 1; ; i++ {
			numberedPath := fmt.Sprintf("%s_%02d.json", basePath, i)
			if _, err := os.Stat(numberedPath); os.IsNotExist(err) {
				filePath = numberedPath
				break
			}
		}
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(filePath, data, 0o644)
}

// LoadAllReports reads all crash report files from the reports directory.
func LoadAllReports() ([]CrashReport, error) {
	pending, err := loadReports()
	if err != nil {
		return nil, err
	}
	reports := make([]CrashReport, 0, len(pending))
	for _, p := range pending {
		reports = append(reports, p.Report)
	}
	return reports, nil
}

// PendingReport is a crash report together with the file it was read from, so callers can update that exact
// occurrence. Several report files can share a hash (repeat crashes), which makes the hash alone ambiguous as
// an identifier for "this one".
type PendingReport struct {
	Report CrashReport
	Path   string
}

// loadReports reads every crash report in the directory. State dumps live in a subdirectory and are skipped,
// so they are never mistaken for reports.
func loadReports() ([]PendingReport, error) {
	if reportsDir == "" {
		return nil, fmt.Errorf("crash reports directory not configured")
	}
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read crash reports directory: %w", err)
	}
	var reports []PendingReport
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(reportsDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var r CrashReport
		if json.Unmarshal(data, &r) == nil {
			reports = append(reports, PendingReport{Report: r, Path: path})
		}
	}
	return reports, nil
}

// LoadReport loads a single crash report by its hash.
func LoadReport(hash string) (*CrashReport, error) {
	r, _, err := LoadReportWithPath(hash)
	return r, err
}

// LoadReportWithPath loads a single crash report by its hash and also returns the file it came from, for
// callers that need to update that exact occurrence (see MarkSubmittedPath).
func LoadReportWithPath(hash string) (*CrashReport, string, error) {
	if reportsDir == "" {
		return nil, "", fmt.Errorf("crash reports directory not configured")
	}
	path := filepath.Join(reportsDir, hash+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", fmt.Errorf("crash report not found: %s", hash)
	}
	var r CrashReport
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, "", fmt.Errorf("failed to parse crash report: %w", err)
	}
	return &r, path, nil
}

// hashString returns the SHA256 hex digest of s.
func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}

// filterStack removes frames from this package, logz, and runtime/ for human-readable display.
// debug.Stack() output looks like:
//
//	goroutine 1 [running]:
//	github.com/pkg/foo.Bar(...)
//		/path/to/file.go:10 +0x...
//	main.main()
//		/path/to/main.go:20 +0x...
func filterStack(stack []byte) string {
	lines := strings.Split(string(stack), "\n")
	var out []string
	skip := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "goroutine ") || strings.HasPrefix(trimmed, "created by ") {
			out = append(out, line)
			continue
		}
		if strings.HasPrefix(trimmed, "/") && strings.Contains(trimmed, ".go:") {
			// file:line row — skip if previous frame was filtered
			if skip {
				skip = false
				continue
			}
			out = append(out, line)
			continue
		}
		// function-name row — check if we should filter this frame
		if strings.HasPrefix(trimmed, "runtime.") ||
			strings.Contains(trimmed, "runtime/") ||
			strings.Contains(trimmed, "github.com/webbben/2d-game-engine/logz") ||
			strings.Contains(trimmed, "github.com/webbben/2d-game-engine/crashreport") {
			skip = true
			continue
		}
		skip = false
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// sanitizeStackForHash strips runtime-varying content (goroutine IDs, memory
// addresses) and line numbers so the hash is consistent across runs of the same
// crash, even after edits that shift line numbers.
func sanitizeStackForHash(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "goroutine ") || strings.HasPrefix(trimmed, "created by ") {
			continue
		}
		line = hexPattern.ReplaceAllString(line, "0x...")
		line = lineNumberPattern.ReplaceAllString(line, ".go")
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// MarkSubmittedPath marks one specific report file as submitted.
//
// Prefer this over MarkSubmitted: a hash is not a unique report identifier, since a repeat crash writes
// hash_01.json, hash_02.json, and so on. Marking by hash prefix marks every one of those at once, including
// occurrences that haven't been filed yet, so they'd be silently skipped forever.
func MarkSubmittedPath(filePath string) error {
	mu.Lock()
	defer mu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("crash report not found: %s", filePath)
	}
	var r CrashReport
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("failed to parse crash report %s: %w", filePath, err)
	}
	r.Submitted = true
	out, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, out, 0o644)
}

// MarkSubmitted marks all crash report files matching the given hash as submitted.
//
// Note this marks every occurrence of the hash, not just one. Use MarkSubmittedPath when submitting a single
// specific occurrence.
func MarkSubmitted(hash string) error {
	mu.Lock()
	defer mu.Unlock()

	if reportsDir == "" {
		return fmt.Errorf("crash reports directory not configured")
	}

	matches, err := filepath.Glob(filepath.Join(reportsDir, hash+"*"))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("crash report not found: %s", hash)
	}
	for _, filePath := range matches {
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		var r CrashReport
		if err := json.Unmarshal(data, &r); err != nil {
			continue
		}
		r.Submitted = true
		if out, err := json.MarshalIndent(r, "", "  "); err == nil {
			os.WriteFile(filePath, out, 0o644)
		}
	}
	return nil
}

// LoadUnsubmittedReports returns all crash reports that haven't been submitted yet.
func LoadUnsubmittedReports() ([]CrashReport, error) {
	pending, err := LoadUnsubmittedReportsWithPaths()
	if err != nil {
		return nil, err
	}
	reports := make([]CrashReport, 0, len(pending))
	for _, p := range pending {
		reports = append(reports, p.Report)
	}
	return reports, nil
}

// LoadUnsubmittedReportsWithPaths returns the unsubmitted crash reports along with the file each came from.
func LoadUnsubmittedReportsWithPaths() ([]PendingReport, error) {
	all, err := loadReports()
	if err != nil {
		return nil, err
	}
	var unsubmitted []PendingReport
	for _, p := range all {
		if !p.Report.Submitted {
			unsubmitted = append(unsubmitted, p)
		}
	}
	return unsubmitted, nil
}

// init sets a default reports directory from the CRASH_REPORTS_DIR env var.
func init() {
	if dir := os.Getenv("CRASH_REPORTS_DIR"); dir != "" {
		SetReportsDir(dir)
	}
}
