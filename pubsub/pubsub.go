// Package pubsub provides a pub/sub event bus for use throughout the game engine
package pubsub

import (
	"fmt"

	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

const (
	EventScheduleFutureEvent defs.EventType = "schedule_future_event" // for queueing up an event to broadcast in the future. not noticed by global event subscribers.

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

	futureEventSchedule map[clock.GameTime][]defs.Event
}

func NewEventBus() *EventBus {
	logz.Warnln("EVENT BUS", "New event bus created! any previous subscriptions are no longer active.")
	return &EventBus{
		subscribers:         make(map[defs.EventType][]func(defs.Event)),
		subscribeAll:        make(map[string]func(defs.Event)),
		alreadySubscribed:   make(map[string]bool),
		futureEventSchedule: make(map[clock.GameTime][]defs.Event),
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
	if e.Type == EventScheduleFutureEvent {
		// queuing future events will not be broadcast to anyone; it's a special event type that is recorded to be fired later on.
		// useful for doing things like causing an event to happen a day later after talking to an NPC, for example.
		eventData, ok := e.Data["event"].(defs.Event)
		if !ok {
			logz.Panicln("QueueFutureEvent", "tried to queue future event from incoming event, but data could not be type asserted to Event.", e.String())
		}
		futureTime, ok := e.Data["time"].(clock.GameTime)
		if !ok {
			logz.Panicln("QueueFutureEvent", "received event for queuing a future event, but event data didn't include the time.", e.String())
		}
		eb.ScheduleFutureEvent(eventData, futureTime)
		return
	}
	if e.Type == EventTimePass {
		// check if any scheduled events should fire.
		// we fire any event that is scheduled for the current hour or anytime in the past.
		gameTime, ok := e.Data["gameTime"].(clock.GameTime)
		if !ok {
			panic("hour change event didn't have game time")
		}
		eb.FireScheduledEvents(gameTime)
	}
	for _, fn := range eb.subscribers[e.Type] {
		fn(e)
	}
	// broadcast every published event to the "subscribe all" list
	for _, fn := range eb.subscribeAll {
		fn(e)
	}
}

func (eb *EventBus) ScheduleFutureEvent(e defs.Event, futureTime clock.GameTime) {
	if e.Type == EventScheduleFutureEvent {
		logz.Panicln("QueueFutureEvent", "cannot queue a future event that is also a 'queue future event' type.", e)
	}
	if _, exists := eb.futureEventSchedule[futureTime]; !exists {
		eb.futureEventSchedule[futureTime] = []defs.Event{}
	}
	eb.futureEventSchedule[futureTime] = append(eb.futureEventSchedule[futureTime], e)

	logz.Println("EVENT BUS", "Queued future event:", e.Type, "Scheduled for:", futureTime)
}

func (eb *EventBus) FireScheduledEvents(currentTime clock.GameTime) {
	for gt, events := range eb.futureEventSchedule {
		if currentTime.IsAfter(gt) || currentTime.IsEqual(gt) {
			for _, e := range events {
				eb.Publish(e)
			}
			delete(eb.futureEventSchedule, gt)
		}
	}
}
