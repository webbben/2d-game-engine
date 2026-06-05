package dialogv2

import (
	"fmt"
	"slices"
	"strings"

	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/item"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/quest"
)

const (
	TopicSeenKey    string = "TOPIC_SEEN"
	ResponseSeenKey string = "RESPONSE_SEEN"
)

type DialogContext struct {
	Exit       bool // if set, the DialogSession will see it and apply it up there too, causing the dialog to end.
	NPCID      string
	Profile    *state.DialogProfileState
	ProfileDef defs.DialogProfileDef
	GameState  defs.GameDialogContext
	eventBus   *pubsub.EventBus
	dataman    *datamanager.DataManager
	questman   *quest.QuestManager

	seenTopics      map[defs.TopicID]bool // topics the player has already discussed before (with this NPC)
	knowledgeTopics map[defs.TopicID]bool // topics which both the player and NPC have knowledge of
	seenResponses   map[string]bool
}

func NewDialogContext(npcID string, profile *state.DialogProfileState, profDef defs.DialogProfileDef, gameState defs.GameDialogContext, eventBus *pubsub.EventBus, dataman *datamanager.DataManager, questman *quest.QuestManager) DialogContext {
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
		NPCID:      npcID,
		Profile:    profile,
		ProfileDef: profDef,
		GameState:  gameState,
		eventBus:   eventBus,
		dataman:    dataman,
		questman:   questman,
	}
	ds.seenTopics = make(map[defs.TopicID]bool)
	ds.knowledgeTopics = make(map[defs.TopicID]bool)
	ds.seenResponses = make(map[string]bool)

	// parse out data from memory map
	for k := range ds.Profile.Memory {
		parts := strings.Split(k, ":")
		// if a memory key doesn't have parts, it's probably just a misc memory key used for random dialog purposes.
		if len(parts) == 2 {
			// found a key:value combo
			key := parts[0]
			value := parts[1]
			switch key {
			case TopicSeenKey:
				ds.seenTopics[defs.TopicID(value)] = true
			case ResponseSeenKey:
				ds.seenResponses[value] = true
			}
		}
	}

	// find intersection of knowledge topics that both the player and NPC know about
	playerCharState := dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))
	for _, topicID := range profDef.KnowledgeTopics {
		if playerCharState.Knowledge[topicID] {
			ds.knowledgeTopics[topicID] = true
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
	characterstate.AddKnowledge(topicID, ctx.dataman, ctx.eventBus)

	// if the NPC also knows this topic, add it to this session's available knowledge topics
	if slices.Contains(ctx.ProfileDef.KnowledgeTopics, topicID) {
		ctx.knowledgeTopics[topicID] = true
	} else {
		// seems weird to unlock knowledge in a conversation where the other person doesn't have that knowledge...
		logz.Warnln("RecordTopicUnlocked", "unlocked topic isn't in NPC's (dialog profile) knowledge.", topicID)
	}
}

func (ctx *DialogContext) RecordResponseSeen(greetingID string) {
	if greetingID == "" {
		panic("greetingID was empty")
	}
	key := fmt.Sprintf("%s:%s", ResponseSeenKey, greetingID)
	ctx.Profile.Memory[key] = true

	ctx.seenResponses[greetingID] = true
}

func (ctx *DialogContext) GetKnowledgeTopics() []defs.TopicID {
	ids := []defs.TopicID{}
	for topicID := range ctx.knowledgeTopics {
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

func (ctx DialogContext) HasSeenResponse(id string) bool {
	_, exists := ctx.seenResponses[id]
	return exists
}

func (ctx *DialogContext) GetNPCID() string {
	return ctx.NPCID
}

func (ctx DialogContext) GetCharacterDef(id defs.CharacterDefID) defs.CharacterDef {
	return ctx.dataman.GetCharacterDef(id)
}

func (ctx DialogContext) GetPlayerGold() int {
	playerState := ctx.dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))
	return item.CountMoney(playerState.StandardInventory, ctx.dataman)
}

func (ctx DialogContext) GetNPCDialogProfileID() defs.DialogProfileID {
	return ctx.Profile.ProfileID
}

func (ctx DialogContext) GetActiveMapDef() defs.MapDef {
	return ctx.GameState.GetActiveMapDef()
}

func (ctx DialogContext) RecordMiscDialogMemory(key string) {
	ctx.Profile.Memory[key] = true
}

func (ctx DialogContext) HasMemory(key string) bool {
	return ctx.Profile.Memory[key]
}

func (ctx DialogContext) GetCharacterSocialRank(id id.CharacterStateID) defs.SocialRank {
	charState := ctx.dataman.GetCharacterState(id)
	return charState.SocialRank
}

func (ctx DialogContext) GetNPCCharStateID() id.CharacterStateID {
	return id.CharacterStateID(ctx.NPCID)
}

func (ctx DialogContext) CharacterHasRole(id id.CharacterStateID, roleID defs.RoleID) bool {
	charState := ctx.dataman.GetCharacterState(id)
	return charState.Roles[roleID]
}

func (ctx DialogContext) GetQuestStage(qid defs.QuestID) (defs.QuestStageDef, defs.QuestStatus) {
	return ctx.questman.GetQuestStage(qid)
}

func (ctx DialogContext) IsItemEquipped(itemID defs.ItemID) bool {
	playerCharState := ctx.dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))
	return characterstate.IsItemEquipped(itemID, *playerCharState)
}

