# Quest System

The quest system manages narrative progression through event-driven, stage-based quests.

---

## Architecture Overview

```
QuestManager
├── QuestDef (definitions)
│   ├── QuestStageDef (stages)
│   │   └── QuestReactionDef (event responses)
│   └── QuestStartTrigger (start condition)
└── QuestState (runtime state)
```

**Key Principles:**
- Event-driven (no polling or update-loop checks)
- Stage-based (finite progression states)
- Data-indexed (efficient event-to-reaction matching)
- Decoupled from dialog and other systems

---

## Core Types

### QuestDef

```go
type QuestDef struct {
    ID          QuestID
    Name        string                    // Player-visible name
    Description string                    // Initial description
    Stages      map[QuestStageID]QuestStageDef
    StartStage  QuestStageID
    StartTrigger QuestStartTrigger        // What starts this quest
}
```

### QuestStageDef

```go
type QuestStageDef struct {
    ID          QuestStageID
    Title       string                    // Optional stage title
    Objective   string                    // What player sees
    Description string                    // Narrative context
    OnEnter     []QuestAction            // Fires on stage entry
    Reactions   []QuestReactionDef        // Event-based responses
}
```

### QuestReactionDef

```go
type QuestReactionDef struct {
    SubscribeEvent EventType              // Event to listen for
    Conditions     []QuestConditionDef   // When to react
    Actions        []QuestAction         // What to do
    NextStage      QuestStageID           // Progress to this stage
    TerminalStatus QuestTerminalStatus   // Complete/Fail
    Text           string                 // Message displayed when reaction triggers (used for quest endings)
}
```

### QuestStartTrigger

```go
type QuestStartTrigger struct {
    EventType  EventType
    Conditions []QuestConditionDef
}
```

---

## Actions

Quest actions execute when a stage is entered or a reaction triggers.

### AssignTaskAction

Assigns a task to an NPC (character must be unique).

```go
quest.AssignTaskAction{
    CharDefID: "npc_guard_captain",
    TaskDef: defs.TaskDef{
        TaskID:   "goto",
        Priority: defs.PriorityAssigned,
        Params: map[string]any{
            "targetX": 100,
            "targetY": 200,
        },
    },
}
```

### QueueScenarioAction

Loads a scenario (isolated map setup).

```go
quest.QueueScenarioAction{
    ScenarioID: "ambush_scene",
}
```

### UnlockAction

Unlocks a map lock (door, gate).

```go
quest.UnlockAction{
    MapLock: defs.MapLock{
        MapID:  "dungeon_1",
        LockID: "cell_door",
    },
}
```

---

## Conditions

Conditions use AND logic. Define in quest definitions using `QuestConditionDef`:

```go
type QuestConditionDef struct {
    Type   QuestConditionType
    Params map[string]string
}
```

Common condition types (implemented in quest.go):

| Type | Description |
|------|-------------|
| `event_data` | Check event payload data |
| `quest_stage` | Check current quest stage |
| `time_range` | Check game time |

**For dialog integration**, use `EventEffect` to emit custom events from dialog.

---

## Terminal Status

Quest reactions can end a quest using `TerminalStatus`. When set, the quest will terminate with the specified outcome instead of transitioning to a new stage.

### Status Types

```go
const (
    TerminalStatusNone     QuestTerminalStatus = iota  // 0 - Quest continues (requires NextStage)
    TerminalStatusComplete                               // Quest succeeded
    TerminalStatusFail                                  // Quest failed
)
```

### Usage

Set `TerminalStatus` instead of `NextStage` to end the quest. Do not set both.

```go
QuestReactionDef{
    SubscribeEvent: "dragon_defeated",
    Actions: []defs.QuestAction{
        quest.UnlockAction{...},
    },
    TerminalStatus: quest.TerminalStatusComplete,  // Quest ends successfully
    // NextStage: ""  // Do not set this when using TerminalStatus
}
```

### Rules

- `TerminalStatusNone` + `NextStage` → Transition to next stage
- `TerminalStatusComplete` → End quest as success (ignore `NextStage`)
- `TerminalStatusFail` → End quest as failure (ignore `NextStage`)

### Complete Example

```go
"stage_return_to_elder": {
    Objective: "Return to the elder with the pelts.",
    Reactions: []defs.QuestReactionDef{
        {
            SubscribeEvent: "talked_to_elder",
            TerminalStatus: quest.TerminalStatusComplete,  // Success ending
            Actions: []defs.QuestAction{
                quest.UnlockAction{...},  // Reward: unlock treasure room
            },
        },
    },
},
"stage_caught_stealing": {
    Objective: "You were caught!",
    Reactions: []defs.QuestReactionDef{
        {
            SubscribeEvent: "caught_by_guard",
            TerminalStatus: quest.TerminalStatusFail,  // Failure ending
        },
    },
},
```

