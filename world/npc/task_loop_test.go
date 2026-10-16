package npc

import (
	"strings"
	"testing"

	"github.com/webbben/2d-game-engine/data/defs"
)

func gotoDef(x, y int, next *defs.TaskDef) defs.TaskDef {
	return defs.TaskDef{
		TaskID:   TaskGoto,
		Priority: Assign,
		Params:   GotoTaskParams{TileX: x, TileY: y},
		NextTask: next,
	}
}

func TestFlattenTaskChain(t *testing.T) {
	t.Run("single task is its own chain", func(t *testing.T) {
		chain := flattenTaskChain(gotoDef(10, 6, nil))
		if len(chain) != 1 {
			t.Fatalf("expected 1 task in chain, got %d", len(chain))
		}
		if got := chain[0].Params.(GotoTaskParams); got.TileX != 10 || got.TileY != 6 {
			t.Errorf("chain lost the inner task's params: got %+v", got)
		}
	})

	t.Run("follows NextTask in order", func(t *testing.T) {
		// The back-and-forth shape traffic uses: go right, then go left.
		chain := flattenTaskChain(gotoDef(27, 6, &defs.TaskDef{
			TaskID:   TaskGoto,
			Priority: Assign,
			Params:   GotoTaskParams{TileX: 10, TileY: 6},
		}))

		if len(chain) != 2 {
			t.Fatalf("expected 2 tasks in chain, got %d", len(chain))
		}
		first := chain[0].Params.(GotoTaskParams)
		second := chain[1].Params.(GotoTaskParams)
		if first.TileX != 27 {
			t.Errorf("chain[0] should be the far end (x=27), got x=%d", first.TileX)
		}
		if second.TileX != 10 {
			t.Errorf("chain[1] should be the near end (x=10), got x=%d", second.TileX)
		}
	})
}

// A cycle can't be built through normal data (a chain is nested values, so it has to end), but the
// inner def reaches here as runtime input, and walking one forever would hang the main loop.
func TestFlattenTaskChainRejectsCycle(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a cyclic chain to panic, but flattenTaskChain returned")
		}
	}()

	first := defs.TaskDef{TaskID: TaskGoto, Params: GotoTaskParams{TileX: 1, TileY: 1}}
	second := defs.TaskDef{TaskID: TaskGoto, Params: GotoTaskParams{TileX: 2, TileY: 2}}
	first.NextTask = &second
	second.NextTask = &first

	flattenTaskChain(first)
}

func TestValidateLoopParams(t *testing.T) {
	tests := []struct {
		name    string
		params  any    // any, not LoopTaskParams, so the wrong-type case is expressible
		wantErr string // empty means valid
	}{
		{
			name:   "single goto is valid",
			params: LoopTaskParams{Inner: gotoDef(10, 6, nil)},
		},
		{
			name: "chained gotos are valid",
			params: LoopTaskParams{Inner: gotoDef(27, 6, &defs.TaskDef{
				TaskID:   TaskGoto,
				Priority: Assign,
				Params:   GotoTaskParams{TileX: 10, TileY: 6},
			})},
		},
		{
			name:    "params of the wrong type",
			params:  defs.TaskDef{TaskID: TaskGoto, Params: GotoTaskParams{TileX: 1, TileY: 1}},
			wantErr: "LOOP params must be LoopTaskParams",
		},
		{
			name:    "empty inner task ID",
			params:  LoopTaskParams{Inner: defs.TaskDef{}},
			wantErr: "empty inner task ID",
		},
		{
			name: "cannot wrap another loop",
			params: LoopTaskParams{Inner: defs.TaskDef{
				TaskID: TaskLoop,
				Params: LoopTaskParams{Inner: gotoDef(1, 1, nil)},
			}},
			wantErr: "cannot wrap another LOOP",
		},
		{
			name:    "child-only task cannot be wrapped directly",
			params:  LoopTaskParams{Inner: defs.TaskDef{TaskID: TaskFollow}},
			wantErr: "not a registered top-level task",
		},
		{
			name: "StartLocation would re-route every iteration",
			params: LoopTaskParams{Inner: defs.TaskDef{
				TaskID:        TaskGoto,
				Params:        GotoTaskParams{TileX: 1, TileY: 1},
				StartLocation: &defs.TaskStartLocation{MapID: "some_map"},
			}},
			wantErr: "cannot route to a start map on every iteration",
		},
		{
			name: "inner params of the wrong type is caught by the recursive check",
			params: LoopTaskParams{Inner: defs.TaskDef{
				TaskID: TaskGoto,
				Params: LoopTaskParams{Inner: gotoDef(1, 1, nil)},
			}},
			wantErr: "GotoTask params must be GotoTaskParams",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLoopParams(defs.TaskDef{
				TaskID:   TaskLoop,
				Priority: Assign,
				Params:   tt.params,
			})

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected valid, got error: %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("expected an error containing %q, got nil", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}

// ValidateTaskDef is the entry point data tooling uses, and LOOP reaches it through validateParams.
// Check the whole def validates, not just the inner task, so a malformed LOOP is caught by the same
// CI pass that catches every other task.
func TestValidateTaskDefAcceptsLoop(t *testing.T) {
	def := defs.TaskDef{
		TaskID:   TaskLoop,
		Priority: Assign,
		Params: LoopTaskParams{Inner: gotoDef(27, 6, &defs.TaskDef{
			TaskID:   TaskGoto,
			Priority: Assign,
			Params:   GotoTaskParams{TileX: 10, TileY: 6},
		})},
	}
	if err := ValidateTaskDef(def); err != nil {
		t.Errorf("expected a well-formed LOOP def to validate, got: %v", err)
	}
}
