package npc

import (
	"errors"
	"time"

	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

type NPC struct {
	NPCInfo
	Entity *entity.Entity

	TaskMGMT
	OnUpdateFn func(n *NPC)
	World      WorldContext

	DialogID string
}

func (n *NPC) Activate() {
	if n.DialogID != "" {
		n.World.StartDialog(n.DialogID)
		return
	}
	logz.Println(n.DisplayName, "nothing happened on activation")
}

// used for renderables sorting
func (n NPC) Y() float64 {
	return n.Entity.Y
}

type WorldContext interface {
	FindNPCAtPosition(c model.Coords) (NPC, bool)
	StartDialog(dialogID string)
}

// Create new NPC from the given NPC struct. Ensures essential data is set.
func New(n NPC) NPC {
	if n.ID == "" {
		n.ID = general_util.GenerateUUID()
	}
	return n
}

type NPCInfo struct {
	ID          string
	DisplayName string
	Priority    int // priority assigned to this NPC by the map it is added to
}

type TaskMGMT struct {
	Active              bool // if the NPC is actively doing a task right now
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

// Error indicates the NPC is already active with a task
var ErrAlreadyActive error = errors.New("NPC already has an active task")

// Sets a task for this NPC to carry out.
// For setting fundamental tasks, use the respective Set<taskType>Task command.
//
// Directly using this task outside of the game engine would be for setting customly defined tasks.
func (n *NPC) SetTask(t Task) error {
	if t == nil {
		panic("SetTask: task is nil")
	}
	if n.Active {
		logz.Warnln(n.DisplayName, "tried to set task on already active NPC")
		return ErrAlreadyActive
	}

	t.SetOwner(n)
	n.CurrentTask = t
	n.Active = true
	return nil
}

// Ends the current task. Causes the task to run its "end" hook logic.
func (n *NPC) EndCurrentTask() {
	if n.CurrentTask == nil {
		logz.Warnln(n.DisplayName, "tried to cancel current task, but no current task exists.")
		return
	}
	n.CurrentTask.Cleanup()
}

// Interrupt regular NPC updates for a certain duration
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
