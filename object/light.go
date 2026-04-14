package object

import (
	"github.com/webbben/2d-game-engine/internal/lights"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/tiled"
)

type Light struct {
	Light *lights.Light
	On    bool
}

func (obj *Object) loadLightObject(props []tiled.Property) {
	lightProps := tiled.GetLightProps(props)

	l := lights.NewLight(int(obj.xPos+float64(obj.Width/2)), int(obj.yPos+float64(obj.Height/2)), lightProps, nil)

	obj.Light = Light{
		Light: &l,
		On:    true,
	}
}

func (obj *Object) activateLight() ObjectUpdateResult {
	if obj.Type != TypeLight {
		panic("tried to activate light, but object is not a light")
	}
	logz.Println("OBJECT", "light activated")
	obj.Light.On = !obj.Light.On
	return ObjectUpdateResult{UpdateOccurred: true}
}
