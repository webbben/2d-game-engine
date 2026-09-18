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

// SetTimedMemoryEffect records a dialog memory key that expires after the given number of in-game hours.
type SetTimedMemoryEffect struct {
	MemoryKey string
	Hours     int // in-game hours until the memory expires
}

func (e SetTimedMemoryEffect) Apply(ctx defs.DialogEffectContext) {
	ctx.RecordTimedMemory(e.MemoryKey, e.Hours)
}
