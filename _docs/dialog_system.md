# Dialog System

The dialog system manages conversations between the player and NPCs. It supports branching dialogs, conditions, effects, topic unlocking, and memory tracking.

---

## Architecture Overview

```
DialogSession
├── DialogContext      // Tracks session state (memory, topics)
├── DialogProfileDef    // NPC's greeting and topics
├── DialogTopic         // Player-selectable conversation subjects
└── DialogResponse      // NPC's replies with conditions/effects
```

**Key Flow:**
1. Dialog starts → Select greeting (via conditions)
2. Show greeting → Await topic selection
3. Topic selected → Play response chain
4. Response may: show replies, unlock topics, trigger effects
5. Player chooses reply → Next response or back to topics

---

## Core Types

### Condition Evaluation Order

**Critical authoring rule:** When a list of `DialogResponse` (or any dialog concept with multiple options) is evaluated, the engine checks each item **in order**. The **first one that evaluates true is used**, and evaluation stops there.

A response with **no conditions is always true**. This means:

1. Responses with conditions should come **first**, ordered by priority/specificity
2. The **last response should always be a "default"** with no conditions
3. This ensures there's always a valid response to give

**Example - Correct ordering:**
```go
Greeting: []DialogResponse{
    // First: specific condition (higher priority)
    {
        Conditions: []DialogCondition{
            dialogv2.ConditionDialogMemory{Key: "is_friend"},
        },
        Text: "Good to see you again, old friend!",
    },
    // Second: another specific condition
    {
        Conditions: []DialogCondition{
            dialogv2.ConditionCulture{IsCulture: "roman"},
        },
        Text: "Salve, traveler!",
    },
    // Last: default fallback (no conditions = always true)
    {
        Text: "Hello there, traveler.",
    },
},
```

**Example - Incorrect (will cause issues):**
```go
Greeting: []DialogResponse{
    // Default first - will always match!
    {
        Text: "Hello there, traveler.",
    },
    // This will NEVER be reached
    {
        Conditions: []DialogCondition{
            dialogv2.ConditionDialogMemory{Key: "is_friend"},
        },
        Text: "Good to see you again, old friend!",
    },
},
```

### DialogProfileDef

Assigned to NPCs, defines greeting and available topics.

```go
type DialogProfileDef struct {
    ProfileID DialogProfileID
    Greeting  []DialogResponse  // Greetings (checked in order, first true wins)
    TopicsIDs []TopicID         // Topics available in this profile
}
```

#### Dialog Profiles: Unique vs Shared

Dialog profiles define **who** the NPC is in conversation. A profile can be:

**1. Unique (specific NPC)**
```go
ProfileID: "elder_ulric",  // Only one village elder Ulric
```
- Use for named characters with unique dialog
- Memory is tracked per-profile, so only this NPC remembers conversations
- Good for important story characters, or unique characters with their own personality

**2. Shared (generic NPCs)**
```go
ProfileID: "town_guard",  // Shared by all generic guards
```
- Use for generic NPCs that don't have individualized personalities 
- Memory is shared among ALL NPCs using this profile
- Good for background characters or generic styles of characters

**When to use each:**
| Scenario | Profile Type |
|----------|-------------|
| Named quest-giver, story character | Unique profile |
| Generic shopkeeper | Shared profile |
| Background NPCs (guards, villagers) | Shared profile |
| NPCs that should remember YOU specifically | Unique profile |

### DialogTopic

A semantic prompt the player can raise.

```go
type DialogTopic struct {
    ID         TopicID
    Prompt     string              // Text shown to player
    Conditions []DialogCondition   // Whether topic appears (AND logic)
    Responses  []DialogResponse    // NPC's possible replies
}
```

### DialogResponse

A specific NPC response.

```go
type DialogResponse struct {
    ID            string
    Text          string              // What NPC says
    Conditions    []DialogCondition   // When this response appears (AND logic)
    Effects       []DialogEffect      // What happens when shown
    NextTopics    []TopicID           // Topics to unlock/emphasize
    Once          bool                // Show only once
    Goodbye       bool                // Ends conversation after
    Replies       []DialogReply       // Player reply options
    Action        *DialogAction        // Special actions (input modal, show screen)
    NextResponse  *DialogResponse      // Chain to another response
    NextResponseOptions []DialogResponse // Conditional chains
}
```

### DialogReply

A reply the player can give.

```go
type DialogReply struct {
    Text          string
    Conditions    []DialogCondition
    Effects       []DialogEffect
    Goodbye       bool
    NextResponse  *DialogResponse  // NPC's next reply
    NextTopicID   TopicID          // Return to topic selection
}
```

---

## Conditions

Conditions determine when dialog elements appear. All conditions in a list use AND logic.

