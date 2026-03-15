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
			g.gameStage = InGameWorld
		}
	case InGameWorld:
		if g.World == nil {
			logz.Panicln("UPDATE", "World has not been initialized yet! Ensure this happens before InGameWorld stage; The Main Menu should be sure to handle this before it is 'done'.")
		}
		if g.dialogSession != nil {
			g.dialogSession.Update()
			if g.dialogSession.Exit {
				g.dialogSession = nil
			}
			// set last player update to now, so that the time hud doesn't immediately display
			g.World.Player.LastUserInput = time.Now()
		} else if g.ShowPlayerMenu {
			g.PlayerMenu.Update()
			// set last player update to now, so that the time hud doesn't immediately display
			g.World.Player.LastUserInput = time.Now()
		} else if g.ShowTradeScreen {
			g.TradeScreen.Update()
			if g.TradeScreen.Exit {
				g.ShowTradeScreen = false
			}
			// set last player update to now, so that the time hud doesn't immediately display
			g.World.Player.LastUserInput = time.Now()
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
