package textwindow

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/display"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/ui/overlay"
)

// HoverWindow is a simple hover window that shows a title and body text
type HoverWindow struct {
	placeHolderImg *ebiten.Image
	textWindow
	mouse.MouseBehavior
}

func NewHoverWindow(title, bodyText string, textWindowParams TextWindowParams) HoverWindow {
	if title == "" {
		panic("title is empty")
	}
	if bodyText == "" {
		panic("body text is empty")
	}
	hoverWindow := HoverWindow{
		textWindow: newTextWindow(title, bodyText, textWindowParams),
	}
	w, h := hoverWindow.Dimensions()
	hoverWindow.placeHolderImg = ebiten.NewImage(w, h)

	return hoverWindow
}

func (hw *HoverWindow) Update(x, y float64, width, height int) {
	hw.MouseBehavior.Update(int(x), int(y), width, height, false)

	if hw.IsHovering {
		hw.textWindow.Update()
	}
}

func (hw *HoverWindow) Draw(om *overlay.OverlayManager) {
	if !hw.IsHovering {
		return
	}
	hw.placeHolderImg.Clear()

	// capture draw result from text window ui component
	hw.textWindow.Draw(hw.placeHolderImg, 0, 0)

	dx, dy := hw.Dimensions()
	x, y := getPosNearMouse(15, dx, dy)
	om.AddOverlay(hw.placeHolderImg, float64(x), float64(y))
}

func getPosNearMouse(distFromMouse int, dx, dy int) (x, y int) {
	// draw next to the mouse
	mouseX, mouseY := ebiten.CursorPosition()
	// make sure the window doesn't go off screen
	x = mouseX + distFromMouse
	y = mouseY + distFromMouse
	if x+dx > display.SCREEN_WIDTH {
		x = mouseX - distFromMouse - dx
	}
	if y+dy > display.SCREEN_HEIGHT {
		y = mouseY - distFromMouse - dy
	}

	return x, y
}
