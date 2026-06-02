package world

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/object"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/world/npc"
)

func (w *World) Update(showingLoadScreen bool) {
	if w.ActiveMap == nil {
		// If ActiveMap is nil, we assume that the game world is not "active" (i.e. it's in a loading screen, or something)
		// So, we can just quit out of world updates in this case.
		// We don't want NPC's to be doing things in the background world simulation while the player is just waiting on a load screen.
		return
	}
	// when ActiveMap is defined, that means the player is actively in a map.
	// this also means that background world simulation can occur, since the game world at large is also "active".

	// allow time lapses to occur while in a screen, so a "sleep screen" can have time lapses take effect while still in screen.
	if w.AwaitingTimeLapse {
		if !w.SimPaused.Load() {
			logz.Panicln("WORLD", "awaiting time lapse, but simulation hasn't been signaled to pause yet")
		}
		if w.TimeLapseTo == nil {
			logz.Panicln("WORLD", "awaiting time lapse, but TimeLapseTo was nil!")
		}
		if w.SimPauseEffected.Load() {
			// simulation has acknowledged the pause, so we can proceed with the time lapse.
			w.timeLapse(*w.TimeLapseTo)
			w.SimPaused.Store(false) // unpause simulation now that time lapse has occurred
		}
	}

	if showingLoadScreen {
		// we don't allow actual map changes during load screens, but we do allow the above time lapse logic to execute.
		return
	}

	// Note: making this a separate variable since I don't want to control w.BlockPlayerChanges by dialog.
	// other places handle setting that variable (transitions, for example) so shouldn't touch it here.
	blockPlayerChanges := w.BlockPlayerChanges
	if w.ActiveMap.IsDialogActive() {
		blockPlayerChanges = true
	}

	w.ActiveMap.Update(blockPlayerChanges)

	// don't allow the player to do anything while map is showing screens.
	// One major reason to block this is, we don't want the E key to trigger the player menu to close from outside the player menu logic.
	// If that happens, then the screen doesn't save its changes.
	if w.ActiveMap.IsScreenShowing() {
		blockPlayerChanges = true
	}

	w.Player.Update(blockPlayerChanges)

	if !blockPlayerChanges && !w.ActiveMap.InScenario {
		// don't update time while player is in dialog or something where his in-map input is paused
		if w.Clock.Update() {
			// hour just changed
			_, h, _, _, _, _ := w.Clock.GetCurrentDateAndTime()
			w.OnHourChange(h, false, false, true)
		}
	}
}

func (w *World) Draw(screen *ebiten.Image) {
	w.ActiveMap.Draw(screen, w.OverlayManager)

	w.OverlayManager.Draw(screen)
}

func (w *World) startNpcSimulation() {
	logz.Println("SIMULATION", "Initializing NPC tasks...")

	if w.ActiveMap != nil {
		// guessing that if this is called while an active map is defined (and therefore, the player is in a world) then this would cause some
		// problems since it would affect NPCs that are potentially already in the active map doing things.
		// we should assume that this function is only called either at the beginning of the game startup, or while the active map is not defined (i.e. in a loading screen)
		logz.Panicln("SIMULATION", "tried to start NPC simulation, but it seems like the active map already exists. this could cause a problem with NPC task initialization.")
	}

	w.initializeNpcWorldState()

	logz.Println("SIMULATION", "Starting background NPC simulation...")
	go w.npcBackgroundSimulation()
}

// initializes current map and task for all NPCs based on their schedules; all previous tasks are cleared and reset.
// does not actually place NPCs into an active map (just sets map occupancies); that should be handled elsewhere.
func (w *World) initializeNpcWorldState() {
	gameTime := w.Clock.GetCurrentGameTime()

	for id, n := range w.NPCs {
		if n.CharacterStateRef.Temp {
			logz.Panicln("SIMULATION", "temp character found in NPC task initialization:", id)
		}
		// move NPCs to their scheduled map (they are created at first in their home maps where their beds are located)
		startMap := n.GetScheduledMap(gameTime)
		if startMap == "" {
			// No start map? strange...
			logz.Panicln("SIMULATION", "NPC didn't have a start map:", id)
		}
		if startMap != n.CharacterStateRef.CurrentMap {
			w.ChangeMapOccupancy(id, n.CharacterStateRef.CurrentMap, startMap, -1)
		}

		// clear any existing task from this NPC and set the one scheduled for the current hour
		n.CurrentTask = nil
		n.SetupTaskState(gameTime, nil)
	}
}

