package logz

import (
	"fmt"
	"log"
)

func Println(category string, args ...any) {
	if category == "" {
		log.Println(args...)
		return
	}
	fullArgs := []any{fmt.Sprintf("[%s]", category)}
	fullArgs = append(fullArgs, args...)
	log.Println(fullArgs...)
}

func Printf(category string, format string, args ...any) {
	if category == "" {
		log.Printf(format, args...)
		return
	}
	format = fmt.Sprintf("[%s] %s", category, format)
	log.Printf(format, args...)
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
	format := fmt.Sprintf(formatString, args...)
	panic(format)
}
