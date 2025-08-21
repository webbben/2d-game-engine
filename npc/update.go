package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/internal/logz"
)

func (n *NPC) Update() {
	n.npcUpdates()
	n.Entity.Update()
}

// Updates related to NPC behavior or tasks
func (n *NPC) npcUpdates() {
	if time.Until(n.WaitUntil) > 0 {
		logz.Println(n.DisplayName, "npc waiting")
		return
	}
	if n.Active {
		if n.CurrentTask == nil {
			panic("NPC is marked as active, but there is no current task set")
		}
		n.CurrentTask.OnUpdate()
	} else {
		n.SetTask(n.DefaultTask)
	}
}
