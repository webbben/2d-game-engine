package crashreport

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteAndLoadReport(t *testing.T) {
	dir := t.TempDir()
	if err := SetReportsDir(dir); err != nil {
		t.Fatal(err)
	}

	WriteCrashReport("hello from test", []byte("goroutine 1:\ntest.go:10 foo\ntest.go:20 bar\nruntime/asm.s:100 baz"), nil)

	reports, err := LoadAllReports()
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	r := reports[0]
	if r.Message != "hello from test" {
		t.Errorf("wrong message: %q", r.Message)
	}
	if r.Hash == "" {
		t.Error("hash was empty")
	}
	if !contains(r.Stack, "goroutine 1") {
		t.Error("stack should contain goroutine header")
	}
}

func TestDedupCreatesNumberedFiles(t *testing.T) {
	dir := t.TempDir()
	if err := SetReportsDir(dir); err != nil {
		t.Fatal(err)
	}

	stack := []byte("goroutine 1:\ntest.go:10 foo\ntest.go:20 bar")
	WriteCrashReport("same error", stack, nil)
	WriteCrashReport("same error", stack, nil)
	WriteCrashReport("same error", stack, nil)

	reports, err := LoadAllReports()
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 3 {
		t.Fatalf("expected 3 numbered reports, got %d", len(reports))
	}
}

func TestDifferentErrorsNotDeduped(t *testing.T) {
	dir := t.TempDir()
	if err := SetReportsDir(dir); err != nil {
		t.Fatal(err)
	}

	WriteCrashReport("error one", []byte("goroutine 1:\nfoo.go:10 foo"), nil)
	WriteCrashReport("error two", []byte("goroutine 1:\nfoo.go:20 bar"), nil)

	reports, err := LoadAllReports()
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
}

func TestFilterStack(t *testing.T) {
	stack := "goroutine 1 [running]:\n" +
		"github.com/webbben/2d-game-engine/logz.Panicln(...)\n" +
		"\t/path/to/logz.go:94 +0x...\n" +
		"github.com/webbben/2d-game-engine/crashreport.WriteCrashReport(...)\n" +
		"\t/path/to/crashreport.go:42 +0x...\n" +
		"mygame.MyFunc(...)\n" +
		"\t/path/to/mygame.go:42 +0x...\n" +
		"main.main()\n" +
		"\t/path/to/main.go:100 +0x...\n" +
		"runtime.main(...)\n" +
		"\t/runtime/proc.go:250 +0x..."
	filtered := filterStack([]byte(stack))
	if contains(filtered, "logz") {
		t.Error("filtered stack should not contain logz frames")
	}
	if contains(filtered, "crashreport") {
		t.Error("filtered stack should not contain crashreport frames")
	}
	if contains(filtered, "runtime/") || contains(filtered, "runtime.main") {
		t.Error("filtered stack should not contain runtime frames")
	}
	if !contains(filtered, "MyFunc") {
		t.Error("filtered stack should contain user code frames")
	}
	if !contains(filtered, "main.main") {
		t.Error("filtered stack should contain main")
	}
}

func TestReportsDirNotSet(t *testing.T) {
	// Set to empty to disable
	SetReportsDir("")
	// Should not panic
	WriteCrashReport("should not write", []byte("goroutine 1:\nfoo.go:10 foo"), nil)

	// Reload from a temp dir to confirm nothing was written
	dir := t.TempDir()
	SetReportsDir(dir)
	reports, err := LoadAllReports()
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 0 {
		t.Errorf("expected 0 reports, got %d", len(reports))
	}
}

func TestHashStableAcrossLineShifts(t *testing.T) {
	dir := t.TempDir()
	if err := SetReportsDir(dir); err != nil {
		t.Fatal(err)
	}

	base := "github.com/webbben/2d-game-engine/entity.(*Entity).TryMoveMaxPx(...)\n" +
		"\t/Users/me/entity/movement.go:187 +0x16c\n" +
		"github.com/webbben/2d-game-engine/world/npc.(*FightTask).handleCombat(...)\n" +
		"\t/Users/me/npc/task_fight.go:246 +0x5b8\n"
	shifted := strings.ReplaceAll(base, "task_fight.go:246", "task_fight.go:245")

	WriteCrashReport("same crash", []byte(base), nil)
	WriteCrashReport("same crash", []byte(shifted), nil)

	reports, err := LoadAllReports()
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
	if reports[0].Hash != reports[1].Hash {
		t.Errorf("line shifts should not change crash hash:\n  %s\n  %s", reports[0].Hash, reports[1].Hash)
	}
}

func TestHashStableAcrossAddressAndGoroutineVariance(t *testing.T) {
	dir := t.TempDir()
	if err := SetReportsDir(dir); err != nil {
		t.Fatal(err)
	}

	stackA := "goroutine 10 [running]:\n" +
		"mygame.Foo(0x14000abc, 0x0?)\n" +
		"\t/path/to/foo.go:10 +0x1a2\n"
	stackB := "goroutine 12 [running]:\n" +
		"mygame.Foo(0x14000def, 0x0?)\n" +
		"\t/path/to/foo.go:10 +0x3b4\n"

	WriteCrashReport("same crash", []byte(stackA), nil)
	WriteCrashReport("same crash", []byte(stackB), nil)

	reports, err := LoadAllReports()
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
	if reports[0].Hash != reports[1].Hash {
		t.Errorf("goroutine IDs/addresses should not change crash hash:\n  %s\n  %s", reports[0].Hash, reports[1].Hash)
	}
}

func TestSetReportsDir(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "crashreport_test")
	defer os.RemoveAll(dir)

	if err := SetReportsDir(dir); err != nil {
		t.Fatal(err)
	}
	if ReportsDir() != dir {
		t.Errorf("expected %q, got %q", dir, ReportsDir())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
