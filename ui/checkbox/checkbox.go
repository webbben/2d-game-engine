// Package checkbox
package checkbox

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/mouse"
	"github.com/webbben/2d-game-engine/tiled"
)

type Checkbox struct {
	uncheckedImg  *ebiten.Image
	checkedImg    *ebiten.Image
	mouseBehavior mouse.MouseBehavior

	checked bool

	x, y int
	w, h int
}

type CheckboxParams struct {
	TilesetSrc                   string
	UncheckedIndex, CheckedIndex int
}

func NewCheckbox(params CheckboxParams) Checkbox {
	cb := Checkbox{
		uncheckedImg: tiled.GetTileImage(params.TilesetSrc, params.UncheckedIndex, true),
		checkedImg:   tiled.GetTileImage(params.TilesetSrc, params.CheckedIndex, true),
	}

	cb.w = cb.uncheckedImg.Bounds().Dx()
	cb.h = cb.uncheckedImg.Bounds().Dy()

	return cb
}

func (cb *Checkbox) Update() {
	cb.mouseBehavior.Update(cb.x, cb.y, int(float64(cb.w)*config.UIScale), int(float64(cb.h)*config.UIScale), false)

	if cb.mouseBehavior.LeftClick.ClickReleased {
		cb.checked = !cb.checked
	}
}

func (cb Checkbox) IsChecked() bool {
	return cb.checked
}

func (cb *Checkbox) Draw(screen *ebiten.Image, x, y float64) {
	cb.x = int(x)
	cb.y = int(y)

	var img *ebiten.Image
	if cb.checked {
		img = cb.checkedImg
	} else {
		img = cb.uncheckedImg
	}

	rendering.DrawImage(screen, img, x, y, config.UIScale)
}
