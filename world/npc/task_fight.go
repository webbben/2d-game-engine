package npc

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/webbben/2d-game-engine/combat"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/data/id"
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

	status fightStatus
	// targetEntity is who we're currently engaging, not the whole enemy list -- the fight session
	// holds that. This is a field rather than something derived per tick because the melee logic needs
	// one specific opponent at a time (distance to it, whether it's in reach, where to stand beside it),
	// and because holding it keeps us from thrashing between equally-close enemies every tick.
	//
	// It can change during a fight: see pickTarget.
	targetEntity *entity.Entity

	// followingTarget is who the active follow child was told to chase, which isn't necessarily
	// targetEntity any more: the target can switch mid-approach, and the child needs restarting when
	// it does. Only meaningful while status is fightStatusFollow.
	followingTarget *entity.Entity

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
	if t.Owner == nil {
		panic("owner entity must not be nil")
	}
	if t.status != fightStatusIdle {
		logz.Println("FightTask", "status:", t.status)
		logz.Panicf("Start: fight task should be idle when (re)starting")
	}

	// A fight task means there's someone to fight. The target we're constructed with is only a
	// starting suggestion -- it may already be dead or gone by the time we get here, and the real
	// candidate list is whatever the session says.
	if !t.pickTarget() {
		logz.Panicln("FightTask", "starting a fight task with no live enemies in the session",
			"npc:", t.Owner.ID(), "suggested target:", t.targetEntity)
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

// currentSession returns the fight this npc is part of, or nil once that fight is over.
func (t *FightTask) currentSession() combat.SessionView {
	return t.Owner.WorldCtx.SessionFor(t.Owner.CharacterStateRef.ID)
}

// targetCandidate is one of our enemies in the current fight, reduced to just what target selection
// actually looks at.
//
// This is deliberately plain data rather than *entity.Entity: target selection is the part of the
// fight logic we most want pinned down by tests, and keeping it free of entities, the world, and
// physics is what makes that possible without standing up a live npc.
type targetCandidate struct {
	ID        id.CharacterStateID
	Dist      float64
	Alive     bool
	InSession bool
	InWorld   bool
}

// targetIsValid reports whether we're willing to fight this candidate at all.
//
// A candidate has to still be alive, still in the fight with us, and still actually present in the
// world -- an enemy can be removed from the map without ever leaving the session, and there's nothing
// left to hit in that case.
//
// Note that offering to surrender is deliberately *not* disqualifying. A surrendering enemy is still a
// legitimate target, because we're free to refuse them, and we shouldn't wander off mid-duel just
// because they gave in.
func targetIsValid(c targetCandidate) bool {
	return c.Alive && c.InSession && c.InWorld
}

// selectTargetID decides which enemy to fight, returning "" when there's nothing left to fight --
// which is how a fight task learns that it should end.
//
// The rules, in priority order:
//
//  1. Stick with the current target while it's still valid. Re-picking every tick would make npcs
//     dither back and forth between equally-close enemies, which reads as broken.
//  2. Otherwise take the nearest valid candidate. Invalid candidates are skipped rather than chosen,
//     so a closer-but-dead enemy never wins just by being closest.
//
// Traits and other game data are expected to layer on top of this later -- a character with a
// particular trait might be expected to chase a fleeing opponent instead of switching away from it.
// That policy belongs here in the engine, but it has to keep building on these two rules rather than
// replacing them, since "don't thrash" and "never swing at a corpse" should survive it.
func selectTargetID(current id.CharacterStateID, candidates []targetCandidate) id.CharacterStateID {
	if current != "" {
		for _, c := range candidates {
			if c.ID == current && targetIsValid(c) {
				return current
			}
		}
	}

	best := id.CharacterStateID("")
	bestDist := math.MaxFloat64
	for _, c := range candidates {
		if !targetIsValid(c) {
			continue
		}
		if c.Dist < bestDist {
			bestDist = c.Dist
			best = c.ID
		}
	}
	return best
}

// pickTarget points this task at an enemy worth attacking, and reports whether there is one.
//
// This is just the adapter that gathers candidates out of the session and resolves the winner back to
// an entity; the policy itself lives in selectTargetID so it can be tested directly.
func (t *FightTask) pickTarget() bool {
	session := t.currentSession()
	if session == nil {
		return false
	}

	ownerEntity := t.Owner.Entity
	ownerID := t.Owner.CharacterStateRef.ID

	var candidates []targetCandidate
	for _, enemyID := range session.EnemiesOf(ownerID) {
		candidate := targetCandidate{
			ID:        enemyID,
			Alive:     true,
			InSession: session.IsCombatant(enemyID),
		}
		if enemy := t.Owner.WorldCtx.EntityFor(enemyID); enemy != nil {
			candidate.InWorld = true
			candidate.Alive = !enemy.IsDead()
			candidate.Dist = ownerEntity.DistFromEntity(*enemy)
		}
		candidates = append(candidates, candidate)
	}

	var currentID id.CharacterStateID
	if t.targetEntity != nil {
		currentID = t.targetEntity.ID()
	}

	chosen := selectTargetID(currentID, candidates)
	if chosen == "" {
		t.targetEntity = nil
		return false
	}
	t.targetEntity = t.Owner.WorldCtx.EntityFor(chosen)
	return t.targetEntity != nil
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
	// remember who the follow child was told to chase. The target can change mid-approach (the one we
	// were closing on died or left), and the child has no way to know -- it would keep walking to the old
	// one while the rest of this task reasons about the new one.
	t.followingTarget = t.targetEntity
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

	// Whoever we were fighting may be dead, fled, surrendered, or simply gone -- and there may be
	// other enemies still standing. Switching targets is how this task handles all of that, and running
	// out of targets is how the fight ends.
	if !t.pickTarget() {
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
		if t.followingTarget != t.targetEntity {
			// We switched targets since the follow child started -- whoever we were closing on died or
			// left the fight. It's still walking the old route, so point it at the new one.
			t.stopFollowing()
			t.startFollowing()
			return
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

func (t *FightTask) handleCombat() {
	if t.status != fightStatusCombat {
		panic("status is not set to combat")
	}

	if t.Owner.shouldFlee() {
		t.Owner.RunTask(defs.TaskDef{
			TaskID:   TaskFlee,
			Priority: Emergency,
			Params:   FleeTaskParams{TargetEntity: t.targetEntity},
		}, t.Owner)
		// we don't care to preserve the fight task as interrupted, since the NPC is no longer willing to fight
		t.Owner.clearInterruptedTask()
		return
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
