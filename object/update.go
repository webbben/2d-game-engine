package object

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/general_util"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

type ObjectUpdateResult struct {
	ChangeMapID         string // if set, change to this map
	ChangeMapSpawnIndex int    // spawn point index to send player to
	UpdateOccurred      bool   // if true, an update happened to the object
}

func (obj *Object) Update() ObjectUpdateResult {
	drawRect := obj.GetDrawRect()
	obj.MouseBehavior.Update(int(drawRect.X), int(drawRect.Y), int(drawRect.W), int(drawRect.H), false)

	if len(obj.tileData.Frames) > 1 {
		obj.tileData.UpdateFrame()
	}

	switch obj.Type {
	case TYPE_DOOR:
		return obj.updateDoor()
	case TYPE_GATE:
		return obj.updateGate()
	case TYPE_LIGHT:
		return obj.updateLight()
	case TYPE_CONTAINER:
		// TODO
		return ObjectUpdateResult{}
	case TYPE_SPAWN_POINT:
		// spawn points have no update logic
		return ObjectUpdateResult{}
	case TYPE_MISC:
		// misc has no update logic
		return ObjectUpdateResult{}
	default:
		// just putting this here so we always know objects in the map are of a recognized type; but not really necessary tbh.
		panic("unrecognized object type: " + obj.Type)
	}
}

func (obj *Object) Activate() ObjectUpdateResult {
	switch obj.Type {
	case TYPE_DOOR:
		return obj.activateDoor()
	case TYPE_GATE:
		return obj.activateGate()
	case TYPE_LIGHT:
		return obj.activateLight()
	}
	return ObjectUpdateResult{}
}

func (obj *Object) activateDoor() ObjectUpdateResult {
	if obj.Type != TYPE_DOOR {
		panic("tried to activate as door, but object is not a door")
	}
	if obj.Door.targetMapID == "" {
		panic("activated door does not have a target map ID!")
	}
	if obj.Door.targetSpawnIndex < 0 {
		panic("activated door's target spawn index is negative! it should be a positive integer, including 0.")
	}
	if obj.Door.openSound == nil {
		panic("activated door's open sound is nil!")
	}

	log.Println("going to room:", obj.Door.targetMapID)
	obj.Door.openSound.Play()

	return ObjectUpdateResult{
		UpdateOccurred:      true,
		ChangeMapID:         obj.Door.targetMapID,
		ChangeMapSpawnIndex: obj.Door.targetSpawnIndex,
	}
}

func (obj *Object) activateGate() ObjectUpdateResult {
	if obj.Type != TYPE_GATE {
		panic("tried to activate gate, but object is not a gate")
	}
	if obj.Gate.changingState {
		// can't open or close a gate if it's already changing state
		return ObjectUpdateResult{}
	}
	if obj.World.GetPlayerRect().Intersects(obj.CollisionRect) || obj.collidesWithNPC() {
		// don't allow gate to open if the player or any NPC is standing in its way
		return ObjectUpdateResult{}
	}

	obj.Gate.changingState = true
	obj.Gate.open = !obj.Gate.open
	if obj.Gate.openSFX == nil {
		panic("gate has no open SFX set. make sure the 'SFX' property is set for this object in Tiled.")
	}
	obj.Gate.openSFX.Play()

	return ObjectUpdateResult{
		UpdateOccurred: true,
	}
}

func (obj *Object) activateLight() ObjectUpdateResult {
	if obj.Type != TYPE_LIGHT {
		panic("tried to activate light, but object is not a light")
	}
	logz.Println("OBJECT", "light activated")
	obj.Light.On = !obj.Light.On
	return ObjectUpdateResult{UpdateOccurred: true}
}

func (obj *Object) updateLight() ObjectUpdateResult {
	obj.PlayerHovering = false
	if obj.MouseBehavior.IsHovering {
		if general_util.EuclideanDistCenter(obj.World.GetPlayerRect(), obj.GetRect()) < config.TileSize*2 {
			obj.PlayerHovering = true
		}
	}
	return ObjectUpdateResult{}
}

func (obj *Object) updateGate() ObjectUpdateResult {
	obj.PlayerHovering = false

	if obj.Gate.changingState {
		if obj.nextFrame(obj.Gate.open) {
			obj.Gate.changingState = false
		}
		return ObjectUpdateResult{}
	}

	if obj.MouseBehavior.IsHovering {
		if general_util.EuclideanDistCenter(obj.World.GetPlayerRect(), obj.CollisionRect) < config.TileSize*2 {
			obj.PlayerHovering = true
		}
	}

	return ObjectUpdateResult{}
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

func (obj *Object) updateDoor() ObjectUpdateResult {
	switch obj.Door.activateType {
	case "click":
		// do nothing - object clicks are detected and handled within mapInfo handler function
	case "step":
		if obj.World.GetPlayerRect().Intersects(obj.Rect) {
			return obj.activateDoor()
		}
	default:
		panic("invalid activation type for door")
	}
	return ObjectUpdateResult{}
}

func (obj *Object) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	obj.DrawX = (obj.xPos - offsetX) * config.GameScale
	obj.DrawY = (obj.yPos - offsetY) * config.GameScale

	if len(obj.imgFrames) == 0 {
		return
	}

	img := obj.tileData.CurrentFrame
	if obj.imgFrameIndex > 0 {
		img = obj.imgFrames[obj.imgFrameIndex]
	}
	ops := ebiten.DrawImageOptions{}
	if obj.PlayerHovering {
		ops.ColorScale.Scale(1.2, 1.2, 1.2, 1)
	}
	rendering.DrawImageWithOps(screen, img, obj.DrawX, obj.DrawY, config.GameScale, &ops)
}
