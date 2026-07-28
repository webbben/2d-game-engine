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
	mu.Lock()
	defer mu.Unlock()

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
	if reportsDir == "" {
		return nil, fmt.Errorf("crash reports directory not configured")
	}
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read crash reports directory: %w", err)
	}
	var reports []CrashReport
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(reportsDir, e.Name()))
		if err != nil {
			continue
		}
		var r CrashReport
		if json.Unmarshal(data, &r) == nil {
			reports = append(reports, r)
		}
	}
	return reports, nil
}

// LoadReport loads a single crash report by its hash.
func LoadReport(hash string) (*CrashReport, error) {
	if reportsDir == "" {
		return nil, fmt.Errorf("crash reports directory not configured")
	}
	data, err := os.ReadFile(filepath.Join(reportsDir, hash+".json"))
	if err != nil {
		return nil, fmt.Errorf("crash report not found: %s", hash)
	}
	var r CrashReport
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("failed to parse crash report: %w", err)
	}
	return &r, nil
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
// addresses) so the hash is consistent across runs of the same crash.
func sanitizeStackForHash(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "goroutine ") || strings.HasPrefix(trimmed, "created by ") {
			continue
		}
		out = append(out, hexPattern.ReplaceAllString(line, "0x..."))
	}
	return strings.Join(out, "\n")
}

// MarkSubmitted marks all crash report files matching the given hash as submitted.
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
	all, err := LoadAllReports()
	if err != nil {
		return nil, err
	}
	var unsubmitted []CrashReport
	for _, r := range all {
		if !r.Submitted {
			unsubmitted = append(unsubmitted, r)
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
