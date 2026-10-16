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
	"github.com/webbben/2d-game-engine/entity/characterState"
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
// TODO: should this be used somewhere?
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
	// Do NOT set directly; use setTarget instead so that relevant flags are reset too.
	targetEntity *entity.Entity

	// followingTarget is who the active follow child was told to chase, which isn't necessarily
	// targetEntity any more: the target can switch mid-approach, and the child needs restarting when
	// it does. Only meaningful while status is fightStatusFollow.
	followingTarget *entity.Entity

	nextAttackTime time.Time
	shieldEndTime  time.Time

	// when attacking, this is set and once it hits 0 the NPC should finish the attack
	chargeAttackTicks int

	// approach step: record closest distance and how long since progress was made
	// this way we can detect if there is something blocking our approach, and we should switch back to a path-finding step.
	closestApproachDist float64
	closestApproachTime time.Time // when the last closestApproachDist was recorded
}

type FightTaskParams struct {
	TargetEntity *entity.Entity
	// TODO: add params to indicate how "serious" the fight is?
	// e.g. can the NPC surrender or not, etc. Probably a future thing once combat is more advanced.
}

var _ Task = (*FightTask)(nil)

// NewFightTask takes the caller's whole TaskDef rather than rebuilding one from loose arguments.
//
// TaskMGMT records a preempted task via GetDef() so it can be replayed later, so anything the def
// carries that this constructor dropped is lost for good -- replay then fails. Params is the field that
// matters: dropping it means a resumed fight task hits a failed type assertion in the registry and
// panics. See issue #182.
func NewFightTask(targetEnt *entity.Entity, owner *NPC, def defs.TaskDef) *FightTask {
	if targetEnt == nil {
		panic("target is nil")
	}
	if owner == nil {
		panic("owner was nil")
	}
	if def.TaskID != TaskFight {
		logz.Panicln("NewFightTask", "task def has wrong ID", string(def.TaskID))
	}
	ft := &FightTask{
		TaskBase: NewTaskBase(def, "Fight", "Fight another entity", owner),
		status:   fightStatusIdle,
	}
	ft.setTarget(targetEnt)

	return ft
}

