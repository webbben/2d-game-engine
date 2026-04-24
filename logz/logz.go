// Package logz is a set of improved logging functions
package logz

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/hajimehoshi/ebiten/v2"
)

var (
	WarnColor = color.New(color.FgYellow, color.Bold)
	// PanicColor = color.RGB(0, 0, 0).Add(color.BgRed, color.Bold)
	PanicColor = color.New(color.FgHiRed, color.Bold)
	TodoColor  = color.RGB(0, 0, 0).Add(color.BgCyan)
)

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
	printLogLine(PanicColor.Sprint("[Panic!]"))
	panic(fmt.Sprintf(formatString, args...))
}

func Panicln(category string, args ...any) {
	printLogLine(PanicColor.Sprintf("[%s]", category))
	panic(fmt.Sprintln(args...))
}

func Panic(s string) {
	printLogLine(PanicColor.Sprint("[Panic!]"))
	panic(s)
}
