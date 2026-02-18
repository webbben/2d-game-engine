// Package pubsub provides a pub/sub event bus for use throughout the game engine
package pubsub

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/internal/logz"
)

const (
	// General world

	EventVisitMap defs.EventType = "visit_map" // player enters a map
	EventTimePass defs.EventType = "time_pass" // event called on every hour; can be used for tracking time passage

	// Quests

	EventQuestStarted defs.EventType = "quest_started" // a quest is started by the player

	// Dialog

	EventDialogStarted defs.EventType = "dialog_started"
	EventDialogEnded   defs.EventType = "dialog_ended"
)

func NpcAssignTaskType(npcID string) defs.EventType {
	return defs.EventType(fmt.Sprintf("NPC:%s:assign_task", npcID))
}

// NPCAssignTask returns an event definition for assigning a task to an NPC
func NPCAssignTask(npcID string, taskDef defs.TaskDef) defs.Event {
	return defs.Event{
		Type: NpcAssignTaskType(npcID),
		Data: map[string]any{
			"taskDef": taskDef,
		},
	}
}

type EventBus struct {
	alreadySubscribed map[string]bool // tracks what subscriptions have already been registered. used to detect extra, unintended subscriptions.
	subscribers       map[defs.EventType][]func(defs.Event)
	subscribeAll      map[string]func(defs.Event)
}

func NewEventBus() *EventBus {
	logz.Warnln("EVENT BUS", "New event bus created! any previous subscriptions are no longer active.")
	return &EventBus{
		subscribers:       make(map[defs.EventType][]func(defs.Event)),
		subscribeAll:      make(map[string]func(defs.Event)),
		alreadySubscribed: make(map[string]bool),
	}
}

// Subscribe subscribes a function to all events of a certain event type.
//
// subscriberID: describes who/where is subscribed. used for detecting duplicate subscriptions, so if a place subscribes to multiple events,
// give the subscriberID a per-eventType based ID. Will panic if the same subscriberID is given more than once.
func (eb *EventBus) Subscribe(subscriberID string, eventType defs.EventType, fn func(defs.Event)) {
	if subscriberID == "" {
		panic("subscriber ID empty")
	}
	if _, exists := eb.alreadySubscribed[subscriberID]; exists {
		logz.Panicln("EVENT BUS", "duplicate subscription detected:", subscriberID)
	}

	logz.Printf("EVENT BUS", "%s subscribed to event type %s\n", subscriberID, eventType)
	eb.alreadySubscribed[subscriberID] = true
	eb.subscribers[eventType] = append(eb.subscribers[eventType], fn)
}

// SubscribeAll subscribes a function to all events of a certain event type.
//
// subscriberID should be unique and not registered for any other subscription; otherwise we panic, to avoid double subscription of the same thing.
func (eb *EventBus) SubscribeAll(subscriberID string, fn func(defs.Event)) {
	if subscriberID == "" {
		panic("subscriber ID empty")
	}
	if _, exists := eb.alreadySubscribed[subscriberID]; exists {
		logz.Panicln("EVENT BUS", "duplicate subscription detected:", subscriberID)
	}

	logz.Printf("EVENT BUS", "%s subscribed to ALL event types\n", subscriberID)
	eb.alreadySubscribed[subscriberID] = true
	eb.subscribeAll[subscriberID] = fn
}

// SubscribeToNPCEvents subscribes to all events related to a specific NPC
func (eb *EventBus) SubscribeToNPCEvents(subscriberID string, npcID string, fn func(defs.Event)) {
	if npcID == "" {
		panic("npcID is empty")
	}
	subID := fmt.Sprintf("%s_%s_%s", subscriberID, npcID, "assign_task")
	eb.Subscribe(subID, NpcAssignTaskType(npcID), fn)
}

func (eb *EventBus) Publish(e defs.Event) {
	e.Log()
	for _, fn := range eb.subscribers[e.Type] {
		fn(e)
	}
	// broadcast every published event to the "subscribe all" list
	for _, fn := range eb.subscribeAll {
		fn(e)
	}
}
