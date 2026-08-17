package npc

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/logz"
)

type fightStatus int

const (
	fightStatusIdle fightStatus = iota
	fightStatusFollow
	fightStatusCombat
)

// powerAttackChance is the probability that an NPC swings a charged power attack instead of a quick regular attack.
const powerAttackChance = 0.2

// slashStartWindUpTicks is how long the "slash-start" (wind-up) animation takes to play before charge time begins
// counting. chargeAttackTicks must exceed this for the attack to build any power-attack charge.
// (2 frames of slash-start, tick interval of 6 -> 12 ticks.)
const slashStartWindUpTicks = 12

func (fs fightStatus) String() string {
	switch fs {
	case fightStatusIdle:
		return "idle (0)"
	case fightStatusFollow:
		return "follow (1)"
	case fightStatusCombat:
		return "combat (2)"
	default:
		return "unregistered status!"
	}
}

type FightTask struct {
	TaskBase

	status       fightStatus
	targetEntity *entity.Entity

	nextAttackTime time.Time
	shieldEndTime  time.Time

	// when attacking, this is set and once it hits 0 the NPC should finish the attack
	chargeAttackTicks int
}

type FightTaskParams struct {
	TargetEntity *entity.Entity
	// TODO: add params to indicate how "serious" the fight is?
	// e.g. can the NPC surrender or not, etc. Probably a future thing once combat is more advanced.
}

var _ Task = (*FightTask)(nil)

func NewFightTask(targetEnt *entity.Entity, owner *NPC, p defs.TaskPriority, nextTask *defs.TaskDef) *FightTask {
	if targetEnt == nil {
		panic("target is nil")
	}
	if owner == nil {
		panic("owner was nil")
	}
	t := defs.TaskDef{
		TaskID:   TaskFight,
		Priority: p,
		NextTask: nextTask,
	}
	return &FightTask{
		TaskBase:     NewTaskBase(t, "Fight", "Fight another entity", owner),
		status:       fightStatusIdle,
		targetEntity: targetEnt,
	}
}

func init() {
	registerTask(TaskFight, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			fightParams, ok := def.Params.(FightTaskParams)
			if !ok {
				logz.Println("FightTask", def.Params)
				logz.Panicln("FightTask", "tried to run a fight task, but the params could not be converted into FightTaskParams. make sure you are using the right struct")
			}
			return NewFightTask(fightParams.TargetEntity, owner, def.Priority, def.NextTask)
		},
		validateParams: func(def defs.TaskDef) error {
			params, ok := def.Params.(FightTaskParams)
			if !ok {
				return fmt.Errorf("FightTask params must be FightTaskParams, got %T", def.Params)
			}
			if params.TargetEntity == nil {
				return fmt.Errorf("FightTask params has a nil TargetEntity")
			}
			return nil
		},
	})
}

/*
Fight Task:

1. Go near the enemy at normal speed, up to a certain distance (a few tiles or so)
2. Once close to the enemy, be "on guard" and move towards the enemy a bit slower and always face towards the enemy
3. Once in striking distance, attack every few seconds or so
4. If the enemy runs away, go back to [1]

*/

func (t *FightTask) Start() {
	if t.targetEntity == nil {
		panic("target entity must not be nil")
	}
	if t.Owner == nil {
		panic("owner entity must not be nil")
	}
	if t.status != fightStatusIdle {
		logz.Println("FightTask", "status:", t.status)
		logz.Panicf("Start: fight task should be idle when (re)starting")
	}

	t.Status = TaskInProg

	// get real path distance first, to determine if we need to follow
	dist := t.Owner.Entity.DistFromEntity(*t.targetEntity)
	if dist > config.TileSize*3 {
		t.startFollowing()
		return
	}

	// close enough already; start the combat portion of this task
	t.startCombat()
}

func (t *FightTask) startFollowing() {
	if t.status != fightStatusIdle {
		panic("fight status should be idle before trying to follow")
	}
	if !t.ChildDone() {
		logz.Panicf("follow subtask appears to already be active. It should've ended (or not started yet) before Start was called.")
	}
	logz.Println(t.Owner.DisplayName(), "start follow")

	t.RunChild(NewFollowTask(t.targetEntity, 0, t.Owner, Emergency, nil))
	t.status = fightStatusFollow
}

