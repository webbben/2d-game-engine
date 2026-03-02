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
