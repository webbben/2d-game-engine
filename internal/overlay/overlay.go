// Package overlay provides a UI utility for drawing overlays on top of the rest of the UI
package overlay

/*
* Usage:
*
* There is no Update function; all the work is done in Draw.
* In a Draw function, take the image that you want to draw as an overlay and pass it to the overlay manager's AddOverlay function.
* Then, at the end of the main game's draw function, call this overlay manager's draw function so that all overlays are drawn on top of everything else.
 */

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

type overlayImage struct {
	img  *ebiten.Image
	x, y float64
}

type OverlayManager struct {
	overlays []overlayImage
}

func (om *OverlayManager) AddOverlay(img *ebiten.Image, x, y float64) {
	om.overlays = append(om.overlays, overlayImage{
		img: img,
		x:   x,
		y:   y,
	})
}

func (om *OverlayManager) Draw(screen *ebiten.Image) {
	for _, overlay := range om.overlays {
		rendering.DrawImage(screen, overlay.img, overlay.x, overlay.y, 0)
	}
	om.Clear()
}

func (om *OverlayManager) Clear() {
	if len(om.overlays) == 0 {
		return
	}
	om.overlays = make([]overlayImage, 0)
}
