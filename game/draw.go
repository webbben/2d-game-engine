package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/logz"
)

func (g *Game) Draw(screen *ebiten.Image) {
	// if a loading screen is actively showing, don't show anything else.
	// we only show content at the same time as a transition if its opening or closing transition is running.
	if !g.TransitionManager.ShowingLoadingScreen {
		switch g.gameStage {
		case MainMenu:
			g.mainMenuViewer.Draw(screen)
		case InGameWorld:
			if g.World == nil || g.World.ActiveMap == nil {
				logz.Panicln("DRAW", "game stage is InGameWorld, but no map info exists...")
			}
			g.drawWorld(screen)

			// show game debug info, including this scale info
			if config.ShowGameDebugInfo {
				g.showGameDebugInfo(screen)
			}
		}
	}

	if g.TransitionManager.TransitionInProgress {
		g.TransitionManager.Draw(screen)
	}

	if g.globalHud != nil {
		g.globalHud.Draw(screen)
	}
}

func (g *Game) drawWorld(screen *ebiten.Image) {
	g.World.Draw(screen)

	// TODO: move to World?
	if g.hud != nil {
		g.hud.Draw(screen)
	}
}
