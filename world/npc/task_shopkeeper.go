package npc

import (
	"math/rand"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/object"
)

type ShopkeeperTask struct {
	TaskBase
	taskAreaObj     *object.Object
	reachedTaskArea bool
	gotoTask        *GotoTask
}

func (t ShopkeeperTask) ZzInterfaceCheck() {
	_ = append([]Task{}, &t)
}

func NewShopkeeperTask(n *NPC, def defs.TaskDef) *ShopkeeperTask {
	if def.TaskID != TaskShopkeeper {
		panic("def had wrong task ID")
	}
	return &ShopkeeperTask{
		TaskBase: NewTaskBase(
			def,
			"Shopkeeper",
			"NPC goes to a shopkeeper area on the map",
			n,
		),
	}
}

func (t *ShopkeeperTask) Update() {
	if t.IsDone() {
		// When would this happen anyway? I don't think we set the status to TaskEnded anywhere
		return
	}
	t.Status = TaskInProg

	if !t.RouteToStartMap(false) {
		// we are routing to a different map to start
		return
	}
	// we are in the start map (already - it should be the active map, and the character must've entered the map now or was already in it.)

	// we are in the start map - find out if we need to walk to the task area
	if t.taskAreaObj == nil {
		logz.Println("ShopkeeperTask", "finding task area")
		// find task area
		t.findTaskArea()
		if t.taskAreaObj == nil {
			logz.Panicln("ShopkeeperTask", "failed to find task area:", TaskShopkeeper, t.Owner.WhoAmI())
		}
	}
	if !t.reachedTaskArea {
		// go to task area
		if t.gotoTask == nil {
			objPos := t.taskAreaObj.TilePos()
			if objPos.Equals(t.Owner.Entity.TilePos()) {
				// well, apparently we are already at the task area position! interesting...
				t.reachedTaskArea = true
				return
			}
			if objPos.Equals(model.Coords{X: 0, Y: 0}) {
				x, y := t.taskAreaObj.Pos()
				logz.Println("ShopkeeperTask", "objID:", t.taskAreaObj.ID, "object pos:", x, y, "rect:", t.taskAreaObj.GetRect())
				logz.Panicln("ShopkeeperTask", "object position (tile position) came back as 0 0, which seems wrong.")
			}

			t.gotoTask = NewGotoTask(GotoTaskParams{TileX: objPos.X, TileY: objPos.Y}, t.Owner, defs.TaskDef{
				TaskID:   TaskGoto,
				Priority: t.GetPriority(),
			})
		}
		if !t.gotoTask.IsDone() {
			t.gotoTask.Update()
			return
		}
		// reached the task area
		t.reachedTaskArea = true
		t.Owner.Entity.SetDirection(t.taskAreaObj.TaskArea.Dir)
	}

	// at the task area; just stand there
}

func (t *ShopkeeperTask) findTaskArea() {
	for _, obj := range t.Owner.ActiveMapCtx.GetAllObjects() {
		if obj.Type == object.TypeTaskArea {
			if obj.TaskArea.TaskID == string(TaskShopkeeper) {
				if !t.Owner.SatisfiesObjectOwnership(*obj) {
					continue
				}
				// found the task area
				t.taskAreaObj = obj
				return
			}
		}
	}
	logz.Panicln("ShopkeeperTask", "failed to find task area:", TaskShopkeeper, t.Owner.WhoAmI())
}

func (t *ShopkeeperTask) SetupActiveState() {
	if !t.InStartMap() {
		// if we aren't in start map at this function call, then we should already be routing there; setupActiveState for underlying routing task.
		t.RouteToStartMapSetupActiveState()
		return
	}

	t.findTaskArea()
	if t.taskAreaObj == nil {
		panic("task area obj was nil")
	}

	// find a tile in the task area and move to it
	tiles := t.taskAreaObj.GetRect().GetOverlappingTiles()
	targetTile := tiles[rand.Intn(len(tiles))]
	t.Owner.Entity.SetPosition(targetTile)
	t.Owner.Entity.SetDirection(t.taskAreaObj.TaskArea.Dir)
	t.reachedTaskArea = true
}

func (t *ShopkeeperTask) BackgroundAssist() {
	t.RouteToStartMapBgAssist()
}

func (t *ShopkeeperTask) SimulationUpdate() {
	// Ultimately there's nothing to do for this task in the simulation loop if the NPC is already at their start map.
	// Only work to do would be to route the NPC to the start map.
	if t.InActiveMap() {
		logz.Panicln("ShopkeeperTask", "why is SimulationUpdate being called while NPC is already in active map")
	}
	if t.IsDone() {
		return
	}
	t.RouteToStartMap(true)
}
