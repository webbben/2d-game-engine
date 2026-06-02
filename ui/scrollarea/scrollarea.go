// Package scrollarea gives a UI component that lets you scroll content within a fixed box
package scrollarea

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
)

type ScrollArea struct {
	height int

	canvas *ebiten.Image

	scrollOffsetY float64
	lastY         float64
}

type ScrollAreaParams struct {
	Width, Height int
}

func NewScrollArea(params ScrollAreaParams) ScrollArea {
	sa := ScrollArea{
		height: params.Height,
		canvas: ebiten.NewImage(params.Width, params.Height),
	}

	return sa
}

func (sa *ScrollArea) Update() {
	_, wheelY := ebiten.Wheel()

	sa.scrollOffsetY += wheelY * 80

	if sa.scrollOffsetY < 0 {
		sa.scrollOffsetY = 0
	}
	if sa.scrollOffsetY > float64(sa.height) {
		sa.scrollOffsetY = float64(sa.height)
	}
}

func (sa *ScrollArea) Draw(screen *ebiten.Image, content *ebiten.Image, x, y float64) {
	sa.canvas.Clear()

	sa.lastY = (sa.scrollOffsetY - sa.lastY) * 0.2
	rendering.DrawImage(sa.canvas, content, 0, sa.lastY, 0)
	rendering.DrawImage(screen, sa.canvas, x, y, 0)
}
