package logz

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

func printLogLine(s string) {
	log.Printf("T %v %s", ebiten.Tick(), s)
}

func Println(category string, args ...any) {
	if category == "" {
		log.Println(args...)
		return
	}
	fullArgs := []any{fmt.Sprintf("[%s]", category)}
	fullArgs = append(fullArgs, args...)
	printLogLine(fmt.Sprintln(fullArgs...))
}

func Printf(category string, format string, args ...any) {
	if category == "" {
		log.Printf(format, args...)
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
	fullArgs := []any{"WARNING:"}
	fullArgs = append(fullArgs, args...)
	Println(category, fullArgs...)
}

func Warnf(category string, format string, args ...any) {
	format = fmt.Sprintf("WARNING: %s", format)
	Printf(category, format, args...)
}

func Panicf(formatString string, args ...any) {
	Println("Panic!")
	panic(fmt.Sprintf(formatString, args...))
}

func Panicln(category string, args ...any) {
	Println(category, args...)
	panic("Panic!")
}

func Panic(s string) {
	Println("Panic!")
	panic(s)
}
