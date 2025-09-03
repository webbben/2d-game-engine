package display

import "github.com/hajimehoshi/ebiten/v2"

const (
	default_window_width  = 1280
	default_window_height = 720
)

var (
	isFullscreen bool = false
	screenWidth  int
	screenHeight int
)

// does all screen and window setup when starting the game
func SetupGameDisplay(windowTitle string, fullscreen bool) {
	ebiten.SetWindowTitle(windowTitle)
	SetFullscreen(fullscreen)
}

func GetScreenSize() (int, int) {
	return screenWidth, screenHeight
}

func ScreenWidth() int {
	return screenWidth
}

func ScreenHeight() int {
	return screenHeight
}

func IsFullscreen() bool {
	return isFullscreen
}

func SetFullscreen(isFullscreen bool) {
	if isFullscreen {
		w, h := ebiten.ScreenSizeInFullscreen()
		screenWidth = w
		screenHeight = h
		ebiten.SetFullscreen(true)
		isFullscreen = true
	} else {
		screenWidth = default_window_width
		screenHeight = default_window_height
		ebiten.SetWindowSize(default_window_width, default_window_height)
		ebiten.SetFullscreen(false)
		isFullscreen = false
	}
}
