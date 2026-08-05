package entity

import (
	"fmt"
	"image/color"
	"time"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/pubsub"
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
	Attacker      id.CharacterStateID
	Damage        int
	StunTicks     int
	TargetRect    model.Rect
	ExcludeEntIds []string
	Origin        model.Vec2
}

// GetFrontRect returns the tile rect that is right in front of the entity
func (e Entity) GetFrontRect() model.Rect {
	// we want to make a rect that's a little smaller, so that we can't reach things that are too far away.
	// It seems like a rect that is a full tilesize is a little too big
	targetRectSize := config.TileSize * 3 / 4
	collisionRect := e.CollisionRect()
	offsetX := (collisionRect.W - float64(targetRectSize)) / 2
	offsetY := (collisionRect.H - float64(targetRectSize)) / 2
	targetRect := model.Rect{
		W: float64(targetRectSize),
		H: float64(targetRectSize),
		X: e.X + offsetX,
		Y: e.Y + offsetY,
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
		logz.Println(e.DisplayName(), "tried to swing weapon, but no weapon is equiped")
		logz.Panicln("Combat", "tried to swing weapon, but no weapon is equiped")
	}
	if e.IsStunned() {
		return
	}
	if e.IsAttacking() {
		logz.Println(e.DisplayName(), "tried to start melee attack, but entity is already attacking")
		logz.Panicln("Combat", "tried to start melee attack, but entity is already attacking")
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

	if e.characterStateRef.EquipedWeapon != nil {
		weaponDef := e.dataman.GetItemDef(e.characterStateRef.EquipedWeapon.DefID)
		if weaponDef.SwingSFX != "" {
			e.footstepSFX.AudioMgr.PlaySFX(weaponDef.SwingSFX, 0.5)
		}
	}

	e.queueAttack(AttackInfo{
		Attacker:      e.ID(),
		Damage:        10,
		StunTicks:     20,
		TargetRect:    e.GetFrontRect(),
		ExcludeEntIds: []string{string(e.ID())},
		Origin:        model.Vec2{X: e.X, Y: e.Y},
	}, animationInterval*3)
}

func (e *Entity) ReceiveAttack(attack AttackInfo) {
	if e.IsDead() {
		logz.Println(string(e.ID()), "received attack, but entity is dead")
		logz.Panicln("Combat", "received attack, but entity is dead")
	}
	logz.Println(e.DisplayName(), "received attack!")
	if attack.Damage < 0 {
		logz.Println(string(e.ID()), attack)
		logz.Panicln("Combat", "attack can not have negative damage.")
	}
	if attack.Damage == 0 {
		// ineffectual attack
		logz.Println(string(e.ID()), attack)
		logz.Panicln("Combat", "attack had 0 damage.")
	}
	if attack.Attacker == "" {
		logz.Println("ReceiveAttack", e.ID())
		logz.Panicln("ReceiveAttack", "no attacker info")
	}

	eventInfo := map[string]any{
		"attacker":    attack.Attacker,
		"receiver":    e.ID(),
		"receiverPos": model.Vec2{X: e.drawX, Y: e.drawY},
		"damage":      attack.Damage,
	}

	params := FloatTextParams{
		Font:     config.DefaultInfoFont, // TODO: add new font for float text?
		Color:    color.RGBA{255, 0, 0, 0},
		Duration: time.Second * 2,
	}
	txt := fmt.Sprintf("-%v", attack.Damage)

	if e.IsUsingShield() {
		// attack was blocked; still some bump back, but no other change
		eventInfo["blocked"] = true
		moveError := e.TryBumpBack(config.TileSize/2, defaultWalkSpeed, attack.Origin, body.AnimShield, defaultIdleAnimationTickInterval)
		if !moveError.Success {
			// perhaps there was a collision?
			logz.Println(e.DisplayName(), "shielded bump back failed:", moveError)
		}

		// play shield hit sound
		e.playHitSFX(e.characterStateRef.EquipedAuxiliary)
		// TODO: damage shield item

		// TODO: once damage reduction/partial blocks are calculated, the damage dealt (blockedDamage) may be greater than 0.
		blockedDamage := 0
		e.FloatMGMT.AddFloatText(NewFloatText(fmt.Sprintf("%v", blockedDamage), FloatTextParams{
			Font:     config.DefaultInfoFont,
			Color:    color.RGBA{200, 200, 200, 0},
			Duration: time.Second * 2,
		}))

		e.eventBus.Publish(defs.Event{
			Type: pubsub.EventAttackEntity,
			Data: eventInfo,
		})
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

	e.characterStateRef.Health -= attack.Damage
	logz.Println(e.DisplayName(), "current health:", e.characterStateRef.Health)

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

	// play armor hit sound (body armor, or default if none)
	e.playHitSFX(e.characterStateRef.EquipedBodywear)

	e.FloatMGMT.AddFloatText(NewFloatText(txt, params))

	e.eventBus.Publish(defs.Event{
		Type: pubsub.EventAttackEntity,
		Data: eventInfo,
	})
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
	weaponIsNil := e.characterStateRef.EquipedWeapon == nil
	if weaponIsNil == partIsNone {
		return !weaponIsNil
	}
	// uh oh - we have a bugged case here. let's panic so it can be noticed and fixed.
	logz.Println(e.DisplayName(), weaponIsNil, "part is none?:", partIsNone)
	logz.Panicln("Combat", "equiped weapon slot and weapon body part don't seem to match")
	return false
}

func (e Entity) IsShieldEquiped() bool {
	if e.characterStateRef.EquipedAuxiliary != nil {
		return e.equipedAuxiliary.Protection > 0
	}
	return false
}

func (e *Entity) UseShield() {
	if !e.IsShieldEquiped() {
		logz.Println(e.DisplayName(), "tried to use shield, but shield is not equipped")
		logz.Panicln("Combat", "tried to use shield, but shield is not equipped")
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
		logz.Println(e.DisplayName(), "trying to stop using shield, but shield isn't being used")
		logz.Panicln("Combat", "trying to stop using shield, but shield isn't being used")
	}

	res := e.Body.SetAnimation(body.AnimIdle, body.SetAnimationOps{Force: true})
	if !res.Success {
		logz.Println(e.DisplayName(), "failed to unset shield animation...")
		logz.Panicln("Combat", "failed to unset shield animation...")
	}
}

func (e Entity) IsUsingShield() bool {
	return e.Body.GetCurrentAnimation() == body.AnimShield
}

func (e Entity) IsAttacking() bool {
	return e.Body.IsAttacking()
}

func (e *Entity) playHitSFX(equipedItem *state.ItemState) {
	if equipedItem != nil {
		itemDef := e.dataman.GetItemDef(equipedItem.DefID)
		if itemDef.HitSFX != "" {
			e.footstepSFX.AudioMgr.PlaySFX(itemDef.HitSFX, 0.5)
			return
		}
	}
	if config.DefaultHitSfx != "" {
		e.footstepSFX.AudioMgr.PlaySFX(config.DefaultHitSfx, 0.5)
	}
}
