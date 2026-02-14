package defs

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
)

type QuestDef struct {
	ID         QuestID
	Name       string
	Stages     []QuestStageDef
	StartStage QuestStageID
}

type QuestStageDef struct {
	ID QuestStageID
	// OnEnter []QuestActionDef
	// Reactions []QuestReactionDef
}
