package state

import "github.com/webbben/2d-game-engine/data/defs"

type QuestStatus string

// QuestState defines the current state of a quest.
type QuestState struct {
	DefID        defs.QuestID
	CurrentStage defs.QuestStageID
	Status       QuestStatus
}
