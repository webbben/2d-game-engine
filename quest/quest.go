package quest

import (
	"log"

	"github.com/webbben/2d-game-engine/internal/pubsub"
)

type QuestManager struct {
	IncompleteQuests []Quest
	CompletedQuests  []Quest
}

func NewQuestManager(eventBus *pubsub.EventBus) *QuestManager {
	qm := QuestManager{}

	eventBus.Subscribe(pubsub.Event_EntityAttacked, qm.OnAttackEvent)
	eventBus.Subscribe(pubsub.Event_EntityKilled, qm.OnKillEvent)
	eventBus.Subscribe(pubsub.Event_EntityDied, func(e pubsub.Event) { log.Println("entity died: no quest event listener!") })
	eventBus.Subscribe(pubsub.Event_StartDialog, func(e pubsub.Event) { log.Println("start dialog: no quest event listener!") })
	eventBus.Subscribe(pubsub.Event_GetItem, func(e pubsub.Event) { log.Println("get item: no quest event listener!") })
	eventBus.Subscribe(pubsub.Event_UseItem, func(e pubsub.Event) { log.Println("use item: no quest event listener!") })
	eventBus.Subscribe(pubsub.Event_VisitMap, func(e pubsub.Event) { log.Println("visit map: no quest event listener!") })
	eventBus.Subscribe(pubsub.Event_TimePass, func(e pubsub.Event) { log.Println("time pass: no quest event listener!") })

	return &qm
}

func (qm *QuestManager) OnAttackEvent(e pubsub.Event) {
	for _, q := range qm.IncompleteQuests {
		for _, obj := range q.Objectives {
			if obj.Complete {
				continue
			}
			if obj.Type == ObjectiveAttack {
				npcID := e.Data["targetNPC"]
				entID := e.Data["targetEntity"]
				if npcID == "" && entID == "" {
					panic("Attack objective: no npc or entity ID found in event data")
				}
				if npcID == obj.Target || entID == obj.Target {
					obj.Progress++
					if obj.Progress >= obj.ProgressCompete {
						obj.Complete = true
					}
				}
			}
		}
	}
}

func (qm *QuestManager) OnKillEvent(e pubsub.Event) {
	for _, q := range qm.IncompleteQuests {
		for _, obj := range q.Objectives {
			if obj.Complete {
				continue
			}
			if obj.Type == ObjectiveKill {
				npcID := e.Data["targetNPC"]
				entID := e.Data["targetEntity"]
				if npcID == "" && entID == "" {
					panic("Kill objective: no npc or entity ID found in event data")
				}
				if npcID == obj.Target || entID == obj.Target {
					obj.Progress++
					if obj.Progress >= obj.ProgressCompete {
						obj.Complete = true
					}
				}
			}
		}
	}
}

type QuestState string

const (
	QuestNotStarted QuestState = "not_started"
	QuestInProgress QuestState = "in_progress"
	QuestCompleted  QuestState = "completed"
	QuestFailed     QuestState = "failed"
)

type Quest struct {
	ID          string
	State       QuestState
	Title       string
	Description string
	Objectives  []*Objective
}

type ObjectiveType string

const (
	ObjectiveAttack   ObjectiveType = "attack"    // trigger on attacking an entity
	ObjectiveKill     ObjectiveType = "kill"      // trigger on killing an entity
	ObjectiveEntDeath ObjectiveType = "ent_death" // trigger on entity death (not directly killed by player)

	ObjectiveTalk        ObjectiveType = "talk"         // trigger on talking to an entity
	ObjectiveDialogTopic ObjectiveType = "dialog_topic" // trigger on starting a specific dialog topic

	ObjectiveGetItem ObjectiveType = "get_item" // trigger on obtaining an item
	ObjectiveUseItem ObjectiveType = "use_item" // trigger on using (consuming or equiping) an item

	ObjectiveVisitMap ObjectiveType = "visit_map" // trigger on entering a map

	ObjectiveTimePass ObjectiveType = "time_pass" // trigger on passage of time
)

type Objective struct {
	ID              string
	Type            ObjectiveType
	Target          string // probably an ID of some kind, be it an NPC, entity, item, etc.
	Progress        int
	ProgressCompete int // value for progress to reach in order for this objective to be complete
	Complete        bool
}
