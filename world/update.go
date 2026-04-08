package world

import (
	"time"

	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/object"
)

func (w *World) Update() {
	if w.ActiveMap == nil {
		// If ActiveMap is nil, we assume that the game world is not "active" (i.e. it's in a loading screen, or something)
		// So, we can just quit out of world updates in this case.
		// We don't want NPC's to be doing things in the background world simulation while the player is just waiting on a load screen.
		return
	}
	// when ActiveMap is defined, that means the player is actively in a map.
	// this also means that background world simulation can occur, since the game world at large is also "active".

	// Note: making this a separate variable since I don't want to control w.BlockPlayerChanges by dialog.
	// other places handle setting that variable (transitions, for example) so shouldn't touch it here.
	blockPlayerChanges := w.BlockPlayerChanges
	if w.ActiveMap.IsDialogActive() {
		blockPlayerChanges = true
	}
	w.Player.Update(blockPlayerChanges)
	w.ActiveMap.Update(blockPlayerChanges)

	if blockPlayerChanges {
		// also block NPC simulation if player changes are blocked.
	} else {
		// don't update time while player is in dialog or something where his in-map input is paused
		if w.Clock.Update() {
			// hour just changed
			_, h, _, _, _, _ := w.Clock.GetCurrentDateAndTime()
			w.OnHourChange(h, false, true)
		}
	}
}

func (w *World) startNpcSimulation() {
	w.simPaused = false
	w.simStopped = false
	logz.Println("SIMULATION", "Initializing NPC tasks...")

	if w.ActiveMap != nil {
		// guessing that if this is called while an active map is defined (and therefore, the player is in a world) then this would cause some
		// problems since it would affect NPCs that are potentially already in the active map doing things.
		// we should assume that this function is only called either at the beginning of the game startup, or while the active map is not defined (i.e. in a loading screen)
		logz.Panicln("SIMULATION", "tried to start NPC simulation, but it seems like the active map already exists. this could cause a problem with NPC task initialization.")
	}

	for id, n := range w.NPCs {
		if n.CharacterStateRef.Temp {
			// TODO: do temp characters exist yet? and if so, would they be in this NPCs map?
			logz.Panicln("SIMULATION", "temp character found in NPC task initialization:", id)
		}
		// move NPCs to their scheduled map (they are created at first in their home maps where their beds are located)
		startMap := n.GetScheduledMap(w.Clock.GetCurrentGameTime())
		if startMap == "" {
			// No start map? strange...
			logz.Warnln("SIMULATION", "NPC didn't have a start map:", id)
		} else {
			if startMap != n.CharacterStateRef.CurrentMap {
				w.ChangeMapOccupancy(id, n.CharacterStateRef.CurrentMap, startMap)
			}
		}
		n.SetupTaskState(w.Clock.GetCurrentGameTime(), nil)
	}

	logz.Println("SIMULATION", "Starting background NPC simulation...")
	go w.npcBackgroundSimulation(w.cmdCh)
}

func (w *World) pauseNpcSimulation() {
	if w.simPaused {
		// simulation should already be paused; don't send the command
		logz.Warnln("SIMULATION", "tried to pause NPC simulation, but simulation was already paused.")
		return
	}

	w.sendSimCmd(SimPause)
}

func (w *World) resumeNpcSimulation() {
	if !w.simPaused {
		// simulation isn't paused... don't send the command
		logz.Warnln("SIMULATION", "tried to resume NPC simulation, but simulation wasn't paused.")
	}

	w.sendSimCmd(SimResume)
}

func (w *World) stopNpcSimulation() {
	if w.simStopped {
		// simulation already stopped...
		logz.Warnln("SIMULATION", "tried to stop NPC simulation, but simulation was already stopped.")
	}

	w.sendSimCmd(SimStop)
}

func (w *World) sendSimCmd(cmd SimCommand) {
	select {
	case w.cmdCh <- cmd:
	// ok
	default:
		panic("uh oh")
	}
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

func (w *World) npcBackgroundSimulation(cmdCh <-chan SimCommand) {
	paused := false

	logz.Println("SIMULATION", "NPC background simulation started!")

	var lastTick time.Time
	tickSpeed := time.Second
	lastHour := w.Clock.GetCurrentGameTime().Hour

	for {
		select {
		case cmd := <-cmdCh:
			switch cmd {
			case SimPause:
				logz.Println("SIMULATION", "Simulation pausing...")
				paused = true
			case SimResume:
				logz.Panicln("SIMULATION", "a resume command came through, but the background simulation wasn't paused. is the logic sending these commands working correctly?")
				paused = false
			case SimStop:
				logz.Println("SIMULATION", "Simulation stopping!")
				return
			}

		default:
			if paused {
				// block until we get a command to resume
				cmd := <-cmdCh
				switch cmd {
				case SimResume:
					logz.Println("SIMULATION", "Simulation resuming...")
					paused = false
				case SimPause:
					logz.Panicln("SIMULATION", "a pause command came through, but the background simulation is already paused. is the logic sending these commands working correctly?")
				case SimStop:
					logz.Println("SIMULATION", "Simulation stopping!")
					return
				}
				continue
			}

			if w.ActiveMap == nil {
				// if active map is nil at the moment, then don't do any background simulations
				time.Sleep(time.Second)
				continue
			}

			lastTick = time.Now()

			newHour := lastHour != w.Clock.GetCurrentGameTime().Hour
			lastHour = w.Clock.GetCurrentGameTime().Hour

			// Do simulation work
			for _, n := range w.NPCs {
				if n.CharacterStateRef.CurrentMap == w.ActiveMap.MapID {
					// do not do simulation updates for NPCs in the active map
					continue
				}
				// TODO: do we need to use a mutex to ensure safety between the threads?
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

			tickElapsed := time.Since(lastTick)
			if tickElapsed < tickSpeed {
				time.Sleep(tickSpeed - tickElapsed)
			} else if tickElapsed > tickSpeed {
				logz.Warnln("SIMULATION", "simulation tick was longer than the expected duration:", tickElapsed)
			}
		}
	}
}
