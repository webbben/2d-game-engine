package world

import (
	"fmt"
	"slices"
	"sync/atomic"

	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/internal/debug"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/quest"
	"github.com/webbben/2d-game-engine/screen"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/utils"
	activemap "github.com/webbben/2d-game-engine/world/activeMap"
	"github.com/webbben/2d-game-engine/world/npc"
	"github.com/webbben/2d-game-engine/worldgraph"
)

// The World represents the overall game world (or "universe" as I sometimes call it).
// It has all the important state data about things like which NPCs are where, what their current tasks are, etc.
// It also knows about what is in the current map that the player is in.
type World struct {
	// Screens

	// we just keep this here to pass to the active map. TODO: should we just keep screen ID instead?
	playerMenu screen.Screen

	// Data managers and contexts

	Dataman   *datamanager.DataManager
	Audioman  *audio.AudioManager
	EventBus  *pubsub.EventBus
	Screenman *screen.ScreenManager
	Questman  *quest.QuestManager
	GameCtx   defs.GameContext

	OverlayManager *overlay.OverlayManager

	// General world information

	Clock clock.Clock

	WorldGraph *worldgraph.WorldGraph

	// Characters in the World

	Player             *player.Player
	BlockPlayerChanges bool
	NPCs               map[id.CharacterStateID]*npc.NPC

	// simulation

	SimPaused        atomic.Bool // set this as the flag to get the simulation to pause
	SimPauseEffected atomic.Bool // if true, then the sim has successfully paused and is no longer processing NPC simulation updates.

	// time lapse - for things like sleeping or waiting in-game.

	AwaitingTimeLapse bool            // when a time lapse action comes in, set this flag and then wait until sim pause has been effected before doing time lapse.
	TimeLapseTo       *clock.GameTime // the time we should lapse to

	// Map Information

	ActiveMap *activemap.ActiveMap

	// tracks which NPCs are in which maps; this is what is checked to determine if an NPC should show up in an ActiveMap or not.
	MapOccupancy map[defs.MapID][]id.CharacterStateID
}

// NewWorld returns a World that is ready to run. Assumes that all data definitions and player state has already been loaded/created.
func NewWorld(
	initTime clock.GameTime,
	dataman *datamanager.DataManager,
	audioman *audio.AudioManager,
	eventBus *pubsub.EventBus,
	screenman *screen.ScreenManager,
	questman *quest.QuestManager,
	gameCtx defs.GameContext,
	playerMenu screen.Screen,
) *World {
	w := &World{
		Dataman:        dataman,
		Audioman:       audioman,
		EventBus:       eventBus,
		Screenman:      screenman,
		Questman:       questman,
		GameCtx:        gameCtx,
		playerMenu:     playerMenu,
		OverlayManager: &overlay.OverlayManager{},
	}

	// no need to setup things like lighting or post events; active map doesn't exist yet,
	// and those things are handled at the time of creating the active map.
	w.Clock = clock.NewClock(
		config.HourSpeed,
		initTime.Hour,
		initTime.Minute,
		initTime.Season,
		initTime.DayOfSeason,
		initTime.Year,
		config.DaysInSeason,
	)

	playerEnt := entity.LoadCharacterStateIntoEntity(id.CharacterStateID(defs.PlayerID), w.Dataman, w.Audioman)
	p := player.NewPlayer(w.Dataman, playerEnt)
	w.Player = &p

	// this ensures all map states exist, and also effectively that all NPC states exist, since everytime it encounters a bed for an NPC it ensures that NPC's state
	// has been created too.
	// Note: this does NOT handle map generators; those are built as they are found during the world graph building process (below).
	for id, def := range w.Dataman.MapDefs {
		if def.IsMapGenTemplate {
			// don't initialize a map state for template map defs; they aren't directly representing a specific map
			continue
		}
		w.EnsureMapStateExists(id)
	}

	for id, def := range w.Dataman.ShopkeeperDefs {
		if !w.Dataman.ShopkeeperStateExists(id) {
			w.Dataman.LoadShopkeeperState(state.NewShopKeeperState(*def))
		}
	}

	w.BuildWorldGraph()
	if w.WorldGraph == nil {
		panic("world graph was nil")
	}

	w.populateNPCMap()

	w.startNpcSimulation()

	w.EventBus.SubscribeToWorldEvents("WORLD", w.OnEvent)

	return w
}

