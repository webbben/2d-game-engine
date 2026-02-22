# Game Data Definitions

This document describes the core data definitions in `data/defs/`. These definitions form the static "blueprint" data that drives the game—the types of characters, items, quests, dialogs, and more that exist in your game world.

Definitions are distinct from **runtime state** (stored in `data/state/`). Definitions describe *what* exists; state describes *what is currently happening*.

---

## Overview of All Definition Types

| Definition | File | Purpose |
|------------|------|---------|
| `CharacterDef` | character.go | Blueprint for NPCs and other characters |
| `BodyDef` | body.go | Visual appearance (skin, eyes, hair) |
| `ItemDef` | item.go | All item types (weapons, armor, consumables) |
| `DialogProfileDef` | dialog.go | NPC conversation data |
| `QuestDef` | quest.go | Quest stages, triggers, and actions |
| `TaskDef` | task.go | NPC AI behaviors |
| `ScheduleDef` | task.go | NPC daily routines |
| `ScenarioDef` | scenario.go | Controlled map setups |
| `ClassDef` | skills.go | Character classes and skill categoriesDef` | skills |
| `Skill.go | Individual skills |
| `TraitDef` | skills.go | Personality traits with buffs/debuffs |
| `Event` | event.go | Game events for system communication |
| `ShopkeeperDef` | shopkeeper.go | Merchant inventory and settings |
| `MapDef` | map.go | Map metadata |
| `FootstepSFXDef` | audio.go | Footstep sound assignments |

---

# Characters

## CharacterDef

```go
type CharacterDef struct {
    ID             CharacterDefID
    Unique         bool                    // singleton character
    BodyDef        BodyDef                 // visual appearance
    DisplayName    string                  // short name (e.g., "Scipio Africanus")
    FullName       string                  // extended name (e.g., "Publius Cornelius Scipio Africanus")
    ClassName      string                  // display class name
    ClassDefID     ClassDefID              // class blueprint reference
    InitialInventory StandardInventory     // starting items
    DialogProfileID DialogProfileID        // conversation data
    FootstepSFXDefID FootstepSFXDefID      // footstep sounds
    ScheduleID     ScheduleID              // daily routine
    BaseAttributes map[AttributeID]int     // starting stats
    BaseSkills     map[SkillID]int         // starting skill levels
    InitialTraits  []TraitID               // personality traits
    BaseVitals     Vitals                 // starting health/stamina
}
```

**Key Concepts:**

- **Unique**: If `true`, only one instance of this character can exist at a time. This allows looking up the character state by `CharacterDefID`. Non-unique characters can have multiple instances.
- **ClassName vs ClassDefID**: `ClassName` is for display purposes (what the player sees), while `ClassDefID` references the actual class definition with skill categories and favored attributes.
- **InitialInventory**: Starting equipment and items. Includes both backpack items and equipped slots.
- **BaseAttributes/Skills**: Starting values before modifiers from traits, equipment, or other sources.

---

## BodyDef

```go
type BodyDef struct {
    BodyHSV HSV         // skin color
    BodyID  BodyPartID  // body shape/type
    EyesHSV HSV         // eye color
    EyesID  BodyPartID  // eye style
    HairHSV HSV        // hair color
    HairID  BodyPartID  // hair style
    ArmsID  BodyPartID  // arm style
    LegsID  BodyPartID  // leg style
}
```

**Key Concepts:**

- **HSV Values**: Hue, Saturation, Value for procedural color variation. Allows characters to have different skin, eye, and hair colors without needing separate sprite sheets.
- **BodyPartID**: References to specific body part definitions that define animations (idle, walk, run, attack) and visual properties.
- **AnimationParams**: Each body part has animations for different states:
  - `IdleAnimation` - standing still
  - `WalkAnimation` - walking
  - `RunAnimation` - running
  - `SlashAnimation` - melee attack (right-handed)
  - `BackslashAnimation` - melee attack (left-handed)
  - `ShieldAnimation` - blocking

**SelectedPartDef** contains the actual animation frame data:
- `TilesLeft`, `TilesRight`, `TilesUp`, `TilesDown` - frame indices for each direction
- `AuxLeft`, etc. - frames that include auxiliary items (weapons, torches)
- `FlipRForL` - whether to mirror right-hand frames for left-hand
- `CropHairToHead` - whether hair should be hidden by headwear

---

# Items

## ItemDef Interface

