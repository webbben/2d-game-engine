package entity

import (
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
	"github.com/webbben/2d-game-engine/data/state"
	"github.com/webbben/2d-game-engine/entity/body"
	characterstate "github.com/webbben/2d-game-engine/entity/characterState"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/model"
	"github.com/webbben/2d-game-engine/pubsub"
)

const ticksPerSecond = 60 // ebiten ticks run at 60 per second

type attackManager struct {
	attackQueued          bool
	waitingToFinishAttack bool // gets set if FinishMeleeAttack is called before the attack start animation is complete
	queuedAttack          AttackInfo
	waitingToAttack       bool // set when entity should trigger attack once movement or other things are done

	chargeStartTick int64 // tick at which the attack began charging (when the wind-up/start animation finished)
}

func (am *attackManager) clearAttack() {
	am.attackQueued = false
	am.queuedAttack = AttackInfo{}
	am.chargeStartTick = 0
}

func (am *attackManager) queueAttack(attackInfo AttackInfo) {
	if am.attackQueued {
		return
	}
	am.attackQueued = true
	am.queuedAttack = attackInfo
}

func (e *Entity) updateAttackManager() {
	// once the wind-up (start) animation has finished and is holding its pose, charging has begun.
	// record the tick this first happens so the power attack multiplier reflects actual hold time.
	if e.chargeStartTick == 0 && e.Body.GetCurrentAnimation() == body.AnimSlashStart && e.Body.AnimationFinished() {
		e.chargeStartTick = ebiten.Tick()
	}

	if e.waitingToAttack {
		if !e.Movement.IsMoving && e.Body.GetCurrentAnimation() == body.AnimIdle {
			e.StartMeleeAttack()
			e.waitingToAttack = false
		}
		return
	}

	if e.waitingToFinishAttack {
		if !e.IsAttacking() {
			logz.PanicCtx("updateAttackManager", "waiting to finish attack, but entity is not attacking...", e.ID(), e.Body.GetCurrentAnimation(), e.queuedAttack)
		}
		if e.Body.AnimationFinished() {
			// attack start animation is done, so proceed to finish the attack
			e.waitingToFinishAttack = false
			e.FinishMeleeAttack()
		}
	}
}

type AttackInfo struct {
	StartTick     int64
	Attacker      id.CharacterStateID
	Damage        defs.RealDamage
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

// TargetInMeleeReach returns true if the target is within the area this entity's melee attack would hit.
func (e Entity) TargetInMeleeReach(target *Entity) bool {
	return e.GetFrontRect().Intersects(target.CollisionRect())
}

// StartMeleeAttack begins a melee attack. Once started, we await "FinishMeleeAttack" to actually perform the attack and do damage.
// Until that function is called, it just holds the "start melee attack" pose.
// This is split into two functions to allow "charge ups" for power attacks.
func (e *Entity) StartMeleeAttack() {
	if !e.IsWeaponEquiped() {
		logz.PanicCtx("Combat", "tried to swing weapon, but no weapon is equiped", e.DisplayName())
	}
	if e.IsStunned() {
		return
	}
	if e.IsAttacking() {
		logz.PanicCtx("Combat", "tried to start melee attack, but entity is already attacking", e.DisplayName())
	}
	e.chargeStartTick = 0

	// TODO: why are we directly calling Body.SetAnimation instead of the Entity.SetAnimation function?
	animationInterval := 6
	e.Body.SetAnimationTickCount(animationInterval)
	res := e.Body.SetAnimation(body.AnimSlashStart, body.SetAnimationOps{
		HoldLastFrame: true,
	})
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
		StartTick:     ebiten.Tick(),
		Attacker:      e.ID(),
		StunTicks:     20,
		TargetRect:    e.GetFrontRect(),
		ExcludeEntIds: []string{string(e.ID())},
		Origin:        model.Vec2{X: e.X, Y: e.Y},
	})
}

