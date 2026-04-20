package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/ui/overlay"
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
			g.drawWorld(screen, g.OverlayManager)

			// show game debug info, including this scale info
			if config.ShowGameDebugInfo {
				g.showGameDebugInfo(screen)
			}
		}
	}

	if g.TransitionManager.TransitionInProgress {
		g.TransitionManager.Draw(screen)
	}
}

func (g *Game) drawWorld(screen *ebiten.Image, om *overlay.OverlayManager) {
	g.World.ActiveMap.Draw(screen, om)

	g.OverlayManager.Draw(screen)

	// TODO: at some point, logic for showing screens should be fully handled by an action that is triggered from the game project, not here.
	if g.ShowPlayerMenu {
		g.playerMenuViewer.Draw(screen)
	}

	if g.hud != nil {
		g.hud.Draw(screen)
	}
}
