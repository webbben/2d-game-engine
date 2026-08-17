// Package logz is a set of improved logging functions
package logz

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/crashreport"
)

// MaxRecentLogs is the max number of recent log lines included in a crash report.
const MaxRecentLogs = 50

var (
	WarnColor = color.New(color.FgYellow, color.Bold)
	// PanicColor = color.RGB(0, 0, 0).Add(color.BgRed, color.Bold)
	PanicColor = color.New(color.FgHiRed, color.Bold)
	TodoColor  = color.RGB(0, 0, 0).Add(color.BgCyan)
	InfoColor  = color.New(color.FgHiBlue)
)

// ring buffer for capturing recent log lines
type logCapture struct {
	mu    sync.Mutex
	lines []string
	max   int
}

func (c *logCapture) Write(p []byte) (n int, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, line := range strings.Split(string(p), "\n") {
		if line == "" {
			continue
		}
		c.lines = append(c.lines, line)
		if len(c.lines) > c.max {
			c.lines = c.lines[len(c.lines)-c.max:]
		}
	}
	return len(p), nil
}

func (c *logCapture) Lines() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]string, len(c.lines))
	copy(result, c.lines)
	return result
}

var logCaptureBuffer *logCapture

func init() {
	logCaptureBuffer = &logCapture{max: MaxRecentLogs}
	log.SetOutput(io.MultiWriter(os.Stderr, logCaptureBuffer))
}

func GetRecentLogs() []string {
	if logCaptureBuffer == nil {
		return nil
	}
	return logCaptureBuffer.Lines()
}

func PrintFancy(s string) {
	colors := []*color.Color{
		color.New(color.FgHiMagenta, color.Bold),
		color.New(color.FgHiYellow, color.Bold),
		color.New(color.FgRed, color.Bold),
		color.New(color.FgGreen, color.Bold),
		color.New(color.FgCyan, color.Bold),
	}
	out := ""
	for i, r := range s {
		c := colors[i%(len(colors)-1)]
		out += c.Sprint(string(r))
	}
	fmt.Println(out)
}

func printLogLine(s string) {
	log.Printf("T %v %s", ebiten.Tick(), s)
}

func Println(category string, args ...any) {
	if category == "" {
		printLogLine(fmt.Sprintln(args...))
		return
	}
	fullArgs := []any{fmt.Sprintf("[%s]", category)}
	fullArgs = append(fullArgs, args...)
	printLogLine(fmt.Sprintln(fullArgs...))
}

func TODO(category string, args ...any) {
	category = TodoColor.Sprintf("[TODO: %s]", category)
	ln := fmt.Sprintln(args...)
	printLogLine(fmt.Sprintln(category, ln))
}

// Printf is the same as fmt.Printf, but adds a \n at the end for convenience. so, don't add one in your string you pass in.
func Printf(category string, format string, args ...any) {
	if category == "" {
		printLogLine(fmt.Sprintf(format, args...))
		return
	}
	format = fmt.Sprintf("[%s] %s\n", category, format)
	printLogLine(fmt.Sprintf(format, args...))
}

func Errorln(category string, args ...any) {
	fullArgs := []any{"ERROR:"}
	fullArgs = append(fullArgs, args...)
	Println(category, fullArgs...)
}

func Errorf(category string, format string, args ...any) {
	format = fmt.Sprintf("ERROR: %s", format)
	Printf(category, format, args...)
}

func Warnln(category string, args ...any) {
	category = WarnColor.Sprintf("[%s]", category)
	s := fmt.Sprintln(args...)
	printLogLine(fmt.Sprintf("%s %s", category, s))
}

func Warnf(category string, format string, args ...any) {
	category = WarnColor.Sprintf("[%s]", category)
	format = fmt.Sprintf("%s WARNING: %s", category, format)
	printLogLine(fmt.Sprintf(format, args...))
}

func Panicf(formatString string, args ...any) {
	msg := fmt.Sprintf(formatString, args...)
	printLogLine(PanicColor.Sprint("[Panic!]"))
	crashreport.WriteCrashReport(msg, debug.Stack(), GetRecentLogs())
	panic(msg)
}

func Panicln(category string, args ...any) {
	msg := fmt.Sprintln(args...)
	printLogLine(PanicColor.Sprintf("[%s]", category))
	crashreport.WriteCrashReport(fmt.Sprintf("[%s] %s", category, msg), debug.Stack(), GetRecentLogs())
	panic(msg)
}

// PanicCtx is the preferred panic call because it allows you to include a category and message, but also variable context all in one call.
// Category and Msg should not include variables, because they are used in a hash function to dedup specific errors, when logging them by a hash.
// Ctx will be logged to the console before the panic and crash report, so that it's visible to investigators.
func PanicCtx(category string, msg string, ctx ...any) {
	printLogLine(PanicColor.Sprintf("[%s]", category))
	ctxMsg := fmt.Sprintln(ctx...)
	fmt.Println(ctxMsg)
	crashreport.WriteCrashReport(fmt.Sprintf("[%s] %s", category, msg), debug.Stack(), GetRecentLogs())
	panic(msg)
}

func Panic(s string) {
	printLogLine(PanicColor.Sprint("[Panic!]"))
	crashreport.WriteCrashReport(s, debug.Stack(), GetRecentLogs())
	panic(s)
}
