package npc

import (
	"log/slog"
	"time"

	"github.com/webbben/2d-game-engine/entity"
)

type NPC struct {
	NPCInfo
	Entity *entity.Entity

	TaskMGMT
}

type NPCInfo struct {
	ID string
}

type TaskMGMT struct {
	Active      bool // if the NPC is actively doing a task right now
	CurrentTask *Task
	WaitUntil   time.Time
}

func (n *NPC) SetTask(t Task) {
	if n.Active {
		slog.Warn("tried to set task on already active NPC")
		return
	}

	t.Owner = n
	n.CurrentTask = &t
	n.Active = true
}

// Interrupt regular NPC updates for a certain duration
func (n *NPC) Wait(d time.Duration) {
	n.WaitUntil = time.Now().Add(d)
}
