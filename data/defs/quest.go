package defs

import (
	"github.com/webbben/2d-game-engine/logz"
)

type (
	QuestID      string
	QuestStageID string

	QuestActionType     string
	QuestConditionType  string
	QuestTerminalStatus int // represents if a quest stage triggers the end of a quest (failure, success, or still going).
	QuestStatus         string
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
	if qd.StartStage == "" {
		logz.Panicln(string(qd.ID), "quest def had no start stage defined")
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
	ID             QuestStageID
	Title          string        // OPT: shows a title for the quest stage to the player
	Objective      string        // REQ: what the player sees as the current objective. Should be somewhat brief and to the point.
	Description    string        // OPT: longer context about the current objective, and adds some narrative to the events.
	OnEnter        []WorldEffect // a list of (non-conditional) effects that execute right when this stage is reached.
	Reactions      []QuestReactionDef
	TerminalStatus QuestTerminalStatus // Determines if this reaction is a "quest end". If this is set, there should be no reactions
}

func (stage QuestStageDef) Validate() {
	if stage.ID == "" {
		panic("stage ID was empty")
	}
	if len(stage.Reactions) == 0 && stage.TerminalStatus == 0 {
		logz.Panicln("QuestStageDef", "no reactions, and no terminal status set; one of these must be set, or else there is no conclusion to the quest.", stage.ID)
	}
	if stage.TerminalStatus != 0 {
		if len(stage.Reactions) != 0 {
			logz.Panicln("QuestStageDef", "a terminal status is set, but reactions are also set. only one or the other should be set.", stage.ID)
		}
		if stage.Objective != "" {
			logz.Panicln("QuestStageDef", "an objective is set for a terminal stage.", stage.ID)
		}
	}
	for _, reaction := range stage.Reactions {
		reaction.Validate()
	}
}

// QuestReactionDef defins how a quest stage reacts, based on conditions.
//
// This means: when event X happens, if conditions pass, run actions and possibly move to next stage.
// Quest Reactions do not directly end a quest; to end a quest (to complete or fail it), the reaction must move it to a terminal stage.
type QuestReactionDef struct {
	SubscribeEvent EventType // The event type to listen to, that triggers this Reaction to run, and check its conditions. This is so that we aren't constantly checking conditions on every loop.
	Conditions     []QuestConditionDef
	Effects        []WorldEffect
	NextStage      QuestStageID // Points to the next quest stage.
}

func (qr QuestReactionDef) Validate() {
	if qr.SubscribeEvent == "" {
		panic("no subscribe event set; no way for this reaction to be triggered for condition evaluation")
	}
	if qr.NextStage == "" {
		logz.Panicln("QuestReactionDef", "no next stage defined.", qr.SubscribeEvent)
	}
}

type QuestConditionDef struct {
	Type   QuestConditionType
	Params map[string]string
}
