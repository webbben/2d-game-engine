package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

type NPC struct {
	NPCInfo
	Entity *entity.Entity

	TaskMGMT
	World WorldContext
}

type WorldContext interface {
	FindNPCAtPosition(c model.Coords) (NPC, bool)
}

// Create new NPC from the given NPC struct. Ensures essential data is set.
func New(n NPC) NPC {
	if n.ID == "" {
		n.ID = general_util.GenerateUUID()
	}
	return n
}

type NPCInfo struct {
	ID          string
	DisplayName string
	Priority    int // priority assigned to this NPC by the map it is added to
}

type TaskMGMT struct {
	Active      bool // if the NPC is actively doing a task right now
	CurrentTask *Task
	waitUntil   time.Time
	DefaultTask Task
	// number of ticks this NPC has been stuck (failing to move to its goal).
	// TODO implement this if needed. so far, haven't needed to report stuck NPCs.
	StuckCount int
}

func (n *NPC) SetTask(t Task) {
	if n.Active {
		logz.Warnln(n.DisplayName, "tried to set task on already active NPC")
		return
	}

	t.Owner = n
	n.CurrentTask = &t
	n.Active = true
}

// Interrupt regular NPC updates for a certain duration
func (n *NPC) Wait(d time.Duration) {
	n.waitUntil = time.Now().Add(d)
}
