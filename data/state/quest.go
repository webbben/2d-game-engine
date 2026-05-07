package state

import "github.com/webbben/2d-game-engine/data/defs"

// QuestState defines the current state of a quest.
type QuestState struct {
	DefID        defs.QuestID
	CurrentStage defs.QuestStageID
	Status       defs.QuestStatus
}
