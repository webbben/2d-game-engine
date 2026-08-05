package npc

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
)

type StartDialogTask struct {
	TaskBase
	dialogProfileID defs.DialogProfileID
	dialogChain     *defs.DialogResponse
	started         bool
	subID           string
}

type StartDialogTaskParams struct {
	ProfileID   defs.DialogProfileID // if set, dialog will launch with the given profile ID.
	DialogChain *defs.DialogResponse // if defined (and profileID is empty), then instead of launching a profile you can do an ad-hoc dialog sequence.
}

func NewStartDialogTask(params StartDialogTaskParams, owner *NPC, def defs.TaskDef) *StartDialogTask {
	if params.ProfileID != "" && params.DialogChain != nil {
		logz.Panicln("NewStartDialogTask", "both profileID and dialogChain are defined; it should be only one or the other!")
	}
	if params.ProfileID == "" && params.DialogChain == nil {
		logz.Panicln("NewStartDialogTask", "both profileID and dialogChain were undefined!")
	}
	if params.DialogChain != nil {
		logz.Panicln("TODO", "implement dialog ad-hoc sequences (currently, only profile-based dialog is supported)")
	}
	if def.TaskID != TaskStartDialog {
		panic("task def has wrong ID")
	}

	return &StartDialogTask{
		TaskBase:        NewTaskBase(def, "Start dialog", "Start dialog with the player", owner),
		dialogProfileID: params.ProfileID,
		dialogChain:     params.DialogChain,
	}
}

func init() {
	registerTask(TaskStartDialog, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			params, ok := def.Params.(StartDialogTaskParams)
			if !ok {
				logz.Println("StartDialogTask", def.Params)
				logz.Panicln("StartDialogTask", "tried to run a start-dialog task, but the params could not be converted into StartDialogTaskParams. make sure you are using the right struct")
			}
			return NewStartDialogTask(params, owner, def)
		},
		validateParams: func(def defs.TaskDef) error {
			params, ok := def.Params.(StartDialogTaskParams)
			if !ok {
				return fmt.Errorf("StartDialogTask params must be StartDialogTaskParams, got %T", def.Params)
			}
			if params.ProfileID != "" && params.DialogChain != nil {
				return fmt.Errorf("StartDialogTask params set both ProfileID and DialogChain; it should be only one or the other")
			}
			if params.ProfileID == "" && params.DialogChain == nil {
				return fmt.Errorf("StartDialogTask params set neither ProfileID nor DialogChain; one must be set")
			}
			if params.DialogChain != nil {
				return fmt.Errorf("StartDialogTask ad-hoc DialogChain sequences are not yet supported; use a ProfileID")
			}
			return nil
		},
	})
}

// Start launches the dialog and subscribes to its end event. It's called by the task manager when this task becomes current.
func (t *StartDialogTask) Start() {
	if t.started {
		panic("StartDialogTask was started twice")
	}
	if t.dialogProfileID != "" {
		t.Owner.ActiveMapCtx.StartDialog(t.dialogProfileID, t.Owner.ID())
	} else {
		// TODO: create way to run ad-hoc dialog sequences
		logz.Panicln("TODO", "implement dialog ad-hoc sequences (currently, only profile-based dialog is supported)")
	}
	t.started = true
	t.subID = fmt.Sprintf("%s_%s", t.Owner.ID(), t.Def.TaskID)
	t.Owner.eventBus.Subscribe(t.subID, pubsub.EventDialogEnded, t.OnDialogEnd)
	t.Owner.activeMapSubscriptionIDs[t.subID] = true
	t.TaskBase.Start()
}

// Finish unsubscribes the dialog-end event, then records the result. This guarantees the subscription is cleaned up
// whether the dialog ends naturally or the task is preempted.
func (t *StartDialogTask) Finish(result TaskResult) {
	t.unsubscribe()
	t.TaskBase.Finish(result)
}

// unsubscribe removes this task's dialog-ended subscription, if any. Safe to call more than once.
func (t *StartDialogTask) unsubscribe() {
	if t.subID == "" {
		return
	}
	delete(t.Owner.activeMapSubscriptionIDs, t.subID)
	t.Owner.eventBus.Unsubscribe(t.subID)
	t.subID = ""
}

func (t *StartDialogTask) Update() {
	if t.IsDone() {
		return
	}
	// The dialog is launched in Start(); there's nothing to do on each tick until the dialog ends.
}

func (t *StartDialogTask) OnDialogEnd(e defs.Event) {
	if e.Type != pubsub.EventDialogEnded {
		return
	}
	profileID, ok := e.Data["profileID"]
	if !ok {
		panic("tried to get profileID, but data didn't include the key")
	}
	if profileID == t.dialogProfileID {
		// dialog has ended; mark the task done (Finish cleans up the subscription).
		t.FinishSuccess()
	} else {
		logz.Warnln(t.Owner.ID(), "dialogStartTask is listening for a dialog ended event, and one came - but it was the wrong profile ID.",
			"Unless there are multiple NPCs with this task type running, there might be a problem.")
	}
}

func (t *StartDialogTask) SetupActiveState() {
	panic("not yet implemented!")
}
