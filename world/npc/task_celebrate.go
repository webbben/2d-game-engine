package npc

import (
	"time"

	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/entity/body"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/utils"
)

// celebrateAnimationTickInterval is how many ticks each frame of the victory animation is held for.
//
// The celebrate animation is seven frames of an arm coming up and waving, so this sets the pace of the
// wave rather than the length of it -- the task's own countdown ends the celebration. Roughly a second
// and a bit per loop at 60 ticks/sec, which reads as a person waving rather than a flickering sprite.
const celebrateAnimationTickInterval = 10

// CelebrateTask is the victory lap: an NPC who won a big enough fight throws their arms up for a few
// seconds before getting back on with their day.
//
// It isn't started by the fight. The world queues a celebration when a fight ends (see
// NPC.QueueVictoryCelebration) and the task loop runs this once the NPC is free, because at the moment
// the fight ends they're usually still finishing the swing that ended it.
//
// Nothing gets to interrupt the animation except the NPC being attacked or otherwise given something
// more important to do, which preempts this at Emergency priority. Finish puts the animation back
// whatever happens, so a celebration cut short never leaves the NPC stuck in the pose.
type CelebrateTask struct {
	TaskBase

	// endTime is when the celebration gives up and the NPC goes back to their schedule.
	endTime time.Time
}

func NewCelebrateTask(n *NPC, def defs.TaskDef) *CelebrateTask {
	utils.PanicAssert(def.TaskID == TaskCelebrate, "task def has wrong ID")
	return &CelebrateTask{
		TaskBase: NewTaskBase(def, "Celebrate", "Victory celebration after winning a big fight", n),
	}
}

func init() {
	registerTask(TaskCelebrate, taskMeta{
		build: func(def defs.TaskDef, owner *NPC) Task {
			return NewCelebrateTask(owner, def)
		},
	})
}

// Start raises the arms and, if the NPC has anything to say about it, says it.
//
// Movement is stopped first: the animation loops until the task ends, and an NPC still walking a path
// would slide across the map in the victory pose, since nothing else is going to interrupt the walk
// animation.
func (t *CelebrateTask) Start() {
	t.Owner.Entity.StopMovement()

	animRes := t.Owner.Entity.SetAnimation(entity.AnimationOptions{
		AnimationName:         body.AnimCelebrate,
		AnimationTickInterval: celebrateAnimationTickInterval,
		SetAnimationOps: body.SetAnimationOps{
			// Force, because the body doesn't have to be idle. Being kept out of the way of an
			// attack in flight is the task loop's job, not this call's; what Force covers is any
			// other non-idle state -- a walk the finished fight task didn't get to cancel --
			// which would make this fail and leave the npc celebrating in their walking frames.
			Force: true,
		},
	})
	if !animRes.Success && !animRes.AlreadySet {
		// Not worth taking the fight down over. The celebration is a few seconds of flavour, and the
		// body will be back on idle when the countdown ends either way.
		logz.Println(t.Owner.WhoAmI(), "couldn't start the celebrate animation:", animRes)
	}

	t.Owner.showCelebrationSpeechBubble()
	t.endTime = time.Now().Add(config.CelebrationDuration)
	t.TaskBase.Start()
}

func (t *CelebrateTask) Update() {
	if t.IsDone() {
		return
	}
	t.Status = TaskInProg
	if !time.Now().Before(t.endTime) {
		t.Finish(TaskResult{Status: ResultSuccess})
	}
}

// Finish always puts the victory pose away.
//
// This runs on preemption as well as on the countdown, and that first case is the reason it exists.
// IsAttacking is no help here: it doesn't consider the victory pose attacking, and it shouldn't. But a
// preempting task's own animation only replaces this one cleanly if the body is idle, and a task that
// merely walks away sets AnimWalk without Force -- which fails against a non-idle body. An NPC pulled
// into a new fight would otherwise be left sliding around the map with its arms up.
//
// Scoped to the celebrate animation so an NPC preempted mid-swing has its attack left alone; the fight
// task throwing its own animation is what should win there.
func (t *CelebrateTask) Finish(result TaskResult) {
	if t.Owner.Entity.Body.GetCurrentAnimation() == body.AnimCelebrate {
		t.Owner.Entity.Body.StopAnimation()
	}
	t.TaskBase.Finish(result)
}

// DisableDefaultSpeechBubbles stops greetings and the like from appearing over the victory animation.
// The celebration's own bubble is shown explicitly in Start, so nothing is lost by muting the rest.
func (t CelebrateTask) DisableDefaultSpeechBubbles() bool { return true }

// SetupActiveState: a victory is decided in the fight, so there is nothing to set up if the player comes
// back to this map afterwards.
func (t *CelebrateTask) SetupActiveState() {}

var _ Task = (*CelebrateTask)(nil)
