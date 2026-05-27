// Package pubsub provides a pub/sub event bus for use throughout the game engine
package pubsub

import (
	"fmt"

	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/utils"
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

type subscriberFn struct {
	subscriberID string
	fn           func(defs.Event)
}

type EventBus struct {
	alreadySubscribed map[string]defs.EventType // tracks what subscriptions have already been registered. used to detect extra, unintended subscriptions.
	subscribers       map[defs.EventType][]subscriberFn

	subscribeAll map[string]func(defs.Event)

	futureEventSchedule map[clock.GameTime][]defs.Event

	queue chan defs.Event
}

func (eb EventBus) GetFutureEventSchedule() map[clock.GameTime][]defs.Event {
	return eb.futureEventSchedule
}

func NewEventBus() *EventBus {
	logz.Warnln("EVENT BUS", "New event bus created! any previous subscriptions are no longer active.")
	return &EventBus{
		subscribers:         make(map[defs.EventType][]subscriberFn),
		subscribeAll:        make(map[string]func(defs.Event)),
		alreadySubscribed:   make(map[string]defs.EventType),
		futureEventSchedule: make(map[clock.GameTime][]defs.Event),
		queue:               make(chan defs.Event, 256),
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

	logz.Printf("EVENT BUS", "%s subscribed to event type %s", subscriberID, eventType)
	eb.alreadySubscribed[subscriberID] = eventType
	eb.subscribers[eventType] = append(eb.subscribers[eventType], subscriberFn{
		subscriberID: subscriberID,
		fn:           fn,
	})
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

	logz.Printf("EVENT BUS", "%s subscribed to ALL event types", subscriberID)
	eb.alreadySubscribed[subscriberID] = "ALL"
	eb.subscribeAll[subscriberID] = fn
}

// SubscribeToNPCEvents subscribes to all events related to a specific NPC
// TODO: are we unsubscribing from these? or do we need to? I guess we'd only need to unsubscribe if:
// A) the task is for activeMap only (so, when an NPC leaves active map)
// B) the NPC died
func (eb *EventBus) SubscribeToNPCEvents(subscriberID string, npcID string, fn func(defs.Event)) {
	if npcID == "" {
		panic("npcID is empty")
	}
	subID := fmt.Sprintf("%s_%s_%s", subscriberID, npcID, "assign_task")
	eb.Subscribe(subID, NpcAssignTaskType(npcID), fn)
}

// Publish handles enqueing an event to be published on the next update tick.
// This is the correct, safe way to publish an event from any thread.
func (eb *EventBus) Publish(e defs.Event) {
	eb.queue <- e
}

// dispatches a single event.
func (eb *EventBus) dispatch(e defs.Event) {
	e.Log()
	if e.Type == EventScheduleFutureEvent {
		// queuing future events will not be broadcast to anyone; it's a special event type that is recorded to be fired later on.
		// useful for doing things like causing an event to happen a day later after talking to an NPC, for example.
		eventDataVal, ok := e.Data["event"]
		if !ok {
			logz.Panicln("QueueFutureEvent", "tried to queue future event from incoming event, but data did not include future event.", e.String())
		}
		eventData, ok := eventDataVal.(defs.Event)
		if !ok {
			logz.Panicln("QueueFutureEvent", "tried to queue future event from incoming event, but data could not be type asserted to Event.", e.String(), eventDataVal)
		}
		futureTimeVal, ok := e.Data["time"]
		if !ok {
			logz.Panicln("QueueFutureEvent", "received event for queuing a future event, but event data didn't include the time.", e.String())
		}
		futureTime, ok := futureTimeVal.(clock.GameTime)
		if !ok {
			logz.Panicln("QueueFutureEvent", "received event for queuing a future event, but time data couldn't be type asserted to gametime (note: pointers can mess this up)", e.String(), futureTimeVal)
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
	// track if a (non-all) subscriber was listening for this event
	subFound := false
	for _, sub := range eb.subscribers[e.Type] {
		sub.fn(e)
		subFound = true
	}
	// broadcast every published event to the "subscribe all" list
	for _, fn := range eb.subscribeAll {
		fn(e)
	}

	if e.RequireSubscriber && !subFound {
		logz.Panicln("EVENT BUS", "Event requires subscribers, but no (non-all) subscribers were listening.", e.Type, e.Data)
	}
}

// ProcessEvents is called from the main Update loop to process all queued events.
// This must never be called asynchronously from the main update loop!
func (eb *EventBus) ProcessEvents() {
	count := 0
	for {
		select {
		case e := <-eb.queue:
			eb.dispatch(e)
			count++
		default:
			if count >= 255 {
				logz.Warnln("EVENT BUS", "Event channel buffer was maxed out! Something must be sending a lot of events... This could cause simulation lag.")
			} else if count > 50 {
				logz.Warnln("EVENT BUS", "High volume of events processed:", count)
			}
			return
		}
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

// Unsubscribe removes an event subscription. Panics if the subscriber ID isn't registered, so only use this if you are sure
// the subscription exists.
func (eb *EventBus) Unsubscribe(subID string) {
	logz.Println("EVENT BUS", "unsubscribing:", subID)

	eventID, exists := eb.alreadySubscribed[subID]
	if !exists {
		logz.Panicln("EVENT BUS", "tried to unsubscribe, but subscriber ID not found:", subID)
	}

	delete(eb.alreadySubscribed, subID)

	if eventID == "ALL" {
		if _, exists := eb.subscribeAll[subID]; !exists {
			logz.Panicln("EVENT BUS", "subscriber was mapped to the subscribeAll list, but subscriber ID not found in that map. subID:", subID)
		}
		delete(eb.subscribeAll, subID)
		return
	}

	if _, exists := eb.subscribers[eventID]; !exists {
		logz.Panicln("EVENT BUS", "event ID mapped to this subscriber doesn't appear to have any subscriptions mapped to it. subID:", subID, "mapped eventID:", eventID)
	}

	// find the specific subscriber in the map
	for i, sub := range eb.subscribers[eventID] {
		if sub.subscriberID == subID {
			// found the match; remove this index.
			eb.subscribers[eventID] = utils.RemoveIndexUnordered(eb.subscribers[eventID], i)
			return
		}
	}

	logz.Panicln("EVENT BUS", "failed to find subscriber in subscribers map; subID:", subID, "event ID mapped to this subID:", eventID)
}
