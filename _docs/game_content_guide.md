# Game Content Authoring Guide

This guide helps AI agents write game content using this engine. It covers how to define items, dialog, and quests that work together to create game experiences.

---

## Quick Reference

| System | Package | Purpose | Defines |
|--------|---------|---------|---------|
| Dialog | `dialogv2/` | NPC conversations | Profiles, Topics, Responses |
| Quests | `quest/` | Narrative progression | Quests, Stages, Reactions |
| Items | `item/` | Equipment and objects | Weapons, Armor, Consumables |

---

## Common Patterns

### Pattern: Quest from Dialog

The most common pattern: dialog starts a quest.

```go
// Define a dialog topic that starts a quest
DialogTopic{
    ID:     "quest_hook",
    Prompt: "I need help finding my lost cat.",
    Responses: []DialogResponse{
        {
            Text: "My cat Mittens ran off into the forest. Can you find her?",
            NextTopics: []TopicID{"quest_details"},  // Unlock more questions
            Effects: []DialogEffect{
                // This event triggers the quest
                EventEffect{
                    Event: defs.Event{
                        Type: "quest_activate_find_cat",
                    },
                },
            },
            Replies: []DialogReply{
                {
                    Text: "I'll find her for you.",
                    Effects: []DialogEffect{
                        SetDialogMemoryEffect{MemoryKey: "accepted_cat_quest"},
                    },
                },
                {Text: "Sorry, I'm busy.", Goodbye: true},
            },
        },
    },
}
```

### Pattern: Quest Progress via Dialog

```go
// Quest stage reaction listens for dialog event
QuestReactionDef{
    SubscribeEvent: "dialog_topic_selected",
    Conditions: []QuestConditionDef{
        {Type: "event_data", Params: map[string]string{
            "key":   "topic",
            "value": "cat_found",
        }},
    },
    NextStage: "stage_return_cat",
}
```

### Pattern: Reward via Dialog

```go
// Final quest stage - dialog gives rewards
DialogTopic{
    ID:     "quest_complete",
    Prompt: "Quest complete!",
    Responses: []DialogResponse{
        {
            Text: "Thank you so much! Here's your reward.",
            Effects: []DialogEffect{
                AddGoldEffect{Amount: 100},
                EventEffect{
                    Event: defs.Event{Type: "quest_completed"},
                },
            },
        },
    },
}
```

### Pattern: Item as Quest Objective

```go
// Quest stage checks for item
QuestReactionDef{
    SubscribeEvent: "item_collected",
    Conditions: []QuestConditionDef{
        {Type: "event_data", Params: map[string]string{
            "key":   "itemID",
            "value": "magical_key",
        }},
    },
    NextStage: "stage_door_unlocked",
}
```

---

## Step-by-Step: Creating a Side Quest

Here's a complete example of creating a fetch quest:

### 1. Define the Item (if needed)

```go
// item/definitions.go
func NewMagicHerb() *item.ItemBase {
    return item.NewItemBase(item.ItemBaseParams{
        ID:                "magic_herb",
        Name:              "Magic Herb",
        Description:       "A rare herb that glows faintly in the dark.",
        Type:              item.TypeConsumable,
        Value:             50,
        Weight:            0.1,
        Groupable:         true,
        TileImgTilesetSrc: "items/herbs.png",
        TileImgIndex:      7,
    })
}
```

### 2. Define the Dialog Profile

```go
// dialog/profiles.go
DialogProfileDef{
    ProfileID: "herbalist_dialog",
    Greeting: []DialogResponse{
        {
            Text: "Welcome, traveler. Looking for herbs?",
        },
    },
    TopicsIDs: []TopicID{
        "herbalist_quest",
        "herbalist_shop",
    },
}
```

### 3. Define Dialog Topics

```go
// dialog/topics.go
DialogTopic{
    ID:     "herbalist_quest",
    Prompt: "Do you need any help?",
    Responses: []DialogResponse{
        {
            Conditions: []DialogCondition{
                // Only offer quest if not already on it
                ConditionNOT{
                    Arg: ConditionDialogMemory{Key: "herb_quest_accepted"},
                },
            },
            Text: "Actually, yes! I need a Moonpetal flower from the Darkwood Grove. The forest is dangerous though...",
            Effects: []DialogEffect{
                // Start the quest
                EventEffect{
                    Event: defs.Event{
                        Type: "quest_activate_herb_gathering",
                    },
                },
                SetDialogMemoryEffect{MemoryKey: "herb_quest_accepted"},
            },
            Replies: []DialogReply{
                {
                    Text: "I'll get it for you.",
                    NextResponse: &DialogResponse{
                        Text: "Thank you! Be careful out there.",
                        Effects: []DialogEffect{
                            // Give player something to help
                            AddItemEffect{ItemID: "torch"},
                        },
                    },
                },
                {Text: "Sounds too dangerous.", Goodbye: true},
            },
        },
        {
            // Already on quest
            Conditions: []DialogCondition{
                ConditionDialogMemory{Key: "herb_quest_accepted"},
            },
            Text: "Still looking for the Moonpetal? Remember, it's in the Darkwood Grove!",
        },
    },
}
```

