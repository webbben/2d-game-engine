package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/utils"
)

func calculateVisibility(observer, target entity.EntityInfo, lightIntensity float32) float64 {
	dist := utils.EuclideanDistCoords(observer.TilePos, target.TilePos)
	if dist >= SightDist {
		return 0
	}
	distanceFactor := 1 - dist/SightDist

	// facing cone (~90 degrees): dot product of normalized to-target vector vs facing direction vector
	toTarget := model.Vec2{
		X: float64(target.TilePos.X - observer.TilePos.X),
		Y: float64(target.TilePos.Y - observer.TilePos.Y),
	}
	if toTarget.Len() == 0 {
		// same tile; trivially visible
		return 1
	}

	// basic formula for visibility is distance * brightness of light
	visibility := distanceFactor * float64(lightIntensity)

	norm := toTarget.Normalize()
	var forward model.Vec2
	switch observer.Direction {
	case 'R':
		forward = model.Vec2{X: 1, Y: 0}
	case 'L':
		forward = model.Vec2{X: -1, Y: 0}
	case 'U':
		forward = model.Vec2{X: 0, Y: -1}
	case 'D':
		forward = model.Vec2{X: 0, Y: 1}
	default:
		logz.Panic("invalid direction")
	}

	// check if target is outside of vision cone
	// cos(45 deg) ~= 0.7071: within 90 deg cone iff dot > 0.7071
	if (norm.X*forward.X)+(norm.Y*forward.Y) <= 0.7071 {
		// if outside of the main vision cone, reduce visibility further
		visibility *= 0.5
	}

	// sneak multiplier
	// TODO: actually factor in sneak skill here
	if target.Sneaking {
		visibility *= 0.5
	}

	if visibility < 0 {
		visibility = 0
	}
	return visibility
}

func (n *NPC) updateVisibility() {
	n.initialPlayerSightingThisTick = false

	observer := n.Entity.GetEntityInfo()

	// check if player is visible
	playerInfo := n.ActiveMapCtx.GetEntityInfo(id.PlayerStateID)
	lightIntensity := n.ActiveMapCtx.GetLightIntensity(playerInfo.TilePos)
	vis := calculateVisibility(observer, playerInfo, lightIntensity)
	if vis > 0 {
		if _, alreadySeen := n.visibleEntities[playerInfo.ID]; !alreadySeen {
			n.visibleEntities[playerInfo.ID] = time.Now()
			n.eventBus.Publish(defs.Event{
				Type: pubsub.EventEntitySpotted,
				Data: map[string]any{
					"observer": observer.ID,
					"target":   playerInfo.ID,
				},
			})
		}
		if !n.hasSeenPlayerYet {
			n.initialPlayerSightingThisTick = true
			logz.Println(n.ID(), "first player sighting")
		}
		n.hasSeenPlayerYet = true
		n.lastPlayerSightingTime = time.Now()
	} else {
		delete(n.visibleEntities, playerInfo.ID)
	}

	// check if other NPCs are visible
	for _, otherNpc := range n.ActiveMapCtx.GetAllNPCs() {
		if otherNpc.ID() == n.ID() {
			continue
		}
		target := n.ActiveMapCtx.GetEntityInfo(otherNpc.GetInfo().CharID)
		lightIntensity := n.ActiveMapCtx.GetLightIntensity(target.TilePos)
		vis := calculateVisibility(observer, target, lightIntensity)
		if vis > 0 {
			if _, alreadySeen := n.visibleEntities[target.ID]; !alreadySeen {
				n.visibleEntities[target.ID] = time.Now()
				n.eventBus.Publish(defs.Event{
					Type: pubsub.EventEntitySpotted,
					Data: map[string]any{
						"observer": observer.ID,
						"target":   target.ID,
					},
				})
			}
		} else {
			delete(n.visibleEntities, target.ID)
		}
	}
}
