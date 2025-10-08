package overlay

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
	om.overlays = make([]overlayImage, 0)
}
