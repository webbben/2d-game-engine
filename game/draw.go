package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/ui/overlay"
)

func (g *Game) Draw(screen *ebiten.Image) {
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

func (g *Game) drawWorld(screen *ebiten.Image, om *overlay.OverlayManager) {
	g.World.ActiveMap.Draw(screen, om)

	g.OverlayManager.Draw(screen)

	// TODO: these should probably be turned into Screens, right?
	if g.dialogSession != nil {
		g.dialogSession.Draw(screen)
	} else if g.ShowPlayerMenu {
		g.PlayerMenu.Draw(screen, om)
	} else if g.ShowTradeScreen {
		g.TradeScreen.Draw(screen, om)
	}

	if g.hud != nil {
		g.hud.Draw(screen)
	}
}
