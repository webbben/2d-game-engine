package entity

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/imgutil/rendering"
)

func (e *Entity) Draw(screen *ebiten.Image, offsetX float64, offsetY float64) {
	if !e.Loaded {
		return
	}

	drawX, drawY := e.DrawPos(offsetX, offsetY)
	// apparently, when drawing the body, gameScale isn't automatically factored in so we need to multiply it in here
	// however, I see another place that relies on DrawPos, so I'm leaving this out of that function for now. (TODO?)
	drawX *= config.GameScale
	drawY *= config.GameScale
	e.Body.Draw(screen, drawX, drawY, config.GameScale)
	e.drawX = drawX
	e.drawY = drawY
}

// DrawPos returns the actual absolute position where the entity will be drawn
func (e Entity) DrawPos(offsetX, offsetY float64) (drawX, drawY float64) {
	dx, dy := e.Body.Dimensions()
	rect := model.NewRect(0, 0, float64(dx), float64(dy))
	drawX, drawY = rendering.GetRectDrawPos(rect, e.X, e.Y, offsetX, offsetY)
	drawY -= 6 // move up a little, since we want the entity to look like its standing in the middle of the tile
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

	e.SyncBodyToState()

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

	e.Body.Update()

	e.updateAttackManager()
}
