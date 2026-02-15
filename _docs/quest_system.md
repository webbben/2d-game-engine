# Quest System Architecture

This document describes the structure and behavior of the quest system.

The quest system is:

- Definition-driven (written in Go code)
- Event-driven (no polling or update-loop checks)
- Stage-based (finite progression states)
- Data-indexed (efficient event-to-reaction matching)
- Decoupled from dialog and other systems

---

# High-Level Overview

A quest consists of:

- A unique ID
- A name and description
- A set of stages
- A start trigger
- Event-driven reactions
- Optional terminal states (success/failure)

Quests progress only when:

1. A relevant event occurs.
2. A reaction subscribed to that event evaluates true.
3. Actions execute.
4. The stage transitions (or ends).

There is **no continuous condition checking** in update loops.

---

# Core Concepts

## 1. QuestDef

```go
type QuestDef struct {
    ID          QuestID
    Name        string
    Description string
    Stages      map[QuestStageID]QuestStageDef
    StartStage  QuestStageID

    StartTrigger QuestStartTrigger
}
```

A `QuestDef` defines the entire lifecycle of a quest.

### Required Components

- `ID` — unique identifier
- `Name` — player-facing name
- `Stages` — all quest stages
- `StartStage` — initial stage
- `StartTrigger` — event that causes the quest to begin

Validation ensures:

- ID is set
- Name is set
- At least one stage exists
- Start stage exists
- StartTrigger has an EventType

---

## 2. QuestStartTrigger

```go
type QuestStartTrigger struct {
    EventType  EventType
    Conditions []QuestConditionDef
}
```

Defines how a quest begins.

When `EventType` is fired:

1. Conditions are evaluated.
2. If true → quest is started.
3. The `StartStage` becomes active.

This is fully event-driven.

Common start triggers include:

- Narrative events from dialog
- NPC death events
- Item acquisition
- Location discovery

---

# Quest Stages

## 3. QuestStageDef

```go
type QuestStageDef struct {
    ID                QuestStageID
    Title             string
    Description       string
    LongerDescription string
    OnEnter           []QuestActionDef
    Reactions         []QuestReactionDef
}
```

A quest stage represents a single objective state.

Each stage:

- Has a player-visible description.
- May run actions when entered.
- Defines reactions to events.
- Determines how progression occurs.

### OnEnter

Executed immediately when the stage becomes active.

Examples:

- Show quest update notification
- Spawn NPC
- Unlock door
- Trigger cutscene
- Set quest variable

These actions are unconditional.

---

# Reactions (Core of Progression)

## 4. QuestReactionDef

```go
type QuestReactionDef struct {
    SubscribeEvent EventType
    Conditions     []QuestConditionDef
    Actions        []QuestActionDef
    NextStage      *QuestStageID
    TerminalStatus QuestTerminalStatus
}
```

Reactions define how a stage responds to events.

When `SubscribeEvent` occurs:

1. Conditions are evaluated.
2. If true:
   - Actions execute.
   - Stage may transition.
   - Quest may terminate.

---

### Reaction Rules

- `SubscribeEvent` is required.
- If `NextStage` is set:
  - The quest transitions to that stage.
  - `TerminalStatus` must not be set.
- If `TerminalStatus` is set:
  - Quest ends (success or failure).
  - `NextStage` must be nil.

---

# Event-Driven Execution Model

The quest system does NOT poll conditions every frame.

Instead:

1. QuestManager subscribes to gameplay events.
2. When an event occurs:
   - Look up all start triggers for that event.
   - Look up all stage reactions subscribed to that event.
3. Evaluate only relevant quests.

This prevents scanning all quests on every event.

---

# Internal Indexing Strategy (Recommended)

At load time, build:

```
startTriggersByEvent map[EventType][]QuestID
stageReactionsByEvent map[EventType][]ReactionRef
```

Where `ReactionRef` identifies:

- QuestID
- StageID
- Reaction index

On event:

1. Map lookup for event type.
2. Process only relevant quests/reactions.
3. Immediate return if none found.

This ensures scalability.

---

# Quest Actions

## 5. QuestActionDef

```go
type QuestActionDef struct {
    Type   QuestActionType
    Params map[string]string
}
```

Defines actions executed:

- When entering a stage (`OnEnter`)
- When a reaction triggers

Actions are interpreted by QuestManager or a related system.

Examples of possible action types:

- ShowNotification
- SpawnNPC
- UnlockDoor
- SetVariable
- EmitEvent
- GiveItem
- TriggerCutscene

Actions are data-driven and extensible.

---

# Quest Conditions

## 6. QuestConditionDef

```go
type QuestConditionDef struct {
    Type   QuestConditionType
    Params map[string]string
}
```

Conditions determine whether:

- A quest can start
- A reaction should trigger

All conditions use AND logic.

Conditions may inspect:

- Event data
- Quest variables
- World state
- NPC state
- Inventory
- Time
- Faction reputation

Condition evaluation happens only when the subscribed event occurs.

---

# Terminal Status

```go
type QuestTerminalStatus int
```

Represents quest end states:

- Success
- Failure
- Still ongoing

A reaction may mark the quest terminal instead of transitioning stages.

---

# Lifecycle Example

1. Dialog emits `EventGuardApproaching`.
2. QuestStartTrigger matches event.
3. Conditions pass.
4. Quest starts at `StartStage`.
5. `OnEnter` actions execute.
6. Later, `EventNPCDied` occurs.
7. Stage reaction subscribed to `EventNPCDied` runs.
8. Conditions pass.
9. Actions execute.
10. Quest transitions to next stage or ends.

No polling.
No global scanning.
Fully event-driven.

---

# Architectural Principles

## 1. Event-Driven, Not Polled

Quests react only when relevant events occur.

No update-loop condition checking.

---

## 2. Decoupled From Dialog

Dialog may emit events.

Quest system listens to events.

Dialog does NOT directly manipulate quest state.

---

## 3. Data-Driven

All progression logic lives in definitions.

QuestManager executes behavior based on data.

---

## 4. Explicit Stage Ownership

Each stage defines:

- What it cares about
- What happens when triggered
- Where it transitions

No hidden transitions.

---

## 5. Deterministic Flow

Given the same event sequence, quest progression is deterministic.

No hidden randomness unless explicitly defined.

---

# Strengths of This Architecture

- Highly scalable
- Easy to debug (clear event names)
- No polling overhead
- Clean separation of concerns
- Clear progression logic
- Easy to add new quest types
- Safe refactoring via Go constants
- IDE-supported validation

---

# Intended System Boundaries

The quest system:

- Does not render UI directly
- Does not own dialog
- Does not directly manipulate world objects

It issues actions.

Other systems execute those actions.

---

# Summary

The quest system is a:

- Stage-based
- Event-driven
- Definition-driven
- Data-indexed
- Decoupled progression engine

It integrates cleanly with dialog and gameplay systems through events and actions, while remaining modular and scalable.

This design supports simple linear quests as well as complex multi-stage branching narratives.
