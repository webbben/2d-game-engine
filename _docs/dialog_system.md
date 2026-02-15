# Dialog System Architecture

This document describes the structure and behavior of the dialog system, including how dialog definitions, conditions, effects, and actions work together.

The dialog system is:

- Definition-driven (written in Go code)
- Context-aware (uses conditions)
- Side-effect capable (via effects)
- Event-compatible (effects can emit events for quests or other systems)
- UI-extensible (via dialog actions)

---

# High-Level Overview

The dialog system is composed of:

- **Dialog Profiles** (assigned to NPCs)
- **Dialog Topics** (semantic prompts the player can select)
- **Dialog Responses** (NPC replies)
- **Dialog Replies** (player responses to NPC lines)
- **Conditions** (control availability)
- **Effects** (apply state changes)
- **Dialog Actions** (UI interruptions such as input modals)

The system is intentionally modular and avoids tight coupling with quests or other systems. Instead, dialog may trigger changes indirectly via `Effect` implementations.

---

# Core Concepts

## 1. DialogProfileDef

```go
type DialogProfileDef struct {
    ProfileID DialogProfileID
    Greeting  []DialogResponse
    TopicsIDs []TopicID
}
```

A dialog profile represents the full conversational identity of an NPC.

It defines:

- **Greeting responses** — lines spoken when conversation begins.
- **Topic IDs** — which dialog topics are available to this NPC.

Profiles may be:
- Shared across multiple NPCs
- Unique to a specific NPC

---

## 2. DialogTopic

```go
type DialogTopic struct {
    ID         TopicID
    Prompt     string
    Conditions []Condition
    Responses  []DialogResponse
    Metadata   map[string]string
}
```

A dialog topic represents a semantic subject the player can raise (e.g., `"Rumors"`, `"Background"`).

It defines:

- The prompt shown to the player
- Conditions for visibility
- All possible NPC responses to that topic
- Optional metadata (for UI hints or tagging)

### Topic Visibility

A topic appears only if:

- All `Conditions` return `true`
- The topic is unlocked (via `EffectContext.UnlockTopic`)
- The topic is not hidden by other game state

Conditions use AND logic.

---

## 3. DialogResponse

```go
type DialogResponse struct {
    Text       string
    Conditions []Condition
    Effects    []Effect
    NextTopics []TopicID
    Once       bool

    Replies     []DialogReply
    Action      *DialogAction
    NextResponse *DialogResponse
}
```

A dialog response is a single possible NPC reply to a topic.

Responses are **contextual** and intentionally not centralized in a global definitions manager.

### Response Selection

When a topic is selected:

1. All responses are evaluated.
2. Only responses whose `Conditions` return `true` are eligible.
3. One response is chosen (selection strategy is implementation-dependent).

### Once Flag

If `Once == true`:
- The response becomes ineligible after it has been shown once.

---

### Response Flow Order

When a response is triggered:

1. If `Action` is set → execute it first.
2. Display `Text`.
3. Apply `Effects`.
4. Unlock any `NextTopics`.
5. If `Replies` exist → show reply options.
6. If `NextResponse` exists → automatically continue.

---

## 4. DialogReply

```go
type DialogReply struct {
    Text       string
    Conditions []Condition
    Effects    []Effect

    NextResponse *DialogResponse
    NextTopicID  TopicID
}
```

A dialog reply represents a player’s response to an NPC line.

When selected:

1. Conditions are validated.
2. Effects are applied.
3. If `NextResponse` is set → NPC immediately responds.
4. If `NextTopicID` is set → dialog shifts to that topic.

Replies allow branching conversational flow.

---

# Conditions

```go
type Condition interface {
    IsMet(ctx ConditionContext) bool
}
```

Conditions control:

- Topic visibility
- Response eligibility
- Reply availability

All conditions use AND logic.

---

## ConditionContext

```go
type ConditionContext interface {
    HasSeenTopic(id TopicID) bool
    IsTopicUnlocked(id TopicID) bool
}
```

Conditions operate only on context passed in.  
They should not directly access global systems.

This keeps them testable and modular.

---

## Example: MemoryCondition

```go
type MemoryCondition struct {
    Key  string
    Seen bool
}
```

Used for checking simple memory state, such as:

- Whether a topic has been seen
- Whether a narrative beat occurred

---

# Effects

```go
type Effect interface {
    Apply(ctx EffectContext)
}
```

Effects are responsible for changing state.

Dialog does not directly modify systems.  
It applies `Effect`s, and those effects perform the changes.

---

## EffectContext

```go
type EffectContext interface {
    MarkTopicSeen(id TopicID)
    UnlockTopic(id TopicID)

    AddGold(amount int)
}
```

Effects are applied through this interface to prevent direct coupling.

Effects may:

- Unlock topics
- Mark topics as seen
- Add gold
- Set memory flags
- Emit narrative events
- Give items
- Trigger quest events (indirectly)

---

## Example: SetMemoryEffect

```go
type SetMemoryEffect struct {
    Key  string
    Seen bool
}
```

Used to mark narrative state for later condition checks.

---

# DialogAction

```go
type DialogAction struct {
    Type   DialogActionType
    Scope  DialogActionResultScope
    Params any
}
```

A DialogAction temporarily interrupts dialog flow.

Examples:

- Showing a text input modal
- Presenting a trade interface
- Triggering a cutscene
- Opening a special UI panel

### Execution Rules

- Actions execute before dialog text or effects.
- They may return data, depending on `Scope`.
- The UI layer handles rendering and user interaction.

---

# Dialog Flow Lifecycle

1. Player initiates conversation.
2. Greeting response is selected and executed.
3. Available topics are shown.
4. Player selects topic.
5. Eligible response is chosen.
6. Effects are applied.
7. Replies are presented (if any).
8. Flow continues via `NextResponse` or topic selection.

---

# Architectural Principles

The dialog system follows these design principles:

### 1. Definition-Driven

All dialog is written as Go definitions.  
This allows:

- Compile-time safety
- IDE validation
- Easy refactoring of IDs
- No runtime parsing

---

### 2. Decoupled from Quests

Dialog does not directly manipulate quests.

Instead:

- Dialog may emit events via an `Effect`.
- QuestManager listens to events.
- Quest state changes occur externally.

This preserves separation of concerns.

---

### 3. Context-Based Logic

All eligibility decisions are driven by:

- `Condition`
- `ConditionContext`

No hardcoded branching inside the dialog system itself.

---

### 4. Explicit State Mutation

All state changes occur via `Effect`.

No hidden side effects.

---

# Intended Strengths

This dialog architecture supports:

- Conditional branching
- Stateful conversations
- One-time responses
- Unlockable topics
- Event-driven quest progression
- UI interruptions
- Narrative chaining
- Reusable profiles

It is scalable, testable, and suitable for complex narrative systems.

---

# Summary

The dialog system is:

- Declarative
- Modular
- Extensible
- Event-compatible
- Architecturally clean

It separates:

- **What is said** (definitions)
- **When it is said** (conditions)
- **What happens because of it** (effects)
- **How the UI reacts** (actions)

This structure ensures long-term maintainability as narrative complexity grows.
