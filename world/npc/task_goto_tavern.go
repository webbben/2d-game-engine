package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/utils"
)

type GotoTavernTask struct {
	TaskBase
	// this task is basically just a wrapper around lounge task that just finds the start location first;
	// the lounge task runs as this task's child once the tavern is found.
}

var _ Task = (*GotoTavernTask)(nil)

func NewGotoTavernTask(n *NPC, def defs.TaskDef) *GotoTavernTask {
	return &GotoTavernTask{
		TaskBase: NewTaskBase(
			def,
			"Go to Tavern",
			"NPC finds a nearby tavern and goes there to lounge",
			n,
		),
	}
}

func init() {
	registerTask(TaskGoToTavern, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			return NewGotoTavernTask(owner, def)
		},
	})
}

func (t *GotoTavernTask) Update() {
	if !t.HasChild() {
		// waiting for background assist to find the tavern
		return
	}
	t.TaskBase.Update()
}

func (t *GotoTavernTask) SimulationUpdate() {
	if !t.HasChild() {
		t.findTavern()
	}
	t.TaskBase.SimulationUpdate()
}

func (t *GotoTavernTask) BackgroundAssist() {
	if !t.HasChild() {
		t.findTavern()
	}
	t.TaskBase.BackgroundAssist()
}

func (t *GotoTavernTask) findTavern() {
	fromID := t.Owner.CharacterStateRef.CurrentMap
	targetMapID := t.resolveTavernFrom(fromID)

	if !t.HasChild() {
		t.RunChild(NewLoungeTask(t.Owner, defs.TaskDef{
			TaskID:   TaskLounge,
			Priority: t.Def.Priority,
			StartLocation: &defs.TaskStartLocation{
				MapID: targetMapID,
			},
		}))
	}
}

// resolveTavernFrom finds the nearest tavern map to the given map, or stays in place if the map already is a tavern.
// Shared by runtime (findTavern, anchored on the NPC's current map) and placement (ResolveStartMap, anchored on the
// schedule) so both always agree on which tavern an NPC should be at.
func (t *GotoTavernTask) resolveTavernFrom(fromID defs.MapID) defs.MapID {
	// get mapDefID from mapState if this is a generated map
	mapState := t.Owner.dataman.GetMapState(fromID)
	mapDefID := fromID
	if mapState.IsGenerated {
		mapDefID = mapState.GeneratedMapDefID
	}
	fromMapDef := t.Owner.dataman.GetMapDef(mapDefID)

	if fromMapDef.Type == defs.MapTypeTavern {
		// already in a tavern! let's just lounge here then.
		logz.Warnln("GotoTavernTask", "NPC is already in a tavern... is this expected?", t.Owner.WhoAmI())
		return fromID
	}

	targetMapID, found := t.Owner.WorldCtx.FindClosestMapType(fromID, defs.MapTypeTavern)
	if !found {
		logz.Println("GotoTavernTask", fromID)
		logz.Panicln("GotoTavernTask", "failed to find tavern map")
	}
	utils.PanicAssert(targetMapID != "", "target map was empty")
	return targetMapID
}

// ResolveStartMap answers the scheduler's placement question: which map does this task put the NPC in?
// GO_TO_TAVERN has no declared start location, so it resolves to the nearest tavern to the anchor.
func (t *GotoTavernTask) ResolveStartMap(anchor defs.MapID) defs.MapID {
	return t.resolveTavernFrom(anchor)
}

func (t *GotoTavernTask) SetupActiveState() {
	if !t.HasChild() {
		t.findTavern()
	}
	t.TaskBase.SetupActiveState()
}
