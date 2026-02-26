package object

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/general_util"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
)

type ObjectUpdateResult struct {
	ChangeMapID         defs.MapID // if set, change to this map
	ChangeMapSpawnIndex int        // spawn point index to send player to
	UpdateOccurred      bool       // if true, an update happened to the object
}

func (obj *Object) Update() ObjectUpdateResult {
	drawRect := obj.GetDrawRect()
	obj.MouseBehavior.Update(int(drawRect.X), int(drawRect.Y), int(drawRect.W), int(drawRect.H), false)

	if len(obj.tileData.Frames) > 1 {
		obj.tileData.UpdateFrame()
	}

	switch obj.Type {
	case TypeDoor:
		return obj.updateDoor()
	case TypeGate:
		return obj.updateGate()
	case TypeLight:
		return obj.updateLight()
	case TypeContainer:
		// TODO
		return ObjectUpdateResult{}
	case TypeSpawnPoint:
		// spawn points have no update logic
		return ObjectUpdateResult{}
	case TypeMisc:
		// misc has no update logic
		return ObjectUpdateResult{}
	case TypeItem:
		return ObjectUpdateResult{}
	default:
		// just putting this here so we always know objects in the map are of a recognized type; but not really necessary tbh.
		panic("unrecognized object type: " + obj.Type)
	}
}

func (obj *Object) Activate(fromX, fromY float64) ObjectUpdateResult {
	if !obj.IsActivatable() {
		logz.Panicln("Activate", "tried to activate object that is not activatable")
	}
	switch obj.Type {
	case TypeDoor:
		return obj.activateDoor()
	case TypeGate:
		return obj.activateGate(fromX, fromY)
	case TypeLight:
		return obj.activateLight()
	}
	return ObjectUpdateResult{}
}

func (obj *Object) activateDoor() ObjectUpdateResult {
	if obj.Type != TypeDoor {
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
	if obj.lockID != "" {
		// check the lock state from the map state
		mapState := obj.defMgr.GetMapState(obj.mapID)
		lockState := mapState.MapLocks[obj.lockID]
		if !lockState.Unlocked {
			// door is locked; cannot enter
			// TODO: play a locked door sound effect
			return ObjectUpdateResult{}
		}
	}

	log.Println("going to room:", obj.Door.targetMapID)
	obj.Door.openSound.Play()

	return ObjectUpdateResult{
		UpdateOccurred:      true,
		ChangeMapID:         obj.Door.targetMapID,
		ChangeMapSpawnIndex: obj.Door.targetSpawnIndex,
	}
}

func (obj *Object) activateGate(fromX, fromY float64) ObjectUpdateResult {
	if obj.Type != TypeGate {
		panic("tried to activate gate, but object is not a gate")
	}
	if obj.Gate.changingState {
		// can't open or close a gate if it's already changing state
		return ObjectUpdateResult{}
	}
	if obj.World.GetPlayerRect().Intersects(obj.CollisionRect) || obj.collidesWithEntityOrObject() {
		// don't allow gate to open if the player or any NPC is standing in its way
		return ObjectUpdateResult{}
	}

	// ensure the activation origin point (NPC/player position) isn't too far away

	obj.Gate.changingState = true
	obj.Gate.open = !obj.Gate.open
	if obj.Gate.openSFXID == "" {
		panic("gate has no open SFX set. make sure the 'SFX' property is set for this object in Tiled.")
	}
	obj.AudioMgr.PlaySFX(obj.Gate.openSFXID, 0.2)

	return ObjectUpdateResult{
		UpdateOccurred: true,
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

func (obj Object) collidesWithEntityOrObject() bool {
	return obj.World.RectCollidesWithOthers(obj.GetRect(), "", obj.ID)
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