### Quest Ending Text

Use the `Text` field to display a conclusion message to the player when the quest ends. This appears when `TerminalStatus` is set.

```go
QuestReactionDef{
    SubscribeEvent: "talked_to_elder",
    TerminalStatus: quest.TerminalStatusComplete,
    Text: "The elder is pleased with your work. The village is safe once more.",
}
```

**Tips for good ending text:**
- Explain what happened or what was accomplished
- Provide narrative closure
- Can hint at future consequences or follow-up quests
- Keep it concise but satisfying

### Best Practices

1. **Always provide a dummy event** for terminal stages so the status actually fires
2. **Use success for positive outcomes** - quest objectives met, rewards earned
3. **Use failure for negative outcomes** - player failed, time expired, died
4. **Grant rewards in Actions** before `TerminalStatus` executes
5. **Include ending text** - give players narrative closure on terminal stages

---

## Creating a Quest

### Basic Structure

```go
import "github.com/webbben/2d-game-engine/quest"

questDef := quest.QuestDef{
    ID:          "deliver_package",
    Name:        "Package Delivery",
    Description: "Deliver a package to the blacksmith.",
    StartStage:  "intro",
    StartTrigger: defs.QuestStartTrigger{
        EventType: "quest_activate_deliver_package",
    },
    Stages: map[defs.QuestStageID]defs.QuestStageDef{
        "intro": {
            ID:        "intro",
            Objective: "Pick up the package from the merchant.",
            OnEnter: []defs.QuestAction{
                quest.AssignTaskAction{
                    CharDefID: "npc_merchant",
                    TaskDef: defs.TaskDef{
                        TaskID:   "goto",
                        Priority: defs.PriorityAssigned,
                        Params:   map[string]any{"targetX": 50, "targetY": 60},
                    },
                },
            },
            Reactions: []defs.QuestReactionDef{
                {
                    SubscribeEvent: "package_picked_up",
                    NextStage:      "delivering",
                },
            },
        },
        "delivering": {
            ID:        "delivering",
            Objective: "Deliver the package to the blacksmith.",
            Reactions: []defs.QuestReactionDef{
                {
                    SubscribeEvent: "package_delivered",
                    NextStage:      "complete",
                },
            },
        },
        "complete": {
            ID:        "complete",
            Objective: "Quest complete!",
            Reactions: []defs.QuestReactionDef{
                {
                    SubscribeEvent: "dummy_event", // Always fires
                    Actions: []defs.QuestAction{
                        quest.UnlockAction{
                            MapLock: defs.MapLock{
                                MapID:  "dungeon_1",
                                LockID: "cell_door",
                            },
                        },
                    },
                    TerminalStatus: defs.QuestTerminalSuccess,
                },
            },
        },
    },
}
```

### Linear Quest Example

A quest that progresses by talking to NPCs:

```go
// Stage 1: Initial conversation
"stage_talk_to_npc": {
    Objective: "Speak to the village elder.",
    Reactions: []defs.QuestReactionDef{
        {
            SubscribeEvent: "dialog_topic_selected",
            Conditions: []defs.QuestConditionDef{
                {Type: "event_data", Params: map[string]string{
                    "key":   "topic",
                    "value": "elder_quest",
                }},
            },
            NextStage: "stage_gather_items",
        },
    },
},

// Stage 2: Gather items
"stage_gather_items": {
    Objective: "Gather 5 wolf pelts.",
    Reactions: []defs.QuestReactionDef{
        {
            SubscribeEvent: "item_collected",
            Conditions: []defs.QuestConditionDef{
                {Type: "event_data", Params: map[string]string{
                    "key":   "itemID",
                    "value": "wolf_pelt",
                }},
            },
            Actions: []defs.QuestAction{
                // Update quest tracking
            },
            NextStage: "stage_return_to_elder",
        },
    },
},
```

### Branching Quest Example

```go
Stages: map[defs.QuestStageID]defs.QuestStageDef{
    "stage_choose_path": {
        Objective: "Choose your approach.",
        Reactions: []defs.QuestReactionDef{
            {
                SubscribeEvent: "player_chose_action",
                Conditions: []defs.QuestConditionDef{
                    {Type: "event_data", Params: map[string]string{
                        "key":   "choice",
                        "value": "diplomacy",
                    }},
                },
                NextStage: "stage_diplomatic",
            },
            {
                SubscribeEvent: "player_chose_action",
                Conditions: []defs.QuestConditionDef{
                    {Type: "event_data", Params: map[string]string{
                        "key":   "choice",
                        "value": "combat",
                    }},
                },
                NextStage: "stage_combat",
            },
        },
    },
}
```