func (t *FightTask) stopFollowing() {
	if t.status != fightStatusFollow {
		logz.Panic("trying to stop following, but not in the following state")
	}
	if !t.HasChild() || t.ChildDone() {
		logz.Panic("trying to stop following, but the follow child is not active")
	}
	logz.Println(t.Owner.DisplayName(), "stop follow")
	t.EndChild()
	t.status = fightStatusIdle
}

func (t *FightTask) startCombat() {
	if t.status != fightStatusIdle {
		panic("fight status should be idle before trying to start combat")
	}
	logz.Println(t.Owner.DisplayName(), "start combat")
	// nothing to really do here except flip the switch on combat status
	t.status = fightStatusCombat
}

func (t *FightTask) Update() {
	if t.IsDone() {
		return
	}

	if t.targetEntity.IsDead() {
		t.FinishSuccess()
		return
	}

	if t.status == fightStatusIdle {
		// Start is what kicks off the fight task; it transitions out of idle into follow or combat.
		t.Start()
		return
	}

	t.Status = TaskInProg

	switch t.status {
	case fightStatusFollow:
		if !t.HasChild() {
			logz.Panic("supposed to be following, but no follow child is set")
		}
		if t.ChildDone() {
			logz.Panicf("supposed to be following, but the follow child is inactive?")
		}
		t.TaskBase.Update()
		// check if we are close enough to end follow stage. note: this uses actual distance rather
		// than path length, since a freshly-started follow may not have a path yet (it's computed by
		// background assist), and a path-length check would collapse straight into combat.
		if t.Owner.Entity.DistFromEntity(*t.targetEntity) <= config.TileSize*3 {
			t.stopFollowing()
			t.startCombat()
			return
		}
	case fightStatusCombat:
		// the real "meat and potatoes" of this task's logic
		t.handleCombat()
	}
}

// In combat, we mainly want to do the following:
//
// 1. creep towards the enemy until in striking range
// 2. slowly waver back and forth a little bit
// 3. at times, hold up a shield (TODO - shields not implemented yet)
// 4. strike! then return to 1
func (t *FightTask) handleCombat() {
	if t.status != fightStatusCombat {
		panic("status is not set to combat")
	}

	if t.Owner.Entity.Body.IsAttacking() {
		// handle NPC power attacks, and finish the melee attack if needed
		if t.chargeAttackTicks > 0 {
			t.chargeAttackTicks--
			if t.chargeAttackTicks == 0 {
				// now that attack charging is done, trigger the attack finish and set a 1-2 second cooldown
				t.Owner.Entity.FinishMeleeAttack()
				t.nextAttackTime = time.Now().Add(time.Second + time.Duration(rand.Intn(1000))*time.Millisecond)
			}
		}
		return
	}

	dist := t.Owner.Entity.DistFromEntity(*t.targetEntity)
	if dist > config.TileSize*5 {
		t.status = fightStatusIdle
		t.startFollowing()
		return
	}

	t.Owner.Entity.FaceTowardsEntity(*t.targetEntity)

	if t.Owner.Entity.IsUsingShield() {
		// keep using shield until time has expired
		if time.Now().After(t.shieldEndTime) {
			t.Owner.Entity.StopUsingShield()
			// slight pause after dropping shield before attacking
			t.nextAttackTime = time.Now().Add(time.Millisecond * 500)
		}
		return
	}

	if time.Now().Before(t.nextAttackTime) {
		// wait until it's attack time before approaching the enemy
		return
	}

	// only strike when our melee attack would actually land; otherwise keep approaching.
	if !t.Owner.Entity.TargetInMeleeReach(t.targetEntity) {
		t.approachToReach()
		return
	}

	// in striking range: decide to raise shield or attack
	if t.Owner.Entity.IsShieldEquiped() && rand.Float32() < 0.4 {
		t.Owner.Entity.UseShield()
		t.shieldEndTime = time.Now().Add(time.Duration(1+rand.Intn(3)) * time.Second)
		return
	}

	// sanity check: we believe we're in striking range, so the swing must land. if not, the
	// approach logic lined us up wrong (or something else moved us out of range). surface it
	// instead of whiffing silently.
	if !t.Owner.Entity.TargetInMeleeReach(t.targetEntity) {
		logz.PanicCtx(t.Owner.DisplayName(), "handleCombat: believed in striking range, but attack would miss",
			"myPos:", t.Owner.Entity.TilePos(), "targetPos:", t.targetEntity.TilePos(), "target rect:", t.targetEntity.CollisionRect())
	}

	// attack
	t.Owner.Entity.StartMeleeAttack()
	if rand.Float64() < powerAttackChance {
		// power attack: hold the wind-up long enough to build actual charge. charge only starts counting
		// after the "slash-start" wind-up finishes (slashStartWindUpTicks of overhead), and the multiplier
		// ramps from 15 to 90 ticks (250ms to 1500ms). aim for ~30-100 ticks so the real charge lands in
		// that ramp (~18-88 ticks -> mult 1.0-2.0) with some randomness.
		t.chargeAttackTicks = int(30 + rand.Float64()*70)
	} else {
		// regular attack: fire as soon as the wind-up completes (uncharged). 1 tick is enough to trigger
		// FinishMeleeAttack; it will wait for the animation to finish on its own.
		t.chargeAttackTicks = 1
	}
}

