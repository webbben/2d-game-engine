// Package defs definitively defines definite definitions
package defs

type (
	TopicID         string
	DialogProfileID string
)

// DialogProfileDef represents a body of greetings, dialog topics, etc that can be assigned to an NPC.
// A dialog profile can be shared between entities or be entity-specific.
type DialogProfileDef struct {
	ProfileID DialogProfileID
	Greeting  []DialogResponse
	TopicsIDs []TopicID
}

// DialogTopic defines a specific dialog topic, and all the various ways that NPCs may discuss this topic.
// A semantic question or prompt the player can raise (e.g. "Rumors", "Background", "Little Advice")
type DialogTopic struct {
	ID         TopicID
	Prompt     string            // Text shown to the player
	Conditions []Condition       // Whether topic appears. AND logic.
	Responses  []DialogResponse  // NPC replies
	Metadata   map[string]string // Optional (UI hints, tags, etc)
}

// DialogResponse defines a specific response to a topic prompt, according to an NPC's role or identity, a quest state, etc.
// It represents a single possible response that can be given to a topic/prompt, and defines the conditions for that response.
type DialogResponse struct {
	// Notes:
	// - DONT centralize in definitions manager as independent DialogResponses. These are contextual and often one-off. not good for global re-use.

	Text       string
	Conditions []Condition // Whether this response is given to a topic. AND logic.
	Effects    []Effect
	NextTopics []TopicID // IDs to unlock/emphasize
	Once       bool      // if set, this response will only be possible to show once. after that, it is no longer eligible.

	Replies []DialogReply // if the response should pose possibly replies by the player, set them here.

	// if you want to weave another dialog response directly after this one, put it here.
	// ex: you want the first section of a speech to introduce some things, then you want the second section (the NextResponse) to actually
	// trigger effects, quest updates, etc. Allows for enhanced pacing and dramatic effect.
	NextResponse *DialogResponse
}

// DialogReply is a reply the player can give to a certain dialog response.
type DialogReply struct {
	Text       string // The text for the reply option (which represents how the player is responding)
	Conditions []Condition
	Effects    []Effect

	NextResponse *DialogResponse // once this reply is chosen, this is how the NPC reacts. If nil, no response is given by the NPC.
	NextTopicID  TopicID
}

type Condition interface {
	IsMet(ctx ConditionContext) bool
}

type ConditionContext interface {
	HasSeenTopic(id TopicID) bool
	IsTopicUnlocked(id TopicID) bool
}

type MemoryCondition struct {
	Key  string
	Seen bool
}

type Effect interface {
	Apply(ctx EffectContext)
}

type EffectContext interface {
	MarkTopicSeen(id TopicID)
	UnlockTopic(id TopicID)

	AddGold(amount int)
}

type SetMemoryEffect struct {
	Key  string
	Seen bool
}
