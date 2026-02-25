package quest

import "github.com/webbben/2d-game-engine/data/defs"

type AssignTaskAction struct {
	CharDefID defs.CharacterDefID
	TaskDef   defs.TaskDef
}

func (a AssignTaskAction) Fire(ctx defs.QuestActionContext) {
	ctx.AssignTaskToNPC(a.CharDefID, a.TaskDef)
}

type QueueScenarioAction struct {
	ScenarioID defs.ScenarioID
}

func (a QueueScenarioAction) Fire(ctx defs.QuestActionContext) {
	ctx.QueueScenario(a.ScenarioID)
}

type UnlockAction struct {
	MapID  defs.MapID
	LockID string
}

func (a UnlockAction) Fire(ctx defs.QuestActionContext) {
	ctx.UnlockMapLock(a.MapID, a.LockID)
}