### ConditionDialogMemory

Check if a dialog memory key exists.

```go
DialogCondition: dialogv2.ConditionDialogMemory{
    Key: "told_player_secret",
}
```

### ConditionCulture

Check character culture.

```go
DialogCondition: dialogv2.ConditionCulture{
    CharDefID: "npc_blacksmith",
    IsCulture: "barbarian",
}
```

### ConditionHasGold

Check player's gold amount.

```go
DialogCondition: dialogv2.ConditionHasGold{
    Amount: 100,
}
```

### ConditionMapID

Check current map.

```go
DialogCondition: dialogv2.ConditionMapID{
    MapID: "main_town",
}
```

### ConditionRand

Random chance (0.0 to 1.0).

```go
DialogCondition: dialogv2.ConditionRand{
    Percent: 0.5, // 50% chance
}
```

### ConditionSocialRank

Check character social rank.

```go
DialogCondition: dialogv2.ConditionSocialRank{
    Player: true,           // Check player instead of NPC
    Rank:   "noble",
    GEQ:    true,           // Greater than or equal
}
```

### ConditionHasRole

Check character has a role.

```go
DialogCondition: dialogv2.ConditionHasRole{
    Player: false,          // Check NPC
    RoleID: "guard_captain",
}
```

### ConditionQuestStage

Check the current stage of a quest.

```go
DialogCondition: dialogv2.ConditionQuestStage{
    QuestID: "deliver_package",
    StageID: "delivering",  // Specific stage
}

// Or check quest status without specific stage:
DialogCondition: dialogv2.ConditionQuestStage{
    QuestID:    "deliver_package",
    Completed: true,  // Check if quest is completed
}

DialogCondition: dialogv2.ConditionQuestStage{
    QuestID:    "deliver_package",
    NotStarted: true,  // Check if quest hasn't started
}

DialogCondition: dialogv2.ConditionQuestStage{
    QuestID: "deliver_package",
    Failed:  true,  // Check if quest failed
}
```

### ConditionItemEquipped

Check if the player has a specific item equipped.

```go
DialogCondition: dialogv2.ConditionItemEquipped{
    ItemID: "iron_sword",
}
```

Useful for: requiring certain equipment for dialog options, recognizing hero status from equipment.

### ConditionKnowledge

Check if the player has learned about a topic (from previous conversations or other sources).

```go
DialogCondition: dialogv2.ConditionKnowledge{
    TopicID: "dark_ritual_secret",
}
```

Useful for: tracking what lore/ secrets the player has discovered, unlocking advanced dialog about learned topics.

#### What Roles Are For

Roles are flexible identifiers for character status. They are **not** for character class (barbarian, warlock, etc.)—that's handled separately. Roles are for:

**1. Faction membership**
```go
RoleID: "thieves_guild_member"
RoleID: "fighters_guild_rank_journeyman"
RoleID: "royal_court_favorite"
```

**2. Faction rank/hierarchy**
```go
RoleID: "fighters_guild_rank_apprentice"
RoleID: "fighters_guild_rank_journeyman"
RoleID: "fighters_guild_rank_master"
```

**3. Temporary access/permissions**
```go
RoleID: "lions_den_inn_guest"      // Can sleep at the inn
RoleID: "library_member"             // Can access restricted books
RoleID: "barracks_resident"         // Can enter military area
```

**4. Special status**
```go
RoleID: "wanted_criminal"
RoleID: "messenger_of_the_king"
RoleID: "blessed_by_priest"
```

Roles can be granted and revoked dynamically during gameplay (e.g., joining a guild, renting a room, completing a quest).

### ConditionNOT

Negates a condition.

```go
DialogCondition: dialogv2.ConditionNOT{
    Arg: dialogv2.ConditionDialogMemory{
        Key: "already_paid",
    },
}
```

### ConditionOR

OR logic over conditions.

```go
DialogCondition: dialogv2.ConditionOR{
    Args: []defs.DialogCondition{
        dialogv2.ConditionCulture{IsCulture: "roman"},
        dialogv2.ConditionCulture{IsCulture: "greek"},
    },
}
```

---

## Effects

Effects trigger actions during dialog.

### EventEffect

Broadcast an event (can trigger quests!).

```go
DialogEffect: dialogv2.EventEffect{
    Event: defs.Event{
        Type: "quest_started",
        Data: map[string]any{
            "questID": "deliver_package",
        },
    },
},
```

### AddGoldEffect

Give gold to player.

```go
DialogEffect: dialogv2.AddGoldEffect{
    Amount: 50,
},
```

### RemoveGoldEffect

Take gold from player.

```go
DialogEffect: dialogv2.RemoveGoldEffect{
    Amount: 25,
},
```

### SetDialogMemoryEffect

