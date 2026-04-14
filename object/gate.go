package object

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/tiled"
)

type Gate struct {
	open          bool
	changingState bool
	openSFXID     defs.SoundID
}

func (g Gate) IsOpen() bool {
	return g.open && !g.changingState
}

func (obj *Object) loadGateObject(props []tiled.Property) {
	for _, prop := range props {
		switch prop.Name {
		case "SFX":
			gateSoundID := defs.SoundID(prop.GetStringValue())
			if gateSoundID == "" {
				panic("no gate sound ID found. TODO: should we make a default one?")
			}
			obj.Gate.openSFXID = gateSoundID
		}
	}

	if obj.Gate.openSFXID == "" {
		panic("no open SFX ID set for gate. make sure to set the 'SFX' property for this object in Tiled.")
	}
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
