// Package quest defines logic for managing quests
package quest

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/pubsub"
)

const (
	TerminalStatusNone defs.QuestTerminalStatus = iota
	TerminalStatusComplete
	TerminalStatusFail
)

const (
	Active     state.QuestStatus = "ACTIVE"
	Completed  state.QuestStatus = "COMPLETED"
	Failed     state.QuestStatus = "FAILED"
	NotStarted state.QuestStatus = "NOT_STARTED"
)

// Quest event data keys
const (
	QuestIDKey string = "QUEST_ID"
)

type QuestManager struct {
	// we define quest defs here instead of the definitionManager, just because it makes sense to store them here.
	// I don't think any other place in the game will need to be able to get quest defs from the definition manager, except for maybe
	// some UI pages that want to display quest info.
	questDefs map[defs.QuestID]defs.QuestDef

	startTriggersByEvent  map[defs.EventType][]defs.QuestID // maps event types to all quests that have start triggers that listen to that event
	stageReactionsByEvent map[defs.EventType][]defs.QuestID // maps event types to all quests that have stage reactions that listen to that event

	notStarted map[defs.QuestID]bool // tracks which quests have not started yet, and therefore should have their start triggers checked
	active     map[defs.QuestID]*state.QuestState
	completed  map[defs.QuestID]*state.QuestState
	failed     map[defs.QuestID]*state.QuestState

	world WorldController

	eventBus *pubsub.EventBus
}

type WorldController interface {
	AssignTaskToNPC(id defs.CharacterDefID, taskDef defs.TaskDef)
	QueueScenario(id defs.ScenarioID)
	UnlockMapLock(mapID defs.MapID, lockID string)
}

func NewQuestManager(eventBus *pubsub.EventBus, worldRef WorldController) *QuestManager {
	if eventBus == nil {
		panic("event bus was nil")
	}
	if worldRef == nil {
		panic("worldRef was nil")
	}
	qm := QuestManager{
		eventBus: eventBus,
		world:    worldRef,

		questDefs: make(map[defs.QuestID]defs.QuestDef),
		// NOTE: I purposely did not create the index maps here, since I want to be able to tell
		// when they've explicitly been created (which is done after loading all quest data).

		// quest buckets
		notStarted: make(map[defs.QuestID]bool),
		active:     make(map[defs.QuestID]*state.QuestState),
		completed:  make(map[defs.QuestID]*state.QuestState),
		failed:     make(map[defs.QuestID]*state.QuestState),
	}

	qm.eventBus.SubscribeAll("QuestManager", qm.OnEvent)

	return &qm
}

func (qm *QuestManager) OnEvent(event defs.Event) {
	if len(qm.questDefs) == 0 {
		logz.Panicln("QuestManager", "OnEvent: no quest defs found. are you sure you loaded quest data yet?")
	}
	if qm.startTriggersByEvent == nil || qm.stageReactionsByEvent == nil {
		logz.Panicln("QuestManager", "OnEvent: event indices were nil. are you sure you created event indices/loaded quest data yet?")
	}

	logz.Println("QuestManager:OnEvent", "event incoming:", event.Type)

	// check for quest start triggers
	notStartedQuests, exists := qm.startTriggersByEvent[event.Type]
	if exists {
		logz.Println("QuestManager:OnEvent", "event type matches with a quest start trigger")
		for _, questID := range notStartedQuests {
			// it's possible some of these could've started, so skip those
			if qm.GetQuestStatus(questID) != NotStarted {
				continue
			}
			questDef := qm.questDefs[questID]
			if questDef.StartTrigger.EventType != event.Type {
				panic("the start trigger event type doesn't match... did we mess up the indexing somehow?")
			}
			if qm.conditionsMet(questDef.StartTrigger.Conditions, questID, event) {
				qm.StartQuest(questID)
				// NOTE: I considered if we should expect a single event to cause more than one quest to start.
				// I think the answer is "yes": perhaps an event occurs where the player is now at a crossroads and needs to decide
				// one path/questline vs another. It could start two questlines that are mutually exclusive.
			}
		}
	}

	activeQuests, exists := qm.stageReactionsByEvent[event.Type]
	if exists {
		logz.Println("QuestManager:OnEvent", "event type matches with a stage reaction")
		for _, questID := range activeQuests {
			if qm.GetQuestStatus(questID) != Active {
				// possibly, this quest was active at load time but has already finished
				continue
			}
			stage := qm.GetActiveQuestStage(questID)
			for _, reaction := range stage.Reactions {
				if reaction.SubscribeEvent != event.Type {
					continue
				}
				if qm.conditionsMet(reaction.Conditions, questID, event) {
					qm.RunReaction(questID, reaction, event)
					// only one reaction can run per stage; first one that executes is the one we use
					break
				}
			}
			// NOTE: it's possible that no reactions will have a matching event type.
			// The reason is, we only index to questID, but we don't say which stage. So, the reaction for this event might just
			// be on a different stage.
		}
	}
}