func (e *Entity) FinishMeleeAttack() {
	if !e.attackQueued {
		logz.PanicCtx("FinishMeleeAttack", "no attack was in the queue", e.ID(), "current anim:", e.Body.GetCurrentAnimation())
	}
	if !e.IsAttacking() {
		logz.PanicCtx("FinishMeleeAttack", "entity is not currently attacking...", e.ID(), "current anim:", e.Body.GetCurrentAnimation(), "queued attack:", e.queuedAttack)
	}
	if !e.Body.AnimationFinished() {
		// wait until the start animation is done first; player or NPC probably called StartMeleeAttack but didn't try to charge the attack at all.
		e.waitingToFinishAttack = true
		return
	}

	// power attacks: the longer the attack was charged (held in the wind-up pose after the start
	// animation finished), the higher the damage multiplier. this is computed outside of the engine
	// via CombatSystemCalc since different games may want different balance.
	currentTick := ebiten.Tick()
	chargeTicks := currentTick - e.chargeStartTick
	if e.chargeStartTick == 0 {
		// charge start was never recorded (e.g. release landed on the same tick the wind-up finished);
		// treat this as an uncharged attack.
		chargeTicks = 0
	}
	if chargeTicks < 0 {
		// tick overflow/wrap-around; shouldn't happen, but don't let it break the attack.
		logz.Warnln("FinishMeleeAttack", "charge start tick was greater than current tick! did the tick integer overflow/wrap back to 0?", "charge start tick:", e.chargeStartTick, "current tick:", currentTick)
		chargeTicks = 0
	}
	chargeDuration := time.Duration(chargeTicks) * time.Second / ticksPerSecond
	mult := e.dataman.CombatSystemCalc.PowerAttackMultiplier(chargeDuration)

	// calculate damage and multiplier now that melee attack charging is done
	weaponID := e.equipedWeapon.ID
	condition := e.characterStateRef.EquipedWeapon.Durability
	weaponType := e.equipedWeapon.GoverningSkill
	skills, attrs := characterstate.CalculateSkillsAndAttributes(e.characterStateRef.ID, e.dataman)
	dmg := e.dataman.CombatSystemCalc.MeleeWeaponDamage(weaponID, condition, mult, weaponType, attrs, skills)
	e.queuedAttack.Damage = dmg

	if e.characterStateRef.EquipedWeapon != nil {
		weaponDef := e.dataman.GetItemDef(e.characterStateRef.EquipedWeapon.DefID)
		if weaponDef.SwingSFX != "" {
			e.footstepSFX.AudioMgr.PlaySFX(weaponDef.SwingSFX, 0.5)
		}
	}

	e.World.AttackArea(e.queuedAttack)
	e.clearAttack()

	e.SetAnimation(AnimationOptions{
		AnimationName:         body.AnimSlashFinish,
		AnimationTickInterval: 6, // TODO: should this be a const or calculated from something? we already used this value before for melee attacks
		SetAnimationOps: body.SetAnimationOps{
			DoOnce: true,
			Force:  true, // TODO: is force needed?
		},
	})
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

	realDamage := attack.Damage
	finalDamage := e.dataman.CombatSystemCalc.CalculateFinalDamage(realDamage, e.equippedArmorProtection)
	// damage is applied to health and shown to the player as a whole number; the calculation itself stays in floats.
	damageDealt := int(finalDamage)

	eventInfo := map[string]any{
		"attacker":    attack.Attacker,
		"receiver":    e.ID(),
		"receiverPos": model.Vec2{X: e.drawX, Y: e.drawY},
		"damage":      damageDealt,
	}

	params := FloatTextParams{
		Font:     config.DefaultInfoFont, // TODO: add new font for float text?
		Color:    color.RGBA{255, 0, 0, 0},
		Duration: time.Second * 2,
	}
	txt := fmt.Sprintf("-%v", damageDealt)

	if e.IsUsingShield() {
		if !e.IsShieldEquiped() {
			logz.Panicln("ReceiveAttack", "entity is using shield, but no shield is equipped")
		}

		// attack was blocked; still some bump back, but no other change
		eventInfo["blocked"] = true
		moveError := e.TryBumpBack(config.TileSize/2, defaultWalkSpeed, attack.Origin, body.AnimShield, defaultIdleAnimationTickInterval)
		if !moveError.Success {
			// perhaps there was a collision?
			logz.Println(e.DisplayName(), "shielded bump back failed:", moveError)
		}

		// play shield hit sound
		e.playHitSFX(e.characterStateRef.EquipedAuxiliary)

		wear := e.dataman.CombatSystemCalc.ShieldBlockDurabilityLoss(realDamage)
		e.characterStateRef.EquipedAuxiliary.Durability = max(0, e.characterStateRef.EquipedAuxiliary.Durability-wear)

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
	e.waitingToFinishAttack = false

	// apply armor deterioration
	// each equipped armor item rolls independently to take wear, weighted by its share of the total
	// base (authored) protection. base protection is used rather than real (condition-scaled) protection,
	// so the selection probabilities stay stable regardless of each piece's current durability.
	type armorWearCandidate struct {
		armorItem *state.ItemState
		itemDef   defs.ItemDef
	}
	var candidates []armorWearCandidate
	var totalBaseProtection defs.BaseProtection
	armorItems := []*state.ItemState{
		e.characterStateRef.EquipedHeadwear,
		e.characterStateRef.EquipedBodywear,
		e.characterStateRef.EquipedFootwear,
		e.characterStateRef.EquipedAuxiliary,
	}
	for _, armorItem := range armorItems {
		if armorItem == nil {
			continue
		}
		itemDef := e.dataman.GetItemDef(armorItem.DefID)
		if itemDef.Protection <= 0 {
			// only do armor items
			continue
		}
		candidates = append(candidates, armorWearCandidate{armorItem: armorItem, itemDef: itemDef})
		totalBaseProtection += itemDef.Protection
	}
	armorWorn := false
	for _, c := range candidates {
		tookWear, wear := e.dataman.CombatSystemCalc.ArmorDurabilityLoss(c.itemDef.Protection, totalBaseProtection, realDamage)
		if tookWear {
			c.armorItem.Durability = max(0, c.armorItem.Durability-wear)
			armorWorn = true
			// TODO: add some kind of effect (breaking sfx?) to show that the armor is broken.
			// we could even unequip it, but not sure if that's the best option or not. Maybe there's some kind of visual effect and/or float text we could show though.
		}
	}
	if armorWorn {
		// armor protection scales with durability, so refresh the cached value
		e.equippedArmorProtection = e.calculateArmorProtection()
	}

	e.characterStateRef.Health -= damageDealt
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

func (e *Entity) HandleWeaponHit(target *Entity) {
	if e.characterStateRef.EquipedWeapon == nil {
		logz.Println(string(e.ID()))
		logz.Panicln("HandleWeaponHit", "no weapon was equipped")
	}
	wear := e.dataman.CombatSystemCalc.WeaponDurabilityLoss(target.equippedArmorProtection)
	e.characterStateRef.EquipedWeapon.Durability = max(0, e.characterStateRef.EquipedWeapon.Durability-wear)
}
