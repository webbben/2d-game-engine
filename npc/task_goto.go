package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/internal/logz"
)

// Start hook for a GoTo task. Just sets the entity on course for the goal position.
func (t *Task) startGoto() {
	if t.GotoTask.GoalPos.Equals(t.Owner.Entity.TilePos) {
		panic("goto task: tried to go to the position the NPC is already at.")
	}
	actualGoal, me := t.Owner.Entity.GoToPos(t.GotoTask.GoalPos, true)
	if !me.Success {
		logz.Println(t.Owner.DisplayName, "goto task: failed to call GoToPos:", me)
		t.Owner.Wait(time.Second)
		t.Restart = true
	}
	// since the goal position could've been changed (due to path being blocked), update it here
	t.GotoTask.GoalPos = actualGoal
}

func (t *Task) updateGoto() {
	t.handleNPCCollision(t.startGoto)
}

func (t Task) isCompleteGoto() bool {
	return t.Owner.Entity.TilePos.Equals(t.GotoTask.GoalPos)
}