func (w *World) populateNPCMap() {
	w.NPCs = make(map[id.CharacterStateID]*npc.NPC)
	w.MapOccupancy = make(map[defs.MapID][]id.CharacterStateID)

	for charID, charState := range w.Dataman.CharacterStates {
		if charID == id.CharacterStateID(defs.PlayerID) {
			// we don't want to make an NPC for the player
			continue
		}
		if _, exists := w.NPCs[charID]; exists {
			logz.Panicln("World", "loading NPC, but an existing NPC with the same ID was found...", charID)
		}
		if charState.Temp {
			// don't use temp char states, since those are just for scenarios
			continue
		}

		npcParams := npc.NPCParams{
			CharStateID:             charID,
			SpeechBubbleTileset:     config.SpeechBubbleBox.TilesetSrc,
			SpeechBubbleOriginIndex: config.SpeechBubbleBox.OriginIndex,
			SpeechBubbleFont:        config.SpeechBubbleFont,
		}
		n := npc.NewNPC(npcParams, w.Dataman, w.Audioman, w.EventBus, w)
		w.NPCs[charID] = n

		currentMap := charState.CurrentMap
		if currentMap == "" {
			logz.Panicln("World", "charState didn't have a current map. charStateID:", charID)
		}
		w.MapOccupancy[currentMap] = append(w.MapOccupancy[currentMap], charID)
	}
}

func (w *World) EnterMapAtPosition(mapID defs.MapID, x, y float64, doTransition bool) {
	loadFunc := func(ctx defs.GameContext) {
		w.setupNewMap(mapID)
		w.ActiveMap.PlacePlayerAtPosition(w.Player, x, y)

		// only unpause simulation if we aren't in a scenario (simulation doesn't run in scenarios)
		if !w.ActiveMap.InScenario {
			w.SimPaused.Store(false)
		}
	}

	w.SimPaused.Store(true)

	if doTransition {
		w.GameCtx.StartLoadScreen(loadFunc)
	} else {
		loadFunc(w.GameCtx)
	}
}

// EnterMap sets up a map and puts the player in it at the given position. meant for use once player already exists in game state
func (w *World) EnterMap(mapID defs.MapID, playerSpawnIndex int, doTransition bool) {
	loadFunc := func(ctx defs.GameContext) {
		w.setupNewMap(mapID)
		w.ActiveMap.PlacePlayerAtSpawnPoint(w.Player, playerSpawnIndex)

		// only unpause simulation if we aren't in a scenario (simulation doesn't run in scenarios)
		if !w.ActiveMap.InScenario {
			w.SimPaused.Store(false)
		}
	}

	// pause the simulation while loading
	w.SimPaused.Store(true)

	if doTransition {
		// block player changes so they can't accidentally enter the same map twice
		w.BlockPlayerChanges = true
		w.GameCtx.StartLoadScreen(loadFunc)
	} else {
		loadFunc(w.GameCtx)
	}
}

func (w *World) setupNewMap(mapID defs.MapID) {
	if !w.SimPaused.Load() {
		logz.Panicln("setupNewMap", "setting up new map, but simulation isn't paused.")
	}
	if !w.SimPauseEffected.Load() {
		logz.Warnln("setupNewMap", "setting up new map, but simulation pause doesn't seem to have taken effect yet")
	}

	debug.StartTimer("setupNewMap")
	logz.Println("WORLD", "setting up map:", mapID)
	if w.ActiveMap != nil {
		w.CloseMap()
	}

	w.ActiveMap = activemap.NewActiveMap(
		w.Dataman,
		w.Audioman,
		w.EventBus,
		w.Screenman,
		w.Questman,
		w.OverlayManager,
		w.GameCtx,
		w,
		mapID,
		w.playerMenu,
		false,
	)

	// initialize day light and send out hour event
	// skip NPC check since that is handled below in the NPC loading functions
	w.OnHourChange(w.Clock.Hour, true, true, true)

	// figure out which NPCs should be added to the map
	// check if a scenario should be loaded into the map
	mapState := w.Dataman.GetMapState(mapID)
	if len(mapState.QueuedScenarios) > 0 {
		scenarioID := mapState.QueuedScenarios[0]
		mapState.QueuedScenarios = mapState.QueuedScenarios[1:]
		scenarioDef := w.Dataman.GetScenarioDef(scenarioID)
		w.loadScenario(scenarioDef)
	} else {
		w.loadRegularMapNPCs()
	}

	debug.StopTimer("setupNewMap")
}

