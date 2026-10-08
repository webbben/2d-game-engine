package world

import (
	"time"

	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/combat"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/tiled"
	"github.com/webbben/2d-game-engine/utils"
	"github.com/webbben/2d-game-engine/world/npc"
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
		logz.Panic("item quantity was <= 0")
	}
	playerCharState := w.Dataman.GetCharacterState(id.PlayerStateID)
	itemToAdd := w.Dataman.NewItemState(itemID, quantity)
	success, remaining := characterstate.AddItemToInventory(playerCharState, *itemToAdd, w.Dataman)
	addedQuantity := quantity - remaining.Quantity
	if !success {
		// inventory must be full
		w.EventBus.Publish(defs.Event{
			Type: pubsub.EventInventoryFull,
		})
		// drop the remaining items on the ground
		w.DropItemOnGround(itemID, remaining.Quantity, remaining.Durability)
	}

	if addedQuantity > 0 {
		w.EventBus.Publish(defs.Event{
			Type: pubsub.EventAddItem,
			Data: map[string]any{
				"itemID":   itemID,
				"quantity": addedQuantity,
			},
		})
	}
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

func (w *World) AssignTaskToNPC(id id.CharacterDefID, taskDef defs.TaskDef, requireListener bool) {
	logz.Println("AssignTaskToNPC", "assigning task to NPC", id, ":", taskDef.TaskID)
	// confirm this is the ID of a unique characterDef
	charDef := w.Dataman.GetCharacterDef(id)
	if !charDef.Unique {
		logz.Panicln("AssignTaskToNPC", "characterDef of ID given is not unique; can only assign tasks to specific characters if they are unique.")
	}

	// send an event to the NPC, assuming he exists...
	w.EventBus.Publish(pubsub.NPCAssignTask(string(id), taskDef))
}

func (w *World) InitiateCombat(charStateID, targetCharStateID id.CharacterStateID) {
	npcRef := w.getInWorldNPC(charStateID)
	if npcRef == nil {
		logz.Panicln("InitiateCombat", "initiating NPC is not a current in-world NPC:", charStateID)
	}

	var targetEntity *entity.Entity
	if targetCharStateID == id.PlayerStateID {
		if w.Player == nil {
			logz.Panicln("InitiateCombat", "target is the player, but the player is nil")
		}
		targetEntity = w.Player.Entity
	} else {
		targetNPC := w.getInWorldNPC(targetCharStateID)
		if targetNPC == nil {
			logz.Panicln("InitiateCombat", "target NPC is not a current in-world NPC:", targetCharStateID)
		}
		targetEntity = targetNPC.Entity
	}

	taskDef := defs.TaskDef{
		TaskID:   npc.TaskFight,
		Priority: npc.Emergency,
		Params:   npc.FightTaskParams{TargetEntity: targetEntity},
	}

	// Record the fight before assigning the task, so that by the time the NPC starts fighting it's
	// already findable as a combatant. Note this can legitimately do nothing (if the initiator is
	// already fighting someone else) -- the task assignment below is unchanged either way.
	w.StartCombat(charStateID, targetCharStateID, combat.IntentKill, combat.IntentSelfDefense)

	w.EventBus.Publish(pubsub.NPCAssignTask(string(charStateID), taskDef))
}

// HasNPCSurrenderOffer returns whether an npc has a surrender offer pending. False if the npc isn't
// on the map (a dialog could still be closing out as the player changes maps).
func (w *World) HasNPCSurrenderOffer(charStateID id.CharacterStateID) bool {
	npcRef := w.getInWorldNPC(charStateID)
	if npcRef == nil {
		return false
	}
	return npcRef.IsOfferingSurrender()
}

