package object

import (
	"log"
)

type ObjectUpdateResult struct {
	ChangeMapID         string // if set, change to this map
	ChangeMapSpawnIndex int    // spawn point index to send player to
}

func (obj *Object) Update() ObjectUpdateResult {
	result := ObjectUpdateResult{}
	obj.MouseBehavior.Update(int(obj.DrawX), int(obj.DrawY), obj.Width, obj.Height, true)

	if obj.Type == TYPE_DOOR {
		switch obj.activateType {
		case "click":
			if obj.MouseBehavior.LeftClick.ClickReleased {
				log.Println("going to room:", obj.Door.targetMapID)
				obj.openSound.Play()
				result.ChangeMapID = obj.Door.targetMapID
				result.ChangeMapSpawnIndex = obj.Door.targetSpawnIndex
			}
		case "step":
			if obj.WorldContext.GetPlayerRect().Intersects(obj.Rect) {
				log.Println("going to room:", obj.Door.targetMapID)
				obj.openSound.Play()
				result.ChangeMapID = obj.Door.targetMapID
				result.ChangeMapSpawnIndex = obj.Door.targetSpawnIndex
			}
		default:
			panic("invalid activation type for door")
		}
	}

	return result
}

func (obj *Object) Draw(offsetX, offsetY float64) {
	obj.DrawX = (obj.X - offsetX)
	obj.DrawY = (obj.Y - offsetY)
}