func (qm *QuestManager) RunReaction(questID defs.QuestID, reaction defs.QuestReactionDef, event defs.Event) {
	for _, action := range reaction.Actions {
		action.Fire(qm.world)
	}

	switch reaction.TerminalStatus {
	case TerminalStatusNone:
		if reaction.NextStage == "" {
			logz.Panicln("QuestManager", "questReaction has no terminal status or next stage set:", questID)
		}
		qm.SetQuestStage(questID, reaction.NextStage)
	case TerminalStatusComplete:
		qm.CompleteQuest(questID)
	case TerminalStatusFail:
		qm.FailQuest(questID)
	default:
		logz.Panicln("QuestManager", "terminal status not recognized:", reaction.TerminalStatus)
	}
}

func (qm *QuestManager) SetQuestStage(questID defs.QuestID, nextStage defs.QuestStageID) {
	questDef := qm.GetQuestDef(questID)

	// get stage def
	stageDef := questDef.Stages[nextStage]

	for _, action := range stageDef.OnEnter {
		action.Fire(qm.world)
	}

	// set current stage ID in quest state
	questState := qm.GetActiveQuestState(questID)
	questState.CurrentStage = nextStage
}

func (qm QuestManager) GetActiveQuestStage(questID defs.QuestID) defs.QuestStageDef {
	questState := qm.active[questID]
	currentStageID := questState.CurrentStage
	questDef := qm.GetQuestDef(questID)
	return questDef.Stages[currentStageID]
}

func (qm *QuestManager) loadQuestDef(d defs.QuestDef) {
	d.Validate()
	qm.questDefs[d.ID] = d

	// add it here, and then as quest states are loaded in later, we can delete them from this map.
	qm.notStarted[d.ID] = true
}

// LoadAllQuestData loads all the quest data, ensuring the proper validations and index creation is done.
func (qm *QuestManager) LoadAllQuestData(questDefs []defs.QuestDef, questStates []state.QuestState) {
	for _, questDef := range questDefs {
		qm.loadQuestDef(questDef)
	}
	for _, questState := range questStates {
		qm.loadQuestState(questState)
	}

	logz.Println("QuestManager:LoadAllQuestData", "Loaded", len(questDefs), "quest defs and", len(questStates), "quest states")

	// now that all quest defs and states are loaded in, create the indices
	qm.CreateEventTypeIndices()
}

// CreateEventTypeIndices creates all the event indices that we use for knowing which quests are listening to which events.
// This should only be done once all quest defs and quest states have been loaded.
// It should probably be done once at the beginning of every play session, and also whenever a new quest is started.
// This is because when a new quest is started, we now need to find its stage reaction events and add them.
// NOTE: if we decide there's a performance hit on creating these indices, we can adapt to just adding the new quest's
// stage reaction events instead of recalculating all indices.
func (qm *QuestManager) CreateEventTypeIndices() {
	logz.Println("QuestManager", "creating event type indices")
	qm.stageReactionsByEvent = make(map[defs.EventType][]defs.QuestID)
	qm.startTriggersByEvent = make(map[defs.EventType][]defs.QuestID)

	// setup start triggers index for quests that haven't been started yet
	for questID := range qm.notStarted {
		d := qm.GetQuestDef(questID)
		if _, exists := qm.startTriggersByEvent[d.StartTrigger.EventType]; !exists {
			qm.startTriggersByEvent[d.StartTrigger.EventType] = []defs.QuestID{}
		}
		qm.startTriggersByEvent[d.StartTrigger.EventType] = append(qm.startTriggersByEvent[d.StartTrigger.EventType], d.ID)
	}

	// setup stage reaction events index for quests that are active
	for questID := range qm.active {
		d := qm.GetQuestDef(questID)
		for _, stage := range d.Stages {
			for _, reaction := range stage.Reactions {
				if _, exists := qm.stageReactionsByEvent[reaction.SubscribeEvent]; !exists {
					qm.stageReactionsByEvent[reaction.SubscribeEvent] = []defs.QuestID{}
				}
				// we only index by quest ID, even though these reactions are more specifically associated to a specific stage.
				// it should work fine; what it means is, when an event happens and it matches to a questID in this map, we have to look up the
				// quest state's current stage, and then check if the reactions in that stage ID are valid.
				// I don't think this will lead to too much inefficiency since the event types will usually be pretty unique for quests.
				// But, if we notice a problem, maybe we can consider mapping to a struct that has both quest ID and stage ID or something more precise.
				qm.stageReactionsByEvent[reaction.SubscribeEvent] = append(qm.stageReactionsByEvent[reaction.SubscribeEvent], d.ID)
			}
		}
	}

	if len(qm.stageReactionsByEvent) == 0 && len(qm.startTriggersByEvent) == 0 {
		panic("both indices are empty... this doesn't seem right")
	}

	logz.Println("QuestManager", "start triggers index:", qm.startTriggersByEvent)
	logz.Println("QuestManager", "stage reaction triggers index:", qm.stageReactionsByEvent)
}

