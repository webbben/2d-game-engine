package npc

import (
	"math"
	"time"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/entity"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/utils"
)

const (
	// VisionConeHalfAngle is the half-angle (in radians) of an entity's vision cone.
	// A target is in the ~90 degree cone when its direction dot product exceeds cos(VisionConeHalfAngle).
	// Shared with debug rendering so the drawn cone always matches the visibility calculation.
	VisionConeHalfAngle = math.Pi / 4

	// VisibilityThreshold determines the visibility level when an NPC can see another entity
	VisibilityThreshold = 0.01

	// OutsideConeFactor is applied to visibility when the target lies outside the observer's
	// vision cone. Tuned independently from SightBlockFactor; kept as a const so tests track it.
	OutsideConeFactor float64 = 0.3
)

func calculateVisibility(observer, target entity.EntityInfo, lightIntensity float32, sneakMult float64, sightBlocked bool) float64 {
	// TODO: factor in target movement speed (movement makes target more visible)
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
	// within the ~90 deg cone iff dot > cos(VisionConeHalfAngle)
	if (norm.X*forward.X)+(norm.Y*forward.Y) <= math.Cos(VisionConeHalfAngle) {
		// if outside of the main vision cone, reduce visibility further
		visibility *= OutsideConeFactor
	}

	// sneak multiplier
	if target.Sneaking {
		visibility *= sneakMult
	}

	// line of sight blocked
	// TODO: I notice a discrepancy here: why would sightBlocked result in 0 visibility, but being outside of vision cone
	// only divides it by 3? shouldn't these be the same or at least similar?
	if sightBlocked {
		visibility *= SightBlockFactor
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
	n.updateTargetVisibility(observer, playerInfo)

	// check if other NPCs are visible
	for _, otherNpc := range n.ActiveMapCtx.GetAllNPCs() {
		if otherNpc.ID() == n.ID() {
			continue
		}
		target := n.ActiveMapCtx.GetEntityInfo(otherNpc.GetInfo().CharID)
		n.updateTargetVisibility(observer, target)
	}
}

func (n *NPC) updateTargetVisibility(observer, target entity.EntityInfo) {
	lightIntensity := n.ActiveMapCtx.GetLightIntensity(target.TilePos)
	sneakMult := 1.0
	if target.Sneaking {
		if n.dataman.StealthSystemCalc == nil {
			logz.Panic("stealth system calc isn't defined")
		}
		skills, attrs := characterstate.CalculateSkillsAndAttributes(target.ID, n.dataman)
		sneakMult = n.dataman.StealthSystemCalc.SneakVisibilityMultiplier(target.ID, attrs, skills)
	}
	sightBlocked := n.ActiveMapCtx.IsLineOfSightBlocked(observer.TilePos, target.TilePos)
	vis := calculateVisibility(observer, target, lightIntensity, sneakMult, sightBlocked)
	if vis > VisibilityThreshold {
		if info, alreadySeen := n.visibleEntities[target.ID]; alreadySeen {
			// already seen; just update the visibility value
			info.Visibility = vis
			n.visibleEntities[target.ID] = info
		} else {
			// not seen yet; create new entry and publish event
			n.visibleEntities[target.ID] = VisibleEntityInfo{
				FirstSeen:  time.Now(),
				Visibility: vis,
			}
			n.eventBus.Publish(defs.Event{
				Type: pubsub.EventEntitySpotted,
				Data: map[string]any{
					"observer": observer.ID,
					"target":   target.ID,
				},
			})
		}
		// handle player specific logic
		if target.ID == id.PlayerStateID {
			if !n.hasSeenPlayerYet {
				n.initialPlayerSightingThisTick = true
				logz.Println(n.ID(), "first player sighting")
			}
			n.hasSeenPlayerYet = true
			n.lastPlayerSightingTime = time.Now()
		}
	} else {
		delete(n.visibleEntities, target.ID)
	}
}