### 4. Define the Quest

```go
// quest/definitions.go
quest.QuestDef{
    ID:          "herb_gathering",
    Name:        "The Herbalist's Request",
    Description: "Find a Moonpetal flower in Darkwood Grove for the village herbalist.",
    StartStage:  "find_herb",
    StartTrigger: defs.QuestStartTrigger{
        EventType: "quest_activate_herb_gathering",
    },
    Stages: map[defs.QuestStageID]defs.QuestStageDef{
        "find_herb": {
            ID:        "find_herb",
            Objective: "Find a Moonpetal flower in Darkwood Grove.",
            OnEnter: []defs.QuestAction{
                // Could spawn a marker on the map
                quest.QueueScenarioAction{ScenarioID: "darkwood_quest_area"},
            },
            Reactions: []defs.QuestReactionDef{
                {
                    SubscribeEvent: "item_collected",
                    Conditions: []QuestConditionDef{
                        {Type: "event_data", Params: map[string]string{
                            "key":   "itemID",
                            "value": "moonpetal_flower",
                        }},
                    },
                    NextStage: "return_herb",
                },
            },
        },
        "return_herb": {
            ID:        "return_herb",
            Objective: "Return the Moonpetal to the herbalist.",
            Reactions: []defs.QuestReactionDef{
                {
                    SubscribeEvent: "dialog_topic_selected",
                    Conditions: []QuestConditionDef{
                        {Type: "event_data", Params: map[string]string{
                            "key":   "topic",
                            "value": "herb_quest_complete",
                        }},
                    },
                    NextStage: "complete",
                },
            },
        },
        "complete": {
            ID:        "complete",
            Objective: "Quest complete!",
            Reactions: []defs.QuestReactionDef{
                {
                    SubscribeEvent: "dummy_event",
                    Actions: []defs.QuestAction{
                        // Unlock something, give bonus, etc.
                    },
                    TerminalStatus: defs.QuestTerminalSuccess,
                },
            },
        },
    },
}
```

---

## Condition Cheat Sheet

### Dialog Conditions

| Condition | Use Case |
|-----------|----------|
| `ConditionDialogMemory` | Check conversation history |
| `ConditionCulture` | NPC/player culture check |
| `ConditionHasGold` | Player gold check |
| `ConditionMapID` | Location check |
| `ConditionRand` | Random chance |
| `ConditionSocialRank` | Status check |
| `ConditionHasRole` | Role/permission check |
| `ConditionQuestStage` | Check player's current quest stage |
| `ConditionNOT` | Negate condition |
| `ConditionOR` | OR multiple conditions |

### Quest Conditions

Quest conditions are typically event-data checks. Use `EventEffect` with data:

```go
EventEffect{
    Event: defs.Event{
        Type: "item_collected",
        Data: map[string]any{
            "itemID":   "magic_key",
            "quantity": 1,
        },
    },
}
```

---

## Effect Cheat Sheet

| Effect | Use Case |
|--------|----------|
| `EventEffect` | Trigger quests, emit custom events |
| `AddGoldEffect` | Give money |
| `RemoveGoldEffect` | Take money |
| `SetDialogMemoryEffect` | Store conversation state |
| `ScheduleFutureEventEffect` | Delayed events |
| `StartLoadScreenEffect` | Trigger transitions |

---

## Validation Checklist

Before publishing content, verify:

- [ ] Item IDs are unique
- [ ] Dialog topic IDs are unique  
- [ ] Quest IDs are unique
- [ ] All referenced items exist
- [ ] All referenced dialog profiles exist
- [ ] Quest stages have reactions (or will be stuck)
- [ ] Terminal quests have `TerminalStatus` not `NextStage`
- [ ] `Once` responses have IDs
- [ ] Items with durability are not groupable
- [ ] Visible equipment has `BodyPartDef`

---

## Event Bus

All systems communicate via the event bus (`pubsub/`). Subscribe to events:

```go
eventBus.Subscribe(eventType, func(event defs.Event) {
    // Handle event
})
```

Publish events:

```go
eventBus.Publish(defs.Event{
    Type: "my_custom_event",
    Data: map[string]any{
        "key": "value",
    },
})
```

Common built-in events:
- `EventDialogStarted` / `EventDialogEnded`
- `EventQuestStarted` / `EventQuestStageAdvanced`
- `EventVisitMap`
- `EventTimePass`
