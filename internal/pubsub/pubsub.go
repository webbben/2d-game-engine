package pubsub

import (
	"fmt"

	"github.com/webbben/2d-game-engine/internal/logz"
)

const (
	Event_EntityAttacked string = "entity_attacked" // (TODO) entity was attacked by player's attacks
	Event_EntityKilled   string = "entity_killed"   // (TODO) entity was directly killed by player's attacks
	Event_EntityDied     string = "entity_died"     // (TODO) entity died (but not by player's hand)
	Event_StartDialog    string = "start_dialog"    // player started dialog
	Event_StartTopic     string = "start_topic"     // player has started a topic in a dialog
	Event_EndDialog      string = "end_dialog"      // player ended dialog
	Event_GetItem        string = "get_item"        // (TODO) player gets item
	Event_UseItem        string = "use_item"        // (TODO) player uses (or equips) an item
	Event_VisitMap       string = "visit_map"       // (TODO) player enters a map
	Event_TimePass       string = "time_pass"       // (TODO) event called on every hour; can be used for tracking time passage
)

type Event struct {
	Type string
	Data map[string]any
}

func (e Event) log() {
	logz.Println("EVENT:", e.Type)
	for k, v := range e.Data {
		fmt.Printf("%s: %s\n", k, v)
	}
}

type EventBus struct {
	subscribers map[string][]func(Event)
}

func NewEventBus() *EventBus {
	logz.Warnln("EVENT BUS", "New event bus created! any previous subscriptions are no longer active.")
	return &EventBus{
		subscribers: make(map[string][]func(Event)),
	}
}

func (eb *EventBus) Subscribe(eventType string, fn func(Event)) {
	eb.subscribers[eventType] = append(eb.subscribers[eventType], fn)
}

func (eb *EventBus) Publish(e Event) {
	e.log()
	for _, fn := range eb.subscribers[e.Type] {
		fn(e)
	}
}