func (w *World) loadRegularMapNPCs() {
	if w.ActiveMap == nil {
		panic("map was nil")
	}
	debug.StartTimer("loadRegularMapNPCs")

	logz.Println("loadRegularMapNPCs", "loading regular NPCs for map:", w.ActiveMap.MapID)

	currentHour := w.Clock.GetCurrentGameTime()
	x0, y0, found := w.ActiveMap.GetSpawnPosition(0)
	if !found {
		logz.Panicln("loadRegularMapNPCs", "failed to find spawn point 0. all maps must have this spawn point index!")
	}
	spawnPoint0 := model.ConvertPxToTilePos(x0, y0)

	for _, id := range w.MapOccupancy[w.ActiveMap.MapID] {
		// add an NPC to the map for this character, and then call its initial task state setter.
		// we can start by placing each NPC at the main spawn point (index=0).
		// if the initial task starter is successful, it should necessarily be moved somewhere else.
		n := w.NPCs[id]
		w.ActiveMap.AddNPCToMap(n, spawnPoint0)
		if n.CurrentTask == nil {
			n.SetupTaskState(currentHour, nil)
		}
		// We do this separately from SetupTaskState since SetupTaskState is also for initializing NPC tasks for background simulation
		if n.CurrentTask != nil {
			logz.Println("loadRegularMapNPCs", "NPC setting up task active state:", n.CurrentTask.GetID(), n.WhoAmI())
			n.CurrentTask.SetupActiveState()
		} else {
			logz.Warnln("loadRegularMapNPCs", "NPC has no active task:", n.WhoAmI())
		}
	}

	debug.StopTimer("loadRegularMapNPCs")
}

// CloseMap handles all work that should be done when an ActiveMap is left by the player.
// All runtime data that could possibly persist between active maps should be reset, to prevent bugs or unexpected behavior.
//
// Especially things like:
//   - event subscriptions for objects, NPCs, or anything else that is tied to this active map (to prevent future possible duplicate subscriptions)
//   - Player, NPC or other entity activity states, like sitting, sleeping, etc. (could affect NPC tasks, for example)
func (w *World) CloseMap() {
	logz.Println("CloseMap", w.ActiveMap.MapID)
	for _, n := range w.ActiveMap.NPCs {
		n.PrepareLeaveActiveMap()
	}

	for _, obj := range w.ActiveMap.Objects {
		obj.OnMapClose()
	}
	w.ActiveMap = nil
}

// OnHourChange handles any hourly changes that should occur; such as lighting, event publishing, etc.
// postEvent: this exists to suppress the event when we initialize the first time. the main reason being that the quest manager
// won't have its data set yet, and will panic if it receives events beforehand.
func (w *World) OnHourChange(hour int, skipFade, skipNpcCheck, postEvent bool) {
	if hour < 0 || hour > 23 {
		panic("invalid hour")
	}
	currentTime := w.Clock.GetCurrentGameTime()
	if hour != currentTime.Hour {
		logz.Panicln("OnHourChange", "given hour doesn't match actual game time. hour:", hour, "gameTime hour:", currentTime.Hour)
	}

	if w.ActiveMap != nil {
		if w.ActiveMap.InScenario {
			logz.Panicln("OnHourChange", "Time is not supposed to pass while in scenarios.")
		}
		w.ActiveMap.OnHourChange(hour, skipFade)

		if !skipNpcCheck {
			// check if NPCs in the map need to change their current task due to their schedule
			for _, n := range w.ActiveMap.NPCs {
				if n == nil {
					panic("npc was nil?")
				}
				n.OnHourChange(hour)
			}
		}
	}

	if postEvent {
		w.EventBus.Publish(defs.Event{
			Type: pubsub.EventTimePass,
			Data: map[string]any{
				"hour":     hour, // we send the hour too, just because that's how we originally did it
				"gameTime": currentTime,
			},
		})
	}
}

func (w *World) SetPlayerName(name string) {
	if w.Player == nil {
		panic("player was nil")
	}
	if w.Player.Entity == nil {
		panic("player entity was nil")
	}
	if w.Player.CharacterStateRef == nil {
		panic("player character state was nil")
	}
	w.Player.CharacterStateRef.DisplayName = name
}