Store arbitrary memory.

```go
DialogEffect: dialogv2.SetDialogMemoryEffect{
    MemoryKey: "met_blacksmith",
},
```

### AddItemEffect

Add an item to the player's inventory.

```go
DialogEffect: dialogv2.AddItemEffect{
    ItemID:   "health_potion",
    Quantity: nil,  // defaults to 1
}

// Or give multiple items:
DialogEffect: dialogv2.AddItemEffect{
    ItemID:   "wolf_pelt",
    Quantity: &[]int{5},
}
```

Note: If the player's inventory is full, the item appears on the ground next to the player.

### ScheduleFutureEventEffect

Schedule an event for later.

```go
DialogEffect: dialogv2.ScheduleFutureEventEffect{
    Event: defs.Event{
        Type: "payment_overdue",
        Data: map[string]any{},
    },
    WaitDays: 7,       // 7 days from now
    UntilHour: utils.Int(9), // At 9 AM
},
```

### StartLoadScreenEffect

Trigger a load screen (ends dialog).

```go
DialogEffect: dialogv2.StartLoadScreenEffect{
    LoadFunction: func(ctx defs.GameContext) {
        // Your load logic here
    },
},

```

NOTE: this should only be used in special situations. Most dialogs will not use this effect.

---

## Actions

Actions interrupt dialog flow for UI interactions.

### GetUserInput

Shows a text input modal.

```go
Action: &defs.DialogAction{
    Type:   dialogv2.ActionTypeGetUserInput,
    Scope:  dialogv2.ActionScopePlayerName,
    Params: dialogv2.GetUserInputActionParams{
        ModalTitle:        "Enter Your Name",
        ConfirmButtonText:  "Confirm",
    },
},
```

NOTE: this should only be used in special situations. Most dialogs will never need this.

### ShowScreen

Display another screen (e.g., trade, inventory).

```go
Action: &defs.DialogAction{
    Type:   dialogv2.ActionTypeShowScreen,
    Params: dialogv2.ShowScreenActionParams{
        ScreenID: "trade_screen",
    },
},
```

NOTE: this should only be used for special situations. Most dialogs will not use this effect.

---

## Response Chaining

Responses can chain together for dramatic pacing:

```go
Responses: []DialogResponse{
    {
        Text: "I have something to tell you...",
        Effects: []DialogEffect{...},
        NextResponse: &DialogResponse{
            Text: "Actually, never mind. Forget I said anything.",
            Effects: []DialogEffect{...},
        },
    },
},
```

### Groupers

Responses without text act as condition-based routers:

```go
DialogResponse{
    Conditions: []DialogCondition{
        dialogv2.ConditionDialogMemory{Key: "knows_secret"},
    },
    NextResponseOptions: []DialogResponse{
        {
            Conditions: []DialogCondition{
                dialogv2.ConditionHasRole{RoleID: "trusted_ally"},
            },
            Text: "Since you're a trusted ally, I'll tell you...",
            Effects: []DialogEffect{...},
        },
        {
            Text: "Actually, it's not important.",
        },
    },
},
```

---

## Contextual Responses (Profile/Role-Based)

A powerful pattern: shared topics (like "Background" or "Rumors") can have different responses depending on who is speaking. Use groupers with `ConditionDialogProfile` or `ConditionHasRole` to route to the appropriate response.

### Pattern: Responses by Dialog Profile

```go
DialogTopic{
    ID:     "background",
    Prompt: "What is your background?",
    Responses: []DialogResponse{
        // Grouper with no conditions - evaluated in order until a match
        {
            // Check for specific NPC profiles first
            NextResponseOptions: []DialogResponse{
                {
                    Conditions: []DialogCondition{
                        dialogv2.ConditionDialogProfile{ProfileID: "elder_dialog"},
                    },
                    Text: "I have served as elder of this village for thirty winters. I remember when the old fortress was still inhabited...",
                },
                {
                    Conditions: []DialogCondition{
                        dialogv2.ConditionDialogProfile{ProfileID: "blacksmith_dialog"},
                    },
                    Text: "I learned my trade in the imperial capital. The finest smiths in the land trained me before I set out on my own.",
                },
                {
                    Conditions: []DialogCondition{
                        dialogv2.ConditionDialogProfile{ProfileID: "guard_dialog"},
                    },
                    Text: "I served in the king's army for ten years. Now I keep the peace here in this village.",
                },
            },
        },
        // Fallback: generic responses based on role
        {
            NextResponseOptions: []DialogResponse{
                {
                    Conditions: []DialogCondition{
                        dialogv2.ConditionHasRole{RoleID: "merchant"},
                    },
                    Text: "I've traveled far and wide, trading goods across many lands.",
                },
                {
                    Conditions: []DialogCondition{
                        dialogv2.ConditionHasRole{RoleID: "scholar"},
                    },
                    Text: "I devoted my life to studying the ancient texts. Knowledge is my greatest treasure.",
                },
                {
                    Conditions: []DialogCondition{
                        dialogv2.ConditionHasRole{RoleID: "guard"},
                    },
                    Text: "I patrol these lands, keeping travelers safe from bandits.",
                },
            },
        },
        // Final fallback: generic default
        {
            Text: "I'm just a simple folk, living my life one day at a time.",
        },
    },
}
```

