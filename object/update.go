package object

import (
	"log"

	"github.com/webbben/2d-game-engine/internal/config"
)

func (obj *Object) Update() {
	obj.MouseBehavior.Update(int(obj.DrawX), int(obj.DrawY), obj.Width, obj.Height)

	if obj.Type == TYPE_DOOR {
		if obj.MouseBehavior.LeftClick.ClickReleased {
			log.Println("going to room:", obj.Door.TargetMapID)
			obj.openSound.Play()
		}
	}
}

func (obj *Object) Draw(offsetX, offsetY float64) {
	obj.DrawX = (obj.X - offsetX) * config.GameScale
	obj.DrawY = (obj.Y - offsetY) * config.GameScale
}
