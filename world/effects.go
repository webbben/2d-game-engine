package world

import (
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
)

// Effects are used in places like dialogs and quests. They represent an interface of core effects
// that can be applied to the game world, and have centralized logic here.

// AddItemEffect adds an item to the player's inventory.
// TODO: if the player's inventory is full, then the item just appears on the ground next to the player
type AddItemEffect struct {
	ItemID   defs.ItemID
	Quantity *int // if unset, defaults to 1. use this to enable giving larger quantities.
}

func (e AddItemEffect) Apply(ctx defs.WorldEffectContext) {
	quantity := 1
	if e.Quantity != nil {
		quantity = *e.Quantity
		if quantity <= 0 {
			panic("custom quantity was <= 0")
		}
	}
	ctx.AddItem(e.ItemID, quantity)
}

type AddRoleEffect struct {
	RoleID defs.RoleID
}

func (e AddRoleEffect) Apply(ctx defs.WorldEffectContext) {
	ctx.AddRole(e.RoleID)
}

type RemoveRoleEffect struct {
	RoleID defs.RoleID
}

func (e RemoveRoleEffect) Apply(ctx defs.WorldEffectContext) {
	ctx.RemoveRole(e.RoleID)
}

// ScheduleFutureEventEffect is for scheduling a future event. Events can be used to schedule certain world effects too, but only very specific ones
// where it seems likely to want to "undo" a current effect. For example, you can use the AddRole effect, then schedule an event to undo that via an event
// that triggers the RemoveRole effect. I'll try to fill these out with more world effect support as the need arises.
//
// Supported world effects so far:
//   - RemoveRole
//
// Q: Why not add functionality for scheduling future effects? Then, you don't need to worry about adding event infrastructure for each effect type, right?
//   - A: True, but the messy part is saving and loading scheduled effects. Effects are an interface type, so we would need a way to preserve knowledge of the original
//     Data type/struct. Without that, the implementation of Apply is lost. Not that crazy hard to do, but would require a lot of infrastructure to ensure saving
//     and loading these effects is handled right. Also, I anticipate only a handful of world effects will actually ever need to be scheduled for the future, like role stuff.
type ScheduleFutureEventEffect struct {
	Event               defs.Event
	SpecificDate        *clock.GameTime // you can optionally define a specific, hard-coded time. otherwise, this event uses relative times.
	WaitDays, WaitHours int             // for waiting relative times; can wait either a specific number of days, or a specific number of hours (but not both)
	// for use either by itself, or with 'WaitDays' only; If used with WaitDays, it will schedule the event for the number of days in the future, at this specific hour.
	// useful for scheduling events that are a relative number of days, but should fire at a certain time of day, like "tomorrow at 9AM"
	UntilHour *int
}

func (e ScheduleFutureEventEffect) Apply(ctx defs.WorldEffectContext) {
	gt := e.SpecificDate
	if gt == nil {
		if e.WaitDays != 0 && e.WaitHours != 0 {
			panic("cannot use relative wait times of days and hours at the same time")
		}

		currentTime := ctx.GetCurrentGameTime()
		gt = &currentTime
		if e.WaitDays != 0 {
			gt.AddTime(24 * e.WaitDays)
			if e.UntilHour != nil {
				gt.Hour = *e.UntilHour
			}
		} else if e.WaitHours != 0 {
			gt.AddTime(e.WaitHours)
		} else if e.UntilHour != nil {
			if gt.Hour < *e.UntilHour {
				gt.Hour = *e.UntilHour
			} else {
				gt.AddTime(24)
				gt.Hour = *e.UntilHour
			}
		} else {
			logz.Panicln("ScheduleFutureEventEffect", "time parameters for future event were invalid.", e)
		}
	}

	if gt == nil {
		logz.Panicln("ScheduleFutureEventEffect", "gametime was nil. event:", e.Event.Type)
	}

	ctx.BroadcastEvent(defs.Event{
		Type: pubsub.EventScheduleFutureEvent,
		Data: map[string]any{
			"event": e.Event,
			"time":  *gt,
		},
	})
}

// StartCustomLoadScreenEffect causes a load screen (transition + execute load script) to occur in a dialog.
// Using this effect will cause dialog to end.
type StartCustomLoadScreenEffect struct {
	TransitionScreenID              defs.ScreenID
	OpenTransition, CloseTransition defs.Transition
	LoadFunction                    func(ctx defs.GameContext)
}

type StartLoadScreenEffect struct {
	LoadFunction func(ctx defs.GameContext)
}

type AddGoldEffect struct {
	Amount int
}

func (e AddGoldEffect) Apply(ctx defs.WorldEffectContext) {
	ctx.AddGold(e.Amount)
}

type RemoveGoldEffect struct {
	Amount int
}

func (e RemoveGoldEffect) Apply(ctx defs.WorldEffectContext) {
	ctx.RemoveGold(e.Amount)
}

type EventEffect struct {
	Event defs.Event
}

func (e EventEffect) Apply(ctx defs.WorldEffectContext) {
	ctx.BroadcastEvent(e.Event)
}

type AssignTaskEffect struct {
	CharDefID         defs.CharacterDefID
	TaskDef           defs.TaskDef
	PanicIfNoListener bool // if no NPC receives this assign task event, set this to trigger a panic
}

func (e AssignTaskEffect) Apply(ctx defs.WorldEffectContext) {
	ctx.AssignTaskToNPC(e.CharDefID, e.TaskDef, e.PanicIfNoListener)
}

type QueueScenarioEffect struct {
	ScenarioID defs.ScenarioID
}

func (a QueueScenarioEffect) Apply(ctx defs.WorldEffectContext) {
	ctx.QueueScenario(a.ScenarioID)
}

type UnlockEffect struct {
	MapLock defs.MapLock
}

func (a UnlockEffect) Apply(ctx defs.WorldEffectContext) {
	ctx.UnlockMapLock(a.MapLock.MapID, a.MapLock.LockID)
}
