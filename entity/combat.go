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
	queuedAttack         AttackInfo
	waitingToAttack      bool // set when entity should trigger attack once movement or other things are done
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
		if !e.Movement.IsMoving && e.Body.GetCurrentAnimation() == "" {
			e.StartMeleeAttack()
			e.waitingToAttack = false
		}
	}
	if !e.attackManager.attackQueued {
		return
	}

	e.attackManager.attackTicksRemaining--
	if e.attackManager.attackTicksRemaining <= 0 {
		e.World.AttackArea(e.attackManager.queuedAttack)
		e.attackManager.clearAttack()
	}
}

type AttackInfo struct {
	Damage        int
	TargetRect    model.Rect
	ExcludeEntIds []string
	Origin        model.Vec2
}

// returns the tile rect that is right in front of the entity
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
	if e.Body.WeaponSet.None {
		panic("tried to swing weapon, but no weapon is equiped")
	}
	animationInterval := 6
	e.Body.SetAnimationTickCount(animationInterval)
	res := e.Body.SetAnimation(body.ANIM_SLASH, body.SetAnimationOps{DoOnce: true})
	if !res.Success {
		if !res.AlreadySet {
			// if not already attacking, then just wait to do the attack once whatever the current animation is finishes
			e.waitingToAttack = true
		}
		// already attacking - need to wait until the animation is done before attacking again
		return
	}

	e.attackManager.queueAttack(AttackInfo{
		Damage:        10,
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

	e.Vitals.Health.CurrentVal -= attack.Damage
	logz.Println(e.DisplayName, "current health:", e.Vitals.Health.CurrentVal)

	e.Body.SetDamageFlicker(15)

	moveError := e.TryBumpBack(config.TileSize, defaultRunSpeed, attack.Origin, "", 0)
	if !moveError.Success {
		logz.Println(e.DisplayName, "failed to bump back:", moveError)
	}
}

func (e *Entity) UnequipWeaponFromBody() {
	e.Body.WeaponSet.None = true
	e.Body.WeaponFxSet.None = true
	e.Body.Load()
}

func (e *Entity) EquipWeapon(weaponDef body.SelectedPartDef, weaponFxDef body.SelectedPartDef) {
	e.Body.SetWeapon(weaponDef, weaponFxDef)
}

func (e Entity) IsWeaponEquiped() bool {
	return !e.Body.WeaponSet.None
}
