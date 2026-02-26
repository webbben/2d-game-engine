package npc

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/internal/pubsub"
)

type StartDialogTask struct {
	TaskBase
	dialogProfileID defs.DialogProfileID
	dialogChain     *defs.DialogResponse
	started         bool
}

func (t *StartDialogTask) BackgroundAssist() {
}

type StartDialogTaskParams struct {
	ProfileID   defs.DialogProfileID // if set, dialog will launch with the given profile ID.
	DialogChain *defs.DialogResponse // if defined (and profileID is empty), then instead of launching a profile you can do an ad-hoc dialog sequence.
}

func NewStartDialogTask(params StartDialogTaskParams, owner *NPC, p defs.TaskPriority, nextTask *defs.TaskDef) *StartDialogTask {
	if params.ProfileID != "" && params.DialogChain != nil {
		logz.Panicln("NewStartDialogTask", "both profileID and dialogChain are defined; it should be only one or the other!")
	}
	if params.ProfileID == "" && params.DialogChain == nil {
		logz.Panicln("NewStartDialogTask", "both profileID and dialogChain were undefined!")
	}
	if params.DialogChain != nil {
		logz.Panicln("TODO", "implement dialog ad-hoc sequences (currently, only profile-based dialog is supported)")
	}

	t := StartDialogTask{
		TaskBase:        NewTaskBase(TaskStartDialog, "Start dialog", "Start dialog with the player", owner, p, nextTask),
		dialogProfileID: params.ProfileID,
		dialogChain:     params.DialogChain,
	}

	return &t
}

func (t *StartDialogTask) Update() {
	if t.IsDone() {
		return
	}

	if !t.started {
		if t.dialogProfileID != "" {
			t.Owner.World.StartDialog(t.dialogProfileID, t.Owner.ID)
		} else {
			// TODO: create way to run ad-hoc dialog sequences
			logz.Panicln("TODO", "implement dialog ad-hoc sequences (currently, only profile-based dialog is supported)")
		}
		t.started = true
		t.Status = TaskInProg
		// TODO: might need an unsubscribe function, or else we need to move event handling to happen outside the task level (duplicate subscribers could happen otherwise)
		t.Owner.eventBus.Subscribe(fmt.Sprintf("%s_%s", t.Owner.ID, t.ID), pubsub.EventDialogEnded, t.OnDialogEnd)
		return
	}
}

func (t *StartDialogTask) OnDialogEnd(e defs.Event) {
	if e.Type == pubsub.EventDialogEnded {
		profileID, ok := e.Data["profileID"]
		if !ok {
			panic("tried to get profileID, but data didn't include the key")
		}
		if profileID == t.dialogProfileID {
			// dialog has ended, so task is done.
			t.Status = TaskEnded
		} else {
			logz.Warnln(t.Owner.ID, "dialogStartTask is listening for a dialog ended event, and one came - but it was the wrong profile ID.",
				"Unless there are multiple NPCs with this task type running, there might be a problem.")
		}
	}
}