func init() {
	registerTask(TaskFight, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			fightParams, ok := def.Params.(FightTaskParams)
			if !ok {
				logz.Println("FightTask", def.Params)
				logz.Panicln("FightTask", "tried to run a fight task, but the params could not be converted into FightTaskParams. make sure you are using the right struct")
			}
			// The def is passed through whole, so the Params that were just parsed here survive into
			// TaskBase and therefore into any preemption record. Rebuilding it would lose them.
			return NewFightTask(fightParams.TargetEntity, owner, def)
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
	//
	// No session at all is still a caller error: nothing should be running a fight outside a fight.
	// But a session with nobody worth chasing in it is ordinary -- every enemy broke off and this npc
	// isn't the sort to follow. Standing down is the answer then, not a crash.
	if t.currentSession() == nil {
		logz.Panicln("FightTask", "starting a fight task with no combat session",
			"npc:", t.Owner.ID(), "suggested target:", t.targetEntity)
	}
	if !t.pickTarget() {
		t.standDown()
		return
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

// standDown ends this fight task because there is nothing left for this npc to fight -- either every
// enemy is gone, or they broke off running and this npc won't give chase.
//
// It also leaves the combat session. Standing down is not the same as being defeated: the fight only
// ends for us when we're no longer in it, and leaving is what makes that true. Without this the npc
// would linger in a session no longer participating in, and the fight would stay findable.
//
// Ending the task hands the npc back to the decision loop, which returns them to their hourly schedule.
// That's the same path a fleeing npc already takes, so a guard who declines a chase resumes patrolling
// rather than standing in the street refusing to work.
// shouldYieldCombat answers whether this npc should stop standing their ground and start running.
//
// Unarmed is a reason to go no matter how hard someone is: that is a fact about equipment, not a
// judgment call. Being outmatched is the judgment, and it is the one game data gets to make, so
// bravery is only consulted once there is actually something to fight with. Trying to arm first also
// means an npc carrying a weapon fights rather than runs for no reason.
//
// This lives here rather than in whoever started the fight so that every way into a fight asks the
// same question, and so the answer can change partway through one.
func (t *FightTask) shouldYieldCombat() bool {
	if !t.Owner.ensureWeaponEquipped() {
		return true
	}
	return t.Owner.shouldFlee()
}

// yieldToFlee hands this npc off to the flee task.
//
// The fight task is deliberately not recorded as the interrupted one: they have decided they don't
// want this fight, so resuming it when the flee ends would be wrong.
func (t *FightTask) yieldToFlee() {
	t.Owner.RunTask(defs.TaskDef{
		TaskID:   TaskFlee,
		Priority: Emergency,
		Params:   FleeTaskParams{TargetEntity: t.targetEntity},
	}, t.Owner)
	t.Owner.clearInterruptedTask()
}

func (t *FightTask) standDown() {
	t.setTarget(nil)
	t.Owner.WorldCtx.LeaveCombat(t.Owner.CharacterStateRef.ID)
	t.Finish(TaskResult{Status: ResultSuccess})
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

	// Fleeing reports whether this enemy has broken off and is running. WorthPursuing is the game's
	// answer to whether this npc would chase them anyway.
	//
	// Both default to false, which reads as "standing their ground and therefore always worth
	// fighting" -- so a candidate that never mentions pursuit is a normal one.
	Fleeing       bool
	WorthPursuing bool
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
	if !c.Alive || !c.InSession || !c.InWorld {
		return false
	}
	// A fleeing enemy is only worth chasing if this npc would pursue them. Someone standing their ground
	// is always valid -- there's no decision to make about them.
	return !c.Fleeing || c.WorthPursuing
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

// handles setting flags and other things when a (new) target is set.
// can pass in nil to clear the target.
func (t *FightTask) setTarget(e *entity.Entity) {
	if e == nil {
		logz.Println("setTarget", "target was cleared", t.Owner.WhoAmI())
	}
	t.targetEntity = e
	t.closestApproachDist = math.MaxFloat64
	t.closestApproachTime = time.Now()
}

// pickTarget points this task at an enemy worth attacking, and reports whether there is one.
//
// This is just the adapter that gathers candidates out of the session and resolves the winner back to
// an entity; the policy itself lives in selectTargetID so it can be tested directly.
//
// Runs every tick, so it checks the cheap thing first. Most ticks the current target is still perfectly
// good, and rebuilding the whole candidate list -- an entity lookup and a distance for every enemy in
// the fight -- only to conclude "keep what I have" is work thrown away. The fast path validates the
// current target alone and returns.
//
// That path also bounds how often the pursuit hook is consulted. A current target who isn't fleeing has
// no pursuit question, so no hook call happens; the hook fires only when someone actually starts
// running, which is the moment a target switch is genuinely on the table.
func (t *FightTask) pickTarget() bool {
	session := t.currentSession()
	if session == nil {
		return false
	}
	sessionID := session.SessionID()

	// Fast path: is what we're already fighting still someone we'll keep fighting?
	if t.targetEntity != nil {
		if candidate := t.buildCandidate(session, t.targetEntity.ID()); targetIsValid(t.withPursuit(sessionID, candidate)) {
			// still targeting the same entity, so no need to change anything
			return true
		}
	}

	ownerID := t.Owner.CharacterStateRef.ID
	var candidates []targetCandidate
	for _, enemyID := range session.EnemiesOf(ownerID) {
		candidate := t.buildCandidate(session, enemyID)
		candidates = append(candidates, t.withPursuit(sessionID, candidate))
	}

	var currentID id.CharacterStateID
	if t.targetEntity != nil {
		currentID = t.targetEntity.ID()
	}

	chosen := selectTargetID(currentID, candidates)
	if chosen == "" {
		t.setTarget(nil)
		return false
	}
	if t.targetEntity != nil && chosen == t.targetEntity.ID() {
		// we chose the same one as the existing target, so nothing to do
		logz.PanicCtx("pickTarget", "the existing target was chosen, even though it should be invalid according to earlier logic in this function.",
			t.targetEntity.ID())
	}
	targetEntity := t.Owner.WorldCtx.EntityFor(chosen)
	if targetEntity == nil {
		logz.Panic("target entity was unexpectedly nil")
	}
	t.setTarget(targetEntity)
	return true
}

// buildCandidate gathers everything target selection needs to know about one enemy. Pure description of
// the world -- no policy.
func (t *FightTask) buildCandidate(session combat.SessionView, enemyID id.CharacterStateID) targetCandidate {
	candidate := targetCandidate{
		ID:        enemyID,
		Alive:     true,
		InSession: session.IsCombatant(enemyID),
	}
	if enemy := t.Owner.WorldCtx.EntityFor(enemyID); enemy != nil {
		candidate.InWorld = true
		candidate.Alive = !enemy.IsDead()
		candidate.Dist = t.Owner.Entity.DistFromEntity(*enemy)
	}
	return candidate
}

// withPursuit fills in the pursuit half of a candidate: whether this enemy is running, and whether this
// npc would chase them if they are.
//
// A standing enemy is left as-is, because the answer for them is always "worth pursuing" and asking the
// question would mean building two profiles to get a value nobody needs.
//
// Dead candidates return early and unmarked. This runs before liveness is known to the caller -- on both
// pickTarget paths, since here it is either an argument to targetIsValid or applied straight to a list of
// session members -- so this is the one place that can stop a corpse being profiled. Alive is set false as
// well as short-circuiting, so targetIsValid rejects them on the merits rather than by luck of what the
// entity lookup happened to find.
func (t *FightTask) withPursuit(sessionID combat.CombatSessionID, candidate targetCandidate) targetCandidate {
	if characterstate.IsDeadOrDying(candidate.ID, t.Owner.dataman) {
		candidate.Alive = false
		return candidate
	}

	if !t.Owner.WorldCtx.CombatProfile(sessionID, candidate.ID).Fleeing {
		return candidate
	}
	candidate.Fleeing = true
	candidate.WorthPursuing = t.Owner.shouldPursue(sessionID, candidate.ID, t.Owner.countStandingEnemies(sessionID, candidate.ID))
	return candidate
}

func (t *FightTask) dropShield() {
	if t.Owner.Entity.IsUsingShield() {
		t.Owner.Entity.StopUsingShield()
		t.shieldEndTime = time.Now()
	}
}

func (t *FightTask) startFollowing() {
	if t.status != fightStatusIdle {
		panic("fight status should be idle before trying to follow")
	}
	if !t.ChildDone() {
		logz.Panicf("follow subtask appears to already be active. It should've ended (or not started yet) before Start was called.")
	}
	logz.Println(t.Owner.ID(), "start follow")

	// not allowed to shield while moving
	t.dropShield()

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
	logz.Println(t.Owner.ID(), "stop follow")
	t.EndChild()
	t.status = fightStatusIdle
}

func (t *FightTask) startCombat() {
	if t.status != fightStatusIdle {
		panic("fight status should be idle before trying to start combat")
	}
	logz.Println(t.Owner.ID(), "start combat")
	// nothing to really do here except flip the switch on combat status
	t.status = fightStatusCombat
	// reset approach flags
	t.closestApproachDist = math.MaxFloat64
	t.closestApproachTime = time.Now()
}

func (t *FightTask) Update() {
	if t.IsDone() {
		return
	}

	// Whoever we were fighting may be dead, fled, surrendered, or simply gone -- and there may be
	// other enemies still standing. Switching targets is how this task handles all of that, and running
	// out of targets is how the fight ends.
	if !t.pickTarget() {
		t.standDown()
		return
	}

	// Re-check every tick whether this npc still wants to be here at all. This sits above the status
	// switch, and above Start, so it applies from the very first tick -- which is what lets a coward
	// break off before landing a swing -- and while following as well as while fighting.
	if t.shouldYieldCombat() {
		t.yieldToFlee()
		return
	}

	if t.status == fightStatusIdle {
		// Start is what kicks off the fight task; it transitions out of idle into follow or combat.
		t.Start()
		return
	}

	t.Status = TaskInProg

	// Catch bad state: an npc blocking with its shield must not be walking. Scoped to path following,
	// because a blocked hit legitimately recoils the blocker while the shield animation is still up
	// (ReceiveAttack -> TryBumpBack). That recoil moves toward TargetX/TargetY directly and never sets
	// a path, so HasPath is what separates the recoil from task-driven locomotion, which is the
	// actual bug: an npc that stopped following used to keep walking the follow path.
	if t.Owner.Entity.IsUsingShield() && t.Owner.Entity.IsMoving() && t.Owner.Entity.HasPath() {
		logz.PanicCtx("FightTask", "entity is walking a path while blocking with shield", t.Owner.WhoAmI())
	}

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
		pathLength := len(t.Owner.Entity.Movement.TargetPath)
		if pathLength == 0 {
			// TODO: (2026-10-15) do we want to do anything here? as of now, it seems to work fine like this.
		} else if pathLength <= 3 {
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

	// Note: whether this npc wants to be fighting at all is decided in Update, before we get here --
	// see shouldYieldCombat. An npc that has decided to run has already been handed to the flee task.

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
		if t.Owner.Entity.UseShield() {
			t.shieldEndTime = time.Now().Add(time.Duration(1+rand.Intn(3)) * time.Second)
			return
		}
		// if shield failed to be set, then fall through to striking logic
	}

	// sanity check: we believe we're in striking range, so the swing must land. if not, the
	// approach logic lined us up wrong (or something else moved us out of range). surface it
	// instead of whiffing silently.
	if !t.Owner.Entity.TargetInMeleeReach(t.targetEntity) {
		logz.PanicCtx(t.Owner.ID(), "handleCombat: believed in striking range, but attack would miss",
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
	myC := e.CollisionRect().GetCenter()
	tc := t.targetEntity.CollisionRect().GetCenter()

	// record when we last hit the closest distance, for detecting "dead end" positions where
	// no real progress is possible
	dist := myC.Dist(tc)
	if dist < t.closestApproachDist {
		t.closestApproachDist = dist
		t.closestApproachTime = time.Now()
	}

	if time.Since(t.closestApproachTime) > time.Second*3 {
		// switch back to pathfinding, since we haven't been able to make any real progress
		logz.Println("approachToReach", "switching back to pathfinding, since current area seems to be a dead end.", t.Owner.WhoAmI())
		t.status = fightStatusIdle
		t.startFollowing()
		return
	}

	// the attack strikes one tile in the facing direction, so the ideal standing spot is the
	// target's center, offset by TileSize along the attack (dominant) axis, with the
	// perpendicular axis centered on the target.
	dx := tc.X - myC.X
	dy := tc.Y - myC.Y
	if dx == 0 && dy == 0 {
		return // on top of the target; shouldn't happen, but avoid the division/panic below
	}

	desiredX, desiredY := tc.X, tc.Y
	if math.Abs(dx) >= math.Abs(dy) {
		desiredX -= math.Copysign(config.TileSize, dx) // stand beside the target
	} else {
		desiredY -= math.Copysign(config.TileSize, dy) // stand above/below the target
	}

	moveX := desiredX - myC.X
	moveY := desiredY - myC.Y
	if math.Abs(moveX) >= math.Abs(moveY) {
		moveY = 0
		moveX = math.Copysign(min(math.Abs(moveX), config.TileSize), moveX)
	} else {
		moveX = 0
		moveY = math.Copysign(min(math.Abs(moveY), config.TileSize), moveY)
	}

	// try all three directions, and if they all fail then we know there is something truly blocking us from
	// reaching the target
	if err := t.approachStep(moveX, moveY); err.Success {
		return
	}
	if err := t.approachStep(-moveY, moveX); err.Success {
		return
	}
	if err := t.approachStep(moveY, -moveX); err.Success {
		return
	}
	logz.Println("approachToReach", "approach blocked in all directions.", t.Owner.WhoAmI())
	// TODO: should we immediately go back to the follow step if this is reached?
}

func (t *FightTask) approachStep(dx, dy float64) entity.MoveError {
	if dx == 0 && dy == 0 {
		logz.PanicCtx("approachStep", "dx and dy were 0", t.Owner.WhoAmI())
	}
	speed := t.Owner.CharacterStateRef.WalkSpeed() / 2
	tickInterval := t.Owner.Entity.Movement.WalkAnimationTickInterval * 2
	moveError := t.Owner.Entity.TryMoveMaxPx(dx, dy, speed)
	if moveError.Success {
		animRes := t.Owner.Entity.SetAnimation(entity.AnimationOptions{
			AnimationName:         body.AnimWalk,
			AnimationTickInterval: tickInterval,
		})
		if !animRes.Success && !animRes.AlreadySet {
			// just want to know if this ever happens
			logz.Panicln("approachStep", "failed to set walk animation.", animRes.String())
		}
	}
	return moveError
}

// Finish stops the follow child if it's still running (so its path and target cleanup run), then records the result.
// This runs whether the fight ends naturally (target died) or the task is preempted.
func (t *FightTask) Finish(result TaskResult) {
	if t.status == fightStatusFollow {
		t.stopFollowing()
	}
	t.dropShield()
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
