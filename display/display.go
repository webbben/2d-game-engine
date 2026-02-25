package display

import "github.com/hajimehoshi/ebiten/v2"

/*
Ebiten manages resizing and scaling on its own.
So, all we have to do is set an internal fixed screen width/height, and let ebiten do the rest.
*/

const (
	SCREEN_WIDTH  int = 1512 // Macbook pro's "effective" resolution
	SCREEN_HEIGHT int = 945

	// this looks really bad when scaled down even just a bit. so, lets shoot for a smaller size and scaling up for now.
	// SCREEN_WIDTH  int = 1920 // full HD "1080p"
	// SCREEN_HEIGHT int = 1080
)

// does all screen and window setup when starting the game
func SetupGameDisplay(windowTitle string, fullscreen bool) {
	ebiten.SetWindowTitle(windowTitle)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	ebiten.SetWindowSize(SCREEN_WIDTH, SCREEN_HEIGHT)
	ebiten.SetFullscreen(fullscreen)
}
