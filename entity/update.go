package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/ui/overlay"
)

func (e *Entity) Draw(screen *ebiten.Image, om *overlay.OverlayManager, offsetX float64, offsetY float64) {
	if !e.Loaded {
		return
	}

	drawX, drawY := e.drawPos(offsetX, offsetY)
	e.Body.Draw(screen, drawX, drawY, config.GameScale)
	e.drawX = drawX
	e.drawY = drawY

	if e.speechBubble != nil {
		if om == nil {
			logz.Panic("overlay manager was empty")
		}
		sbX := (drawX + config.TileSize) * config.GameScale
		sbY := (drawY - config.TileSize*2) * config.GameScale
		e.speechBubble.DrawOverlay(om, sbX, sbY)
	}

	e.FloatMGMT.Draw(screen, e.drawX, e.drawY)
}

// DrawPos returns the actual absolute position where the entity will be drawn
func (e Entity) drawPos(offsetX, offsetY float64) (drawX, drawY float64) {
	dx, dy := e.Body.Dimensions()
	rect := model.NewRect(0, 0, float64(dx), float64(dy))
	drawX, drawY = rendering.GetRectDrawPos(rect, e.X, e.Y, offsetX, offsetY)
	drawY -= 6 // move up a little, since we want the entity to look like its standing in the middle of the tile
	drawX *= config.GameScale
	drawY *= config.GameScale
	return drawX, drawY
}

func (e Entity) GetDrawRect() model.Rect {
	dx, dy := e.Body.Dimensions()
	dx = int(float64(dx) * config.GameScale)
	dy = int(float64(dy) * config.GameScale)
	return model.NewRect(e.drawX, e.drawY, float64(dx), float64(dy))
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
			logz.Println(e.DisplayName(), "x:", e.X, "y:", e.Y, "targetX:", e.TargetX, "targetY:", e.TargetY)
			panic("entity is not moving but hasn't met its goal yet. hint: if you are setting the entity position, use the SetPosition function to ensure Target is updated too.")
		}
		if e.Body.IsMoving() {
			logz.Panicln(e.DisplayName(), "entity is not moving, but body is still doing movement animations")
		}
	}

	movementResult := e.updateMovement()

	if movementResult.UnexpectedCollision {
		e.Movement.Interrupted = true
		e.StopMovement()
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
				logz.Println(e.DisplayName(), "failed to set next target path:", res)
				if res.AlreadyMoving {
					logz.Panicf("movement failed because we are already moving... but IsMoving is false? %s", res)
				}
				e.Movement.Interrupted = true
				e.StopMovement()
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
}

func (e Entity) IsDead() bool {
	return e.characterStateRef.Dead
}