func (w *World) timeLapse(newTime clock.GameTime) {
	if !w.SimPaused.Load() {
		logz.Panicln("WORLD", "tried to do timelapse, but simulation was not paused yet.")
	}
	if !w.SimPauseEffected.Load() {
		logz.Panicln("WORLD", "tried to do timelapse, but simulation pause was not yet effected.")
	}

	logz.Println("WORLD", "applying time lapse to", newTime)

	currentTime := w.Clock.GetCurrentGameTime()
	if !newTime.IsAfter(currentTime) {
		logz.Panicln("WORLD", "attempted time lapse, but new time was not after current time. new time:", newTime, "current time:", currentTime)
	}

	w.Clock.SetGameTime(newTime)
	// sanity check
	if !w.Clock.GetCurrentGameTime().IsEqual(newTime) {
		logz.Panic("clock isn't set to the correct new time...")
	}

	// this handles figuring out which NPC's should be in which maps
	w.initializeNpcWorldState()

	// once we know which NPC's go in which maps, now we can:
	// - remove all NPC's from active map
	// - add in all the ones that are supposed to be there

	// Note: I guess we don't call CloseMap here since we don't want to set it to nil;
	// we just want to reset the NPC state
	for _, n := range w.ActiveMap.NPCs {
		n.PrepareLeaveActiveMap()
	}
	w.ActiveMap.NPCs = []*npc.NPC{}

	w.loadRegularMapNPCs()

	h := newTime.Hour
	// skip NPC check since NPC's already have their tasks initialized from initializeNpcWorldState
	// post event since some events may be scheduled for the future, and therefore will need to know the current time
	w.OnHourChange(h, true, true, true)
	w.EventBus.Publish(defs.Event{
		Type: pubsub.SysTimeLapse,
		Data: map[string]any{
			"time": newTime,
		},
	})

	w.AwaitingTimeLapse = false
	w.TimeLapseTo = nil
}

func (w *World) HandleMapDoor(result object.ObjectUpdateResult) {
	if result.ChangeMapID == "" {
		panic("tried to do map door change, but no map ID is set in object update result")
	}
	logz.Println("WORLD", "Handling map door:", result.ChangeMapID, result.ChangeMapSpawnIndex)
	w.EnterMap(result.ChangeMapID, result.ChangeMapSpawnIndex, true)
}

type SimCommand int

const (
	SimPause SimCommand = iota
	SimResume
	SimStop
)

func (w *World) npcBackgroundSimulation() {
	logz.Println("SIMULATION", "NPC background simulation started!")

	var lastTick time.Time
	tickSpeed := time.Second
	lastHour := w.Clock.GetCurrentGameTime().Hour

	for {
		// detect when simulation should be paused.
		if w.SimPaused.Load() {
			// once we've "noticed" the SimPaused flag, set this "sim pause effected" flag so that outside code knows for sure
			// that it is now safe to do things.
			w.SimPauseEffected.Store(true)
			continue
		}
		w.SimPauseEffected.Store(false)

		if w.ActiveMap == nil {
			// if active map is nil, either the first map hasn't been initialized yet, or something is happening to the active map.
			continue
		}
		if w.ActiveMap.InScenario {
			logz.Panicln("SIMULATION", "ActiveMap is in a scenario, but simulation is not paused.")
		}

		lastTick = time.Now()

		newHour := lastHour != w.Clock.GetCurrentGameTime().Hour
		lastHour = w.Clock.GetCurrentGameTime().Hour

		// Do simulation work
		midSimPause := false
		for _, n := range w.NPCs {
			if w.SimPaused.Load() {
				// a pause has occurred mid simulation loop; cancel all further processing of NPC tasks
				midSimPause = true
				break
			}
			if n.CharacterStateRef.CurrentMap == w.ActiveMap.MapID {
				// do not do simulation updates for NPCs in the active map
				continue
			}
			if newHour {
				// check if this NPC should change tasks or not
				n.OnHourChange(lastHour)
				continue
			}
			if n.CurrentTask == nil {
				// TODO: should a schedule task be launched for this NPC?
				continue
			}
			n.CurrentTask.SimulationUpdate()
		}

		if midSimPause {
			// skip the possible sleep below since the simulation was interrupted with a pause
			continue
		}

		tickElapsed := time.Since(lastTick)
		if tickElapsed < tickSpeed {
			time.Sleep(tickSpeed - tickElapsed)
		} else if tickElapsed > tickSpeed {
			logz.Warnln("SIMULATION", "simulation tick was longer than the expected duration:", tickElapsed)
		}
	}
}
