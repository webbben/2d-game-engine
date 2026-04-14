package world

import (
	"fmt"
	"slices"

	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/internal/debug"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/screen"
	"github.com/webbben/2d-game-engine/utils"
	activemap "github.com/webbben/2d-game-engine/world/activeMap"
	"github.com/webbben/2d-game-engine/world/npc"
	"github.com/webbben/2d-game-engine/world/worldgraph"
)

// The World represents the overall game world (or "universe" as I sometimes call it).
// It has all the important state data about things like which NPCs are where, what their current tasks are, etc.
// It also knows about what is in the current map that the player is in.
type World struct {
	// Data managers and contexts

	Dataman   *datamanager.DataManager
	Audioman  *audio.AudioManager
	EventBus  *pubsub.EventBus
	Screenman *screen.ScreenManager
	GameCtx   defs.GameContext

	// General world information

	Clock clock.Clock

	WorldGraph *worldgraph.WorldGraph

	// Characters in the World

	Player             *player.Player
	BlockPlayerChanges bool
	NPCs               map[id.CharacterStateID]*npc.NPC

	// simulation

	cmdCh      chan SimCommand
	simPaused  bool
	simStopped bool

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
	gameCtx defs.GameContext,
) *World {
	w := &World{
		Dataman:   dataman,
		Audioman:  audioman,
		EventBus:  eventBus,
		Screenman: screenman,
		GameCtx:   gameCtx,

		cmdCh: make(chan SimCommand, 1), // purposely making buffer size 1, since commands should be infrequent and not queuing up
	}

	w.SetGameTime(initTime)

	playerEnt := entity.LoadCharacterStateIntoEntity(id.CharacterStateID(defs.PlayerID), w.Dataman, w.Audioman)
	p := player.NewPlayer(w.Dataman, playerEnt)
	w.Player = &p

	for id := range w.Dataman.MapDefs {
		w.EnsureMapStateExists(id)
	}

	w.WorldGraph = worldgraph.BuildWorldGraph(w.Dataman)
	if w.WorldGraph == nil {
		panic("world graph was nil")
	}

	w.populateNPCMap()

	w.startNpcSimulation()

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
		n := npc.NewNPC(npc.NPCParams{CharStateID: charID}, w.Dataman, w.Audioman, w.EventBus, w)
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
	}
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
	}
	if doTransition {
		// block player changes so they can't accidentally enter the same map twice
		w.BlockPlayerChanges = true
		w.GameCtx.StartLoadScreen(loadFunc)
	} else {
		loadFunc(w.GameCtx)
	}
}

func (w *World) setupNewMap(mapID defs.MapID) {
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
		w.GameCtx,
		w,
		mapID,
		false,
	)

	w.ActiveMap.OnHourChange(w.Clock.Hour, true)

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
		n.SetupTaskState(currentHour, nil)
		// We do this separately from SetupTaskState since SetupTaskState is also for initializing NPC tasks for background simulation
		if n.CurrentTask != nil {
			logz.Println("loadRegularMapNPCs", "NPC setting up task active state:", n.CurrentTask.GetID(), n.WhoAmI())
			n.CurrentTask.SetupActiveState()
		} else {
			logz.Println("loadRegularMapNPCs", "NPC has no active task:", n.WhoAmI())
		}
	}
}

// CloseMap handles all work that should be done when an ActiveMap is left by the player.
// All runtime data that could possibly persist between active maps should be reset, to prevent bugs or unexpected behavior.
func (w *World) CloseMap() {
	for _, n := range w.ActiveMap.NPCs {
		n.Entity.ResetActiveMapRuntimeState()
	}
	w.ActiveMap = nil
}

func (w *World) SetGameTime(gameTime clock.GameTime) {
	w.Clock = clock.NewClock(
		config.HourSpeed,
		gameTime.Hour,
		gameTime.Minute,
		gameTime.Season,
		gameTime.DayOfSeason,
		gameTime.Year,
		config.DaysInSeason,
	)

	// make sure lighting is initialized
	w.OnHourChange(gameTime.Hour, true, false)
}

// OnHourChange handles any hourly changes that should occur; such as lighting, event publishing, etc.
// postEvent: this exists to suppress the event when we initialize the first time. the main reason being that the quest manager
// won't have its data set yet, and will panic if it receives events beforehand.
func (w *World) OnHourChange(hour int, skipFade bool, postEvent bool) {
	if hour < 0 || hour > 23 {
		panic("invalid hour")
	}

	if w.ActiveMap != nil {
		w.ActiveMap.OnHourChange(hour, skipFade)
		// check if NPCs in the map need to change their current task due to their schedule
		for _, n := range w.ActiveMap.NPCs {
			if n == nil {
				panic("npc was nil?")
			}
			n.OnHourChange(hour)
		}
	}

	if postEvent {
		w.EventBus.Publish(defs.Event{
			Type: pubsub.EventTimePass,
			Data: map[string]any{
				"hour":     hour, // should we send the hour? I think sending the full game time is probably better.
				"gameTime": w.Clock.GetCurrentGameTime(),
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

// ChangeMapOccupancy handles moving a character from one map to another, including:
//
// 1. update MapOccupancy map (determines which NPCs should load for which maps)
//
// 2. update character state 'currentMap' field
//
// Does NOT change anything in ActiveMap; ActiveMap should handle removing or inserting NPCs to its own NPC slice elsewhere.
func (w *World) ChangeMapOccupancy(charStateID id.CharacterStateID, from, to defs.MapID) {
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
		logz.Panicln("ChangeMapOccupancy", "didn't find character state ID in map occupancy: charStateID:", charStateID, "mapID:", from)
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
}

func (w World) GetActiveMapID() defs.MapID {
	if w.ActiveMap == nil {
		return ""
	}
	return w.ActiveMap.MapID
}

func (w World) getNPC(id id.CharacterStateID) *npc.NPC {
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
