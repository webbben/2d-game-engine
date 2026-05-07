package world

import (
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/logz"
)

// TODO: should this be consolidated into WorldEffectContext?

func (w *World) TimeLapse(newTime clock.GameTime) {
	if w.AwaitingTimeLapse {
		panic("AwaitingTimeLapse was already set; either the last time lapse didn't clear this, or we accidentally called a time lapse before the last one could trigger")
	}
	if w.TimeLapseTo != nil {
		panic("TimeLapseTo wasn't nil; either the last time lapse didn't clear this, or we accidentally called a time lapse before the last one could be triggered")
	}
	// signal a pause to simulation loop
	w.SimPaused.Store(true)
	// set flag for update loop to know that we are awaiting a time lapse.
	// once sim pause has been effected, time lapse will be carried out.
	w.AwaitingTimeLapse = true
	w.TimeLapseTo = &newTime
	logz.Println("WORLD", "Prepared TimeLapse:", newTime)
}
