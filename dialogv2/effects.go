package dialogv2

import (
	"github.com/webbben/2d-game-engine/data/defs"
)

func ZzEffectInterfaceCheck() {
	_ = append([]defs.DialogEffect{}, &EventEffect{}, &AddGoldEffect{}, &SetDialogMemoryEffect{})
}

type EventEffect struct {
	Event defs.Event
}

func (e EventEffect) Apply(ctx defs.EffectContext) {
	ctx.BroadcastEvent(e.Event)
}

type AddGoldEffect struct {
	Amount int
}

func (e AddGoldEffect) Apply(ctx defs.EffectContext) {
	ctx.AddGold(e.Amount)
}

type SetDialogMemoryEffect struct {
	MemoryKey string
}

func (e SetDialogMemoryEffect) Apply(ctx defs.EffectContext) {
}

// StartCustomLoadScreenEffect causes a load screen (transition + execute load script) to occur in a dialog.
// Using this effect will cause dialog to end.
type StartCustomLoadScreenEffect struct {
	TransitionScreenID              defs.ScreenID
	OpenTransition, CloseTransition defs.Transition
	LoadFunction                    func(ctx defs.GameContext)
}

func (e StartCustomLoadScreenEffect) Apply(ctx defs.EffectContext) {
	ctx.StartCustomLoadScreen(e.TransitionScreenID, e.OpenTransition, e.CloseTransition, e.LoadFunction)
}

type StartLoadScreenEffect struct {
	LoadFunction func(ctx defs.GameContext)
}

func (e StartLoadScreenEffect) Apply(ctx defs.EffectContext) {
	ctx.StartLoadScreen(e.LoadFunction)
}
