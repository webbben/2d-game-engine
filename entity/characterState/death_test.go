package characterstate

import (
	"testing"

	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
)

// IsDeadOrDying exists so combat can ask "can this character still act" from a character ID, without
// needing the entity. The cases worth pinning are the two that a Dead-only check gets wrong: health
// already spent but the corpse not yet processed, and the flag set ahead of any other change.
func TestIsDeadOrDying(t *testing.T) {
	tests := []struct {
		name      string
		charState state.CharacterState
		want      bool
	}{
		{
			name:      "alive",
			charState: state.CharacterState{Health: 10},
			want:      false,
		},
		{
			name:      "one health is still alive",
			charState: state.CharacterState{Health: 1},
			want:      false,
		},
		{
			name:      "zero health is dying, even though Dead is not set yet",
			charState: state.CharacterState{Health: 0},
			want:      true,
		},
		{
			name:      "negative health is dying",
			charState: state.CharacterState{Health: -3},
			want:      true,
		},
		{
			name:      "dead flag on its own is enough, with no health set",
			charState: state.CharacterState{Dead: true},
			want:      true,
		},
		{
			name:      "dead flag wins even with health still showing",
			charState: state.CharacterState{Health: 10, Dead: true},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataman := datamanager.NewDataManager()
			charStateID := id.CharacterStateID("test_char_state")
			charState := tt.charState
			charState.ID = charStateID
			dataman.CharacterStates[charStateID] = &charState

			got := IsDeadOrDying(charStateID, dataman)
			if got != tt.want {
				t.Errorf("IsDeadOrDying() = %v, want %v (dead=%v, health=%d)",
					got, tt.want, charState.Dead, charState.Health)
			}
		})
	}
}

// A nil dataman is a wiring bug, not a runtime condition -- every caller in the engine has one. Match
// how the sibling helpers in this package handle it.
func TestIsDeadOrDyingNilDatamanPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected IsDeadOrDying to panic on a nil dataman, but it returned")
		}
	}()

	IsDeadOrDying(id.CharacterStateID("test_char_state"), nil)
}
