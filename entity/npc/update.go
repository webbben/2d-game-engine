package npc

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
)

type debug struct {
	lastDebugPrint time.Time
}

func (n *NPC) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if n.Entity == nil {
		panic("tried to draw NPC that doesn't have an entity!")
	}
	n.Entity.Draw(screen, offsetX, offsetY)
}

func (n *NPC) Update() {
	if time.Since(n.debug.lastDebugPrint) > 10*time.Second {
		n.debug.lastDebugPrint = time.Now()
		logz.Println(n.DisplayName, "== DEBUG PRINT ==")
		logz.Println(n.DisplayName, "IsActive:", n.IsActive())
		if n.CurrentTask != nil {
			logz.Println(n.DisplayName, "Current Task:", n.CurrentTask.GetName())
			logz.Println(n.DisplayName, "Status:", n.CurrentTask.GetStatus())
		}
	}

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
	if n.IsActive() {
		if n.CurrentTask == nil {
			panic("NPC is marked as active, but there is no current task set")
		}
		n.HandleTaskUpdate()
	} else {
		err := n.SetTask(n.DefaultTask, false)
		if err != nil {
			panic(err)
		}
	}
}
