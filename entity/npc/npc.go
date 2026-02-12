// Package npc defines NPC management logic
package npc

import (
	"errors"
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

// TODO:
// - add NPC defs to definition manager

type NPCDef struct{}

type NPC struct {
	debug debug
	NPCInfo
	Entity *entity.Entity

	dialogProfileID defs.DialogProfileID // retrieved from character state, but set here for convenience

	TaskMGMT
	OnUpdateFn func(n *NPC)
	World      WorldContext
}

func (n *NPC) Activate() {
	if n.dialogProfileID != "" {
		n.World.StartDialog(n.dialogProfileID, n.ID)
		return
	}
	logz.Println(n.DisplayName, "nothing happened on activation")
}

// Y is used for renderables sorting
func (n NPC) Y() float64 {
	return n.Entity.Y
}

type WorldContext interface {
	FindNPCAtPosition(c model.Coords) (NPC, bool)
	StartDialog(dialogProfileID defs.DialogProfileID, npcID string)
}

type NPCParams struct {
	CharStateID state.CharacterStateID
}

func NewNPC(params NPCParams, defMgr *definitions.DefinitionManager, audioMgr *audio.AudioManager) NPC {
	if params.CharStateID == "" {
		panic("CharStateID is empty")
	}

	ent := entity.LoadCharacterStateIntoEntity(params.CharStateID, defMgr, audioMgr)

	// get dialog profile ID from character def
	charDef := defMgr.GetCharacterDef(ent.CharacterStateRef.DefID)

	return NPC{
		Entity: ent,
		NPCInfo: NPCInfo{
			ID:          string(ent.ID()),
			DisplayName: ent.DisplayName(),
		},
		dialogProfileID: charDef.DialogProfileID,
	}
}

type NPCInfo struct {
	ID          string
	DisplayName string
	Priority    int // priority assigned to this NPC by the map it is added to
}

type TaskMGMT struct {
	CurrentTask         Task
	TaskQueue           []*Task // TODO queue of tasks to run one after the other. not implemented yet.
	waitUntil           time.Time
	waitUntilDoneMoving bool // if set, will wait until entity has stopped moving before processing next update
	// A default task that will run whenever no other task is active.
	// Useful for if this NPC should always just continuously do one task.
	DefaultTask Task
	// number of ticks this NPC has been stuck (failing to move to its goal).
	// TODO implement this if needed. so far, haven't needed to report stuck NPCs.
	StuckCount int
}

// IsActive checks if the npc is currently working on a task
func (tm TaskMGMT) IsActive() bool {
	if tm.CurrentTask == nil {
		return false
	}
	return !tm.CurrentTask.IsDone()
}

// ErrAlreadyActive indicates the NPC is already active with a task
var ErrAlreadyActive error = errors.New("NPC already has an active task")

// SetTask sets a task for this NPC to carry out.
// For setting fundamental tasks, use the respective Set<taskType>Task command.
//
// Directly using this task outside of the game engine would be for setting customly defined tasks.
func (n *NPC) SetTask(t Task, force bool) error {
	if t == nil {
		panic("SetTask: task is nil")
	}
	if n.IsActive() {
		if force {
			// TODO force quit current task
		} else {
			logz.Warnln(n.DisplayName, "tried to set task on already active NPC")
			return ErrAlreadyActive
		}
	}

	t.SetOwner(n)
	n.CurrentTask = t
	return nil
}

// EndCurrentTask ends the current task. Causes the task to run its "end" hook logic.
func (n *NPC) EndCurrentTask() {
	if n.CurrentTask == nil {
		logz.Warnln(n.DisplayName, "tried to cancel current task, but no current task exists.")
		return
	}
	n.CurrentTask.End()
}

// Wait interrupts regular NPC updates for a certain duration
func (n *NPC) Wait(d time.Duration) {
	n.waitUntil = time.Now().Add(d)
}

func (n *NPC) WaitUntilNotMoving() {
	// remove any existing wait
	if n.waitUntil.After(time.Now()) {
		n.waitUntil = time.Now()
	}
	n.waitUntilDoneMoving = true
}