### Multi-Stage Boss Quest

```go
Stages: map[defs.QuestStageID]defs.QuestStageDef{
    "stage_find_dungeon": {
        Objective: "Find the entrance to the dragon's lair.",
        OnEnter: []defs.QuestAction{
            quest.QueueScenarioAction{ScenarioID: "dungeon_entrance"},
        },
        Reactions: []defs.QuestReactionDef{
            {
                SubscribeEvent: "entered_dungeon",
                NextStage:      "stage_prepare",
            },
        },
    },
    "stage_prepare": {
        Objective: "Prepare for the battle.",
        Reactions: []defs.QuestReactionDef{
            {
                SubscribeEvent: "ready_for_battle",
                NextStage:      "stage_boss_fight",
            },
        },
    },
    "stage_boss_fight": {
        Objective: "Defeat the dragon!",
        OnEnter: []defs.QuestAction{
            quest.QueueScenarioAction{ScenarioID: "dragon_boss_encounter"},
        },
        Reactions: []defs.QuestReactionDef{
            {
                SubscribeEvent: "dragon_defeated",
                NextStage:      "stage_victory",
            },
            {
                SubscribeEvent: "player_died_in_battle",
                NextStage:      "stage_retreat",
            },
        },
    },
    "stage_victory": {
        Objective: "Return to claim your reward.",
        Reactions: []defs.QuestReactionDef{
            {
                SubscribeEvent: "dummy_event",
                Actions: []defs.QuestAction{
                    quest.UnlockAction{MapLock: defs.MapLock{MapID: "treasury", LockID: "dragon_hoard"}},
                },
                TerminalStatus: defs.QuestTerminalSuccess,
            },
        },
    },
}
```

---

## Triggering Quests

Quests are triggered by events. Use `EventEffect` in dialog:

```go
// In dialog response
Effects: []defs.DialogEffect{
    dialogv2.EventEffect{
        Event: defs.Event{
            Type: "quest_activate_deliver_package",
            Data: map[string]any{},
        },
    },
}
```

---

## Event Naming Conventions

Use descriptive, namespaced event types:

```
quest_<questID>                    // Quest start triggers
quest_<questID>_stage_<stageID>    // Stage transitions
dialog_<topicID>                   // Dialog topics
npc_<action>                       // NPC actions
item_<action>                      // Item actions
combat_<action>                    // Combat events
location_<mapID>                   // Location events
time_<period>                      // Time events
```

---

## Best Practices

1. **One event per reaction** - Keep reactions focused
2. **Use descriptive IDs** - `stage_talk_to_merchant` not `stage_1`
3. **Provide fallback stages** - Handle edge cases
4. **Test terminal states** - Ensure quests can complete/fail
5. **OnEnter for setup** - Use OnEnter for stage initialization
6. **Event-driven only** - Don't check conditions in update loops
7. **Clear objectives** - Player should always know what to do

---

## Quest Manager

The engine provides `QuestManager` for managing quest lifecycle:

```go
type QuestManager struct {
    questDefs map[QuestID]QuestDef
    notStarted map[QuestID]bool    // Waiting to be started
    active     map[QuestID]*state.QuestState
    completed  map[QuestID]*state.QuestState
    failed     map[QuestID]*state.QuestState
}
```

**Loading Quests:**
```go
questManager.LoadQuestDef(questDef)
```

---

## Quest State

Runtime state is stored separately from definitions:

```go
type QuestState struct {
    QuestID   QuestID
    StageID   QuestStageID
    Completed bool
    Failed    bool
    // Custom fields for tracking
}
```

---

## Integration Patterns

### Dialog → Quest

```go
// In DialogTopic
{
    Text: "I'll help you find the artifact.",
    Effects: []defs.DialogEffect{
        dialogv2.EventEffect{
            Event: defs.Event{Type: "quest_activate_find_artifact"},
        },
        dialogv2.SetDialogMemoryEffect{MemoryKey: "accepted_artifact_quest"},
    },
}
```

### Quest → NPC Task

```go
// On quest start
OnEnter: []defs.QuestAction{
    quest.AssignTaskAction{
        CharDefID: "npc_ally",
        TaskDef: defs.TaskDef{
            TaskID:   "follow",
            Priority: defs.PriorityAssigned,
            Params:   map[string]any{"target": "player"},
        },
    },
}
```

### Quest → Scenario

```go
// When reaching a stage
OnEnter: []defs.QuestAction{
    quest.QueueScenarioAction{
        ScenarioID: " ambush_at_bridge",
    },
}
```
