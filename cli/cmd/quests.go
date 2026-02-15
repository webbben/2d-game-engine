package cmd

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/quest"
)

const (
	Q001GameStart                 defs.QuestID   = "Q001: Game Start"
	Q001PlayerWakesUp             defs.EventType = "Q001: Player wakes up in prison ship"
	Q001GuardApproaching          defs.EventType = "Q001: Guard is approaching"
	Q001CharacterCreationComplete defs.EventType = "Q001: character creation complete"
)

func GetAllQuestDefs() []defs.QuestDef {
	questDefs := []defs.QuestDef{
		{
			ID:   Q001GameStart,
			Name: "Awakening",
			StartTrigger: defs.QuestStartTrigger{
				EventType: Q001PlayerWakesUp,
			},
			Stages: map[defs.QuestStageID]defs.QuestStageDef{
				"Awakening": {
					ID:    "Awakening",
					Title: "Awakening",
					Objective: `
					Find out where you are and what's happened.
					`,
					Description: `
					I've awoken in a very unexpected place - what appears to be the bottom of some kind of ship.
					What's going on? How did I end up here? Strangely, I can hardly remember anything of my past.
					It almost feels like I've awoken from some kind of coma.
					`,
					Reactions: []defs.QuestReactionDef{
						{
							SubscribeEvent: Q001GuardApproaching,
							NextStage:      "Census & Excise",
						},
					},
				},
				"Census & Excise": {
					ID:    "Census & Excise",
					Title: "A Prisoner?",
					Objective: `
					Follow the guard out of the ship.
					`,
					Description: `
					I've spoken to a man named Jovis who claims we are prisoners, and likely to be sold off into slavery.
					I'm completely dumbfounded by all of this, and I almost thought he had gone mad.
					`,
					Reactions: []defs.QuestReactionDef{
						{
							SubscribeEvent: Q001CharacterCreationComplete,
							TerminalStatus: quest.TerminalStatusComplete,
							// TODO: trigger next quest to start: probably a timelapse occurs and the character is in a legionary camp, having been conscripted
						},
					},
				},
			},
		},
	}

	return questDefs
}
