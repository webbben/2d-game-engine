package npc

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func (n *NPC) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if n.Entity == nil {
		panic("tried to draw NPC that doesn't have an entity!")
	}
	n.Entity.Draw(screen, offsetX, offsetY)
}

func (n *NPC) Update() {
	n.npcUpdates()
	n.Entity.Update()
}

// Updates related to NPC behavior or tasks
func (n *NPC) npcUpdates() {
	if time.Until(n.waitUntil) > 0 {
		return
	}
	if n.waitUntilDoneMoving {
		if n.Entity.Movement.IsMoving {
			return
		}
		n.waitUntilDoneMoving = false
	}
	if n.Active {
		if n.CurrentTask == nil {
			panic("NPC is marked as active, but there is no current task set")
		}
		n.HandleTaskUpdate()
	} else {
		n.SetTask(n.DefaultTask)
	}
}
