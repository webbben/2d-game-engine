package dialogv2

import (
	"github.com/webbben/2d-game-engine/clock"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
)

type EventEffect struct {
	Event defs.Event
}

func (e EventEffect) Apply(ctx defs.EffectContext) {
	ctx.BroadcastEvent(e.Event)
}

type AddGoldEffect struct {
	Amount int
}

func (e AddGoldEffect) Apply(ctx defs.EffectContext) {
	ctx.AddGold(e.Amount)
}

type RemoveGoldEffect struct {
	Amount int
}

func (e RemoveGoldEffect) Apply(ctx defs.EffectContext) {
	ctx.RemoveGold(e.Amount)
}

// SetDialogMemoryEffect sets a specific dialog memory key.
// This should be used sparingly, only when other effects or mechanisms won't accomplish what you need to do.
type SetDialogMemoryEffect struct {
	MemoryKey string
}

func (e SetDialogMemoryEffect) Apply(ctx defs.EffectContext) {
	ctx.RecordMiscDialogMemory(e.MemoryKey)
}

// StartCustomLoadScreenEffect causes a load screen (transition + execute load script) to occur in a dialog.
// Using this effect will cause dialog to end.
type StartCustomLoadScreenEffect struct {
	TransitionScreenID              defs.ScreenID
	OpenTransition, CloseTransition defs.Transition
	LoadFunction                    func(ctx defs.GameContext)
}

func (e StartCustomLoadScreenEffect) Apply(ctx defs.EffectContext) {
	ctx.StartCustomLoadScreen(e.TransitionScreenID, e.OpenTransition, e.CloseTransition, e.LoadFunction)
}

type StartLoadScreenEffect struct {
	LoadFunction func(ctx defs.GameContext)
}

func (e StartLoadScreenEffect) Apply(ctx defs.EffectContext) {
	ctx.StartLoadScreen(e.LoadFunction)
}

type ScheduleFutureEventEffect struct {
	Event               defs.Event
	SpecificDate        *clock.GameTime // you can optionally define a specific, hard-coded time. otherwise, this event uses relative times.
	WaitDays, WaitHours int             // for waiting relative times; can wait either a specific number of days, or a specific number of hours (but not both)
	// for use either by itself, or with 'WaitDays' only; If used with WaitDays, it will schedule the event for the number of days in the future, at this specific hour.
	// useful for scheduling events that are a relative number of days, but should fire at a certain time of day, like "tomorrow at 9AM"
	UntilHour *int
}

func (e ScheduleFutureEventEffect) Apply(ctx defs.EffectContext) {
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

// AddItemEffect adds an item to the player's inventory.
// TODO: if the player's inventory is full, then the item just appears on the ground next to the player
type AddItemEffect struct {
	ItemID   defs.ItemID
	Quantity *int // if unset, defaults to 1. use this to enable giving larger quantities.
}

func (e AddItemEffect) Apply(ctx defs.EffectContext) {
	quantity := 1
	if e.Quantity != nil {
		quantity = *e.Quantity
		if quantity <= 0 {
			panic("custom quantity was <= 0")
		}
	}
	ctx.AddItem(e.ItemID, quantity)
}
