package npc

import (
	"math/rand"
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
)

// IdleTask is a default task where the NPC doesn't really do anything besides turn a different direction here and there, take a step forward or backwards, etc.
// Just a good default for an NPC that shouldn't be doing anything in particular, but still puts them in the map and occupies them with something.
type IdleTask struct {
	TaskBase
	NoBackgroundWork

	timer idleTimer
}

type idleTimer struct {
	lastChange     time.Time     // the time when this NPC last made a misc change
	nextChangeWait time.Duration // how long to wait until doing another misc change
}

func (t *idleTimer) setChangeTimer() {
	t.lastChange = time.Now()
	r := rand.Intn(8) + 5
	t.nextChangeWait = time.Second * time.Duration(r)
}

func (t idleTimer) timeExpired() bool {
	return time.Since(t.lastChange) >= t.nextChangeWait
}

func NewIdleTask(n *NPC, def defs.TaskDef) *IdleTask {
	if def.TaskID != TaskIdle {
		panic("def had wrong task ID")
	}
	return &IdleTask{
		TaskBase: NewTaskBase(def, "Idle", "Just str8 chillin breh", n),
	}
}

func (it *IdleTask) SetupActiveState() {
	// For the idle task, the active state should just to be placed somewhere in the map.
	var pos model.Coords
	startLocation := it.GetStartLocation()
	if startLocation != nil {
		// a specific location is defined; start as close as possible to there.
		if startLocation.TileX == nil || startLocation.TileY == nil {
			panic("no start location set")
		}
		c := model.Coords{X: *startLocation.TileX, Y: *startLocation.TileY}
		var found bool
		pos, found = it.Owner.getNearestOpenTile(c, 3, false)
		if !found {
			panic("didn't find start position for idle task")
		}
	} else {
		pos = it.Owner.ActiveMapCtx.GetValidMapPosition(*it.Owner)
	}
	it.Owner.Entity.SetPosition(pos)
	it.timer.setChangeTimer()
}

func (it IdleTask) ZzCompileCheck() {
	_ = append([]Task{}, &it)
}

func (it *IdleTask) Update() {
	if it.IsDone() {
		return
	}

	it.Status = TaskInProg
	if it.timer.timeExpired() {
		// time to do another random change
		switch rand.Intn(2) {
		case 0:
			// just change directions, but don't move
			switch rand.Intn(4) {
			case 0: // L
				it.Owner.Entity.SetDirection('L')
			case 1: // R
				it.Owner.Entity.SetDirection('R')
			case 2: // U
				it.Owner.Entity.SetDirection('U')
			case 3: // D
				it.Owner.Entity.SetDirection('D')
			}
		default:
			// take a single step in an open, adjacent position.
			// get open adjacent positions
			currentPos := it.Owner.Entity.TilePos()
			choices := []model.Coords{}
			l := currentPos.GetAdj('L')
			if !it.Owner.ActiveMapCtx.IsTileCollision(l) && !it.Owner.ActiveMapCtx.IsTileEntityCollision(l, it.Owner.ID()) {
				choices = append(choices, l)
			}
			r := currentPos.GetAdj('R')
			if !it.Owner.ActiveMapCtx.IsTileCollision(r) && !it.Owner.ActiveMapCtx.IsTileEntityCollision(r, it.Owner.ID()) {
				choices = append(choices, r)
			}
			u := currentPos.GetAdj('U')
			if !it.Owner.ActiveMapCtx.IsTileCollision(u) && !it.Owner.ActiveMapCtx.IsTileEntityCollision(u, it.Owner.ID()) {
				choices = append(choices, u)
			}
			d := currentPos.GetAdj('D')
			if !it.Owner.ActiveMapCtx.IsTileCollision(d) && !it.Owner.ActiveMapCtx.IsTileEntityCollision(d, it.Owner.ID()) {
				choices = append(choices, d)
			}
			if len(choices) == 0 {
				logz.Warnln("IdleTask", "NPC appears to be blocked in! Idle task tried to find an adjacent tile, but all were collisions. npcID:", it.Owner.ID(), "pos:", currentPos)
			} else {
				target := choices[rand.Intn(len(choices))]
				it.Owner.Entity.GoToPos(target, true)
			}
		}

		it.timer.setChangeTimer()
	}
}
