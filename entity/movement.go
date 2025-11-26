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

// Attempts to put the entity on a path to reach the given target.
// If the path to the target is blocked, you can conditionally go as close as possible with the "close enough" flag.
// Returns the actual goal target (in case it was changed due to a conflict and the "close enough" flag).
func (e *Entity) GoToPos(c model.Coords, closeEnough bool) (model.Coords, MoveError) {
	if e.Movement.IsMoving {
		logz.Println(e.DisplayName, "GoToPos: entity is already moving")
		return c, MoveError{AlreadyMoving: true}
	}
	if len(e.Movement.TargetPath) > 0 {
		logz.Warnln(e.DisplayName, "GoToPos: entity already has a target path. Target path should be cancelled first.")
	}
	if e.TilePos.Equals(c) {
		logz.Errorln(e.DisplayName, "entity attempted to GoToPos for position it is already in")
		return c, MoveError{Cancelled: true, Info: "already in target position"}
	}

	path, found := e.World.FindPath(e.TilePos, c)
	if !found {
		if !closeEnough {
			return c, MoveError{Cancelled: true, Info: "path not found, and 'close enough' not enabled"}
		}
		logz.Warnln(e.DisplayName, "going a partial path since original path is blocked.", "start:", e.TilePos, "path:", path, "original goal:", c)
	}
	if len(path) == 0 {
		fmt.Println("tile pos:", e.TilePos, "goal:", c)
		logz.Warnln(e.DisplayName, "GoToPos: calculated path is empty. Is the entity completely blocked in?")
		return c, MoveError{Cancelled: true, Info: "calculated path is empty"}
	}

	e.Movement.TargetPath = path
	return path[len(path)-1], MoveError{Success: true}
}

// Tells the entity to stop moving along its set path once it has finished its current tile movement.
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

// This function points the entity in the direction corresponding to the given movement.
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
			if e.Body.GetCurrentAnimation() == body.ANIM_IDLE {
				panic("failed to set movement animation, but current animation seems to be empty...?")
			}
		}
	}
	return animRes
}

// same as TryMovePx, but lets the entity still move in the direction even if a collision is encountered,
// as long as there is some space in that direction
func (e *Entity) TryMoveMaxPx(dx, dy int, speed float64) MoveError {
	if dx == 0 && dy == 0 {
		panic("TryMoveMaxPx called with no distance given")
	}
	moveError := e.TryMovePx(dx, dy, speed)
	if moveError.Collision {
		// a collision occurred; try to adjust the target by the intersection area
		cr := moveError.CollisionResult
		cx, cy := 0, 0

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
					cx = int(max(cr.BottomLeft.Dx, cr.TopLeft.Dx))
				}
			} else if dx > 0 {
				// right
				if right != fullOpen {
					cx = -int(max(cr.BottomRight.Dx, cr.TopRight.Dx))
				}
			}
			if dy < 0 {
				// up
				if top != fullOpen {
					cy = int(max(cr.TopLeft.Dy, cr.TopRight.Dy))
				}
			} else if dy > 0 {
				// down
				if bottom != fullOpen {
					cy = -int(max(cr.BottomLeft.Dy, cr.BottomRight.Dy))
				}
			}
		} else {
			// moving in two directions; more special cases logic to consider
			yDir := 0
			var yDirCorner1, yDirCorner2 model.IntersectionResult
			var xDirCorner1, xDirCorner2 model.IntersectionResult
			xDirFactor, yDirFactor := 1, 1
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

			if yDir != fullOpen || xDir != fullOpen {
				// there is blockage in one (or both) directions. let's suss it out.
				switch xDir {
				case fullBlock:
					// get as close as possible, but ultimately block this direction
					cx = xDirFactor * int(max(xDirCorner1.Dx, xDirCorner2.Dx))
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
							cy = yDirFactor * int(max(yDirCorner1.Dy, yDirCorner2.Dy))
						} else {
							cx = xDirFactor * int(max(xDirCorner1.Dx, xDirCorner2.Dx))
						}
					case fullOpen:
						// about to turn around a corner
						// this direction should be cancelled for now
						cx = xDirFactor * int(max(xDirCorner1.Dx, xDirCorner2.Dx))
					}
				}
				switch yDir {
				case fullBlock:
					cy = yDirFactor * int(max(yDirCorner1.Dy, yDirCorner2.Dy))
				case partialOpen:
					switch xDir {
					case fullBlock:
						// sliding along a wall; continue freely
					case fullOpen:
						// about to turn around a corner
						// this direction should be cancelled for now
						cy = yDirFactor * int(max(yDirCorner1.Dy, yDirCorner2.Dy))
					}
				}
			}
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

