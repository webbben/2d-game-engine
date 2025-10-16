package object

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

type ObjectUpdateResult struct {
	ChangeMapID         string // if set, change to this map
	ChangeMapSpawnIndex int    // spawn point index to send player to
}

func (obj *Object) Update() ObjectUpdateResult {
	result := ObjectUpdateResult{}
	obj.MouseBehavior.Update(int(obj.DrawX), int(obj.DrawY), obj.Width, obj.Height, true)

	switch obj.Type {
	case TYPE_DOOR:
		obj.updateDoor(&result)
	case TYPE_GATE:
		obj.updateGate(&result)
	}

	return result
}

func (obj *Object) updateGate(result *ObjectUpdateResult) {
	obj.PlayerHovering = false

	if obj.MouseBehavior.LeftClick.ClickReleased || obj.MouseBehavior.IsHovering {
		if !obj.Gate.changingState {
			if general_util.EuclideanDistCenter(obj.World.GetPlayerRect(), obj.CollisionRect) < config.TileSize*2 {
				obj.PlayerHovering = true
				if obj.MouseBehavior.LeftClick.ClickReleased {
					if !obj.World.GetPlayerRect().Intersects(obj.CollisionRect) && !obj.collidesWithNPC() {
						obj.Gate.changingState = true
						obj.Gate.open = !obj.Gate.open
					}
				}
			}
		}
	}

	if obj.Gate.changingState {
		if obj.nextFrame(obj.Gate.open) {
			obj.Gate.changingState = false
		}
	}
}

func (obj Object) collidesWithNPC() bool {
	for _, n := range obj.World.GetNearbyNPCs(
		obj.CollisionRect.X+(obj.CollisionRect.W/2), // use center of collision rect
		obj.CollisionRect.Y+(obj.CollisionRect.H/2),
		obj.CollisionRect.W+obj.CollisionRect.H,
	) {
		if obj.CollisionRect.Intersects(n.Entity.CollisionRect()) {
			return true
		}
	}
	return false
}

func (obj *Object) nextFrame(forwards bool) (done bool) {
	if time.Since(obj.animLastUpdate) < time.Millisecond*time.Duration(obj.animSpeedMs) {
		return false
	}
	obj.animLastUpdate = time.Now()

	if forwards {
		if obj.imgFrameIndex == len(obj.imgFrames)-1 {
			return true
		}
		obj.imgFrameIndex++
	} else {
		if obj.imgFrameIndex == 0 {
			return true
		}
		obj.imgFrameIndex--
	}

	if obj.imgFrameIndex < 0 || obj.imgFrameIndex > len(obj.imgFrames)-1 {
		panic("imgFrameIndex out of range")
	}

	return false
}

func (obj *Object) updateDoor(result *ObjectUpdateResult) {
	switch obj.Door.activateType {
	case "click":
		if obj.MouseBehavior.LeftClick.ClickReleased {
			log.Println("going to room:", obj.Door.targetMapID)
			obj.Door.openSound.Play()
			result.ChangeMapID = obj.Door.targetMapID
			result.ChangeMapSpawnIndex = obj.Door.targetSpawnIndex
		}
	case "step":
		if obj.World.GetPlayerRect().Intersects(obj.Rect) {
			log.Println("going to room:", obj.Door.targetMapID)
			obj.Door.openSound.Play()
			result.ChangeMapID = obj.Door.targetMapID
			result.ChangeMapSpawnIndex = obj.Door.targetSpawnIndex
		}
	default:
		panic("invalid activation type for door")
	}
}

func (obj *Object) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	obj.DrawX = (obj.xPos - offsetX)
	obj.DrawY = (obj.yPos - offsetY)

	if len(obj.imgFrames) > 0 {
		ops := ebiten.DrawImageOptions{}
		if obj.PlayerHovering {
			ops.ColorScale.Scale(1.2, 1.2, 1.2, 1)
		}
		rendering.DrawImageWithOps(screen, obj.imgFrames[obj.imgFrameIndex], obj.DrawX*config.GameScale, obj.DrawY*config.GameScale, config.GameScale, &ops)
	}
}
