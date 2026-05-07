package dialogv2

import (
	"github.com/webbben/2d-game-engine/data/defs"
)

// SetDialogMemoryEffect sets a specific dialog memory key.
// This should be used sparingly, only when other effects or mechanisms won't accomplish what you need to do.
type SetDialogMemoryEffect struct {
	MemoryKey string
}

func (e SetDialogMemoryEffect) Apply(ctx defs.DialogEffectContext) {
	ctx.RecordMiscDialogMemory(e.MemoryKey)
}
