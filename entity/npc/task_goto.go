package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

type GotoTask struct {
	TaskBase

	goalPos   model.Coords
	isGoingTo bool

	restart bool
}

// Start hook for a GoTo task. Just sets the entity on course for the goal position.
func (t *GotoTask) Start() {
	if t.goalPos.Equals(t.Owner.Entity.TilePos) {
		panic("goto task: tried to go to the position the NPC is already at.")
	}
	actualGoal, me := t.Owner.Entity.GoToPos(t.goalPos, true)
	if !me.Success {
		logz.Println(t.Owner.DisplayName, "goto task: failed to call GoToPos:", me)
		t.Owner.Wait(time.Second)
		t.restart = true
		return
	}
	// since the goal position could've been changed (due to path being blocked), update it here
	t.goalPos = actualGoal
}

func (t *GotoTask) Update() {
	t.HandleNPCCollision(t.Start)
}

func (t GotoTask) IsComplete() bool {
	return t.Owner.Entity.TilePos.Equals(t.goalPos)
}
