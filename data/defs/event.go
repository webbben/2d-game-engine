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

	// if true, this event will panic if there are no (non-all) subscribers.
	// meant for events that are crucial to be listened to, and to which critical logic or behavior is tied to.
	RequireSubscriber bool
}

func (e Event) String() string {
	return fmt.Sprintf("%s (%v)", e.Type, e.Data)
}

func (e Event) Log() {
	logz.Println("EVENT", e)
}
