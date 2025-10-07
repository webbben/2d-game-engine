package rendering

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Utility for fading an image as an animation
type Fader struct {
	TargetScale  float32
	currentScale float32
	fadeFactor   float32
}

// fadeFactor should be a decimal value in the range (0, 1], but probably no bigger than 0.1 or else it will be very fast
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

// A fader that goes between two points continuously
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
