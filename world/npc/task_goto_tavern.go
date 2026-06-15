package npc

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/utils"
)

type GotoTavernTask struct {
	TaskBase

	// this task is basically just a wrapper around lounge task that just finds the start location first
	loungeTask *LoungeTask
}

func (t *GotoTavernTask) ZzInterfaceCheck() {
	_ = append([]Task{}, t)
}

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

func (t *GotoTavernTask) Update() {
	if t.loungeTask == nil {
		if t.BgAssistUnplugged() {
			logz.Panicln("GotoTavernTask", "BackgroundAssist is unplugged!")
		}
		// waiting for background assist
		return
	}

	t.loungeTask.Update()
}

func (t *GotoTavernTask) SimulationUpdate() {
	if t.loungeTask == nil {
		t.findTavern()
	}
	utils.PanicAssert(t.loungeTask != nil, "lounge task was nil")

	t.loungeTask.SimulationUpdate()
}

func (t *GotoTavernTask) BackgroundAssist() {
	if t.loungeTask == nil {
		t.findTavern()
	}
	utils.PanicAssert(t.loungeTask != nil, "lounge task was nil")

	t.loungeTask.BackgroundAssist()
}

func (t *GotoTavernTask) findTavern() {
	var targetMapID defs.MapID
	// find nearest tavern map

	fromID := t.Owner.CharacterStateRef.CurrentMap
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
		targetMapID = fromID
	} else {
		var found bool
		targetMapID, found = t.Owner.WorldCtx.FindClosestMapType(fromID, defs.MapTypeTavern)
		if !found {
			logz.Panicln("GotoTavernTask", "failed to find tavern map. fromID:", fromID)
		}
	}

	utils.PanicAssert(targetMapID != "", "target map was empty")

	if t.loungeTask == nil {
		t.loungeTask = NewLoungeTask(t.Owner, defs.TaskDef{
			TaskID:   TaskLounge,
			Priority: t.Def.Priority,
			StartLocation: &defs.TaskStartLocation{
				MapID: targetMapID,
			},
		})
	}
}

func (t *GotoTavernTask) SetupActiveState() {
	if t.loungeTask == nil {
		t.findTavern()
	}

	utils.PanicAssert(t.loungeTask != nil, "loung task was nil")

	t.loungeTask.SetupActiveState()
}
