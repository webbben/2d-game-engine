package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
)

func (e Entity) Draw(screen *ebiten.Image, offsetX float64, offsetY float64) {
	if !e.Loaded {
		return
	}
	if e.CurrentFrame == nil {
		panic("tried to draw entity with no set frame")
	}

	op := &ebiten.DrawImageOptions{}
	drawX, drawY := rendering.GetImageDrawPos(e.CurrentFrame, e.X, e.Y, offsetX, offsetY)
	op.GeoM.Translate(drawX, drawY)
	op.GeoM.Scale(config.GameScale, config.GameScale)
	screen.DrawImage(e.CurrentFrame, op)
}

// If the entity is following a target path, this function handles setting the next target tile on the path.
// Only meant to be called when element is on a target path and is ready to set the next target tile.
//
// If already moving, it's required that the next position is the same direction as the current movement.
// This is to prevent an awkward diagonal movement from occurring.
func (e *Entity) trySetNextTargetPath() bool {
	if len(e.Movement.TargetPath) == 0 {
		panic("tried to set next target along path for entity that has no set target path")
	}
	var moveError MoveError
	lastTilePos := e.TilePos.Copy()
	if e.Movement.IsMoving {
		if getRelativeDirection(e.Movement.TargetTile, e.Movement.TargetPath[0]) != e.Movement.Direction {
			// don't allow a seamless transition to the next tile if entity is moving but next tile isn't the same direction.
			// this is because the entity would start to go diagonally.
			return false
		}
		// if entity is already moving (i.e. hasn't landed on a tile yet), use its current target tile as the new tile position.
		// this is because we are then setting the next tile beyond that position as the next target, basically "getting slightly ahead".
		// not waiting to stop first improves smoothness of walking, since we don't have to stop and wait a tick first (if direction is the same).
		lastTilePos = e.Movement.TargetTile.Copy()
		moveError = e.tryQueueNextMove(e.Movement.TargetPath[0])
	} else {
		// if the entity is not moving, then use the normal method of initiating a movement towards the next tile in the target path.
		moveError = e.TryMove(e.Movement.TargetPath[0])
	}
	if moveError.Success {
		e.TilePos = lastTilePos
		// shift target path
		e.Movement.TargetPath = e.Movement.TargetPath[1:]
		return true
	}

	e.Movement.Interrupted = true // mark move as interrupted, so that NPCs can react as needed
	logz.Println(e.DisplayName, "movement interrupted")
	logz.Println(e.DisplayName, "trySetNextTargetPath: failed to set next target tile:", moveError)
	return false
}

func (e *Entity) Update() {
	if !e.Movement.IsMoving {
		if len(e.Movement.TargetPath) > 0 {
			e.trySetNextTargetPath()
		}
	}
	if e.Movement.IsMoving {
		e.updateMovement()
	}

	e.updateCurrentFrame()
}

func (e *Entity) updateMovement() {
	if e.Movement.Speed == 0 {
		panic("updateMovement called when speed is 0; speed was not set wherever entity movement was started")
	}
	if e.Movement.TargetTile.Equals(e.TilePos) && len(e.Movement.TargetPath) == 0 {
		panic("updateMovement called when entity has no target tile or target path")
	}
	targetPx := float64(e.Movement.TargetTile.X * config.TileSize)
	targetPy := float64(e.Movement.TargetTile.Y * config.TileSize)
	dx := targetPx - e.X
	dy := targetPy - e.Y

	dist := math.Hypot(dx, dy)

	finishMove := false

	if dist <= e.Movement.Speed {
		// entity is already within range of the target tile.
		// check if there is a next move ready that is the same direction as the current move
		// if so, then keep on trucking. this helps to avoid a little blip and makes walking go smoother
		if len(e.Movement.TargetPath) > 0 {
			// Case: entity has path it is following
			success := e.trySetNextTargetPath()
			if !success {
				// failed to go to queue next target in path; finish current movement
				finishMove = true
			}
		} else if e.IsPlayer && playerStillWalking(e.Movement.Direction) {
			// Case: player is still moving in the same direction
			nextTarget := e.Movement.TargetTile.GetAdj(e.Movement.Direction)
			lastTarget := e.Movement.TargetTile.Copy()
			moveError := e.tryQueueNextMove(nextTarget)
			if moveError.Success {
				e.TilePos = lastTarget
			} else {
				logz.Println(e.DisplayName, "tryQueueNextMove failed:", moveError)
				finishMove = true
			}
		} else {
			finishMove = true
		}
	}
	if finishMove {
		// finish movement
		e.X, e.Y = targetPx, targetPy
		e.TilePos = e.Movement.TargetTile.Copy()
		e.Movement.IsMoving = false
	} else {
		moveDx := e.Movement.Speed * dx / dist
		moveDy := e.Movement.Speed * dy / dist
		if moveDx == 0 && moveDy == 0 {
			panic("somehow, movement distance calculation is 0! entity is stuck and not moving towards its goal")
		}
		e.X += moveDx
		e.Y += moveDy
	}

	// update animation
	e.Movement.AnimationTimer++
	if e.Movement.AnimationTimer > 10 {
		_, frameCount := e.getMovementAnimationInfo()
		e.Movement.AnimationFrame = (e.Movement.AnimationFrame + 1) % frameCount
		e.Movement.AnimationTimer = 0
	}
}

func playerStillWalking(direction byte) bool {
	switch direction {
	case 'L':
		return ebiten.IsKeyPressed(ebiten.KeyA)
	case 'R':
		return ebiten.IsKeyPressed(ebiten.KeyD)
	case 'U':
		return ebiten.IsKeyPressed(ebiten.KeyW)
	case 'D':
		return ebiten.IsKeyPressed(ebiten.KeyS)
	default:
		panic("playerStillWalking: invalid direction passed")
	}
}

func (e *Entity) updateCurrentFrame() {
	if e.Movement.IsMoving {
		animationName, _ := e.getMovementAnimationInfo()
		e.CurrentFrame = e.getAnimationFrame(animationName, e.Movement.AnimationFrame)
		return
	}
	// idle
	animationName, _ := e.getMovementAnimationInfo()
	e.CurrentFrame = e.getAnimationFrame(animationName, 0)
}
