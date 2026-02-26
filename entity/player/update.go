// Package player is the layer that contains all the logic for the player
package player

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
)

// MovementMechanics keep track of variables or state related to managing movement mechanics
type MovementMechanics struct {
	ticksSinceLastMouseDirect int
}

func (p Player) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if p.Entity == nil {
		panic("tried to draw player that doesn't have entity")
	}
	p.Entity.Draw(screen, offsetX, offsetY)
}

func (p *Player) Update() {
	if p.handleMovement() {
		p.LastUserInput = time.Now()
	}

	if p.handleActions() {
		p.LastUserInput = time.Now()
	}

	p.Entity.Update()
}

func (p *Player) handleMovement() bool {
	if p.Entity.Body.IsAttacking() {
		return false
	}

	curPos := model.Vec2{X: p.Entity.X, Y: p.Entity.Y}
	targetPos := model.Vec2{X: p.Entity.TargetX, Y: p.Entity.TargetY}

	// if we are still far from the target position, there's no reason to allow more player movement input
	if curPos.Dist(targetPos) > p.Entity.Movement.Speed {
		return false
	}

	v := model.Vec2{}

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		v.X -= 1
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		v.X += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		v.Y -= 1
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		v.Y += 1
	}

	running := ebiten.IsKeyPressed(ebiten.KeyShift)
	faceMouse := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	if faceMouse {
		// can't run while sidleing/facing mouse position
		running = false
	}
	if p.ticksSinceLastMouseDirect < 100 {
		p.ticksSinceLastMouseDirect++ // don't need to increment forever, just used to check for a small number of ticks
	}

	isDiagonal := v.X != 0 && v.Y != 0

	animationTickInterval := p.Entity.Movement.WalkAnimationTickInterval
	animation := body.AnimWalk
	speed := p.Entity.CharacterStateRef.WalkSpeed()
	if running {
		animationTickInterval = p.Entity.Movement.RunAnimationTickInterval
		animation = body.AnimRun
		speed = p.Entity.CharacterStateRef.RunSpeed()
	}

	travelDistance := speed * 2

	scaled := v.Normalize().Scale(travelDistance)

	if isDiagonal {
		scaled.X = math.Round(scaled.X * 2)
		scaled.Y = math.Round(scaled.Y * 2)
	}

	// we need to detect if the player is actually trying to move, for tracking if a movement action is occurring
	tryingToMove := false

	// if there is movement input and we are not blocking with a shield, move
	if v.X != 0 || v.Y != 0 {
		animRes := p.Entity.SetAnimation(entity.AnimationOptions{
			AnimationName:         animation,
			AnimationTickInterval: animationTickInterval,
		})
		if !animRes.Success && !animRes.AlreadySet {
			logz.Println(p.Entity.DisplayName(), "failed to set movement animation:", animRes)
			return false
		}
		e := p.Entity.TryMoveMaxPx(scaled.X, scaled.Y, speed)
		if !e.Success && !e.Collision {
			logz.Println(p.Entity.DisplayName(), "player failed to move:", e)
			logz.Println("", e.CollisionResult)
			p.Entity.Body.StopAnimation()
		}
		tryingToMove = true
	}
	if !p.Entity.Body.IsAttacking() {
		// if using faceMouse, make sure the player can't do it too rapidly
		if faceMouse {
			if p.ticksSinceLastMouseDirect > 5 {
				p.ticksSinceLastMouseDirect = 0
				mouseX, mouseY := ebiten.CursorPosition()
				r := p.Entity.GetDrawRect()
				p.Entity.FaceTowards(float64(mouseX)-r.X, float64(mouseY)-r.Y)
				tryingToMove = true
			}
		} else {
			p.Entity.FaceTowards(scaled.X, scaled.Y)
		}
	}

	return tryingToMove
}

// handleActions handles all user actions. returns true if an action occurred.
func (p *Player) handleActions() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		p.Entity.CharacterStateRef.UnequipWeapon()
		return true
	}

	if p.Entity.IsWeaponEquiped() {
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && !p.Entity.IsAttacking() {
			if p.Entity.IsUsingShield() {
				p.Entity.StopUsingShield()
			}
			// attack
			p.Entity.StartMeleeAttack()
			return true
		}
	}
	if p.Entity.IsShieldEquiped() {
		if p.Entity.IsUsingShield() {
			// detect if we should stop using the shield
			// - entity is moving
			// - no longer holding the right click
			if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) || isTryingToMove() {
				p.Entity.StopUsingShield()
			}
		} else {
			// entity is not using shield; detect if we should start blocking
			if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
				if !p.Entity.Movement.IsMoving {
					p.Entity.UseShield()
				}
			}
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		x, y := p.Entity.X, p.Entity.Y
		if p.World.ActivateArea(p.Entity.GetFrontRect(), x, y) {
			return true
		}
	}

	if !p.Entity.IsWeaponEquiped() {
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			mouseX, mouseY := ebiten.CursorPosition()
			p.World.HandleMouseClick(mouseX, mouseY)
			return true
		}
	}

	return false
}

// isTryingToMove detects if there is currently input related to moving
func isTryingToMove() bool {
	return ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyD)
}