// AcceptNPCSurrender ends the fight. The npc is flagged as surrendered so it won't take another
// surrender offer, and gets TaskDoNothing, which clears its current task; the task decision loop then
// picks up whatever this hour's schedule says next.
func (w *World) AcceptNPCSurrender(charStateID id.CharacterStateID) {
	npcRef := w.getInWorldNPC(charStateID)
	if npcRef == nil {
		logz.Panicln("AcceptNPCSurrender", "npc is not a current in-world NPC:", charStateID)
	}
	npcRef.MarkSurrendered()

	// Accepting ends this character's part in the fight: they've given it up and won't be attacked
	// again, so they stop being a combatant. Note the *other* side stays in the session -- they may
	// still be fighting someone else, or the fight may now be over, which LeaveCombat sorts out.
	w.LeaveCombat(charStateID)

	// Finish the surrender task rather than swapping in another one. Once it reports done, the task
	// decision loop picks this hour's schedule up on its own. Assigning TaskDoNothing instead would go
	// through clearTask, which nils the task without ever calling Finish -- so the task's dialog-ended
	// subscription would stay registered, and the next time any dialog with this npc closed it would
	// react and send them running for no reason.
	if surrenderTask, ok := npcRef.CurrentTask.(*npc.SurrenderTask); ok {
		surrenderTask.Finish(npc.TaskResult{Status: npc.ResultSuccess})
	}
}

// getInWorldNPC resolves an NPC by char state ID, first from the world's regular
// NPC registry, and falling back to temp scenario NPCs placed in the active map.
func (w *World) getInWorldNPC(charStateID id.CharacterStateID) *npc.NPC {
	if n, exists := w.NPCs[charStateID]; exists {
		return n
	}
	if w.ActiveMap != nil {
		for _, n := range w.ActiveMap.GetAllNPCs() {
			if n.CharacterStateRef.ID == charStateID {
				return n
			}
		}
	}
	return nil
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
			logz.Println("QueueScenario", id)
			logz.Panicln("QueueScenario", "tried to queue a scenario, but its ID was already in the scenario queue for this map")
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
		logz.Println("UnlockMapLock", mapID, "lockID:", lockID)
		logz.Panicln("UnlockMapLock", "given lock ID was not found in map")
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
		logz.Println("UnlockMapLock", mapID, "lockID:", lockID)
		logz.Panicln("UnlockMapLock", "given lock ID was not found in map")
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

func (w *World) AddOpinionModifier(holder, subject id.CharacterStateID, mod defs.OpinionModifier) {
	characterstate.AddOpinionModifier(holder, subject, mod, w.Dataman)
}

func (w *World) GetDialogNPC() id.CharacterStateID {
	if w.ActiveMap == nil {
		return ""
	}
	return w.ActiveMap.GetDialogNPC()
}

func (w *World) DropItemOnGround(itemID defs.ItemID, quantity int, durability float64) {
	utils.PanicAssert(itemID != "", "itemID was empty")
	utils.PanicAssert(quantity > 0, "quantity was <= 0")
	utils.PanicAssert(w.ActiveMap != nil, "active map is nil")
	utils.PanicAssert(w.Player != nil, "player was nil")

	w.EnsureMapStateExists(w.ActiveMap.MapID)

	mapState := w.Dataman.GetMapState(w.ActiveMap.MapID)

	// dropped items land on the player's tile
	playerTile := w.Player.Entity.GetEntityInfo().TilePos
	x := float64(playerTile.X * config.TileSize)
	y := float64(playerTile.Y * config.TileSize)

	expiresAt := w.Clock.GetFutureGameTime(24)
	st := state.MapItemState{
		ID:        nextRuntimeItemObjID(mapState, *w.ActiveMap.Map),
		X:         x,
		Y:         y,
		ExpiresAt: &expiresAt,
		Dropped:   true,
		ItemState: state.ItemState{
			DefID:      itemID,
			Quantity:   quantity,
			Durability: durability,
		},
	}
	st.ItemState.Validate()

	mapState.MapItems = append(mapState.MapItems, st)
	w.ActiveMap.AddDroppedItem(st)

	logz.Println("DropItemOnGround", "item dropped:", itemID, quantity)
	w.EventBus.Publish(defs.Event{
		Type: pubsub.EventDropItem,
		Data: map[string]any{
			"itemID":    itemID,
			"quantity":  quantity,
			"droppedBy": id.PlayerStateID,
		},
	})
}

func nextRuntimeItemObjID(ms *state.MapState, m tiled.Map) int {
	// TODO: if we get to the point that we want to support spawning tons of items at once, we should probably just track NextObjectID
	// on map state. that way we don't have to iterate over all map items each time.
	// for example, this could cause noticeable lag if we tried to spawn hundreds of items in a single tick. guessing that will never happen though.
	id := m.NextObjectID
	for _, is := range ms.MapItems {
		if is.ID >= id {
			id = is.ID + 1
		}
	}
	return id
}
