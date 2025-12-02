package entity

import (
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
)

type attackManager struct {
	attackQueued         bool
	attackTicksRemaining int
	// for "queuing" attacks to happen, when the attack should not register at the start of the attack animation.
	// for example, if a swing animation has a cock-back portion that is two frames, you may want to "queue" the attack to happen after two ticks
	// so that the damage is applied at the right timing.
	queuedAttack    AttackInfo
	waitingToAttack bool // set when entity should trigger attack once movement or other things are done
}

func (am *attackManager) clearAttack() {
	am.attackQueued = false
	am.attackTicksRemaining = 0
	am.queuedAttack = AttackInfo{}
}

func (am *attackManager) queueAttack(attackInfo AttackInfo, delayTicks int) {
	if am.attackQueued {
		return
	}
	am.attackQueued = true
	am.attackTicksRemaining = delayTicks
	am.queuedAttack = attackInfo
}

func (e *Entity) updateAttackManager() {
	if e.waitingToAttack {
		if !e.Movement.IsMoving && e.Body.GetCurrentAnimation() == body.AnimIdle {
			e.StartMeleeAttack()
			e.waitingToAttack = false
		}
	}
	if !e.attackQueued {
		return
	}

	e.attackTicksRemaining--
	if e.attackTicksRemaining <= 0 {
		e.World.AttackArea(e.queuedAttack)
		e.clearAttack()
	}
}

type AttackInfo struct {
	Damage        int
	StunTicks     int
	TargetRect    model.Rect
	ExcludeEntIds []string
	Origin        model.Vec2
}

// GetFrontRect returns the tile rect that is right in front of the entity
func (e Entity) GetFrontRect() model.Rect {
	targetRect := model.Rect{
		W: config.TileSize,
		H: config.TileSize,
		X: e.X,
		Y: e.Y,
	}
	switch e.Movement.Direction {
	case model.Directions.Left:
		targetRect.X -= config.TileSize
	case model.Directions.Right:
		targetRect.X += config.TileSize
	case model.Directions.Up:
		targetRect.Y -= config.TileSize
	case model.Directions.Down:
		targetRect.Y += config.TileSize
	}
	return targetRect
}

func (e *Entity) StartMeleeAttack() {
	if e.Body.WeaponSet.PartSrc.None {
		panic("tried to swing weapon, but no weapon is equiped")
	}
	if e.IsStunned() {
		return
	}

	animationInterval := 6
	e.Body.SetAnimationTickCount(animationInterval)
	res := e.Body.SetAnimation(body.AnimSlash, body.SetAnimationOps{DoOnce: true})
	if !res.Success {
		if !res.AlreadySet {
			// if not already attacking, then just wait to do the attack once whatever the current animation is finishes
			e.waitingToAttack = true
		}
		// already attacking - need to wait until the animation is done before attacking again
		return
	}

	e.queueAttack(AttackInfo{
		Damage:        10,
		StunTicks:     20,
		TargetRect:    e.GetFrontRect(),
		ExcludeEntIds: []string{e.ID},
		Origin:        model.Vec2{X: e.X, Y: e.Y},
	}, animationInterval*3)
}

func (e *Entity) ReceiveAttack(attack AttackInfo) {
	logz.Println(e.DisplayName, "received attack!")
	if attack.Damage < 0 {
		panic("attack can not have negative damage")
	}
	if attack.Damage == 0 {
		// ineffectual attack
		return
	}

	// unset all attacks or pending attack logic
	if e.Body.IsAttacking() {
		// attack animations should be interrupted
		e.Body.StopAnimation()
	}
	if e.attackQueued {
		// if an attack is interrupted, clear the queued damage signal
		e.clearAttack()
	}
	e.waitingToAttack = false

	e.Vitals.Health.CurrentVal -= attack.Damage
	logz.Println(e.DisplayName, "current health:", e.Vitals.Health.CurrentVal)

	e.Body.SetDamageFlicker(15)

	moveError := e.TryBumpBack(config.TileSize, defaultRunSpeed, attack.Origin, body.AnimIdle, defaultIdleAnimationTickInterval)
	if !moveError.Success {
		logz.Println(e.DisplayName, "failed to bump back:", moveError)
		if !moveError.Collision {
			logz.Panic("bump back failed, but it wasn't due to a collision. the bump back should always succeed unless the entity is up against a wall")
		}
	}

	if attack.StunTicks > 0 {
		e.stun(attack.StunTicks)
	}
}

func (e *Entity) stun(ticks int) {
	e.stunTicks = ticks
}

func (e Entity) IsStunned() bool {
	return e.stunTicks > 0
}

func (e *Entity) UnequipWeaponFromBody() {
	if e.IsStunned() {
		return
	}
	e.Body.WeaponSet.PartSrc.None = true
	e.Body.WeaponFxSet.PartSrc.None = true
	e.Body.Load()
}

func (e *Entity) EquipWeapon(weaponDef body.SelectedPartDef, weaponFxDef body.SelectedPartDef) {
	if e.IsStunned() {
		return
	}
	e.Body.SetWeapon(weaponDef, weaponFxDef)
}

func (e Entity) IsWeaponEquiped() bool {
	return !e.Body.WeaponSet.PartSrc.None
}