### When to Use Each Condition

| Condition | Use When |
|----------|----------|
| `ConditionDialogProfile` | A specific NPC has unique dialog that shouldn't apply to others |
| `ConditionHasRole` | A role (like "guard" or "scholar") shares common dialog across NPCs |
| `ConditionCulture` | NPCs of a certain culture share background/history |
| `ConditionMapID` | The NPC's location affects their background |

### Rule of Thumb

1. **Most specific first** - Specific NPCs (via `ConditionDialogProfile`)
2. **General categories next** - Roles or cultures
3. **Default last** - Generic fallback with no conditions

This ensures the most appropriate response is always selected, but no one is left without an answer.

---

## Topic Unlocking

Topics can be unlocked dynamically via `NextTopics`:

```go
DialogResponse{
    Text: "I heard rumors about a treasure in the old mine.",
    NextTopics: []TopicID{"mine_location"},
    Effects: []DialogEffect{...},
},
```

The player will then see "mine_location" as an available topic in future conversations.

---

## Dialog Memory

Dialog state persists via `DialogProfileState.Memory` (a `map[string]bool`).

**Special Memory Keys:**
- `TOPIC_SEEN:<TopicID>` - Topic was discussed
- `TOPIC_UNLOCKED:<TopicID>` - Topic was unlocked
- `RESPONSE_SEEN:<ResponseID>` - Response was shown (for `Once` tracking)

**Custom Memory:**
```go
// Set
SetDialogMemoryEffect{MemoryKey: "met_at_tavern"}

// Check
ConditionDialogMemory{Key: "met_at_tavern"}
```

---

## Variable Substitution

Text can include variables that are replaced at runtime:

```go
Text: "Welcome, {player_name} of {player_culture}!",
```

Available variables:
- `{player_name}` - Player's name
- `{player_culture}` - Player's culture display name

---

## Creating a Dialog Profile

```go
DialogProfileDef{
    ProfileID: "blacksmith_dialog",
    Greeting: []DialogResponse{
        // First: conditional responses in priority order
        {
            Conditions: []DialogCondition{
                dialogv2.ConditionDialogMemory{Key: "completed_repair"},
            },
            Text: "Back again? Good to see you!",
        },
        // Last: default fallback (no conditions)
        {
            Text: "Good day, traveler! Need any weapons forged?",
        },
    },
    TopicsIDs: []TopicID{
        "blacksmith_weapons",
        "blacksmith_repair",
        "blacksmith_rumors",
    },
}
```

---

## Creating Topics

```go
DialogTopic{
    ID:     "blacksmith_weapons",
    Prompt: "Show me your weapons",
    Responses: []DialogResponse{
        {
            Text: "Take a look at these fine blades!",
            Effects: []DialogEffect{
                dialogv2.EventEffect{Event: defs.Event{Type: "open_shop"}},
            },
            Replies: []DialogReply{
                {Text: "I'll take the sword.", Goodbye: true},
                {Text: "Too expensive for me.", Goodbye: true},
            },
        },
    },
}
```

---

## Integration with Quests

Dialog effects can trigger quests via events:

```go
Effects: []DialogEffect{
    dialogv2.EventEffect{
        Event: defs.Event{
            Type: "quest_started",
            Data: map[string]any{
                "questID": "deliver_message",
            },
        },
    },
},
```

The quest system listens for this event and starts the quest if conditions are met.
The quest system will also listen for specific events to advance a quest to its next stage. For more information about quests, see `quest_system.md`. 

---

## Best Practices

1. **Condition order matters** - Responses are checked in order; first true wins. Put specific conditions first, defaults last.
2. **Always include a default** - Every response list should end with a fallback that has no conditions
3. **Use contextual routing for shared topics** - When multiple NPCs share a topic (like "Background"), use groupers with profile/role conditions to provide varied responses
4. **Keep responses concise** - Long text should be broken into chained responses
5. **Use conditions sparingly** - Each condition adds evaluation overhead
6. **Group related topics** - Topics the NPC can always discuss should be in `ProfileDef.TopicsIDs`
7. **Unlock topics strategically** - Don't unlock everything at once
8. **Use memory for state** - Track important conversation outcomes