```go
type ItemDef interface {
    GetID() ItemID
    GetName() string
    GetDescription() string
    GetValue() int          // sell price
    GetWeight() float64    // inventory weight
    GetMaxDurability() int
    
    GetTileImg() *ebiten.Image           // inventory icon
    GetEquipedTiles() []*ebiten.Image     // equipped appearance
    
    GetBodyPartDef() *SelectedPartDef    // body appearance (armor)
    GetLegsPartDef() *SelectedPartDef    // leg appearance
    
    // Item type checks
    IsGroupable() bool       // stacks with same items
    IsWeapon() bool
    IsHeadwear() bool
    IsBodywear() bool
    IsFootwear() bool
    IsAmulet() bool
    IsRing() bool
    IsAmmunition() bool
    IsAuxiliary() bool       // left-hand items (shields, torches)
    IsConsumable() bool
    IsMiscItem() bool
    IsCurrencyItem() bool
    IsEquipable() bool
    
    GetItemType() ItemType
    Load()                    // load images
    Validate()                // validate definition
}
```

**Item Types:**

| Type | Equip Slot | Notes |
|------|------------|-------|
| Weapon | Right hand | Melee or ranged |
| Headwear | Head | Helmets, hats, helmets |
| Bodywear | Body | Armor, clothing |
| Footwear | Feet | Boots, shoes |
| Amulet | Neck | Necklaces, amulets |
| Ring | Finger | Two rings allowed |
| Auxiliary | Left hand | Shields, torches |
| Ammunition | Quiver | Arrows, bolts |
| Consumable | Inventory | Food, potions |
| Currency | Coin purse | Gold, coins |
| MiscItem | Inventory | Quest items, general items |

**InventoryItem** and **ItemInstance**:

```go
type ItemInstance struct {
    DefID      ItemID   // reference to ItemDef
    Durability int     // current condition (degrades with use)
}

type InventoryItem struct {
    Instance ItemInstance
    Def      ItemDef
    Quantity int       // for stackable items
}
```

**StandardInventory** contains:
- `CoinPurse` - currency items
- `InventoryItems` - backpack items
- `Equipment` - equipped items in all slots

---

# Skills & Attributes

## ClassDef

```go
type ClassDef struct {
    ID                  ClassDefID
    Name                string
    SkillCategories     map[SkillID]SkillCategory  // which category each skill belongs to
    FavoredAttributes   []AttributeID               // attribute bonus
    AboutMe             string                      // description
}
```

**Skill Categories:**
- Major - primary skills, level up faster
- Minor - secondary skills
- Misc - miscellaneous skills

The class defines which skills are Major/Minor/Misc, affecting how character level is calculated from skill progress.

## SkillDef

```go
type SkillDef struct {
    ID                  SkillID
    DisplayName         string
    GoverningAttributes []AttributeID  // attributes that affect this skill
    Description         string
}
```

Each skill is governed by one or more attributes. When an attribute increases, related skills may also improve.

## Trait Interface

```go
type Trait interface {
    GetID() TraitID
    GetName() string
    GetDescription() string
    GetTilesetSrc() string
    GetTileID() int
    GetConflictTraitIDs() []string           // mutually exclusive
    GetSkillChanges() map[SkillID]int       // skill bonuses/penalties
    GetAttributeChanges() map[AttributeID]int // attribute bonuses/penalties
    GetOpinionChangeToTraitHolder(factors OpinionFactors) int
    GetOpinionChangeToOther(factors OpinionFactors) int
}
```

**Traits** represent personality characteristics that modify:
- Skills (e.g., "Athletic" boosts physical skills)
- Attributes (e.g., "Strong" boosts Strength)
- NPC opinions of the character

**Conflict Traits**: Some traits cannot coexist (e.g., "Greedy" and "Generous").

## LevelSystemParameters

Defines the math behind character leveling:
- `MajorRate/MinorRate/MiscRate` - XP needed per level for each category
- `MajorWeight/MinorWeight/MiscWeight` - how much each category contributes to level
- `AttributeGrowth` - how much attributes increase with governed skills
- `FavoredBonus` - extra attribute points for favored attributes

---

# Vitals

```go
type CurMax struct {
    CurrentVal int
    MaxVal     int
}

type Vitals struct {
    Health  CurMax
    Stamina CurMax
}
```

Vitals track current and maximum values for health and stamina.

---

# Dialog

See [`dialog_system.md`](dialog_system.md) for detailed documentation.

**Key Types:**
- `DialogProfileDef` - assigned to NPCs, defines greeting and available topics
- `DialogTopic` - subject player can ask about
- `DialogResponse` - NPC's answer
- `DialogReply` - player's response option
- `DialogCondition` - controls visibility/availability
- `DialogEffect` - state changes from dialog
- `DialogAction` - UI interruptions (input modals, trade, etc.)

---

# Quests

See [`quest_system.md`](quest_system.md) for detailed documentation.

**Key Types:**

