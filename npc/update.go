package npc

import (
	"log"
	"math/rand"
	"time"

	"github.com/webbben/2d-game-engine/internal/model"
)

var task Task = Task{
	Description: "wandering aimlessly",
	Context:     make(map[string]interface{}),
	startFn: func(t *Task) {
		// set a random location to walk to
		width, height := t.Owner.Entity.World.MapDimensions()
		c := model.Coords{
			X: rand.Intn(width),
			Y: rand.Intn(height),
		}
		t.Context["goal"] = c
		log.Println("npc traveling to:", t.Context["goal"])
		me := t.Owner.Entity.GoToPos(c)
		if !me.Success {
			log.Println("failed to call GoToPos:", me)
		}
	},
	isCompleteFn: func(t Task) bool {
		return t.Owner.Entity.TilePos.Equals(t.Context["goal"].(model.Coords)) && !t.Owner.Entity.Movement.IsMoving
	},
}

func (n *NPC) Update() {
	n.npcUpdates()
	n.Entity.Update()
}

// Updates related to NPC behavior or tasks
func (n *NPC) npcUpdates() {
	if time.Until(n.WaitUntil) > 0 {
		return
	}
	if n.Active {
		if n.CurrentTask == nil {
			panic("NPC is marked as active, but there is no current task set")
		}
		n.CurrentTask.OnUpdate()
	} else {
		n.SetTask(task)
	}
}
