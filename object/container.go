package object

import (
	"fmt"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/screen"
	"github.com/webbben/2d-game-engine/tiled"
)

type Container struct {
	open          bool
	changingState bool
	openSFXID     defs.SoundID
	closeSFXID    defs.SoundID
}

func (obj *Object) activateContainer() ObjectUpdateResult {
	if obj.Type != TypeContainer {
		panic("tried to activate container, but object is not a container")
	}
	if obj.Container.changingState {
		// can't open or close a gate if it's already changing state
		return ObjectUpdateResult{}
	}

	obj.Container.changingState = true
	obj.Container.open = !obj.Container.open

	if obj.Container.open {
		if obj.Container.openSFXID != "" {
			obj.AudioMgr.PlaySFX(obj.Container.openSFXID, 0.5)
		}
	} else {
		if obj.Container.closeSFXID != "" {
			obj.AudioMgr.PlaySFX(obj.Container.closeSFXID, 0.5)
		}
	}

	return ObjectUpdateResult{
		UpdateOccurred: true,
	}
}

func (obj *Object) updateContainer() ObjectUpdateResult {
	if obj.Container.changingState {
		if obj.nextFrame(obj.Container.open) {
			obj.Container.changingState = false
			if obj.Container.open {
				// send event to show open container screen
				obj.eventBus.Publish(defs.Event{
					Type: pubsub.SysShowScreen,
					Data: map[string]any{
						"screen_id": config.OpenContainerScreen,
						"params": screen.OpenContainerScreenParams{
							MapID:       obj.mapID,
							ContainerID: obj.ID,
							DisplayName: obj.DisplayName,
						},
					},
				})
			}
		}
		obj.PlayerHovering = false
	}

	return ObjectUpdateResult{}
}

const (
	// data: "obj_id" (int)
	ObjCommandCloseContainer defs.EventType = "obj_cmd_close_container"
)

func (obj *Object) loadContainerObject(props []tiled.Property) {
	if openSFX, found := tiled.GetStringProperty("sfx_open", props); found {
		obj.Container.openSFXID = defs.SoundID(openSFX)
	}
	if closeSFX, found := tiled.GetStringProperty("sfx_close", props); found {
		obj.Container.closeSFXID = defs.SoundID(closeSFX)
	}

	subID := fmt.Sprintf("%s_obj_%v", obj.mapID, obj.ID)
	obj.eventBus.Subscribe(subID, ObjCommandCloseContainer, obj.closeContainerCmd)
	obj.subIDs = append(obj.subIDs, subID)
}

func (obj *Object) closeContainerCmd(e defs.Event) {
	objID, ok := e.Data["obj_id"]
	if !ok {
		logz.Panicln("Object", "obj_id not present in closeContainer command event data.", e.Data)
	}
	if objID != obj.ID {
		// wrong container object
		return
	}

	if obj.Container.open {
		obj.Container.open = false
		obj.Container.changingState = true
		if obj.Container.closeSFXID != "" {
			obj.AudioMgr.PlaySFX(obj.Container.closeSFXID, 0.5)
		}
	}
}
