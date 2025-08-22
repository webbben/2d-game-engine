package npc

import (
	"fmt"
	"time"

	"github.com/webbben/2d-game-engine/internal/logz"
)

// Start hook for a GoTo task. Just sets the entity on course for the goal position.
func (t *Task) startGoto() {
	if t.GotoTask.GoalPos.Equals(t.Owner.Entity.TilePos) {
		panic("goto task: tried to go to the position the NPC is already at.")
	}
	me := t.Owner.Entity.GoToPos(t.GotoTask.GoalPos)
	if !me.Success {
		logz.Println(t.Owner.DisplayName, "goto task: failed to call GoToPos:", me)
		t.Owner.Wait(time.Second)
		t.Restart = true
	}
}

func (t *Task) updateGoto() {
	if t.Owner.Entity.Movement.Interrupted {
		// path entity was moving on has been interrupted.
		// if interrupted by NPC, try to negotiate resolution to collision.
		if !t.Owner.Entity.Movement.TargetTile.Equals(t.Owner.Entity.TilePos) {
			panic("Goto task: since NPC movement was interrupted, we expect its target position to be the same as its current position")
		}
		// get NPCs that are at the next target tile - i.e. the next position in the target path
		if len(t.Owner.Entity.Movement.TargetPath) == 0 {
			panic("Goto task: npc movement was interrupted, but there is no next step in target path")
		}
		nextTarget := t.Owner.Entity.Movement.TargetPath[0]
		collisionNpc, found := t.Owner.World.FindNPCAtPosition(nextTarget)
		if collisionNpc.ID == t.Owner.ID {
			panic("Goto task: npc collided with itself?! that shouldn't be happening")
		}
		if !found {
			// TODO check if NPC collided with player. NPC doesn't collide with player yet.

			// if NPC didn't collide with player, and no collision NPC found, then it seems the blocking NPC is now gone, so proceed
			t.startGoto() // re-route now that the coast is clear
			return
		}
		if collisionNpc.Priority > t.Owner.Priority {
			// other NPC is higher priority; let him go first.
			logz.Println(t.Owner.DisplayName, "waiting for other NPC to go first")
			t.Owner.Wait(time.Second) // wait a second before we check logic again for this NPC
			return
		} else {
			if collisionNpc.Priority == t.Owner.Priority {
				fmt.Println(collisionNpc.DisplayName, collisionNpc.Priority, " / ", t.Owner.DisplayName, t.Owner.Priority)
				panic("updateGoto: other NPC has same priority as this one! that's not suppposed to happen.")
			}
			// this NPC has higher priority, so it can re-route first
			t.startGoto()
			return
		}
		// TODO add logic for reporting stuck NPC, if this stuck state goes on for too long.
	}
}

func (t Task) isCompleteGoto() bool {
	return t.Owner.Entity.TilePos.Equals(t.GotoTask.GoalPos)
}
