package defs

import (
	"github.com/webbben/2d-game-engine/internal/logz"
)

/*
*
* What the Quest System is responsible for:
* - Maintain persistent narrative state
* - react to events
* - advance through stages
* - trigger world-side effects
* - track objectives
*
* It should NOT:
* - execute AI
* - run animations
* - modify NPC internals directly
* - contain low-level behavior logic
*
* it's a state machine that reacts to events.
*
 */

type (
	QuestID      string
	QuestStageID string

	QuestActionType     string
	QuestConditionType  string
	QuestTerminalStatus int // represents if a quest stage triggers the end of a quest (failure, success, or still going).
)

type QuestDef struct {
	ID          QuestID
	Name        string // actual name that the player sees
	Description string // initial description of the quest (shown when accepted)
	Stages      map[QuestStageID]QuestStageDef
	StartStage  QuestStageID

	StartTrigger QuestStartTrigger // REQ: this is what causes the quest to begin.
}

// QuestStartTrigger defines conditions that will cause the quest to start.
// You must define an event for it to listen to, and then when such an event happens, it will check if its conditions are valid.
//
// Many quests are set to listen for "narrative moment" style events. Basically, events that represent a moment in a dialog, or something like that.
type QuestStartTrigger struct {
	EventType  EventType // REQ: event type that causes this trigger to be evaluated
	Conditions []QuestConditionDef
}

func (qd QuestDef) Validate() {
	if qd.ID == "" {
		panic("id is empty")
	}
	if qd.Name == "" {
		logz.Panicln(string(qd.ID), "quest def has no name set")
	}
	if len(qd.Stages) == 0 {
		logz.Panicln(string(qd.ID), "quest def has no stages")
	}
	if qd.StartTrigger.EventType == "" {
		logz.Panicln(string(qd.ID), "quest def has no start trigger event type set")
	}
	startStageFound := false
	for _, stage := range qd.Stages {
		if stage.ID == qd.StartStage {
			startStageFound = true
		}
		stage.Validate()
	}
	if !startStageFound {
		logz.Panicln(string(qd.ID), "quest def start stage not found in quest stages")
	}
}

func NewQuestDef(id QuestID, name string, stages map[QuestStageID]QuestStageDef, startStage QuestStageID) QuestDef {
	qd := QuestDef{
		ID:         id,
		Name:       name,
		Stages:     stages,
		StartStage: startStage,
	}
	qd.Validate()
	return qd
}

// QuestStageDef defines a specific stage in a quest.
//
// Each stage:
//
// - Knows what events it cares about
//
// - Knows what happens when those events occur
//
// - Knows what the next stage is
type QuestStageDef struct {
	ID          QuestStageID
	Title       string        // OPT: shows a title for the quest stage to the player
	Objective   string        // REQ: what the player sees as the current objective. Should be somewhat brief and to the point.
	Description string        // OPT: longer context about the current objective, and adds some narrative to the events.
	OnEnter     []QuestAction // a list of (non-conditional) actions that execute right when this stage is reached.
	Reactions   []QuestReactionDef
}

func (stage QuestStageDef) Validate() {
	if stage.ID == "" {
		panic("stage ID was empty")
	}
	if len(stage.Reactions) == 0 {
		panic("stage has no reactions... so this quest would never be able to progress from here")
	}
	for _, reaction := range stage.Reactions {
		reaction.Validate()
	}
}

// QuestReactionDef defins how a quest stage reacts, based on conditions.
//
// This means: when event X happens, if conditions pass, run actions and possibly move to next stage (if defined).
type QuestReactionDef struct {
	SubscribeEvent EventType // The event type to listen to, that triggers this Reaction to run, and check its conditions. This is so that we aren't constantly checking conditions on every loop.
	Conditions     []QuestConditionDef
	Actions        []QuestAction
	NextStage      QuestStageID        // Points to the next quest stage. If a next stage is set, there should be no terminal status.
	TerminalStatus QuestTerminalStatus // Determines if this reaction is a "quest end". If this is set, there should be no NextStage
	// if this reaction should inform the player about anything, it can go here. especially if this is a terminal reaction, since the quest could use a conclusion.
	// TODO: make sure this is recorded somewhere? I guess the quest state should be able to find this, at least.
	Text string
}

func (qr QuestReactionDef) Validate() {
	if qr.SubscribeEvent == "" {
		panic("no subscribe event set; no way for this reaction to be triggered for condition evaluation")
	}
	// TODO: validate conditions and all that
}

// QuestAction defines actions that occur right when a quest stage is reached.
//
// Examples of actions concepts:
//
// - Show quest update notification
//
// - Spawn an NPC
//
// - Unlock a door
//
// - Set a quest variable
//
// - Trigger cutscene
//
// ... etc.
type QuestAction interface {
	Fire(ctx QuestActionContext)
}

type QuestActionContext interface {
	// assigns a task to an NPC;
	// NPCs are assigned to characters, so you pass a CharacterDefID.
	// The characterDef must be "unique", which means only one instance can exist.
	AssignTaskToNPC(id CharacterDefID, taskDef TaskDef)
	QueueScenario(id ScenarioID)
	UnlockMapLock(mapID MapID, lockID string)
}

type QuestConditionDef struct {
	Type   QuestConditionType
	Params map[string]string
}
