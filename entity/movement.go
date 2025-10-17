package entity

import (
	"fmt"
	"math"

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
}

func (me MoveError) String() string {
	if me.Success {
		return "Success"
	}
	if me.Collision {
		return "Collision"
	}
	if me.AlreadyMoving {
		return "Already Moving"
	}
	if me.Cancelled {
		return "Cancelled"
	}
	return "No value set"
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
		return c, MoveError{Cancelled: true}
	}

	path, found := e.World.FindPath(e.TilePos, c)
	if !found {
		if !closeEnough {
			return c, MoveError{Cancelled: true}
		}
		logz.Warnln(e.DisplayName, "going a partial path since original path is blocked.", "start:", e.TilePos, "path:", path, "original goal:", c)
	}
	if len(path) == 0 {
		fmt.Println("tile pos:", e.TilePos, "goal:", c)
		logz.Warnln(e.DisplayName, "GoToPos: calculated path is empty. Is the entity completely blocked in?")
		return c, MoveError{Cancelled: true}
	}

	e.Movement.TargetPath = path
	return path[len(path)-1], MoveError{Success: true}
}

// gets the frameType and frameCount for the correct animation, based on the entity's direction
func (e Entity) getMovementAnimationInfo() (string, int) {
	var name string
	switch e.Movement.Direction {
	case model.Directions.Left:
		if e.Movement.IsMoving || !e.Movement.movementStopped {
			name = "left_walk"
		} else {
			name = "left_idle"
		}
	case model.Directions.Right:
		if e.Movement.IsMoving || !e.Movement.movementStopped {
			name = "right_walk"
		} else {
			name = "right_idle"
		}
	case model.Directions.Up:
		if e.Movement.IsMoving || !e.Movement.movementStopped {
			name = "up_walk"
		} else {
			name = "up_idle"
		}
	case model.Directions.Down:
		if e.Movement.IsMoving || !e.Movement.movementStopped {
			name = "down_walk"
		} else {
			name = "down_idle"
		}
	default:
		panic("incorrect direction value found during UpdateMovement")
	}

	count, exists := e.AnimationFrameCount[name]
	if !exists {
		panic("animation name has no record in AnimationFrameCount map: " + name)
	}

	return name, count
}

// Tells the entity to stop moving once it has finished its current tile movement.
// Meant for stopping entities that are currently following a path.
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

// same as TryMovePx, but lets the entity still move in the direction even if a collision is encountered,
// as long as there is some space in that direction
func (e *Entity) TryMoveMaxPx(dx, dy int) MoveError {
	if dx == 0 && dy == 0 {
		panic("TryMoveMaxPx called with no distance given")
	}
	moveError := e.TryMovePx(dx, dy)
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

		return e.TryMovePx(dx, dy)
	}
	return moveError
}

func (e *Entity) TryMovePx(dx, dy int) MoveError {
	if dx == 0 && dy == 0 {
		panic("TryMovePx: dx and dy are both 0")
	}
	x := int(e.TargetX) + dx
	y := int(e.TargetY) + dy
	targetRect := model.Rect{X: float64(x), Y: float64(y), W: e.width, H: e.width}

	res := e.World.Collides(targetRect, e.ID, e.IsPlayer)
	if res.Collides() {
		return MoveError{
			Collision:       true,
			CollisionResult: res,
		}
	}

	e.Position.TargetX = float64(x)
	e.Position.TargetY = float64(y)

	if dx != 0 {
		if dx > 0 {
			e.Movement.Direction = model.Directions.Right
		} else {
			e.Movement.Direction = model.Directions.Left
		}
	} else {
		if dy > 0 {
			e.Movement.Direction = model.Directions.Down
		} else {
			e.Movement.Direction = model.Directions.Up
		}
	}

	e.Movement.IsMoving = true
	e.Movement.Speed = e.Movement.WalkSpeed

	return MoveError{Success: true}
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

	e.Movement.movementStopped = false

	pos := model.Vec2{X: e.X, Y: e.Y}
	target := model.Vec2{X: e.TargetX, Y: e.TargetY}

	newPos := moveTowards(pos, target, e.Movement.Speed)
	e.X = newPos.X
	e.Y = newPos.Y

	if math.IsNaN(e.X) || math.IsNaN(e.Y) {
		logz.Println(e.DisplayName, "e.X:", e.X, "e.Y:", e.Y)
		panic("entity position is NaN")
	}

	e.TilePos = model.ConvertPxToTilePos(int(e.X), int(e.Y))

	if target.Equals(newPos) {
		e.Movement.IsMoving = false
	}

	// update animation
	e.Movement.AnimationTimer++
	if e.Movement.AnimationTimer > 10 {
		_, frameCount := e.getMovementAnimationInfo()
		e.Movement.AnimationFrame = (e.Movement.AnimationFrame + 1) % frameCount
		e.Movement.AnimationTimer = 0
	}
	e.footstepSFX.TicksUntilNextPlay--
	if e.footstepSFX.TicksUntilNextPlay <= 0 {
		groundMaterial := e.World.GetGroundMaterial(e.TilePos.X, e.TilePos.Y)
		switch groundMaterial {
		case "wood":
			e.footstepSFX.StepWood()
		case "stone":
			e.footstepSFX.StepStone()
		case "grass":
			e.footstepSFX.StepGrass()
		case "tile":
			e.footstepSFX.StepStone() // TODO add new category for tile material?
		case "":
			e.footstepSFX.StepDefault() // if no string found, use default
		default:
			// if we don't have the string registered (and it's not an empty string) then error
			panic("ground material not recognized: " + groundMaterial)
		}
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

func (e *Entity) trySetNextTargetPath() MoveError {
	if len(e.Movement.TargetPath) == 0 {
		panic("tried to set next target along path for entity that has no set target path")
	}
	nextTarget := e.Movement.TargetPath[0]
	if nextTarget.Equals(e.TilePos) {
		panic("trySetNextTargetPath: next target is the same tile as current position")
	}

	curPos := model.Vec2{X: e.X, Y: e.Y}
	target := model.Vec2{X: float64(nextTarget.X * config.TileSize), Y: float64(nextTarget.Y * config.TileSize)}
	dist := curPos.Dist(target)
	dPos := target.Sub(curPos)

	if dist > 16 {
		logz.Println(e.DisplayName, "curPos:", curPos, "target:", target, "dist:", dist)
		panic("trySetNextTargetPath: next target is not an adjacent tile (dist > 16)")
	}

	moveError := e.TryMovePx(int(dPos.X), int(dPos.Y))

	if !moveError.Success {
		return moveError
	}

	e.Movement.TargetPath = e.Movement.TargetPath[1:]
	e.Movement.IsMoving = true
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
