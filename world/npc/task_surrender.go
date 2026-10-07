package npc

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/utils"
)

type SurrenderTask struct {
	TaskBase

	targetEntity *entity.Entity // who we are surrendering to; who we run from if this falls through
	subID        string
}

type SurrenderTaskParams struct {
	TargetEntity *entity.Entity
}

func init() {
	registerTask(TaskSurrender, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			params, ok := def.Params.(SurrenderTaskParams)
			if !ok {
				logz.PanicCtx("SurrenderTask", "tried to run a surrender task, but the params could not be converted into SurrenderTaskParams", def.Params)
			}
			return NewSurrenderTask(params.TargetEntity, owner, def)
		},
		validateParams: func(def defs.TaskDef) error {
			params, ok := def.Params.(SurrenderTaskParams)
			if !ok {
				return fmt.Errorf("SurrenderTask params must be SurrenderTaskParams, got %T", def.Params)
			}
			if params.TargetEntity == nil {
				return fmt.Errorf("SurrenderTask params had a nil TargetEntity")
			}
			return nil
		},
	})
}

func NewSurrenderTask(targetEnt *entity.Entity, owner *NPC, def defs.TaskDef) *SurrenderTask {
	utils.PanicAssert(def.TaskID == TaskSurrender, "task def has wrong ID")
	utils.PanicAssert(targetEnt != nil, "target was nil")
	return &SurrenderTask{
		TaskBase:     NewTaskBase(def, "Surrender", "Offer to surrender to an enemy", owner),
		targetEntity: targetEnt,
	}
}

// Start announces the offer and starts listening for the dialog to close.
//
// How acceptance and rejection are told apart: accepting is handled by AcceptSurrenderEffect, which
// ends this task before the dialog ever closes. So reaching OnDialogEnd at all means the offer was
// refused or the player just walked away -- there is no outcome payload on the event to check.
func (t *SurrenderTask) Start() {
	t.Owner.Entity.ShowSpeechBubble("I yield! I have had enough!", t.Owner.defaultSpeechBubbleParams())
	t.subID = t.Owner.ID() + "_" + string(TaskSurrender) + "_sub"
	t.Owner.subscribeToEvent(t.subID, pubsub.EventDialogEnded, t.OnDialogEnd)
	t.TaskBase.Start()
}

func (t *SurrenderTask) Update() {
	if t.IsDone() {
		return
	}
	// Deliberately empty. We stand still and wait while the offer is on the table. Accepting preempts
	// this task, so Update only runs while the offer is still pending.
}

func (t *SurrenderTask) OnDialogEnd(e defs.Event) {
	if e.Type != pubsub.EventDialogEnded {
		return
	}

	// Only react to our own dialog closing. The event carries a profile id, but profiles are shared
	// between npcs (legionary01 and legionary02 both use Q001_misc_guard), so a profile check alone
	// would let an unrelated npc's dialog dismissal count as our offer being turned down.
	npcID, ok := e.Data["npcID"].(string)
	if !ok {
		logz.PanicCtx("SurrenderTask", "dialog ended event had no usable npcID key", e.Data, t.Owner.WhoAmI())
	}
	if npcID != t.Owner.ID() {
		return
	}

	// Only act if we're still the npc's task. A task can be dropped without Finish ever running (the
	// hourly schedule does exactly this), which leaves this subscription registered; without this check a
	// later, unrelated conversation with the same npc would send them running.
	if t.Owner.getCurrentTask() != t {
		return
	}

	// The offer is spent either way. Never nag, even if this was closed without a reply chosen.
	t.Owner.surrenderOfferSpent = true

	// Back to running, which is where we'd be if we'd never offered.
	t.Owner.RunTask(defs.TaskDef{
		TaskID:   TaskFlee,
		Priority: Emergency,
		Params:   FleeTaskParams{TargetEntity: t.targetEntity},
	}, t.Owner)
	// RunTask records this task as the interrupted one. We don't want to be resumed if preempted later;
	// the npc has moved on to running. (Must come after RunTask, which is where the recording happens.)
	t.Owner.clearInterruptedTask()
}

// Finish removes the dialog subscription, so a preempted or abandoned surrender task doesn't react to
// some later, unrelated dialog closing.
func (t *SurrenderTask) Finish(result TaskResult) {
	t.unsubscribe()
	t.TaskBase.Finish(result)
}

func (t *SurrenderTask) unsubscribe() {
	if t.subID == "" {
		return
	}
	// Only unsubscribe if we're still registered. Leaving the active map unsubscribes everything
	// (NPC.unsubscribeAll), and a task can outlive the map it was started on, so by the time we're
	// finished we may already be gone from the bus -- which panics if asked to remove us again.
	if t.Owner.activeMapSubscriptionIDs[t.subID] {
		t.Owner.unsubscribe(t.subID)
	}
	t.subID = ""
}

// SetupActiveState: a surrendered npc has nothing to set up when the player returns to the map. The
// offer is decided in the dialog, not by walking around.
func (t *SurrenderTask) SetupActiveState() {}