```go
type QuestDef struct {
    ID          QuestID
    Name        string
    Description string
    Stages      map[QuestStageID]QuestStageDef
    StartStage  QuestStageID
    StartTrigger QuestStartTrigger  // what starts the quest
}

type QuestStageDef struct {
    ID          QuestStageID
    Title       string              // optional stage title
    Objective   string              // shown to player
    Description string              // narrative context
    OnEnter     []QuestAction       // fires on stage entry
    Reactions   []QuestReactionDef  // event responses
}

type QuestReactionDef struct {
    SubscribeEvent EventType
    Conditions     []QuestConditionDef
    Actions        []QuestAction
    NextStage      QuestStageID
    TerminalStatus QuestTerminalStatus
}
```

**Key Concepts:**
- **Stage-based**: Quests progress through defined stages
- **Event-driven**: Reactions trigger on specific events
- **Conditional**: Conditions determine if reactions fire
- **Actions**: Side effects (spawn NPC, give item, unlock topic)

---

# Tasks & Scheduling

## TaskDef

```go
type TaskDef struct {
    TaskID   TaskID      // which task logic to run
    Priority TaskPriority // scheduling priority
    Params   any          // task-specific data
    NextTask *TaskDef     // chain to next task
}
```

**TaskPriority Levels:**
- Schedule - base daily routine
- Assigned - quest/dialog assigned tasks
- Emergency - critical tasks that override all

**Available Tasks** (implemented in `entity/npc/task_*.go`):
- `goto` - move to a location
- `follow` - follow a character
- `startDialog` - initiate conversation
- `fight` - combat behavior

## ScheduleDef

```go
type ScheduleDef struct {
    ID     ScheduleID
    Hourly map[int]TaskDef  // task for each hour (0-23)
}
```

Schedules define what NPCs do at each hour of the day. The `BuildSchedule` helper fills in gaps by carrying forward the last defined task.

---

# Scenarios

```go
type ScenarioDef struct {
    ID    ScenarioID
    MapID MapID
    Characters []ScenarioCharDef
}

type ScenarioCharDef struct {
    CharDefID       CharacterDefID
    DefaultSchedule ScheduleID
    DialogProfileID DialogProfileID
    SpawnCoordX, SpawnCoordY int
}
```

**Purpose:** Create controlled, isolated map setups.

**Use Cases:**
- Quest-specific encounters
- Cutscene setups
- Locked-area content

**Behavior:**
- Characters in a scenario are isolated from the "outside world"
- Actions within a scenario don't affect world state
- Used for quests requiring specific NPCs in specific places

---

# Events

```go
type Event struct {
    Type EventType
    Data map[string]any
}
```

**Purpose:** Decoupled communication between systems.

**Usage:**
- Quests subscribe to events
- Dialog emits events
- Any system can broadcast events

Common event types (defined elsewhere):
- `npc_killed`
- `item_given`
- `dialog_started`
- `quest_stage_advanced`
- Custom quest events

---

# Shops

```go
type ShopkeeperDef struct {
    ID            ShopID
    ShopName      string
    BaseInventory []InventoryItem  // items available for purchase
    BaseGold      int              // merchant's starting gold
}
```

Defines what items a merchant has available and how much gold they have for player sales.

---

# Audio

## FootstepSFXDef

```go
type FootstepSFXDef struct {
    ID             FootstepSFXDefID
    StepDefaultIDs []SoundID  // default surface
    StepWoodIDs    []SoundID  // wooden floor
    StepStoneIDs   []SoundID  // stone
    StepGrassID  // grass
    StepForestIDs  []SoundIDs   []SoundID  // forest floor
    StepSandIDs    []SoundID  // sand
    StepSnowIDs    []SoundID  // snow
}
```

Associates sound IDs with different ground surfaces. The game selects appropriate sounds based on the terrain the character walks on.

---

# Maps

```go
type MapDef struct {
    ID MapID
}
```

Currently minimal. Map loading is primarily handled through Tiled map files. This provides a hook for map-specific metadata if needed.

---

# Summary

The `data/defs/` package provides the static blueprint data for the entire game world:

1. **Characters** define who exists and their base attributes
2. **Body** defines visual appearance and animations
3. **Items** define equipment, consumables, and quest items
4. **Skills/Attributes/Traits** define character progression and personality
5. **Dialog** defines conversations
6. **Quests** define narrative content
7. **Tasks/Schedules** define NPC AI
8. **Scenarios** define controlled encounters
9. **Events** enable system communication
10. **Shops** define merchant inventories
11. **Audio** defines sound assignments

All definitions follow a consistent pattern:
- Use type aliases for IDs (e.g., `type QuestID string`)
- Include validation methods
- Use interfaces for extensible behavior (ItemDef, Trait, DialogCondition, etc.)
