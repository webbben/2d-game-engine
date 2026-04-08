package object

import (
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/utils"
)

type ObjectUpdateResult struct {
	ChangeMapID         defs.MapID // if set, change to this map
	ChangeMapSpawnIndex int        // spawn point index to send player to
	UpdateOccurred      bool       // if true, an update happened to the object

	// For beds or other "usable" objects

	AlreadyInUse bool // if set, the object is being used by someone else already
}

// Update : blockChanges is just to make sure that no actual changes occur; but, animation changes and stuff can continue.
func (obj *Object) Update(blockChanges bool) ObjectUpdateResult {
	drawRect := obj.GetDrawRect()
	obj.MouseBehavior.Update(int(drawRect.X), int(drawRect.Y), int(drawRect.W), int(drawRect.H), false)

	if len(obj.tileData.Frames) > 1 {
		obj.tileData.UpdateFrame()
	}

	// handle hovering
	if obj.IsHoverable() {
		obj.PlayerHovering = false
		if obj.MouseBehavior.IsHovering {
			if utils.EuclideanDistCenter(obj.World.GetPlayerRect(), obj.GetRect()) < config.TileSize*2 {
				obj.PlayerHovering = true
			}
		}
	}

	switch obj.Type {
	case TypeDoor:
		// if a door is a step activation type, it can cause repeated door opening actions unless blocked.
		if !blockChanges {
			return obj.updateDoor()
		}
	case TypeGate:
		return obj.updateGate()
	}
	return ObjectUpdateResult{}
}

// ObjectActivationParams provides contextual information needed for the logic that handles activating an object.
type ObjectActivationParams struct {
	ActivatorID state.CharacterStateID // the character that is trying to activate the object
	LockIDs     []string               // locks the activator can unlock (has keys, etc)
}

func (obj *Object) Activate(fromX, fromY float64, params ObjectActivationParams) ObjectUpdateResult {
	if !obj.IsActivatable() {
		logz.Panicln("Activate", "tried to activate object that is not activatable")
	}

	logz.Println("Activate Object", "attempting to activate object:", obj.ID, obj.Type, "activated by:", params.ActivatorID)

	// check if object is locked, and if so, ensure the character has the keys
	if obj.lockID != "" {
		// check the lock state from the map state
		mapState := obj.dataman.GetMapState(obj.mapID)
		lockState := mapState.MapLocks[obj.lockID]
		if !lockState.Unlocked {
			// check if the NPC/player activating the door has the key
			if !slices.Contains(params.LockIDs, obj.lockID) {
				// door is locked; cannot enter
				// TODO: play a locked door sound effect
				return ObjectUpdateResult{}
			}
			// we have the key to the lock; so can proceed
		}
	}

	switch obj.Type {
	case TypeDoor:
		return obj.activateDoor(params)
	case TypeGate:
		return obj.activateGate(fromX, fromY)
	case TypeLight:
		return obj.activateLight()
	case TypeBed:
		return obj.activateBed(params)
	case TypeChair:
		return obj.activateChair(params)
	}
	return ObjectUpdateResult{}
}

func (obj *Object) activateDoor(params ObjectActivationParams) ObjectUpdateResult {
	if obj.Type != TypeDoor {
		panic("tried to activate as door, but object is not a door")
	}
	if obj.Door.TargetMapID == "" {
		panic("activated door does not have a target map ID!")
	}
	if obj.Door.TargetSpawnIndex < 0 {
		panic("activated door's target spawn index is negative! it should be a positive integer, including 0.")
	}
	if obj.Door.openSoundID == "" {
		panic("activated door's open sound is empty!")
	}

	// TODO: should we define volume somewhere?
	obj.AudioMgr.PlaySFX(obj.Door.openSoundID, 0.5)

	// check if this is the player, or an NPC
	// It doesn't actually change the logic here, since the calling code will handle it, but good to know.
	if params.ActivatorID == state.CharacterStateID(defs.PlayerID) {
		// player has activated the door
		logz.Println("activateDoor", "Player going to map:", obj.Door.TargetMapID)
	} else {
		logz.Println("activateDoor", "NPC going to map:", obj.Door.TargetMapID)
	}

	return ObjectUpdateResult{
		UpdateOccurred:      true,
		ChangeMapID:         obj.Door.TargetMapID,
		ChangeMapSpawnIndex: obj.Door.TargetSpawnIndex,
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
	if obj.World.GetPlayerRect().Intersects(obj.collisionRect) || obj.collidesWithEntityOrObject() {
		// don't allow gate to open if the player or any NPC is standing in its way
		return ObjectUpdateResult{}
	}

	// ensure the activation origin point (NPC/player position) isn't too far away

	obj.Gate.changingState = true
	obj.Gate.open = !obj.Gate.open
	if obj.Gate.openSFXID == "" {
		panic("gate has no open SFX set. make sure the 'SFX' property is set for this object in Tiled.")
	}
	obj.AudioMgr.PlaySFX(obj.Gate.openSFXID, 0.5)

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

func (obj *Object) updateGate() ObjectUpdateResult {
	if obj.Gate.changingState {
		if obj.nextFrame(obj.Gate.open) {
			obj.Gate.changingState = false
		}
		obj.PlayerHovering = false // don't show hover effect while gate is opening or closing
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
		if obj.World.GetPlayerRect().Intersects(obj.rect) {
			return obj.activateDoor(ObjectActivationParams{ActivatorID: state.CharacterStateID(defs.PlayerID)})
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
