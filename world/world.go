package world

import (
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/player"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/pubsub"
	activemap "github.com/webbben/2d-game-engine/world/activeMap"
	"github.com/webbben/2d-game-engine/world/npc"
)

// The World represents the overall game world (or "universe" as I sometimes call it).
// It has all the important state data about things like which NPCs are where, what their current tasks are, etc.
// It also knows about what is in the current map that the player is in.
type World struct {
	// Data managers and contexts

	Dataman  *datamanager.DataManager
	Audioman *audio.AudioManager
	EventBus *pubsub.EventBus
	GameCtx  activemap.GameContext

	// General world information

	Clock clock.Clock

	WorldGraph *WorldGraph

	// Characters in the World

	Player *player.Player
	NPCs   map[state.CharacterStateID]*npc.NPC

	// Map Information

	ActiveMap *activemap.ActiveMap

	// tracks which NPCs are in which maps; this is what is checked to determine if an NPC should show up in an ActiveMap or not.
	MapOccupancy map[defs.MapID][]state.CharacterStateID
}

// NewWorld returns a World that is ready to run. Assumes that all data definitions and player state has already been loaded/created.
func NewWorld(
	initTime clock.GameTime,
	dataman *datamanager.DataManager,
	audioman *audio.AudioManager,
	eventBus *pubsub.EventBus,
	gameCtx activemap.GameContext,
) *World {
	w := &World{
		Dataman:  dataman,
		Audioman: audioman,
		EventBus: eventBus,
		GameCtx:  gameCtx,
	}

	w.SetGameTime(initTime)

	playerEnt := entity.LoadCharacterStateIntoEntity(state.CharacterStateID(defs.PlayerID), w.Dataman, w.Audioman)
	p := player.NewPlayer(w.Dataman, playerEnt)
	w.Player = &p

	w.populateNPCMap()

	return w
}

func (w *World) populateNPCMap() {
	w.NPCs = make(map[state.CharacterStateID]*npc.NPC)
	w.MapOccupancy = make(map[defs.MapID][]state.CharacterStateID)

	for id, charState := range w.Dataman.CharacterStates {
		if id == state.CharacterStateID(defs.PlayerID) {
			// we don't want to make an NPC for the player
			continue
		}
		if _, exists := w.NPCs[id]; exists {
			logz.Panicln("World", "loading NPC, but an existing NPC with the same ID was found...", id)
		}
		if charState.Temp {
			// don't use temp char states, since those are just for scenarios
			continue
		}
		n := npc.NewNPC(npc.NPCParams{CharStateID: id}, w.Dataman, w.Audioman, w.EventBus)
		w.NPCs[id] = n

		currentMap := charState.CurrentMap
		if currentMap == "" {
			logz.Panicln("World", "charState didn't have a current map. charStateID:", id)
		}
		w.MapOccupancy[currentMap] = append(w.MapOccupancy[currentMap], id)
	}
}

func (w *World) EnterMapAtPosition(mapID defs.MapID, x, y float64) {
	w.setupNewMap(mapID)
	w.ActiveMap.PlacePlayerAtPosition(w.Player, x, y)
}

// EnterMap sets up a map and puts the player in it at the given position. meant for use once player already exists in game state
func (w *World) EnterMap(mapID defs.MapID, playerSpawnIndex int) {
	w.setupNewMap(mapID)
	w.ActiveMap.PlacePlayerAtSpawnPoint(w.Player, playerSpawnIndex)
}

func (w *World) setupNewMap(mapID defs.MapID) {
	if w.ActiveMap != nil {
		w.CloseMap()
	}

	w.ActiveMap = activemap.NewActiveMap(
		w.Dataman,
		w.Audioman,
		w.EventBus,
		w.GameCtx,
		w,
		mapID,
		false,
	)

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
}

func (w *World) loadRegularMapNPCs() {
	if w.ActiveMap == nil {
		panic("map was nil")
	}

	for _, n := range w.MapOccupancy[w.ActiveMap.MapID] {
		// TODO: decide the positions that these NPCs should be at.
		// probably based on their current task? I think this is how it'll go:
		// - Route tasks: calculate the progress, and then place the NPC along the path at an appropriate place.
		// - Other tasks: put the NPC at the "in progress state". E.g. sitting at a table, if their task is to do so.
		// For now, we will just put NPCs at the position where their "in progress state" would be. we will implement more complex things like
		// NPCs sitting in chairs and whatnot later.
		logz.TODO("World", "need to load NPCs for regular maps. NPC/character to load:", n)
	}
}

func (w *World) loadScenario(scenarioDef defs.ScenarioDef) {
	if w.ActiveMap == nil {
		panic("map was nil")
	}

	logz.Println("Loading Scenario", "MapID:", scenarioDef.MapID, "ScenarioID:", scenarioDef.ID)
	if scenarioDef.MapID != w.ActiveMap.MapID {
		logz.Panicln("SetupMap", "found queued scenario for map, but mapID in scenario def doesn't match. mapID:", w.ActiveMap.MapID, "found in scenario def:", scenarioDef.MapID)
	}

	for _, charDef := range scenarioDef.Characters {
		params := entity.NewCharacterStateParams{
			Temp: true,
		}
		charStateID := entity.CreateNewCharacterState(
			charDef.CharDefID,
			params,
			w.Dataman)
		n := npc.NewNPC(npc.NPCParams{
			CharStateID: charStateID,
		}, w.Dataman, w.Audioman, w.EventBus)

		startPos := model.Coords{X: charDef.SpawnCoordX, Y: charDef.SpawnCoordY}
		w.ActiveMap.AddNPCToMap(n, startPos)
	}
}

func (w *World) CloseMap() {
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
		90, // TODO: move this to config variable?
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
	}

	if postEvent {
		w.EventBus.Publish(defs.Event{
			Type: pubsub.EventTimePass,
			Data: map[string]any{
				"hour": hour,
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
