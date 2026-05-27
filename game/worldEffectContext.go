package game

import (
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

// All of these just pass through to the implementations in World. Need them under game too though, since this context feeds
// into GameContext.

func (g *Game) requireWorld() {
	if g.World == nil {
		logz.Panicln("GAME", "Tried to call a WorldEffectContext function, but World hasn't been created yet!")
	}
}

func (g *Game) AddGold(amount int) {
	g.requireWorld()
	g.World.AddGold(amount)
}

func (g *Game) RemoveGold(amount int) {
	g.requireWorld()
	g.World.RemoveGold(amount)
}

func (g *Game) SetPlayerName(name string) {
	g.requireWorld()
	g.World.SetPlayerName(name)
}

func (g *Game) AddItem(itemID defs.ItemID, quantity int) {
	g.requireWorld()
	g.World.AddItem(itemID, quantity)
}

func (g *Game) AssignTaskToNPC(id defs.CharacterDefID, taskDef defs.TaskDef, requireListener bool) {
	g.requireWorld()
	g.World.AssignTaskToNPC(id, taskDef, requireListener)
}

func (g *Game) QueueScenario(id defs.ScenarioID) {
	g.requireWorld()
	g.World.QueueScenario(id)
}

func (g *Game) UnlockMapLock(mapID defs.MapID, lockID string) {
	g.requireWorld()
	g.World.UnlockMapLock(mapID, lockID)
}

// EnterMap adds the player to a map. creates the active map too, in the process.
// Used in the NewGame flow to actually put the player in a map once his character has been created.
func (g *Game) EnterMap(mapID defs.MapID, spawnIndex int, doTransition bool) {
	g.requireWorld()
	g.World.EnterMap(mapID, spawnIndex, doTransition)
}

func (g Game) GetCurrentGameTime() clock.GameTime {
	g.requireWorld()
	return g.World.GetCurrentGameTime()
}

func (g *Game) AddRole(roleID defs.RoleID) {
	g.requireWorld()
	g.World.AddRole(roleID)
}

func (g *Game) RemoveRole(roleID defs.RoleID) {
	g.requireWorld()
	g.World.RemoveRole(roleID)
}

func (g *Game) TravelToMap(mapID defs.MapID, spawnIndex int, hours int) {
	g.requireWorld()
	g.World.TravelToMap(mapID, spawnIndex, hours)
}
