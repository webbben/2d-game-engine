package ui

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/mouse"
	"github.com/webbben/2d-game-engine/internal/overlay"
	"github.com/webbben/2d-game-engine/internal/text"
)

type HoverTooltip struct {
	box      BoxDef
	boxImage *ebiten.Image
	mouse.MouseBehavior
	mouseOffsetX, mouseOffsetY int
	msDelay                    int
}

func NewHoverTooltip(s string, tilesetSrc string, originIndex int, msDelay int, mouseOffsetX, mouseOffsetY int) HoverTooltip {
	if s == "" {
		panic("text is empty")
	}

	hoverTooltip := HoverTooltip{
		box:          NewBox(tilesetSrc, originIndex),
		mouseOffsetX: mouseOffsetX,
		mouseOffsetY: mouseOffsetY,
		msDelay:      msDelay,
	}

	tileSize := int(config.TileSize * config.UIScale)

	height := tileSize * 2
	width, _, _ := text.GetStringSize(s, config.DefaultFont)
	width += int(2.5 * float64(tileSize))
	width -= width % tileSize

	hoverTooltip.boxImage = hoverTooltip.box.BuildBoxImage(width, height)

	sx, _, _ := text.GetStringSize(s, config.DefaultFont)
	sy, _ := text.GetRealisticFontMetrics(config.DefaultFont)
	textX := (width / 2) - (sx / 2)
	textY := (height / 2) + (sy / 2)

	text.DrawShadowText(hoverTooltip.boxImage, s, config.DefaultFont, textX, textY, nil, nil, 0, 0)

	return hoverTooltip
}

func (ht *HoverTooltip) Update(x, y float64, width, height int) {
	ht.MouseBehavior.Update(int(x), int(y), width, height, false)
}

func (ht *HoverTooltip) Draw(om *overlay.OverlayManager) {
	if !ht.MouseBehavior.IsHovering {
		return
	}
	if time.Since(ht.MouseBehavior.HoverStart) < (time.Millisecond * time.Duration(ht.msDelay)) {
		return
	}

	// draw next to the mouse
	mouseX, mouseY := ebiten.CursorPosition()
	om.AddOverlay(ht.boxImage, float64(mouseX+ht.mouseOffsetX), float64(mouseY+ht.mouseOffsetY))
}
