package entity

import (
	"fmt"
	"math"

	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/audio"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

type MoveError struct {
	AlreadyMoving   bool
	Collision       bool
	Cancelled       bool
	Success         bool
	CollisionResult model.CollisionResult
	Info            string
	CollisionPoint  *model.Vec2
}

func (me MoveError) String() string {
	info := ""
	if me.Info != "" {
		info = fmt.Sprintf(" (info: %s)", me.Info)
	}
	if me.Success {
		return "Success" + info
	}
	if me.Collision {
		return "Collision" + info
	}
	if me.AlreadyMoving {
		return "Already Moving" + info
	}
	if me.Cancelled {
		return "Cancelled" + info
	}
	return "No value set" + info
}

// GoToPos Attempts to put the entity on a path to reach the given target.
// If the path to the target is blocked, you can conditionally go as close as possible with the "close enough" flag.
// Returns the actual goal target (in case it was changed due to a conflict and the "close enough" flag).
func (e *Entity) GoToPos(c model.Coords, closeEnough bool) (model.Coords, MoveError) {
	if e.Movement.IsMoving {
		logz.Println(e.DisplayName(), "GoToPos: entity is already moving")
		return c, MoveError{AlreadyMoving: true}
	}
	if len(e.Movement.TargetPath) > 0 {
		logz.Warnln(e.DisplayName(), "GoToPos: entity already has a target path. Target path should be cancelled first.")
	}
	if e.TilePos.Equals(c) {
		logz.Errorln(e.DisplayName(), "entity attempted to GoToPos for position it is already in")
		return c, MoveError{Cancelled: true, Info: "already in target position"}
	}

	path, found := e.World.FindPath(e.TilePos, c)
	if !found {
		if !closeEnough {
			return c, MoveError{Cancelled: true, Info: "path not found, and 'close enough' not enabled"}
		}
		logz.Warnln(e.DisplayName(), "going a partial path since original path is blocked.", "start:", e.TilePos, "path:", path, "original goal:", c)
	}
	if len(path) == 0 {
		fmt.Println("tile pos:", e.TilePos, "goal:", c)
		logz.Warnln(e.DisplayName(), "GoToPos: calculated path is empty. Is the entity completely blocked in?")
		return c, MoveError{Cancelled: true, Info: "calculated path is empty"}
	}

	e.Movement.TargetPath = path
	return path[len(path)-1], MoveError{Success: true}
}

// CancelCurrentPath Tells the entity to stop moving along its set path once it has finished its current tile movement.
func (e *Entity) CancelCurrentPath() {
	if len(e.Movement.TargetPath) == 0 {
		logz.Warnln("tried to cancel path for an entity that has no path")
		return
	}
	e.Movement.TargetPath = []model.Coords{}
}

const (
	fullOpen = iota
	partialOpen
	fullBlock
)

type AnimationOptions struct {
	AnimationName         string
	AnimationTickInterval int
	SetAnimationOps       body.SetAnimationOps
}

// FaceTowards points the entity in the direction corresponding to the given movement.
// Ex: dx=-1 and dy=0 corresponds to facing left, and dx=0 dy=-1 corresponds to facing up.
func (e *Entity) FaceTowards(dx, dy float64) {
	if dx == 0 && dy == 0 {
		// no direction to face towards; just stay as is.
		return
	}
	// since we can't "face" diagonally, first find out which value (dx or dy) is of greater magnitude.
	horizontal := math.Abs(dx) >= math.Abs(dy)

	if horizontal {
		if dx < 0 {
			e.Movement.Direction = model.Directions.Left
		} else {
			e.Movement.Direction = model.Directions.Right
		}
	} else {
		if dy < 0 {
			e.Movement.Direction = model.Directions.Up
		} else {
			e.Movement.Direction = model.Directions.Down
		}
	}

	e.Body.SetDirection(e.Movement.Direction)
}

func (e *Entity) SetAnimation(animOps AnimationOptions) body.SetAnimationResult {
	animRes := e.Body.SetAnimation(animOps.AnimationName, animOps.SetAnimationOps)
	if animRes.Success {
		e.Body.SetAnimationTickCount(animOps.AnimationTickInterval)
	} else {
		if !animRes.AlreadySet {
			if e.Body.GetCurrentAnimation() == body.AnimIdle {
				panic("failed to set movement animation, but current animation seems to be empty...?")
			}
		}
	}
	return animRes
}

// TryMoveMaxPx is the same as TryMovePx, but lets the entity still move in the direction even if a collision is encountered,
// as long as there is some space in that direction
func (e *Entity) TryMoveMaxPx(dx, dy, speed float64) MoveError {
	if dx == 0 && dy == 0 {
		panic("TryMoveMaxPx called with no distance given")
	}
	moveError := e.TryMovePx(dx, dy, speed)
	if moveError.Collision {
		// a collision occurred; try to adjust the target by the intersection area
		// first, reset dx/dy based on specifically where the collision occurred
		// since the collision could've happened on the first step towards moving there (we check that in TryMovePx too)
		if moveError.CollisionPoint == nil {
			logz.Panicln("TryMoveMaxPx", "a collision occurred, but no collision point was set (it was nil) so we don't know where to adjust from.")
		}
		collisionPoint := *moveError.CollisionPoint
		curPos := model.NewVec2(e.X, e.Y)
		delta := collisionPoint.Sub(curPos)
		dx = delta.X
		dy = delta.Y

		cr := moveError.CollisionResult
		cx, cy := 0.0, 0.0

		top := cr.TopLeft.Int() + cr.TopRight.Int()
		left := cr.TopLeft.Int() + cr.BottomLeft.Int()
		right := cr.TopRight.Int() + cr.BottomRight.Int()
		bottom := cr.BottomLeft.Int() + cr.BottomRight.Int()

		// TODO: the below logic seems to work well, but it seems too long and probably can be simplified.
		// let's try to combine the "one direction" logic into the "bi-direction" section
		if dx == 0 || dy == 0 {
			// only moving in one direction; simpler logic
			if dx < 0 {
				// left
				if left != fullOpen {
					// get as close as possible
					cx = max(cr.BottomLeft.Dx, cr.TopLeft.Dx)
				}
			} else if dx > 0 {
				// right
				if right != fullOpen {
					cx = -max(cr.BottomRight.Dx, cr.TopRight.Dx)
				}
			}
			if dy < 0 {
				// up
				if top != fullOpen {
					cy = max(cr.TopLeft.Dy, cr.TopRight.Dy)
				}
			} else if dy > 0 {
				// down
				if bottom != fullOpen {
					cy = -max(cr.BottomLeft.Dy, cr.BottomRight.Dy)
				}
			}
		} else {
			// moving in two directions; more special cases logic to consider
			yDir := 0
			var yDirCorner1, yDirCorner2 model.IntersectionResult
			var xDirCorner1, xDirCorner2 model.IntersectionResult
			xDirFactor, yDirFactor := 1.0, 1.0
			// check how "open" each direction is
			if dy < 0 {
				yDir = top
				yDirCorner1, yDirCorner2 = cr.TopLeft, cr.TopRight
			} else {
				yDir = bottom
				yDirCorner1, yDirCorner2 = cr.BottomLeft, cr.BottomRight
				yDirFactor = -1
			}
			xDir := 0
			if dx < 0 {
				xDir = left
				xDirCorner1, xDirCorner2 = cr.TopLeft, cr.BottomLeft
			} else {
				xDir = right
				xDirCorner1, xDirCorner2 = cr.TopRight, cr.BottomRight
				xDirFactor = -1
			}

			// sanity check: if a collision occurred, then both directions cannot be "full open"
			if yDir == fullOpen && xDir == fullOpen {
				logz.Panicln("TryMoveMaxPx", "collision occurred, but both directions (x & y) appear to be fully open...")
			}

			if yDir != fullOpen || xDir != fullOpen {
				// there is blockage in one (or both) directions. let's suss it out.
				switch xDir {
				case fullBlock:
					// get as close as possible, but ultimately block this direction
					cx = xDirFactor * max(xDirCorner1.Dx, xDirCorner2.Dx)
				case partialOpen:
					switch yDir {
					case fullBlock:
						// sliding along a wall; continue freely
					case partialOpen:
						// walking into an outwardly pointing corner
						// need to decide which side we will go along
						// go along the direction that overlaps the most (smaller overlap gets clamped)
						xOverlap := int(max(xDirCorner1.Dx, xDirCorner2.Dx))
						yOverlap := int(max(yDirCorner1.Dy, yDirCorner2.Dy))
						if xOverlap > yOverlap {
							cy = yDirFactor * max(yDirCorner1.Dy, yDirCorner2.Dy)
						} else {
							cx = xDirFactor * max(xDirCorner1.Dx, xDirCorner2.Dx)
						}
					case fullOpen:
						// about to turn around a corner
						// this direction should be cancelled for now
						cx = xDirFactor * max(xDirCorner1.Dx, xDirCorner2.Dx)
					}
				}
				switch yDir {
				case fullBlock:
					cy = yDirFactor * max(yDirCorner1.Dy, yDirCorner2.Dy)
				case partialOpen:
					switch xDir {
					case fullBlock:
						// sliding along a wall; continue freely
					case fullOpen:
						// about to turn around a corner
						// this direction should be cancelled for now
						cy = yDirFactor * max(yDirCorner1.Dy, yDirCorner2.Dy)
					}
				}
			}
		}

		if cx == 0 && cy == 0 {
			// no changes suggested? this seems wrong
			logz.Panicln("TryMoveMaxPx", "no target adjustment occurred (cx and cy calculated to 0). this probably shouldn't be happening...")
		}

		dx += cx
		dy += cy

		// if no actual change will occur after adjustments, give up
		// entity is probably directly up against some collision rects
		if dx == 0 && dy == 0 {
			return moveError
		}

		return e.TryMovePx(dx, dy, speed)
	}
	return moveError
}

func (e *Entity) TryMovePx(dx, dy, speed float64) MoveError {
	if dx == 0 && dy == 0 {
		panic("TryMovePx: dx and dy are both 0")
	}

	if e.IsStunned() {
		return MoveError{Cancelled: true, Info: "stunned"}
	}

	if e.Body.IsAttacking() || e.waitingToAttack {
		// cannot move (or change directions) while attacking
		info := ""
		if e.Body.IsAttacking() {
			info = "body is attacking;"
		}
		if e.waitingToAttack {
			info += " waiting to attack"
		}
		return MoveError{Cancelled: true, Info: info}
	}

	// Notes: We don't use CollisionRect here since we aren't checking our current position, but the next position we move into
	// The reason we use TargetX/Y (instead of x/y) is because we add onto the existing target when we use TryMovePx
	x := e.TargetX + dx
	y := e.TargetY + dy
	w, h := e.CollisionRect().W, e.CollisionRect().H

	// actual targets
	targetRect := model.Rect{X: x, Y: y, W: w, H: h}
	target := model.Vec2{X: x, Y: y}

	// check if first movement towards target (at given speed) collides
	// if this collides, that means we can't actually move a single step towards the goal (even though the goal itself is open).
	// one instance where this happens is when you are trying to turn around a corner.
	pos := model.Vec2{X: e.X, Y: e.Y}
	newPos := pos.MoveTowards(target, speed)
	nextStepRect := model.NewRect(newPos.X, newPos.Y, w, h)
	res := e.World.Collides(nextStepRect, string(e.ID()))
	if res.Collides() {
		// a collision happened on the first step towards the target (but not the target itself)
		// we need to tell exactly where this first step was, so that TryMoveMaxPx knows precisely where to adjust from.
		return MoveError{
			CollisionPoint:  &newPos,
			Collision:       true,
			CollisionResult: res,
			Info:            "first step towards new target was collision",
		}
	}

	// if first step is clear, then check that the actual target itself isn't a collision
	res = e.World.Collides(targetRect, string(e.ID()))
	if res.Collides() {
		return MoveError{
			CollisionPoint:  &target,
			Collision:       true,
			CollisionResult: res,
			Info:            "new target position was collision",
		}
	}

	e.Movement.IsMoving = true
	e.TargetX = float64(x)
	e.TargetY = float64(y)

	e.Movement.Speed = speed

	if e.Movement.Speed == 0 {
		panic("movement speed is 0")
	}

	return MoveError{Success: true}
}

func (e *Entity) TryBumpBack(px int, speed float64, forceOrigin model.Vec2, anim string, animTickInterval int) MoveError {
	// calculate dx dy values
	origin := model.Vec2{X: e.X, Y: e.Y}
	dest := origin.MoveAway(forceOrigin, float64(px))
	dx := dest.X - origin.X
	dy := dest.Y - origin.Y

	animRes := e.SetAnimation(AnimationOptions{
		AnimationName:         anim,
		AnimationTickInterval: animTickInterval,
		SetAnimationOps: body.SetAnimationOps{
			Force: true,
		},
	})
	if !animRes.Success && !animRes.AlreadySet {
		logz.Println(e.DisplayName(), "TryBumpBack: failed to set animation:", animRes)
		return MoveError{Cancelled: true, Info: "failed to set animation"}
	}

	return e.TryMoveMaxPx(dx, dy, speed)
}

type updateMovementResult struct {
	ReachedTarget           bool
	UnexpectedCollision     bool
	ContinuingTowardsTarget bool
}

func (e *Entity) updateMovement() updateMovementResult {
	if !e.Movement.IsMoving {
		return updateMovementResult{} // not moving, so an empty result will suffice
	}
	if e.Movement.Speed == 0 {
		panic("updateMovement called when speed is 0; speed was not set wherever entity movement was started")
	}

	// check for suggested paths (if entity is currently following a path)
	if len(e.Movement.TargetPath) > 0 && len(e.Movement.SuggestedTargetPath) > 0 {
		e.tryMergeSuggestedPath(e.Movement.SuggestedTargetPath)
		e.Movement.SuggestedTargetPath = []model.Coords{}
	}

	res := e.World.Collides(e.CollisionRect(), string(e.ID()))
	if res.Collides() {
		logz.Panicf("[%s] updateMovement: current position is colliding!", e.DisplayName())
	}

	pos := model.Vec2{X: e.X, Y: e.Y}
	target := model.Vec2{X: e.TargetX, Y: e.TargetY}

	newPos := pos.MoveTowards(target, e.Movement.Speed)
	w, h := e.CollisionRect().W, e.CollisionRect().H
	res = e.World.Collides(model.NewRect(newPos.X, newPos.Y, w, h), string(e.ID()))
	if res.Collides() {
		logz.Println(e.DisplayName(), "updateMovement: next position is colliding! cancelling movement")
		return updateMovementResult{UnexpectedCollision: true}
	}

	e.X = newPos.X
	e.Y = newPos.Y

	if math.IsNaN(e.X) || math.IsNaN(e.Y) {
		logz.Println(e.DisplayName(), "e.X:", e.X, "e.Y:", e.Y)
		panic("entity position is NaN")
	}

	e.TilePos = model.ConvertPxToTilePos(int(e.X), int(e.Y))

	result := updateMovementResult{ContinuingTowardsTarget: true}

	if target.Equals(newPos) {
		// don't return yet, so that footstep sound can play
		result = updateMovementResult{ReachedTarget: true}
	}

	// footstep sound effects
	e.footstepSFX.TicksUntilNextPlay--
	if e.footstepSFX.TicksUntilNextPlay <= 0 {
		groundMaterial := e.World.GetGroundMaterial(e.TilePos.X, e.TilePos.Y)
		var distToPlayer float64
		if e.IsPlayer() {
			distToPlayer = 0
		} else {
			distToPlayer = e.World.GetDistToPlayer(e.X, e.Y)
			maxDist := float64(config.TileSize * 10)
			distToPlayer = min(distToPlayer, maxDist)
			distToPlayer = distToPlayer / maxDist
		}
		volFactor := 1 - distToPlayer
		switch groundMaterial {
		case "wood":
			e.footstepSFX.Step(audio.StepWood, volFactor)
		case "stone":
			e.footstepSFX.Step(audio.StepStone, volFactor)
		case "grass":
			e.footstepSFX.Step(audio.StepGrass, volFactor)
		case "tile":
			e.footstepSFX.Step(audio.StepStone, volFactor) // TODO get new sound specifically for tile?
		case "":
			e.footstepSFX.Step(audio.StepDefault, volFactor)
		default:
			// if we don't have the string registered (and it's not an empty string) then error
			panic("ground material not recognized: " + groundMaterial)
		}
	}
	return result
}

func (e *Entity) StopMovement() {
	if !e.Movement.IsMoving {
		panic("told to stop movement, but entity is not moving?")
	}
	logz.Println(e.DisplayName(), "Stopping movement")
	e.TargetX = e.X
	e.TargetY = e.Y
	e.Movement.IsMoving = false
	if e.Body.IsMoving() {
		e.Body.StopAnimation()
	}
}

func (e *Entity) trySetNextTargetPath() MoveError {
	if len(e.Movement.TargetPath) == 0 {
		panic("tried to set next target along path for entity that has no set target path")
	}
	nextTarget := e.Movement.TargetPath[0]
	if nextTarget.Equals(e.TilePos) {
		panic("trySetNextTargetPath: next target is the same tile as current position")
	}

	if float64(e.TilePos.X) != e.X/config.TileSize || float64(e.TilePos.Y) != e.Y/config.TileSize {
		logz.Println(e.DisplayName(), "trySetNextTargetPath: entity is not at its tile position. Was it bumped by an enemy attack or something?")
		logz.Println(e.DisplayName(), "tilePos:", e.TilePos, "e.X:", e.X/config.TileSize, "e.Y:", e.Y/config.TileSize)
		logz.Println(e.DisplayName(), "Clamping to current tile position. TODO: perhaps we should make a more graceful way of recovering the position than this?")
		e.TargetX = float64(e.TilePos.X * config.TileSize)
		e.TargetY = float64(e.TilePos.Y * config.TileSize)
		e.Movement.IsMoving = true
		return MoveError{Cancelled: true, Info: "entity was unexpectedly not at its tile position - possibly bumped by an enemy attack or something?"}
	}

	curPos := model.Vec2{X: e.X, Y: e.Y}
	target := model.Vec2{X: float64(nextTarget.X * config.TileSize), Y: float64(nextTarget.Y * config.TileSize)}
	dist := curPos.Dist(target)
	dPos := target.Sub(curPos)

	if dist > config.TileSize {
		logz.Println(e.DisplayName(), "curPos:", curPos, "target:", target, "dist:", dist)
		logz.Println(e.DisplayName(), "trySetNextTargetPath: next target is not an adjacent tile (dist > 16). Clearing target path.")
		e.Movement.TargetPath = []model.Coords{}
		return MoveError{Cancelled: true, Info: "next target was not an adjacent tile (dist > tilesize)"}
	}

	moveError := e.TryMovePx(dPos.X, dPos.Y, e.CharacterStateRef.WalkSpeed())

	if !moveError.Success {
		return moveError
	}

	animRes := e.SetAnimation(AnimationOptions{
		AnimationName:         body.AnimWalk,
		AnimationTickInterval: e.Movement.WalkAnimationTickInterval,
	})
	if !animRes.Success && !animRes.AlreadySet {
		logz.Println(e.DisplayName(), "trySetNextTargetPath: failed to set animation:", animRes)
	}

	e.FaceTowards(dPos.X, dPos.Y)

	if !e.Movement.IsMoving {
		panic("movement succeeded, but not moving?")
	}

	e.Movement.TargetPath = e.Movement.TargetPath[1:]
	return MoveError{Success: true}
}

func (e *Entity) tryMergeSuggestedPath(newPath []model.Coords) bool {
	if len(e.Movement.TargetPath) == 0 {
		panic("a path was suggested to an entity with no existing target path to merge it into")
	}
	if len(newPath) == 0 {
		panic("an empty path was suggested to an entity")
	}
	if len(e.Movement.TargetPath) <= 3 {
		return false
	}
	if newPath[0].Equals(e.TilePos) {
		logz.Println("tryMergeSuggestedPath", "error: new path starts at entity's current position. it should start at a position in the target path ahead of the current position.")
		return false
	}

	for i, c := range e.Movement.TargetPath {
		if c.Equals(newPath[0]) {
			// new path starts from this target path position; merge it in by replacing this position
			e.Movement.TargetPath = append(e.Movement.TargetPath[:i], newPath...)
			// logz.Println("tryMergeSuggestedPath", "merged suggested path into current target path")
			return true
		}
		if c.IsAdjacent(newPath[0]) {
			// new path is adjacent to a target path position; merge it in by adding it next to this position
			e.Movement.TargetPath = append(e.Movement.TargetPath[:i+1], newPath...)
			return true
		}
	}
	logz.Println("tryMergeSuggestedPath", "failed to merge suggested path", "suggested path:", newPath, "current path:", e.Movement.TargetPath)
	return false
}
