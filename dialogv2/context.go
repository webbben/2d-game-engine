package dialogv2

import (
	"fmt"
	"strings"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
)

type DialogContext struct {
	NPCID     string
	Profile   *state.DialogProfileState
	GameState GameStateContext

	seenTopics     map[defs.TopicID]bool
	unlockedTopics map[defs.TopicID]bool
}

type GameStateContext interface {
	GetPlayerInfo() PlayerInfo
}

// PlayerInfo is information about the player that dialogs might use
type PlayerInfo struct {
	PlayerName string
}

func NewDialogContext(npcID string, profile *state.DialogProfileState, gameState GameStateContext) DialogContext {
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
	ds := DialogContext{
		NPCID:     npcID,
		Profile:   profile,
		GameState: gameState,
	}
	ds.seenTopics = make(map[defs.TopicID]bool)
	ds.unlockedTopics = make(map[defs.TopicID]bool)

	// parse out data from memory map
	for k := range ds.Profile.Memory {
		parts := strings.Split(k, ":")
		if len(parts) == 2 {
			// found a key:value combo
			key := parts[0]
			value := parts[1]
			switch key {
			case "TOPIC_SEEN":
				ds.seenTopics[defs.TopicID(value)] = true
			case "TOPIC_UNLOCKED":
				ds.unlockedTopics[defs.TopicID(value)] = true
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
	key := fmt.Sprintf("TOPIC_SEEN:%s", topicID)
	ctx.Profile.Memory[key] = true

	ctx.seenTopics[topicID] = true
}

func (ctx *DialogContext) RecordTopicUnlocked(topicID defs.TopicID) {
	if topicID == "" {
		panic("topicID was empty")
	}
	key := fmt.Sprintf("TOPIC_UNLOCKED:%s", topicID)
	ctx.Profile.Memory[key] = true

	ctx.unlockedTopics[topicID] = true
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

func (ctx *DialogContext) GetNPCID() string {
	return ctx.NPCID
}
