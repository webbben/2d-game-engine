// Package defs definitively defines definite definitions
package defs

import (
	"github.com/webbben/2d-game-engine/data/id"
)

type (
	TopicID                 string
	DialogProfileID         string
	DialogActionType        string
	DialogActionResultScope string
)

// DialogProfileDef represents a body of greetings, dialog topics, etc that can be assigned to an NPC.
// A dialog profile can be shared between entities or be entity-specific.
type DialogProfileDef struct {
	ProfileID       DialogProfileID
	Greeting        []DialogResponse
	TopicsIDs       []TopicID // default topics that the NPC can immediately discuss; doesn't require player knowledge
	KnowledgeTopics []TopicID // topics that require player knowledge; if the player doesn't know about this topic yet, it won't be available

	// Speech bubbles aren't part of actual dialog, but defined here since it's related to speech

	SpeechBubbles []SpeechBubbleDef
}

type SpeechBubbleReaction interface {
	Reaction(e Event, ctx SpeechBubbleContext) string
}

type SpeechBubbleDef struct {
	SubscribeEvents []EventType // the events we subscribe to
	Reaction        SpeechBubbleReaction
}

type SpeechBubbleContext interface {
	WorldInfoContext
}

// DialogTopic defines a specific dialog topic, and all the various ways that NPCs may discuss this topic.
// A semantic question or prompt the player can raise (e.g. "Rumors", "Background", "Little Advice")
type DialogTopic struct {
	ID         TopicID
	Prompt     string            // Text shown to the player
	Conditions []DialogCondition // Whether topic appears. AND logic.
	Responses  []DialogResponse  // NPC replies
}

func (dt DialogTopic) Validate() {
	if dt.ID == "" {
		panic("id was empty")
	}
	if dt.Prompt == "" {
		panic("prompt was empty")
	}
	if len(dt.Responses) == 0 {
		panic("responses was empty")
	}
	for _, resp := range dt.Responses {
		resp.Validate()
	}
}

// DialogResponse defines a specific response to a topic prompt, according to an NPC's role or identity, a quest state, etc.
// It represents a single possible response that can be given to a topic/prompt, and defines the conditions for that response.
type DialogResponse struct {
	ID string // only needed if this dialog response is used in any sort of memory tracking; if you want to use the Only flag, this is required, otherwise it's okay to omit.
	// Notes:
	// - DONT centralize in definitions manager as independent DialogResponses. These are contextual and often one-off. not good for global re-use.

	Text         string
	Conditions   []DialogCondition // Whether this response is given to a topic. AND logic.
	Effects      []DialogEffect    // effects that only manipulate dialog state
	WorldEffects []WorldEffect     // effects that manipulate the game world outside of dialog
	NextTopics   []TopicID         // IDs to unlock/emphasize
	Once         bool              // if set, this response will only be possible to show once. after that, it is no longer eligible.
	Goodbye      bool              // if set, this response will only have a 'Goodbye' reply available, which ends the conversation.
	Exit         bool              // if set,  the dialog will exit immediately, without any 'Goodbye' reply prompt or anything

	Replies []DialogReply // if the response should pose possibly replies by the player, set them here.

	Action *DialogAction // if set, will fire first before any text or effects; can cause UI flows to happen, such as getting user text input.

	// if you want to weave another dialog response directly after this one, put it here.
	// ex: you want the first section of a speech to introduce some things, then you want the second section (the NextResponse) to actually
	// trigger effects, quest updates, etc. Allows for enhanced pacing and dramatic effect.
	NextResponse *DialogResponse
	// similar to NextResponse, but allows for different options based on conditions. You can use either NextResponse or this, but not both at the same time.
	NextResponseOptions []DialogResponse
}

func (dr DialogResponse) Validate() {
	if dr.Once && dr.ID == "" {
		panic("responses marked as Once must have an ID")
	}
	if dr.Text == "" {
		// if text is empty, then this must be a grouper
		if len(dr.Conditions) == 0 {
			panic("is this supposed to be a grouper? no text is set, but no conditions are set either.")
		}
		if dr.NextResponse == nil && len(dr.NextResponseOptions) == 0 {
			panic("is this supposed to be a grouper? no text is set, but no next responses are set either.")
		}
	}
	if dr.Goodbye {
		if dr.NextResponse != nil {
			panic("goodbye response has next response linked")
		}
		if len(dr.NextResponseOptions) > 0 {
			panic("goodbye response has next response options set")
		}
		if len(dr.Replies) > 0 {
			panic("goodbye response has replies defined. goodbye responses will automatically define a single 'goodbye' reply, so don't create one yourself.")
		}
	}
	if dr.NextResponse != nil {
		dr.NextResponse.Validate()
		if len(dr.NextResponseOptions) > 0 {
			panic("has next response, but also has next response options")
		}
	} else {
		for _, nr := range dr.NextResponseOptions {
			nr.Validate()
		}
	}
}

// DialogReply is a reply the player can give to a certain dialog response.
type DialogReply struct {
	Text         string // The text for the reply option (which represents how the player is responding)
	Conditions   []DialogCondition
	Effects      []DialogEffect // effects that only manipulate dialog state
	WorldEffects []WorldEffect  // effects that manipulate the game world outside of dialog
	Goodbye      bool           // if set, this reply will cause dialog to end

	NextResponse *DialogResponse // once this reply is chosen, this is how the NPC reacts. If nil, no response is given by the NPC.
}

type DialogCondition interface {
	IsMet(ctx ConditionContext) bool
}

type ConditionContext interface {
	GetQuestStage(qid QuestID) (QuestStageDef, QuestStatus)
	GetNPCCharStateID() id.CharacterStateID
	GetActiveMapDef() MapDef
	HasSeenTopic(id TopicID) bool
	HasMemory(key string) bool
	GetCharacterDef(id CharacterDefID) CharacterDef
	GetCharacterSocialRank(id id.CharacterStateID) SocialRank
	CharacterHasRole(id id.CharacterStateID, roleID RoleID) bool
	GetPlayerGold() int
	GetNPCDialogProfileID() DialogProfileID
	IsItemEquipped(itemID ItemID) bool
	PlayerHasKnowledge(topicID TopicID) bool
}

type MemoryCondition struct {
	Key  string
	Seen bool
}

// DialogEffect is for effects that only manipulate dialog state, such as memory.
type DialogEffect interface {
	Apply(ctx DialogEffectContext)
}

// DialogEffectContext is for effects that only impact in-dialog behavior (like recording dialog memory).
// For effects that influence other aspects of the game world, use WorldEffectContext.
type DialogEffectContext interface {
	RecordTopicSeen(id TopicID)
	RecordTopicUnlocked(id TopicID)
	RecordMiscDialogMemory(key string)
}

type SetMemoryEffect struct {
	Key  string
	Seen bool
}

// A DialogAction is something that interrupts the flow of a dialog to bring different UI elements, workflows, etc. Ex: getting user input in a modal.
type DialogAction struct {
	Type   DialogActionType
	Scope  DialogActionResultScope // what the action's result should be applied to.
	Params any                     // Params that are passed to the action's UI, modal, etc. You should be using an actual params struct defined for this Action type's UI.
}
