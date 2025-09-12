package object

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
)

func (o Object) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	op := ebiten.DrawImageOptions{}
	if o.CanSeeBehind {
		o.fade.SetDrawOps(&op)
	}
	drawX, drawY := o.DrawPos(offsetX, offsetY)
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(o.img, &op)
}

func (o *Object) Update() {
	if o.CanSeeBehind {
		if o.WorldContext.PlayerIsBehindObject(*o) {
			o.fade.TargetScale = 0.5
		} else {
			o.fade.TargetScale = 1
		}
		o.fade.Update()
	}
}