// approachToReach moves the NPC until a melee attack would actually land (TargetInMeleeReach).
// Unlike the old tile-based alignment (dx == 0 || dy == 0), this works off collision-rect centers,
// so entities sitting at different pixel offsets within their tiles still end up lined up.
// One axis at a time (largest gap first), capped at one tile per move, and only when idle.
func (t *FightTask) approachToReach() {
	if t.Owner.Entity.IsMoving() {
		return
	}

	e := t.Owner.Entity
	myCX, myCY := e.CollisionRect().GetCenter()
	tcX, tcY := t.targetEntity.CollisionRect().GetCenter()

	// the attack strikes one tile in the facing direction, so the ideal standing spot is the
	// target's center, offset by TileSize along the attack (dominant) axis, with the
	// perpendicular axis centered on the target.
	dx := tcX - myCX
	dy := tcY - myCY
	if dx == 0 && dy == 0 {
		return // on top of the target; shouldn't happen, but avoid the division/panic below
	}

	desiredX, desiredY := tcX, tcY
	if math.Abs(dx) >= math.Abs(dy) {
		desiredX -= math.Copysign(config.TileSize, dx) // stand beside the target
	} else {
		desiredY -= math.Copysign(config.TileSize, dy) // stand above/below the target
	}

	moveX := desiredX - myCX
	moveY := desiredY - myCY
	if math.Abs(moveX) >= math.Abs(moveY) {
		moveY = 0
		moveX = math.Copysign(min(math.Abs(moveX), config.TileSize), moveX)
	} else {
		moveX = 0
		moveY = math.Copysign(min(math.Abs(moveY), config.TileSize), moveY)
	}

	speed := t.Owner.CharacterStateRef.WalkSpeed() / 2
	tickInterval := t.Owner.Entity.Movement.WalkAnimationTickInterval * 2
	moveError := t.Owner.Entity.TryMoveMaxPx(moveX, moveY, speed)
	if moveError.Success {
		t.Owner.Entity.SetAnimation(entity.AnimationOptions{
			AnimationName:         body.AnimWalk,
			AnimationTickInterval: tickInterval,
		})
	} else {
		logz.Println(t.Owner.DisplayName(), "handleCombat: approach move failed:", moveError)
	}
}

// Finish stops the follow child if it's still running (so its path and target cleanup run), then records the result.
// This runs whether the fight ends naturally (target died) or the task is preempted.
func (t *FightTask) Finish(result TaskResult) {
	if t.status == fightStatusFollow {
		t.stopFollowing()
	}
	t.TaskBase.Finish(result)
}

// BackgroundAssist forwards to the active follow child. The child is read from the atomic slot and the
// follow child's own BackgroundAssist only touches atomic mailboxes (see FollowTask), so this is safe to
// run on the background goroutine.
func (t *FightTask) BackgroundAssist() {
	t.TaskBase.BackgroundAssist()
}

func (t *FightTask) SimulationUpdate() {}

func (t FightTask) DisableDefaultSpeechBubbles() bool {
	// we don't want standard greeting speech bubbles during combat
	return true
}
