package dialogv2

import (
	"fmt"
	"strings"

	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/pubsub"
)

const (
	TopicSeenKey     string = "TOPIC_SEEN"
	TopicUnlockedKey string = "TOPIC_UNLOCKED"
	ResponseSeenKey  string = "RESPONSE_SEEN"
)

type DialogContext struct {
	NPCID     string
	Profile   *state.DialogProfileState
	GameState defs.GameDialogContext
	eventBus  *pubsub.EventBus
	dataman   *datamanager.DataManager

	seenTopics     map[defs.TopicID]bool
	unlockedTopics map[defs.TopicID]bool
	seenResponses  map[string]bool
}

func NewDialogContext(npcID string, profile *state.DialogProfileState, gameState defs.GameDialogContext, eventBus *pubsub.EventBus, dataman *datamanager.DataManager) DialogContext {
	// just confirming dialog context implements the necessary interfaces
	_ = append([]defs.ConditionContext{}, &DialogContext{})
	if profile == nil {
		panic("profile was nil")
	}
	if gameState == nil {
		panic("gameState was nil")
	}
	if npcID == "" {
		panic("npc ID was empty")
	}
	if dataman == nil {
		panic("dataman was nil")
	}
	ds := DialogContext{
		NPCID:     npcID,
		Profile:   profile,
		GameState: gameState,
		eventBus:  eventBus,
		dataman:   dataman,
	}
	ds.seenTopics = make(map[defs.TopicID]bool)
	ds.unlockedTopics = make(map[defs.TopicID]bool)
	ds.seenResponses = make(map[string]bool)

	// parse out data from memory map
	for k := range ds.Profile.Memory {
		parts := strings.Split(k, ":")
		if len(parts) == 2 {
			// found a key:value combo
			key := parts[0]
			value := parts[1]
			switch key {
			case TopicSeenKey:
				ds.seenTopics[defs.TopicID(value)] = true
			case TopicUnlockedKey:
				ds.unlockedTopics[defs.TopicID(value)] = true
			case ResponseSeenKey:
				ds.seenResponses[value] = true
			}
		} else {
			// so far, we only have the above pattern, so panic
			panic("unexpected key pattern found in memory")
		}
	}

	return ds
}

func (ctx *DialogContext) RecordTopicSeen(topicID defs.TopicID) {
	if topicID == "" {
		panic("topicID was empty")
	}
	key := fmt.Sprintf("%s:%s", TopicSeenKey, topicID)
	ctx.Profile.Memory[key] = true

	ctx.seenTopics[topicID] = true
}

func (ctx *DialogContext) RecordTopicUnlocked(topicID defs.TopicID) {
	if topicID == "" {
		panic("topicID was empty")
	}
	key := fmt.Sprintf("%s:%s", TopicUnlockedKey, topicID)
	ctx.Profile.Memory[key] = true

	ctx.unlockedTopics[topicID] = true
}

func (ctx *DialogContext) RecordResponseSeen(greetingID string) {
	if greetingID == "" {
		panic("greetingID was empty")
	}
	key := fmt.Sprintf("%s:%s", ResponseSeenKey, greetingID)
	ctx.Profile.Memory[key] = true

	ctx.seenResponses[greetingID] = true
}

func (ctx *DialogContext) GetUnlockedTopics() []defs.TopicID {
	ids := []defs.TopicID{}
	for topicID := range ctx.unlockedTopics {
		ids = append(ids, topicID)
	}
	return ids
}

func (ctx *DialogContext) GetSeenTopics() []defs.TopicID {
	ids := []defs.TopicID{}
	for topicID := range ctx.seenTopics {
		ids = append(ids, topicID)
	}
	return ids
}

func (ctx DialogContext) HasSeenTopic(id defs.TopicID) bool {
	return ctx.seenTopics[id]
}

func (ctx DialogContext) IsTopicUnlocked(id defs.TopicID) bool {
	return ctx.unlockedTopics[id]
}

func (ctx DialogContext) HasSeenResponse(id string) bool {
	_, exists := ctx.seenResponses[id]
	return exists
}

func (ctx *DialogContext) GetNPCID() string {
	return ctx.NPCID
}

func (ctx *DialogContext) BroadcastEvent(e defs.Event) {
	ctx.eventBus.Publish(e)
}

func (ctx *DialogContext) AddGold(amount int) {
	ctx.GameState.DialogCtxAddGold(amount)
}

func (ctx DialogContext) GetCharacterDef(id defs.CharacterDefID) defs.CharacterDef {
	return ctx.dataman.GetCharacterDef(id)
}

func (ctx DialogContext) RecordMiscDialogMemory(key string) {
	ctx.Profile.Memory[key] = true
}
