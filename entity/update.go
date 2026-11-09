package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/model"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/ui/overlay"
)

func (e *Entity) Draw(screen *ebiten.Image, om *overlay.OverlayManager, offsetX float64, offsetY float64) {
	if !e.Loaded {
		return
	}

	anchorDrawX, anchorDrawY := e.anchorDrawPos(offsetX, offsetY)
	opacity := 1.0
	if e.IsSneaking {
		opacity = 0.5
	}
	e.Body.Draw(screen, anchorDrawX, anchorDrawY, body.DrawBodyParams{
		Scale:   config.GameScale,
		Opacity: opacity,
	})
	e.anchorDrawX = anchorDrawX
	e.anchorDrawY = anchorDrawY

	screenBodyRect := e.GetScreenBodyRect()

	if e.speechBubble != nil {
		if om == nil {
			logz.Panic("overlay manager was empty")
		}
		// make it appear to the top right of the NPC's head
		sbX := screenBodyRect.X + config.TileSize*config.GameScale
		sbY := screenBodyRect.Y - config.TileSize*config.GameScale*2.5
		e.speechBubble.DrawOverlay(om, sbX, sbY)
	}

	e.FloatMGMT.Draw(screen, screenBodyRect.X, screenBodyRect.Y)
}

func (e Entity) anchorDrawPos(offsetX, offsetY float64) (anchorX, anchorY float64) {
	anchorX = (e.X - offsetX) * config.GameScale
	anchorY = (e.Y - offsetY) * config.GameScale
	return anchorX, anchorY
}

// GetScreenBodyRect gives a rect that represents the screen position and space that an entity's body is in on the screen.
// Can be used for things, such as detecting if the mouse is hovering over an entity. This purposely excludes any canvas space beyond the body.
func (e Entity) GetScreenBodyRect() model.Rect {
	bodyDx, bodyDy := e.Body.Dimensions()

	// the anchor tile represents the bottom center tile of the entity's body, so we can derive the actual effective "body" area from it:
	// - subtract height from the bottom of the anchor tile
	// - center the width across the anchor tile (to handle for bodies that are wider than 1 tile)

	widthOverflow := bodyDx - config.TileSize

	// we have to use the anchor draw position since we don't have access to the camera's current offset. that's why all the math below
	// is scaling all the values by game scale.
	r := model.Rect{
		X: e.anchorDrawX - float64(widthOverflow/2)*config.GameScale,
		Y: e.anchorDrawY + config.TileSize*config.GameScale - float64(bodyDy)*config.GameScale,
		W: float64(bodyDx) * config.GameScale,
		H: float64(bodyDy) * config.GameScale,
	}

	return r
}

func (e *Entity) Update() {
	if !e.Loaded {
		panic("entity not loaded yet!")
	}

	if e.speechBubble != nil {
		if e.speechBubble.Done() {
			e.speechBubble = nil
		}
	}

	e.FloatMGMT.Update()

	e.SyncBodyToState()

	e.Body.Update()

	if e.characterStateRef.Dead {
		return
	} else {
		// detect if entity has died
		if e.characterStateRef.Health <= 0 {
			e.Kill()
			return
		}
	}

	if e.stunTicks > 0 {
		e.stunTicks--
	}

	// doing this here so that if the player is still trying to move, their next movement can be set before officially deciding we have stopped.
	if !e.Movement.IsMoving {
		if e.Body.IsMoving() {
			e.Body.StopAnimation()
		}
		// some validation and sanity checks
		if e.TargetX != e.X || e.TargetY != e.Y {
			logz.Println(string(e.ID()), "x:", e.X, "y:", e.Y, "targetX:", e.TargetX, "targetY:", e.TargetY)
			panic("entity is not moving but hasn't met its goal yet. hint: if you are setting the entity position, use the SetPosition function to ensure Target is updated too.")
		}
		if e.Body.IsMoving() {
			logz.Println(string(e.ID()), "entity is not moving, but body is still doing movement animations")
			logz.Panicln("Entity", "entity is not moving, but body is still doing movement animations")
		}
	}

	movementResult := e.updateMovement()

	if movementResult.UnexpectedCollision {
		e.markInterrupted()
	} else if movementResult.ReachedTarget {
		e.Movement.IsMoving = false

		// check if we can queue up the next target in an existing path
		if len(e.Movement.TargetPath) > 0 {
			res := e.trySetNextTargetPath()
			if res.Success {
				if !e.Movement.IsMoving {
					panic("trySetNextTargetPath succeeded, but still not moving?")
				}
			} else {
				// failed to set next path
				logz.Println(string(e.ID()), "failed to set next target path:", res)
				if res.AlreadyMoving {
					logz.Println(string(e.ID()), "movement failed; already moving. res:", res)
					logz.Panicf("movement failed because we are already moving... but IsMoving is false")
				}
				e.markInterrupted()
			}
		}
	} else if movementResult.ContinuingTowardsTarget {
		// sanity check
		if !e.Movement.IsMoving {
			panic("we are supposedly still moving towards target... why is IsMoving false?")
		}
	}

	// TODO: 2026-07-24 moved body.Update from here. if something starts behaving weirdly with body, maybe move it back here.

	e.updateAttackManager()
}

func (e *Entity) Kill() {
	e.characterStateRef.Health = 0
	e.characterStateRef.Dead = true
	logz.Warnln(string(e.ID()), "Entity died!")
	e.SetAnimation(AnimationOptions{
		AnimationName:         body.AnimDead,
		AnimationTickInterval: 1,
		SetAnimationOps: body.SetAnimationOps{
			Force: true,
		},
	})

	// Tell the world. Death cleanup has to happen exactly once and in a fixed order -- drop them out of
	// any fight, stop their tasks, schedule the corpse for removal -- and doing that here would mean the
	// entity package reaching into world state. The event also puts it on the main loop like every other
	// sys-event, so cleanup can't land on whichever goroutine happened to run us.
	e.eventBus.SysEntityDied(e.ID())
}

// IsDead reports whether this entity is dead or has taken a killing blow but not yet processed as dead.
// The reason these two things could diverge on a single tick is, actually being marked as Dead (in character state)
// could be delayed by a tick depending on when entity.Update actually gets called. So, other logic that wants to check
// if this entity is dead should check this function, and not just check the Dead property of character state ref.
func (e Entity) IsDead() bool {
	return e.characterStateRef.Health <= 0 || e.characterStateRef.Dead
}