func (w *World) FindWorldPath(from, to defs.MapID) (pathToGoal worldgraph.WorldPath, found bool) {
	if w.WorldGraph == nil {
		logz.Panicln("WORLD", "tried to get world path, but world graph was nil!")
	}
	return w.WorldGraph.FindPath(from, to)
}

func (w *World) ChangeMapOccupancyEvent(charStateID id.CharacterStateID, from, to defs.MapID, toSpawn int) {
	w.EventBus.SysChangeMapOccupancy(charStateID, from, to, toSpawn)
}

func (w *World) OnEvent(e defs.Event) {
	logz.Println("WORLD", "Incoming event:", e.Type)
	switch e.Type {
	case pubsub.SysEventChangeMapOccupancy:
		if _, ok := e.Data["params"]; ok {
			params, ok := e.Data["params"].(pubsub.SysEventChangeMapOccupancyParams)
			if !ok {
				logz.Panicln("WORLD", "SysEventChangeMapOccupancy:", "event data couldn't be type asserted to params.", e.Data)
			}
			w.ChangeMapOccupancy(params.CharacterStateID, params.From, params.To, params.ToSpawn)
		} else {
			logz.Panicln("WORLD", "SysEventChangeMapOccupancy:", "didn't find params in data.", e.Data)
		}
	case pubsub.SysScheduledWorldEffect:
		effectType, ok := e.Data["type"]
		if !ok {
			logz.Panicln("SysScheduledWorldEffect", "event data was missing effect type.", e.Data)
		}
		effectData, ok := e.Data["effect"]
		if !ok {
			logz.Panicln("SysScheduledWorldEffect", "event data was missing effect.", e.Data)
		}

		var worldEffect defs.WorldEffect

		switch effectType {
		case "RemoveRoleEffect":
			removeRoleEffect, ok := effectData.(RemoveRoleEffect)
			if !ok {
				logz.Panicln("SysScheduledWorldEffect", "effect is supposed to be removeRole, but failed to type assert.", effectType, effectData)
			}
			worldEffect = removeRoleEffect
		}

		if worldEffect == nil {
			logz.Panicln("SysScheduledWorldEffect", "worldEffect was nil")
		}

		worldEffect.Apply(w)
	case pubsub.SysShowScreen:
		if w.ActiveMap == nil {
			logz.Panicln("SysShowScreen", "received show screen event, but active map is nil!")
		}
		screenID, ok := e.Data["screen_id"].(defs.ScreenID)
		if !ok {
			logz.Panicln("SysShowScreen", "screen_id data not found.", e.Data)
		}
		screenParams := e.Data["params"]

		scr := w.Screenman.GetScreen(screenID)
		w.ShowMiscScreen(scr, screenParams)
	}
}