func (qm *QuestManager) loadQuestState(questState state.QuestState) {
	// TODO: validate quest states
	switch questState.Status {
	case Active:
		qm.active[questState.DefID] = &questState
	case Completed:
		qm.completed[questState.DefID] = &questState
	case Failed:
		qm.failed[questState.DefID] = &questState
	case NotStarted:
		logz.Panicln("QuestManager", "loading quest state, but state has a NotStarted status...")
	default:
		logz.Panicln("QuestManager", "loaded quest state has unrecognized status:", questState.DefID, questState.Status)
	}

	// remove it from notStarted
	if _, exists := qm.notStarted[questState.DefID]; !exists {
		logz.Panicln("QuestManager", `
			loaded quest state was not found in notStarted bucket. this should've been set when questDefs were loaded. 
			ensure that all quest defs are loaded into QuestManager before any quest states are.
			`)
	}
	delete(qm.notStarted, questState.DefID)
}

func (qm *QuestManager) GetQuestDef(id defs.QuestID) defs.QuestDef {
	if id == "" {
		panic("id was empty")
	}
	questDef, exists := qm.questDefs[id]
	if !exists {
		logz.Panicln("QuestManager", "tried to get quest def, but id not found:", id)
	}
	return questDef
}

func (qm QuestManager) GetActiveQuestState(id defs.QuestID) *state.QuestState {
	if id == "" {
		panic("id was empty")
	}
	questState, exists := qm.active[id]
	if !exists {
		logz.Panicln("QuestManager", "tried to get active quest state, but it didn't exist:", id)
	}

	return questState
}

// GetQuestStatus checks (and confirms validity) of a quest's status
func (qm QuestManager) GetQuestStatus(id defs.QuestID) state.QuestStatus {
	if questState, exists := qm.active[id]; exists {
		if questState.Status != Active {
			logz.Panicln("QuestManager", "GetQuestStatus: quest is in active bucket, but status is wrong:", questState.Status)
		}
		return Active
	}
	if questState, exists := qm.completed[id]; exists {
		if questState.Status != Completed {
			logz.Panicln("QuestManager", "GetQuestStatus: quest is in completed bucket, but status is wrong:", questState.Status)
		}
		return Completed
	}
	if questState, exists := qm.failed[id]; exists {
		if questState.Status != Failed {
			logz.Panicln("QuestManager", "GetQuestStatus: quest is in failed bucket, but status is wrong:", questState.Status)
		}
		return Failed
	}
	return NotStarted
}

func (qm *QuestManager) StartQuest(id defs.QuestID) {
	if id == "" {
		panic("id was empty")
	}
	questStatus := qm.GetQuestStatus(id)
	if questStatus != NotStarted {
		logz.Panicln("QuestManager", "tried to start a quest that has already started before:", id, questStatus)
	}

	questDef := qm.GetQuestDef(id)

	// instantiate quest state
	questState := state.QuestState{
		DefID:  questDef.ID,
		Status: Active,
	}

	qm.active[id] = &questState

	// recreate indices since we need this quests' events included
	qm.CreateEventTypeIndices()

	qm.eventBus.Publish(defs.Event{
		Type: pubsub.EventQuestStarted,
		Data: map[string]any{
			QuestIDKey: id,
		},
	})

	// start the first stage
	qm.SetQuestStage(id, questDef.StartStage)
}

func (qm *QuestManager) CompleteQuest(id defs.QuestID) {
	if id == "" {
		panic("id was empty")
	}

	// make sure this quest is active
	switch qm.GetQuestStatus(id) {
	case NotStarted:
		logz.Panicln("QuestManager", "tried to complete quest, but it wasn't started yet")
	case Completed:
		logz.Panicln("QuestManager", "tried to complete quest, but it was already completed")
	case Failed:
		logz.Panicln("QuestManager", "tried to complete quest, but it was already failed")
	}

	// move quest from active to completed
	questState := *qm.active[id]
	questState.Status = Completed
	delete(qm.active, id)
	qm.completed[id] = &questState
}

func (qm *QuestManager) FailQuest(id defs.QuestID) {
	if id == "" {
		panic("id was empty")
	}

	// make sure this quest is active
	switch qm.GetQuestStatus(id) {
	case NotStarted:
		logz.Panicln("QuestManager", "tried to fail quest, but it wasn't started yet")
	case Completed:
		logz.Panicln("QuestManager", "tried to fail quest, but it was already completed")
	case Failed:
		logz.Panicln("QuestManager", "tried to fail quest, but it was already failed")
	}

	// move quest from active to failed
	questState := *qm.active[id]
	questState.Status = Failed
	delete(qm.active, id)
	qm.failed[id] = &questState
}

func (qm *QuestManager) conditionsMet(conditions []defs.QuestConditionDef, questID defs.QuestID, event defs.Event) bool {
	// questState := qm.GetActiveQuestState(questID)
	for _, cond := range conditions {
		switch cond.Type {
		// TODO: need to figure out what type of conditions we have, and how they are checked
		}
	}

	return true
}
