package defs

import (
	"fmt"

	"github.com/webbben/2d-game-engine/logz"
)

type (
	EventType string
)

type Event struct {
	Type EventType
	Data map[string]any
}

func (e Event) String() string {
	return fmt.Sprintf("%s (%v)", e.Type, e.Data)
}

func (e Event) Log() {
	logz.Println("EVENT", e)
}
