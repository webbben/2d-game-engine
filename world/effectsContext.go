package world

import (
	"time"

	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
)

// This file holds all the actual implementations for the WorldEffectContext interface.
// This is where the logic goes for all those functions, which can then be used in world effects.

func ZzWorldEffectContextCheck() {
	_ = append([]defs.WorldEffectContext{}, &World{})
}

func (w *World) GetCurrentGameTime() clock.GameTime {
	return w.Clock.GetCurrentGameTime()
}

func (w *World) AddItem(itemID defs.ItemID, quantity int) {
	if quantity <= 0 {
		panic("item quantity was <= 0")
	}
	playerCharState := w.Dataman.GetCharacterState(id.CharacterStateID(defs.PlayerID))
	itemToAdd := w.Dataman.NewItemState(itemID, quantity)
	characterstate.AddItemToInventory(playerCharState, *itemToAdd, w.Dataman)

	w.EventBus.Publish(defs.Event{
		Type: pubsub.EventAddItem,
		Data: map[string]any{
			"itemID":   itemToAdd.DefID,
			"quantity": itemToAdd.Quantity,
		},
	})
}

func (w *World) AddGold(amount int) {
	if amount == 0 {
		return
	}
	characterstate.EarnMoney(&w.Player.CharacterStateRef.StandardInventory, amount, w.Dataman)
	w.EventBus.Publish(defs.Event{
		Type: pubsub.EventGoldChange,
		Data: map[string]any{
			"amount": amount,
		},
	})
}

func (w *World) RemoveGold(amount int) {
	if amount == 0 {
		return
	}
	characterstate.SpendMoney(&w.Player.CharacterStateRef.StandardInventory, amount, w.Dataman)
	w.EventBus.Publish(defs.Event{
		Type: pubsub.EventGoldChange,
		Data: map[string]any{
			"amount": -amount,
		},
	})
}

func (w *World) AddRole(roleID defs.RoleID) {
	if w.Player == nil {
		logz.Panic("Player was nil!")
	}
	w.Player.CharacterStateRef.Roles[roleID] = true
	w.EventBus.Publish(defs.Event{
		Type: pubsub.EventRoleAdded,
		Data: map[string]any{"roleID": roleID},
	})
}

func (w *World) RemoveRole(roleID defs.RoleID) {
	if w.Player == nil {
		logz.Panic("Player was nil!")
	}
	delete(w.Player.CharacterStateRef.Roles, roleID)
	w.EventBus.Publish(defs.Event{
		Type: pubsub.EventRoleRemoved,
		Data: map[string]any{"roleID": roleID},
	})
}

func (w *World) BroadcastEvent(e defs.Event) {
	w.EventBus.Publish(e)
}

func (w *World) AssignTaskToNPC(id defs.CharacterDefID, taskDef defs.TaskDef, requireListener bool) {
	logz.Println("AssignTaskToNPC", "assigning task to NPC", id, ":", taskDef.TaskID)
	// confirm this is the ID of a unique characterDef
	charDef := w.Dataman.GetCharacterDef(id)
	if !charDef.Unique {
		logz.Panicln("AssignTaskToNPC", "characterDef of ID given is not unique; can only assign tasks to specific characters if they are unique.")
	}

	// send an event to the NPC, assuming he exists...
	w.EventBus.Publish(pubsub.NPCAssignTask(string(id), taskDef))
}

func (w *World) QueueScenario(id defs.ScenarioID) {
	scenarioDef := w.Dataman.GetScenarioDef(id)

	mapID := scenarioDef.MapID
	if mapID == "" {
		panic("mapID was empty")
	}

	w.EnsureMapStateExists(mapID)

	mapState := w.Dataman.GetMapState(mapID)

	// ensure this scenario is not already queued up
	for _, scenarioID := range mapState.QueuedScenarios {
		if scenarioID == id {
			logz.Panicln("QueueScenario", "tried to queue a scenario, but its ID was already in the scenario queue for this map:", id)
		}
	}

	mapState.QueuedScenarios = append(mapState.QueuedScenarios, id)

	logz.Println("Scenario Queued", "queued", id, "in map", mapID)
}

func (w *World) UnlockMapLock(mapID defs.MapID, lockID string) {
	w.EnsureMapStateExists(mapID)

	mapState := w.Dataman.GetMapState(mapID)
	lockState, exists := mapState.MapLocks[lockID]
	if !exists {
		logz.Panicln("UnlockMapLock", "given lock ID was not found in map. mapID:", mapID, "lockID:", lockID)
	}
	lockState.Unlocked = true
	mapState.MapLocks[lockID] = lockState

	w.EventBus.Publish(defs.Event{
		Type: pubsub.EventUnlock,
		Data: map[string]any{
			"mapID":  mapID,
			"lockID": lockID,
		},
	})
}

func (w *World) SetMapLock(mapID defs.MapID, lockID string, lockLevel int) {
	w.EnsureMapStateExists(mapID)

	mapState := w.Dataman.GetMapState(mapID)
	lockState, exists := mapState.MapLocks[lockID]
	if !exists {
		logz.Panicln("UnlockMapLock", "given lock ID was not found in map. mapID:", mapID, "lockID:", lockID)
	}
	// if lockLevel is 0, we just set the lock to the original value
	if lockLevel == 0 {
		lockLevel = lockState.OriginalLockLevel
	}
	lockState.LockLevel = lockLevel
	lockState.Unlocked = false

	mapState.MapLocks[lockID] = lockState

	w.EventBus.Publish(defs.Event{
		Type: pubsub.EventUnlock,
		Data: map[string]any{
			"mapID":  mapID,
			"lockID": lockID,
		},
	})
}

func (w *World) TravelToMap(mapID defs.MapID, spawnIndex int, hours int) {
	// Basically, we just do an EnterMap followed by a timelapse.
	// Timelapse expects the player to be in a map, and handles changing the time and picking the correct NPCs to initialize in the map.

	loadFunc := func(ctx defs.GameContext) {
		w.EnterMap(mapID, spawnIndex, false)
		if hours > 0 {
			newTime := w.GetCurrentGameTime()
			newTime.AddTime(hours)
			w.TimeLapse(newTime)
		}
		time.Sleep(time.Second) // since the time lapse might need to wait for background loop to pause, wait a second before ending the loading screen
	}

	// pause the simulation while loading
	w.SimPaused.Store(true)

	// block player changes so they can't accidentally enter the same map twice
	w.BlockPlayerChanges = true
	w.GameCtx.StartLoadScreen(loadFunc)
}