// ChangeMapOccupancy handles moving a character from one map to another, including:
//
// 1. update MapOccupancy map (determines which NPCs should load for which maps)
//
// 2. update character state 'currentMap' field
//
// 3. if entering the active map, also places the NPC in at the spawn point
func (w *World) ChangeMapOccupancy(charStateID id.CharacterStateID, from, to defs.MapID, toSpawn int) {
	if from == "" {
		panic("from was empty!")
	}
	if to == "" {
		panic("to was empty!")
	}
	if from == to {
		logz.Panicln("ChangeMapOccupancy", "from and to were the same! whoever called this should've noticed that and handled it. to/from:", to)
	}
	if _, exists := w.MapOccupancy[from]; !exists {
		// all from maps should have a map occupancy because, of course, a character is apparently already in that map...
		logz.Panicln("ChangeMapOccupancy", "MapOccupancy not defined for 'from' map:", from)
	}
	if _, exists := w.MapOccupancy[to]; !exists {
		// it's not actually a problem if the 'to' doesn't exist yet. That could just mean that there are no
		// character beds in this map, so on initialization nobody was ever placed in it initially as their "home" map.
		w.MapOccupancy[to] = make([]id.CharacterStateID, 0)
	}

	logz.Printf("WORLD", "Change map occupancy (%s) %s -> %s", charStateID, from, to)

	found := false
	for i, id := range w.MapOccupancy[from] {
		if id == charStateID {
			fromOccupancy := w.MapOccupancy[from]
			originalLen := len(fromOccupancy)
			fromOccupancy = utils.RemoveIndexUnordered(fromOccupancy, i)
			// do a quick gut check and confirm the length decreased by 1
			if len(fromOccupancy) != originalLen-1 {
				logz.Panicln("ChangeMapOccupancy", "map occupancy length didn't decrease by 1 after removing character. from:", from)
			}
			w.MapOccupancy[from] = fromOccupancy
			found = true
			break
		}
	}
	if !found {
		logz.Panicln("ChangeMapOccupancy", "didn't find character state ID in from map occupancy. charStateID:", charStateID, "from mapID:", from)
	}

	// ensure ID doesn't already exist in new map (before adding)
	// TODO: if this function ever becomes sufficiently "hot" then we can remove this if it proves to slow things down at all.
	// I don't think it'll be a problem, but maybe it could have some minor impact if maps start having really large numbers of occupants?
	if slices.Contains(w.MapOccupancy[to], charStateID) {
		fmt.Println("from:", w.MapOccupancy[from], "\nto:", w.MapOccupancy[to])
		logz.Panicln("ChangeMapOccupancy", "charStateID already exists in new map occupancy:", charStateID, "to mapID:", to)
	}

	w.MapOccupancy[to] = append(w.MapOccupancy[to], charStateID)

	// update character state
	charState := w.Dataman.GetCharacterState(charStateID)
	charState.CurrentMap = to

	// handle placing at spawn point if NPC is entering active map
	if w.ActiveMap != nil && w.ActiveMap.MapID == to && toSpawn != -1 {
		// if toSpawn is set to -1, then that means the place calling this function doesn't expect the player to be in the map, or at least to have to place the
		// NPC in a specific spot. One reason could be, the game is setting up initial map occupancies for NPC's based on their scheduled task, and later on it will
		// actually run logic to place them at their active locations. Such as during a time lapse.

		// TODO: should we add some extra check to ensure that toSpawn is never "incorrectly" set to -1?
		// Maybe that's the job of the calling code?

		n := w.NPCs[charStateID]
		if n == nil {
			panic("npc was nil")
		}
		x, y, found := w.ActiveMap.GetSpawnPosition(toSpawn)
		if !found {
			logz.Panicln("ChangeMapOccupancy", "spawn point not found! spawn index:", toSpawn)
		}
		startPos := model.ConvertPxToTilePos(x, y)
		w.ActiveMap.AddNPCToMap(n, startPos)
	}

	// send event to notify map movement
	w.EventBus.Publish(defs.Event{
		Type: pubsub.EventMapOccupancyChange,
		Data: map[string]any{
			"charStateID": charStateID,
			"to":          to,
			"from":        from,
		},
	})
}

func (w *World) GetActiveMapID() defs.MapID {
	if w.ActiveMap == nil {
		return ""
	}
	return w.ActiveMap.MapID
}

func (w *World) getNPC(id id.CharacterStateID) *npc.NPC {
	n, exists := w.NPCs[id]
	if !exists {
		logz.Panicln("WORLD", "NPC not found! charStateID:", id)
	}
	return n
}

// AddNPCToActiveMap adds an NPC to the current active map. All it does is place the NPC into the map; it doesn't do any task active state setup or anything.
// This is because this function is used to allow NPCs to enter the map that the player is already in - so it's like they're walking into the map from the doorway.
func (w *World) AddNPCToActiveMap(charStateID id.CharacterStateID, spawnIndex int) {
	n := w.getNPC(charStateID)
	x, y, found := w.ActiveMap.GetSpawnPosition(spawnIndex)
	if !found {
		logz.Panicln("AddNPCToActiveMap", "given spawn index doesn't exist in active map:", spawnIndex, "mapID:", w.ActiveMap.MapID)
	}
	spawnPos := model.ConvertPxToTilePos(x, y)

	w.ActiveMap.AddNPCToMap(n, spawnPos)
}

func (w *World) TogglePlayerMenu() {
	if w.playerMenu == nil {
		panic("player menu was nil")
	}
	w.ActiveMap.TogglePlayerMenu()
}

func (w *World) ShowMiscScreen(scr screen.Screen, params any) {
	if scr == nil {
		logz.Panic("screen was nil!")
	}
	w.ActiveMap.ShowMiscScreen(scr, params)
}
