package entity

import (
	"fmt"
	"math"

	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

type MoveError struct {
	AlreadyMoving bool
	Collision     bool
	Cancelled     bool
	Success       bool
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

func (e *Entity) TryMovePx(dx, dy int) MoveError {
	if dx == 0 && dy == 0 {
		panic("TryMovePx: dx and dy are both 0")
	}
	x := int(e.TargetX) + dx
	y := int(e.TargetY) + dy
	targetRect := model.Rect{X: float64(x), Y: float64(y), W: e.width, H: e.width}

	if e.World.Collides(targetRect, e.ID, e.IsPlayer) {
		return MoveError{
			Collision: true,
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
	if e.IsPlayer {
		e.footstepSFX.TicksUntilNextPlay--
		if e.footstepSFX.TicksUntilNextPlay <= 0 {
			e.footstepSFX.StepDefault()
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
