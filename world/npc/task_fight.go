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

	// attacks only land orthogonally (the front rect is one tile in the facing direction),
	// so we must be aligned with the target (same row or column) before striking.
	myPos := t.Owner.Entity.TilePos()
	targetPos := t.targetEntity.TilePos()
	dx := targetPos.X - myPos.X
	dy := targetPos.Y - myPos.Y
	aligned := dx == 0 || dy == 0

	if !aligned || dist > config.TileSize*1.8 {
		// creep towards the enemy, moving along the axis with the greater tile offset
		// this both approaches the enemy and lines us up so our attacks can land
		if !t.Owner.Entity.IsMoving() {
			speed := t.Owner.CharacterStateRef.WalkSpeed() / 2
			tickInterval := t.Owner.Entity.Movement.WalkAnimationTickInterval * 2
			var moveError entity.MoveError
			if math.Abs(float64(dx)) >= math.Abs(float64(dy)) {
				moveError = t.Owner.Entity.TryMoveMaxPx(float64(dx)/math.Abs(float64(dx))*config.TileSize, 0, speed)
			} else {
				moveError = t.Owner.Entity.TryMoveMaxPx(0, float64(dy)/math.Abs(float64(dy))*config.TileSize, speed)
			}
			if moveError.Success {
				t.Owner.Entity.SetAnimation(entity.AnimationOptions{
					AnimationName:         body.AnimWalk,
					AnimationTickInterval: tickInterval,
				})
			} else {
				logz.Println(t.Owner.DisplayName(), "handleCombat: creep forward failed:", moveError)
			}
		}
		return
	}

	// in striking range (and aligned): decide to raise shield or attack
	if t.Owner.Entity.IsShieldEquiped() && rand.Float32() < 0.4 {
		t.Owner.Entity.UseShield()
		t.shieldEndTime = time.Now().Add(time.Duration(1+rand.Intn(3)) * time.Second)
		return
	}

	// attack, then set a 1-2s cooldown
	t.Owner.Entity.StartMeleeAttack()
	t.nextAttackTime = time.Now().Add(time.Second + time.Duration(rand.Intn(1000))*time.Millisecond)
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
