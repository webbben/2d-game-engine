package npc

import (
	"time"
)

func (n *NPC) Update() {
	n.npcUpdates()
	n.Entity.Update()
}

// Updates related to NPC behavior or tasks
func (n *NPC) npcUpdates() {
	if time.Until(n.waitUntil) > 0 {
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
