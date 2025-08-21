package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/logz"
)

type NPC struct {
	NPCInfo
	Entity *entity.Entity

	TaskMGMT
}

type NPCInfo struct {
	ID          string
	DisplayName string
}

type TaskMGMT struct {
	Active      bool // if the NPC is actively doing a task right now
	CurrentTask *Task
	WaitUntil   time.Time
	DefaultTask Task
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
	n.WaitUntil = time.Now().Add(d)
}