func (e *Entity) TryMovePx(dx, dy int, speed float64) MoveError {
	if dx == 0 && dy == 0 {
		panic("TryMovePx: dx and dy are both 0")
	}

	if e.IsStunned() {
		return MoveError{Cancelled: true, Info: "stunned"}
	}

	if e.Body.IsAttacking() || e.attackManager.waitingToAttack {
		// cannot move (or change directions) while attacking
		info := ""
		if e.Body.IsAttacking() {
			info = "body is attacking;"
		}
		if e.attackManager.waitingToAttack {
			info += " waiting to attack"
		}
		return MoveError{Cancelled: true, Info: info}
	}

	x := int(e.TargetX) + dx
	y := int(e.TargetY) + dy
	targetRect := model.Rect{X: float64(x), Y: float64(y), W: e.width, H: e.width}

	res := e.World.Collides(targetRect, e.ID)
	if res.Collides() {
		return MoveError{
			Collision:       true,
			CollisionResult: res,
		}
	}

	e.Movement.IsMoving = true
	e.Position.TargetX = float64(x)
	e.Position.TargetY = float64(y)

	e.Movement.Speed = speed

	if e.Movement.Speed == 0 {
		panic("movement speed is 0")
	}

	return MoveError{Success: true}
}

func (e *Entity) TryBumpBack(px int, speed float64, forceOrigin model.Vec2, anim string, animTickInterval int) MoveError {
	// calculate dx dy values
	origin := model.Vec2{X: e.X, Y: e.Y}
	dest := moveAway(origin, forceOrigin, float64(px))
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
		logz.Println(e.DisplayName, "TryBumpBack: failed to set animation:", animRes)
		return MoveError{Cancelled: true, Info: "failed to set animation"}
	}

	return e.TryMoveMaxPx(int(dx), int(dy), speed)
}

