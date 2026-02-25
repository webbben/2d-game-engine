package textwindow

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/ui/textbox"
)

// HoverTooltip is a simple component that shows a small amount of text when the user is hovering over something
type HoverTooltip struct {
	textBox textbox.TextBox
	mouse.MouseBehavior
	mouseOffsetX, mouseOffsetY int
	msDelay                    int
}

func NewHoverTooltip(s string, tilesetSrc string, originIndex int, msDelay int, mouseOffsetX, mouseOffsetY int) HoverTooltip {
	if s == "" {
		panic("text is empty")
	}

	hoverTooltip := HoverTooltip{
		textBox:      textbox.NewTextBox(s, tilesetSrc, originIndex, config.DefaultFont, nil, nil),
		mouseOffsetX: mouseOffsetX,
		mouseOffsetY: mouseOffsetY,
		msDelay:      msDelay,
	}

	return hoverTooltip
}

func (ht *HoverTooltip) Update(x, y float64, width, height int) {
	ht.MouseBehavior.Update(int(x), int(y), width, height, false)
}

func (ht *HoverTooltip) Draw(om *overlay.OverlayManager) {
	if !ht.IsHovering {
		return
	}
	if time.Since(ht.HoverStart) < (time.Millisecond * time.Duration(ht.msDelay)) {
		return
	}
	if om == nil {
		panic("draw hover tooltip: overlay manager is nil")
	}

	// draw next to the mouse
	mouseX, mouseY := ebiten.CursorPosition()
	om.AddOverlay(ht.textBox.GetImage(), float64(mouseX+ht.mouseOffsetX), float64(mouseY+ht.mouseOffsetY))
}
