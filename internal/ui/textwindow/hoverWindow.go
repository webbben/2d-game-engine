package textwindow

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/display"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/overlay"
)

type HoverWindow struct {
	placeHolderImg *ebiten.Image
	TextWindow
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
		TextWindow: NewTextWindow(title, bodyText, textWindowParams),
	}
	w, h := hoverWindow.TextWindow.Dimensions()
	hoverWindow.placeHolderImg = ebiten.NewImage(w, h)

	return hoverWindow
}

func (hw *HoverWindow) Update(x, y float64, width, height int) {
	hw.MouseBehavior.Update(int(x), int(y), width, height, false)

	if hw.MouseBehavior.IsHovering {
		hw.TextWindow.Update()
	}
}

func (hw *HoverWindow) Draw(om *overlay.OverlayManager) {
	if !hw.MouseBehavior.IsHovering {
		return
	}
	hw.placeHolderImg.Clear()

	// capture draw result from text window ui component
	hw.TextWindow.Draw(hw.placeHolderImg, 0, 0)

	// draw next to the mouse
	mouseX, mouseY := ebiten.CursorPosition()
	// make sure the window doesn't go off screen
	x := mouseX + 15
	y := mouseY + 15
	dx, dy := hw.Dimensions()
	if x+dx > display.SCREEN_WIDTH {
		x = mouseX - 15 - dx
	}
	if y+dy > display.SCREEN_HEIGHT {
		y = mouseY - 15 - dy
	}
	om.AddOverlay(hw.placeHolderImg, float64(x), float64(y))
}
