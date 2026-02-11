package entity

import (
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/model"
	"github.com/webbben/2d-game-engine/item"
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
	if !e.IsWeaponEquiped() {
		logz.Panicln(e.DisplayName(), "tried to swing weapon, but no weapon is equiped")
	}
	if e.IsStunned() {
		return
	}
	if e.IsAttacking() {
		logz.Panicln(e.DisplayName(), "tried to start melee attack, but entity is already attacking")
	}

	animationInterval := 6
	e.Body.SetAnimationTickCount(animationInterval)
	res := e.Body.SetAnimation(body.AnimSlash, body.SetAnimationOps{DoOnce: true})
	if !res.Success {
		logz.Println(e.DisplayName(), "melee attack failed:", res.String())
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
		ExcludeEntIds: []string{string(e.ID())},
		Origin:        model.Vec2{X: e.X, Y: e.Y},
	}, animationInterval*3)
}

func (e *Entity) ReceiveAttack(attack AttackInfo) {
	logz.Println(e.DisplayName(), "received attack!")
	if attack.Damage < 0 {
		panic("attack can not have negative damage")
	}
	if attack.Damage == 0 {
		// ineffectual attack
		panic("attack had 0 damage")
	}
	if e.IsUsingShield() {
		// attack was blocked; still some bump back, but no other change

		moveError := e.TryBumpBack(config.TileSize/2, defaultWalkSpeed, attack.Origin, body.AnimShield, defaultIdleAnimationTickInterval)
		if !moveError.Success {
			// perhaps there was a collision?
			logz.Println(e.DisplayName(), "shielded bump back failed:", moveError)
		}
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

	e.CharacterStateRef.Vitals.Health.CurrentVal -= attack.Damage
	logz.Println(e.DisplayName(), "current health:", e.CharacterStateRef.Vitals.Health.CurrentVal)

	e.Body.SetDamageFlicker(15)

	moveError := e.TryBumpBack(config.TileSize, defaultRunSpeed, attack.Origin, body.AnimIdle, defaultIdleAnimationTickInterval)
	if !moveError.Success {
		logz.Println(e.DisplayName(), "failed to bump back:", moveError)
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

func (e Entity) IsWeaponEquiped() bool {
	// ensure that weapon set matches equiped weapon
	partIsNone := e.Body.WeaponSet.PartSrc.None
	weaponIsNil := e.CharacterStateRef.Equipment.EquipedWeapon == nil
	if weaponIsNil == partIsNone {
		return !weaponIsNil
	}
	// uh oh - we have a bugged case here. let's panic so it can be noticed and fixed.
	logz.Panicln(e.DisplayName(), "equiped weapon slot and weapon body part don't seem to match... weapon is nil?:", weaponIsNil, "part is none?:", partIsNone)
	return false
}

func (e Entity) IsShieldEquiped() bool {
	if e.CharacterStateRef.Equipment.EquipedAuxiliary != nil {
		if item.IsArmor(e.CharacterStateRef.Equipment.EquipedAuxiliary.Def) {
			return true
		}
	}
	return false
}

func (e *Entity) UseShield() {
	if !e.IsShieldEquiped() {
		logz.Panicln(e.DisplayName(), "tried to use shield, but shield is not equipped")
	}
	if e.IsUsingShield() {
		return
	}
	if e.IsStunned() {
		return
	}

	res := e.Body.SetAnimation(body.AnimShield, body.SetAnimationOps{QueueNext: true})
	if !res.Success {
		// probably fails because shield is already set (or a different action like attack is ongoing).
		// not adding checks unless I find weird behavior in the future
		return
	}
}

func (e *Entity) StopUsingShield() {
	if !e.IsUsingShield() {
		logz.Panicln(e.DisplayName(), "trying to stop using shield, but shield isn't being used")
	}

	res := e.Body.SetAnimation(body.AnimIdle, body.SetAnimationOps{Force: true})
	if !res.Success {
		logz.Panicln(e.DisplayName(), "failed to unset shield animation...")
	}
}

func (e Entity) IsUsingShield() bool {
	return e.Body.GetCurrentAnimation() == body.AnimShield
}

func (e Entity) IsAttacking() bool {
	return e.Body.IsAttacking()
}
