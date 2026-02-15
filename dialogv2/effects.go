package dialogv2

import (
	"github.com/webbben/2d-game-engine/data/defs"
)

type EventEffect struct {
	Event defs.Event
}

func (e EventEffect) Apply(ctx defs.EffectContext) {
	ctx.BroadcastEvent(e.Event)
}

func ZzEffectInterfaceCheck() {
	_ = append([]defs.Effect{}, &EventEffect{})
}
