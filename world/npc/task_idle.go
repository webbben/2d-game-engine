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

	lastChange     time.Time     // the time when this NPC last made a misc change
	nextChangeWait time.Duration // how long to wait until doing another misc change
}

func NewIdleTask(n *NPC, p defs.TaskPriority) *IdleTask {
	d := defs.TaskDef{
		TaskID:   TaskIdle,
		Priority: p,
	}
	return &IdleTask{
		TaskBase: NewTaskBase(d, "Idle", "Just str8 chillin breh", n),
	}
}

func (it *IdleTask) SetupActiveState() {
	// For the idle task, the active state should just to be placed somewhere in the map.
	var pos model.Coords
	startLocation := it.GetStartLocation()
	if startLocation != nil {
		// a specific location is defined; start as close as possible to there.
		c := model.Coords{X: startLocation.TileX, Y: startLocation.TileY}
		var found bool
		pos, found = it.Owner.getNearestOpenTile(c, 3, false)
		if !found {
			panic("didn't find start position for idle task")
		}
	} else {
		pos = it.Owner.World.GetValidMapPosition(*it.Owner)
	}
	it.Owner.Entity.SetPosition(pos)
	it.setChangeTimer()
}

func (it *IdleTask) setChangeTimer() {
	it.lastChange = time.Now()
	r := rand.Intn(8) + 5
	it.nextChangeWait = time.Second * time.Duration(r)
}

func (it IdleTask) ZzCompileCheck() {
	_ = append([]Task{}, &it)
}

func (it *IdleTask) Update() {
	if it.IsDone() {
		return
	}

	it.Status = TaskInProg
	if time.Since(it.lastChange) >= it.nextChangeWait {
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
			if !it.Owner.World.IsTileCollision(l) && !it.Owner.World.IsTileEntityCollision(l, it.Owner.ID()) {
				choices = append(choices, l)
			}
			r := currentPos.GetAdj('R')
			if !it.Owner.World.IsTileCollision(r) && !it.Owner.World.IsTileEntityCollision(r, it.Owner.ID()) {
				choices = append(choices, r)
			}
			u := currentPos.GetAdj('U')
			if !it.Owner.World.IsTileCollision(u) && !it.Owner.World.IsTileEntityCollision(u, it.Owner.ID()) {
				choices = append(choices, u)
			}
			d := currentPos.GetAdj('D')
			if !it.Owner.World.IsTileCollision(d) && !it.Owner.World.IsTileEntityCollision(d, it.Owner.ID()) {
				choices = append(choices, d)
			}
			if len(choices) == 0 {
				logz.Warnln("IdleTask", "NPC appears to be blocked in! Idle task tried to find an adjacent tile, but all were collisions. npcID:", it.Owner.ID(), "pos:", currentPos)
			} else {
				target := choices[rand.Intn(len(choices))]
				it.Owner.Entity.GoToPos(target, true)
			}
		}

		it.setChangeTimer()
	}
}
