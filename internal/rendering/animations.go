package rendering

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/model"
)

// Fader is a utility for fading an image as an animation
type Fader struct {
	TargetScale  float32
	currentScale float32
	fadeFactor   float32
}

// NewFader creates a Fader. fadeFactor should be a decimal value in the range (0, 1], but probably no bigger than 0.1 or else it will be very fast
func NewFader(initialScale float32, fadeFactor float32) Fader {
	if fadeFactor <= 0 {
		fadeFactor = 0.1
	}
	if fadeFactor > 1 {
		fadeFactor = 1
	}
	if initialScale < 0 {
		initialScale = 0
	}
	if initialScale > 1 {
		initialScale = 1
	}
	return Fader{
		currentScale: initialScale,
		TargetScale:  initialScale,
		fadeFactor:   fadeFactor,
	}
}

func (fade *Fader) Update() {
	if fade.currentScale == fade.TargetScale {
		return
	}
	if math.Abs(float64(fade.TargetScale-fade.currentScale)) < 0.01 {
		// snap to target
		fade.currentScale = fade.TargetScale
		return
	}

	diff := (fade.TargetScale - fade.currentScale) * fade.fadeFactor
	fade.currentScale += diff
}

func (fade Fader) SetDrawOps(op *ebiten.DrawImageOptions) {
	op.ColorScale.ScaleAlpha(fade.currentScale)
}

// BounceFader is a fader that goes between two points continuously
type BounceFader struct {
	sourceFader                Fader
	targetScaleA, targetScaleB float32 // the endpoints where this will fade between
}

func NewBounceFader(initialScale, targetScaleA, targetScaleB, fadeFactor float32) BounceFader {
	bounceFader := BounceFader{
		sourceFader:  NewFader(initialScale, fadeFactor),
		targetScaleA: targetScaleA,
		targetScaleB: targetScaleB,
	}
	bounceFader.sourceFader.TargetScale = targetScaleA
	return bounceFader
}

func (bf *BounceFader) Update() {
	bf.sourceFader.Update()
	if bf.sourceFader.currentScale == bf.targetScaleA {
		bf.sourceFader.TargetScale = bf.targetScaleB
	} else if bf.sourceFader.currentScale == bf.targetScaleB {
		bf.sourceFader.TargetScale = bf.targetScaleA
	}
}

func (bf *BounceFader) GetCurrentScale() float32 {
	return bf.sourceFader.currentScale
}

type FadeToPosition struct {
	Target     model.Vec2
	Current    model.Vec2
	FadeFactor float64
}

func NewFadeToPosition(targetX, targetY, curX, curY, fadeFactor float64) FadeToPosition {
	if fadeFactor <= 0 {
		panic("fadeFactor is 0")
	}
	if fadeFactor >= 1 {
		panic("fadeFactor is >= 1; it should be between 0 and 1")
	}
	if targetX == curX && targetY == curY {
		panic("target and current position are the same")
	}
	return FadeToPosition{
		Target:     model.NewVec2(targetX, targetY),
		Current:    model.NewVec2(curX, curY),
		FadeFactor: fadeFactor,
	}
}

func (f *FadeToPosition) moveTowardsTarget(speed float64) {
	f.Current = model.MoveTowards(f.Current, f.Target, speed)
}

func (f *FadeToPosition) Update() {
	if f.Current.Equals(f.Target) {
		return // already at target
	}

	dist := f.Current.Dist(f.Target)
	// continuously move to the position, but move full speed when far away, and ease to a slow down when close to target
	if dist < 0.1 {
		// snap to position
		f.Current = f.Target
	} else {
		f.moveTowardsTarget(dist * f.FadeFactor)
	}
}
