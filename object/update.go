package object

import (
	"log"

	"github.com/webbben/2d-game-engine/internal/config"
)

type ObjectUpdateResult struct {
	ChangeMapID         string // if set, change to this map
	ChangeMapSpawnIndex int    // spawn point index to send player to
}

func (obj *Object) Update() ObjectUpdateResult {
	result := ObjectUpdateResult{}
	obj.MouseBehavior.Update(int(obj.DrawX), int(obj.DrawY), obj.Width, obj.Height, true)

	if obj.Type == TYPE_DOOR {
		if obj.MouseBehavior.LeftClick.ClickReleased {
			log.Println("going to room:", obj.Door.TargetMapID)
			obj.openSound.Play()
			result.ChangeMapID = obj.Door.TargetMapID
			result.ChangeMapSpawnIndex = obj.SpawnIndex
		}
	}

	return result
}

func (obj *Object) Draw(offsetX, offsetY float64) {
	obj.DrawX = (obj.X - offsetX) * config.GameScale
	obj.DrawY = (obj.Y - offsetY) * config.GameScale
}
