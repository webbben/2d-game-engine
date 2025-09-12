package rendering

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Utility for fading an image as an animation
type Fade struct {
	TargetScale  float32
	currentScale float32
	fadeFactor   float32
}

// fadeFactor should be a decimal value in the range (0, 1], but probably no bigger than 0.1 or else it will be very fast
func NewFader(initialScale float32, fadeFactor float32) Fade {
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
	return Fade{
		currentScale: initialScale,
		TargetScale:  initialScale,
		fadeFactor:   fadeFactor,
	}
}

func (fade *Fade) Update() {
	if fade.currentScale == fade.TargetScale {
		return
	}
	if math.Abs(float64(fade.TargetScale-fade.currentScale)) < 0.001 {
		// snap to target
		fade.currentScale = fade.TargetScale
		return
	}

	diff := (fade.TargetScale - fade.currentScale) * fade.fadeFactor
	fade.currentScale += diff
}

func (fade Fade) SetDrawOps(op *ebiten.DrawImageOptions) {
	op.ColorScale.ScaleAlpha(fade.currentScale)
}
