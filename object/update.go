package object

import (
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
)

type ObjectUpdateResult struct {
	ChangeMapID         defs.MapID // if set, change to this map
	ChangeMapSpawnIndex int        // spawn point index to send player to
	UpdateOccurred      bool       // if true, an update happened to the object

	// For beds or other "usable" objects

	AlreadyInUse bool // if set, the object is being used by someone else already
}

// Update : blockChanges is just to make sure that no actual changes occur; but, animation changes and stuff can continue.
func (obj *Object) Update(blockChanges bool, hovering bool) ObjectUpdateResult {
	if len(obj.tileData.Frames) > 1 {
		obj.tileData.UpdateFrame()
	}

	obj.PlayerHovering = hovering
	if obj.PlayerHovering && !obj.IsHoverable() {
		logz.Panicln("Object", "object is set to PlayerHovering, but is apparently not hoverable...", obj.ID, obj.Name)
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
	ActivatorID id.CharacterStateID // the character that is trying to activate the object
	LockIDs     []string            // locks the activator can unlock (has keys, etc)
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

	obj.World.PublishEvent(defs.Event{
		Type: pubsub.EventObjectActivated,
		Data: map[string]any{
			pubsub.DataKey: pubsub.EventObjectActivatedData{
				ObjectType:  string(obj.Type),
				ActivatorID: string(params.ActivatorID),
			},
		},
	})

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