func (e *Entity) updateMovement() {
	if e.Movement.Speed == 0 {
		panic("updateMovement called when speed is 0; speed was not set wherever entity movement was started")
	}

	// check for suggested paths (if entity is currently following a path)
	if len(e.Movement.TargetPath) > 0 && len(e.Movement.SuggestedTargetPath) > 0 {
		e.tryMergeSuggestedPath(e.Movement.SuggestedTargetPath)
		e.Movement.SuggestedTargetPath = []model.Coords{}
	}

	res := e.World.Collides(e.CollisionRect(), e.ID)
	if res.Collides() {
		logz.Panicf("[%s] updateMovement: current position is colliding!", e.DisplayName)
	}

	pos := model.Vec2{X: e.X, Y: e.Y}
	target := model.Vec2{X: e.TargetX, Y: e.TargetY}

	newPos := moveTowards(pos, target, e.Movement.Speed)
	w, h := e.CollisionRect().W, e.CollisionRect().H
	res = e.World.Collides(model.NewRect(newPos.X, newPos.Y, w, h), e.ID)
	if res.Collides() {
		logz.Println(e.DisplayName, "updateMovement: next position is colliding! cancelling movement")
		e.StopMovement()
		return
	}

	e.X = newPos.X
	e.Y = newPos.Y

	if math.IsNaN(e.X) || math.IsNaN(e.Y) {
		logz.Println(e.DisplayName, "e.X:", e.X, "e.Y:", e.Y)
		panic("entity position is NaN")
	}

	e.TilePos = model.ConvertPxToTilePos(int(e.X), int(e.Y))

	if target.Equals(newPos) {
		// we don't use StopMovement here since we might want to keep the animation going (if the entity has another target)
		e.Movement.IsMoving = false
		//logz.Println(e.DisplayName, "movement light stop")
		if len(e.Movement.TargetPath) == 0 {
			e.Body.StopAnimation()
		}
	}

	// footstep sound effects
	e.footstepSFX.TicksUntilNextPlay--
	if e.footstepSFX.TicksUntilNextPlay <= 0 {
		groundMaterial := e.World.GetGroundMaterial(e.TilePos.X, e.TilePos.Y)
		var distToPlayer float64
		if e.IsPlayer {
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
			e.footstepSFX.Step(audio.STEP_WOOD, volFactor)
		case "stone":
			e.footstepSFX.Step(audio.STEP_STONE, volFactor)
		case "grass":
			e.footstepSFX.Step(audio.STEP_GRASS, volFactor)
		case "tile":
			e.footstepSFX.Step(audio.STEP_STONE, volFactor) // TODO get new sound specifically for tile?
		case "":
			e.footstepSFX.Step(audio.STEP_DEFAULT, volFactor)
		default:
			// if we don't have the string registered (and it's not an empty string) then error
			panic("ground material not recognized: " + groundMaterial)
		}
	}
}

func (e *Entity) StopMovement() {
	if !e.Movement.IsMoving {
		panic("told to stop movement, but entity is not moving?")
	}
	logz.Println(e.DisplayName, "Stopping movement")
	e.TargetX = e.X
	e.TargetY = e.Y
	e.Movement.IsMoving = false
	if e.Body.IsMoving() {
		e.Body.StopAnimation()
	}
}

func moveTowards(pos, target model.Vec2, speed float64) model.Vec2 {
	dir := target.Sub(pos)
	dist := dir.Len()
	step := dir.Normalize().Scale(speed)
	if dist < speed {
		// snap to target
		return target
	}

	return pos.Add(step)
}

func moveAway(pos, target model.Vec2, speed float64) model.Vec2 {
	dir := target.Sub(pos)
	step := dir.Normalize().Scale(speed)
	return pos.Sub(step)
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
		logz.Println(e.DisplayName, "trySetNextTargetPath: entity is not at its tile position. Was it bumped by an enemy attack or something?")
		logz.Println(e.DisplayName, "tilePos:", e.TilePos, "e.X:", e.X/config.TileSize, "e.Y:", e.Y/config.TileSize)
		logz.Println(e.DisplayName, "Clamping to current tile position. TODO: perhaps we should make a more graceful way of recovering the position than this?")
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
		logz.Println(e.DisplayName, "curPos:", curPos, "target:", target, "dist:", dist)
		logz.Println(e.DisplayName, "trySetNextTargetPath: next target is not an adjacent tile (dist > 16). Clearing target path.")
		e.Movement.TargetPath = []model.Coords{}
		return MoveError{Cancelled: true, Info: "next target was not an adjacent tile (dist > tilesize)"}
	}

	moveError := e.TryMovePx(int(dPos.X), int(dPos.Y), e.WalkSpeed)

	if !moveError.Success {
		return moveError
	}

	animRes := e.SetAnimation(AnimationOptions{
		AnimationName:         body.ANIM_WALK,
		AnimationTickInterval: e.Movement.WalkAnimationTickInterval,
	})
	if !animRes.Success && !animRes.AlreadySet {
		logz.Println(e.DisplayName, "trySetNextTargetPath: failed to set animation:", animRes)
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
			//logz.Println("tryMergeSuggestedPath", "merged suggested path into current target path")
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
