package object

import (
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/tiled"
)

type Window struct {
	Light *lights.WindowLight
}

func (obj *Object) loadWindowObject(props []tiled.Property) {
	windowProps := tiled.GetWindowProps(props)

	l := lights.NewWindowFromTiledProps(int(obj.xPos+float64(obj.Width/2)), int(obj.yPos), windowProps)

	obj.Window = Window{
		Light: &l,
	}
}