func (ctx DialogContext) PlayerHasKnowledge(topicID defs.TopicID) bool {
	playerCharState := ctx.dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))
	return playerCharState.Knowledge[topicID]
}

// WorldEffectContext functions

func (ctx *DialogContext) BroadcastEvent(e defs.Event) {
	ctx.GameState.BroadcastEvent(e)
}

func (ctx *DialogContext) AddGold(amount int) {
	ctx.GameState.AddGold(amount)
}

func (ctx *DialogContext) RemoveGold(amount int) {
	ctx.GameState.RemoveGold(amount)
}

func (ctx DialogContext) GetCurrentGameTime() clock.GameTime {
	return ctx.GameState.GetCurrentGameTime()
}

func (ctx DialogContext) AddItem(itemID defs.ItemID, quantity int) {
	ctx.GameState.AddItem(itemID, quantity)
}

func (ctx DialogContext) AddRole(roleID defs.RoleID) {
	ctx.GameState.AddRole(roleID)
}

func (ctx DialogContext) RemoveRole(roleID defs.RoleID) {
	ctx.GameState.RemoveRole(roleID)
}

func (ctx DialogContext) AssignTaskToNPC(id defs.CharacterDefID, taskDef defs.TaskDef, requireListener bool) {
	ctx.GameState.AssignTaskToNPC(id, taskDef, requireListener)
}

func (ctx DialogContext) QueueScenario(scnID defs.ScenarioID) {
	ctx.GameState.QueueScenario(scnID)
}

func (ctx DialogContext) UnlockMapLock(mapID defs.MapID, lockID string) {
	ctx.GameState.UnlockMapLock(mapID, lockID)
}

func (ctx DialogContext) TravelToMap(mapID defs.MapID, spawnIndex int, hours int) {
	ctx.GameState.TravelToMap(mapID, spawnIndex, hours)
}

func (ctx DialogContext) PlayerHasItem(itemID defs.ItemID) bool {
	playerState := ctx.dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))

	for _, itemState := range playerState.InventoryItems {
		if itemState != nil {
			if itemState.DefID == itemID {
				return true
			}
		}
	}

	return false
}

func (ctx DialogContext) GetPlayerSkillLevel(skillID defs.SkillID) int {
	playerState := ctx.dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))

	// TODO: need to factor in other things like traits, enchanted items (future), etc
	return playerState.BaseSkills[skillID]
}

func (ctx DialogContext) GetPlayerAttributeLevel(attrID defs.AttributeID) int {
	playerState := ctx.dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))

	// TODO: need to factor in other things like traits, enchanted items (future), etc
	return playerState.BaseAttributes[attrID]
}

func (ctx DialogContext) SetMapLock(mapID defs.MapID, lockID string, lockLevel int) {
	ctx.GameState.SetMapLock(mapID, lockID, lockLevel)
}
