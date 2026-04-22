package game

import (
	"time"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/internal/debug"
	"github.com/webbben/2d-game-engine/logz"
)

func (g *Game) Update() error {
	if g.GlobalKeyBindings != nil {
		g.handleGlobalKeyBindings()
	}

	if g.globalHud != nil {
		g.globalHud.Update(g)
	}

	g.EventBus.ProcessEvents()

	if g.TransitionManager.TransitionInProgress {
		g.TransitionManager.Update(g)
		if g.TransitionManager.ShowingLoadingScreen {
			// if we are showing a loading screen, then no other content should get updates.
			// we do allow updates during opening and closing transitions though, since we don't want things to look oddly "frozen"
			// and snap into update only once the transition has fully finished. that makes it look weird.
			return nil
		}
		if !g.TransitionManager.TransitionInProgress {
			// transition just ended. resume letting player do things in game map
			g.World.BlockPlayerChanges = false
			g.World.SimPaused.Store(false)
		}
	}

	switch g.gameStage {
	case MainMenu:
		if g.MainMenu == nil {
			// No main menu, so just jump straight into the game world
			g.gameStage = InGameWorld
			logz.Warnln("UPDATE", "No main menu found; jumping to InGameWorld...")
			return nil
		}
		g.mainMenuViewer.Update()
		if g.mainMenuViewer.IsDone() {
			if !g.TransitionManager.TransitionInProgress {
				// TODO: When a screen ends, it might be launching a loading screen next. So, we can't switch directly to InGameWorld here.
				// I've wondered if we should have a "Loading" game stage, but, usually when loading we want to keep the existing screen there long enough (at least)
				// for the transitions to finish fading in/out. So, I don't think we want to switch directly into a new game stage.
				// It seems to me like we should rely on the screens (or whoever is calling a transition) to handle switching the game stage, and not assume here.
				// So, I wonder if we should add some validation in here to ensure that the loading process elsewhere is indeed changing the game stage - so we dont' get stuck.
			}
		}
	case InGameWorld:
		if g.World == nil {
			logz.Panicln("UPDATE", "World has not been initialized yet! Ensure this happens before InGameWorld stage; The Main Menu should be sure to handle this before it is 'done'.")
		}

		// TODO: turn these menus and things into Screens
		if g.ShowPlayerMenu {
			// set last player update to now, so that the time hud doesn't immediately display
			g.World.Player.LastUserInput = time.Now()
			g.playerMenuViewer.Update()
			if g.playerMenuViewer.IsDone() {
				g.ShowPlayerMenu = false
			}
		} else {
			g.World.Update()
		}

		g.hud.Update(g)
	default:
		logz.Panicln("UPDATE", "Game stage was invalid! Game stage value:", g.gameStage)
	}

	if config.TrackMemoryUsage {
		debug.UpdatePerformanceMetrics()
	}

	return nil
}
