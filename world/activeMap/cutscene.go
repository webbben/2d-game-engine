package activemap

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
)

// CutsceneSession - TODO: gave up on this for now; once we need a cutscene, we can pick back up on it.
type CutsceneSession struct {
	startClose       bool
	def              defs.CutsceneDef
	currentStep      int
	currentStepStart time.Time
	done             bool
}

func (am *ActiveMap) StartCutscene(def defs.CutsceneDef) {
	def.Validate()

	sesh := CutsceneSession{
		def: def,
	}

	am.cutsceneSession = &sesh

	// start first step of cutscene
	firstStep := def.Steps[0]
	am.startCutsceneStep(firstStep)
}

func (am *ActiveMap) drawCutscene(screen *ebiten.Image) {
	if am.cutsceneSession.done {
		return
	}
	if am.cutsceneSession.startClose {
		if am.cutsceneSession.def.CloseTransition.IsDone() {
			return
		}
		am.cutsceneSession.def.CloseTransition.Draw(screen)
		return
	}
	if !am.cutsceneSession.def.OpenTransition.IsDone() {
		am.cutsceneSession.def.OpenTransition.Draw(screen)
	}
}

func (am *ActiveMap) updateCutscene() {
	if am.cutsceneSession == nil {
		panic("cutscene session was nil, but updateCutscene was called")
	}
	if am.cutsceneSession.done {
		return
	}

	if am.cutsceneSession.startClose {
		if !am.cutsceneSession.def.CloseTransition.IsDone() {
			am.cutsceneSession.def.CloseTransition.Update()
			return
		}
		am.cutsceneSession.done = true
		return
	}

	// check opening transition
	if !am.cutsceneSession.def.OpenTransition.IsDone() {
		am.cutsceneSession.def.OpenTransition.Update()
	}

	// check if currentStep is done
	currentStep := am.cutsceneSession.def.Steps[am.cutsceneSession.currentStep]
	if am.dialogSession != nil {
		// dialog still going, so assume a cutscene dialog is still in progress.
		return
	}
	if currentStep.Duration != 0 {
		if time.Since(am.cutsceneSession.currentStepStart) < currentStep.Duration {
			return
		}
	}

	// time to start the next cutscene step
	am.cutsceneSession.currentStep++
	if am.cutsceneSession.currentStep >= len(am.cutsceneSession.def.Steps) {
		// already at the end; end cutscene
		// TODO: end cutscene
		return
	}

	nextStep := am.cutsceneSession.def.Steps[am.cutsceneSession.currentStep]
	am.startCutsceneStep(nextStep)
}

func (am *ActiveMap) startCutsceneStep(step defs.CutsceneStepDef) {
	if am.cutsceneSession == nil {
		panic("cutscene sesh was nil")
	}
	if am.dialogSession != nil {
		panic("dialog session wasn't nil, but we tried to go to the next cutscene step")
	}

	for _, assignTask := range step.AssignTasks {
		pubsub.NPCAssignTask(string(assignTask.CharDefID), assignTask.TaskDef)
	}

	for _, cinemaFx := range step.Cinematic {
		logz.TODO("Cinematic FX", "need to implement.", cinemaFx)
	}

	if step.DialogDef != nil {
		// start up new dialog
		am.StartAdHocDialog(*step.DialogDef.Dialog)
	}

	am.cutsceneSession.currentStepStart = time.Now()
}
