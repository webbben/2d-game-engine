package quest

import "github.com/webbben/2d-game-engine/data/defs"

type AssignTaskAction struct {
	CharDefID defs.CharacterDefID
	TaskDef   defs.TaskDef
}

func (a AssignTaskAction) Fire(ctx defs.QuestActionContext) {
	ctx.AssignTaskToNPC(a.CharDefID, a.TaskDef)
}
