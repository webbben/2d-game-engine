package npc

import (
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/internal/config"
	"github.com/webbben/2d-game-engine/internal/logz"
)

type fightStatus int

const (
	fight_status_idle fightStatus = iota
	fight_status_follow
	fight_status_combat
)

func (fs fightStatus) String() string {
	switch fs {
	case fight_status_idle:
		return "idle (0)"
	case fight_status_follow:
		return "follow (1)"
	case fight_status_combat:
		return "combat (2)"
	default:
		return "unregistered status!"
	}
}

type FightTask struct {
	TaskBase

	FollowTask FollowTask // task for running after the enemy

	status fightStatus

	targetEntity *entity.Entity
}

func (t FightTask) ZzCompileCheck() {
	_ = append([]Task{}, &t)
}

func NewFightTask(targetEnt *entity.Entity) FightTask {
	if targetEnt == nil {
		panic("target is nil")
	}
	return FightTask{
		TaskBase: TaskBase{
			Name:        "fight task",
			Description: "task for NPC to attack an entity",
		},
		status:       fight_status_idle,
		targetEntity: targetEnt,
	}
}

func (n *NPC) SetFightTask(targetEnt *entity.Entity, force bool) error {
	t := NewFightTask(targetEnt)
	return n.SetTask(&t, force)
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
	if t.status != fight_status_idle {
		logz.Panicf("Start: fight task should be idle when (re)starting. (%s)", t.status)
	}

	t.Status = TASK_STATUS_INPROG

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
	if t.status != fight_status_idle {
		panic("fight status should be idle before trying to follow")
	}
	if t.FollowTask.IsActive() {
		// if the follow task appears to already be active, then that's also a problem
		logz.Panicf("follow subtask appears to already be active (%v). It should've ended (or not started yet) before Start was called.", t.FollowTask.GetStatus())
	}
	logz.Println(t.Owner.DisplayName, "start follow")

	t.FollowTask = NewFollowTask(t.targetEntity, 0)
	t.FollowTask.SetOwner(t.Owner)
	t.FollowTask.Start()
	if !t.FollowTask.IsActive() {
		logz.Panicf("follow subtask should've started, but appears inactive (%s)", t.FollowTask.GetStatus())
	}

	t.status = fight_status_follow
}

func (t *FightTask) stopFollowing() {
	if t.status != fight_status_follow {
		logz.Panic("trying to stop following, but not in the following state")
	}
	if !t.FollowTask.IsActive() {
		// if the follow task appears to already be active, then that's also a problem
		logz.Panicf("trying to stop following, but follow task appears to not be active (%v)", t.FollowTask.GetStatus())
	}
	logz.Println(t.Owner.DisplayName, "stop follow")
	t.FollowTask.End()
	t.status = fight_status_idle
}

func (t *FightTask) startCombat() {
	if t.status != fight_status_idle {
		panic("fight status should be idle before trying to start combat")
	}
	logz.Println(t.Owner.DisplayName, "start combat")
	// nothing to really do here except flip the switch on combat status
	t.status = fight_status_combat
}

func (t *FightTask) Update() {
	t.Status = TASK_STATUS_INPROG
	switch t.status {
	case fight_status_idle:
		// why are you idle? you should've already been directed to a new status from wherever the previous action was cancelled
		panic("fight task appears to be stuck in idle")
	case fight_status_follow:
		if !t.FollowTask.IsActive() {
			logz.Panicf("supposed to be following, but follow task is inactive? (followTask status=%s)", t.FollowTask.GetStatus())
		}
		// check if we are close enough to end follow stage
		pathLen := len(t.Owner.Entity.Movement.TargetPath)
		if pathLen < 3 {
			t.stopFollowing()
			t.startCombat()
			return
		}
		t.FollowTask.Update()
	case fight_status_combat:
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
	if t.status != fight_status_combat {
		panic("status is not set to combat")
	}

	if t.Owner.Entity.Body.IsAttacking() {
		return
	}

	dist := t.Owner.Entity.DistFromEntity(*t.targetEntity)
	if dist > config.TileSize*5 {
		t.status = fight_status_idle
		t.startFollowing()
		return
	}

	t.Owner.Entity.FaceTowardsEntity(*t.targetEntity)

	if dist > config.TileSize*2 {
		// creep forward
		if !t.Owner.Entity.Movement.IsMoving {
			speed := t.Owner.Entity.WalkSpeed / 2
			tickInterval := t.Owner.Entity.Movement.WalkAnimationTickInterval * 2
			moveError := t.Owner.Entity.TryMoveTowardsEntity(*t.targetEntity, config.TileSize, speed)
			if moveError.Success {
				t.Owner.Entity.SetAnimation(entity.AnimationOptions{
					AnimationName:         body.AnimWalk,
					AnimationTickInterval: tickInterval,
				})
			} else {
				logz.Println(t.Owner.DisplayName, "handleCombat: creep forward failed:", moveError)
			}
		}
		return
	}

	// once close, attack
	t.Owner.Entity.StartMeleeAttack()
}

func (t *FightTask) End() {
	t.Status = TASK_STATUS_END
}

func (t FightTask) IsComplete() bool {
	return false
}

func (t FightTask) IsFailure() bool {
	return false
}

func (t *FightTask) BackgroundAssist() {

}
