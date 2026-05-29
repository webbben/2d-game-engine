package pubsub

import "github.com/webbben/2d-game-engine/data/defs"

const (
	EventScheduleFutureEvent defs.EventType = "schedule_future_event" // for queueing up an event to broadcast in the future. not noticed by global event subscribers.

	// General world

	EventVisitMap           defs.EventType = "visit_map"                // data: "MapID" (string), "MapDisplayName" (string)
	EventTimePass           defs.EventType = "time_pass"                // event called on every hour; can be used for tracking time passage
	EventMapOccupancyChange defs.EventType = "npc_map_occupancy_change" // called when an NPC changes maps

	// Maps

	EventUnlock defs.EventType = "unlock" // data: "mapID" (string), "lockID" (string)

	// Quests

	EventQuestStarted defs.EventType = "quest_started" // a quest is started by the player

	// Dialog & Topics

	EventDialogStarted defs.EventType = "dialog_started"
	EventDialogEnded   defs.EventType = "dialog_ended"

	EventNewKnowledgeTopic defs.EventType = "new_knowledge_topic" // data: "topicID" (string), "topicDisplayName" (string)

	// Objects

	EventObjectActivated defs.EventType = "object_activated"

	// Items

	EventAddItem    defs.EventType = "add_item"    // data: "itemID" (string)
	EventGoldChange defs.EventType = "gold_change" // data: "amount" (int, non-zero)

	// Player

	EventRoleAdded   defs.EventType = "role_added"   // data: "roleID" (string)
	EventRoleRemoved defs.EventType = "role_removed" // data: "roleID" (string)
)

// DataKey is the commonly used key in the Data map of an event to store specific structs.
// If this isn't used, then the event type should include specific keys info in its own description.
//
// e.g. event.Data[DataKey]
const DataKey string = "struct"

type EventObjectActivatedData struct {
	ObjectType  string // the type of object that was activated
	ActivatorID string // who activated the object
}
