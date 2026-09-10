// Package particle contains particle emitter functionality for emitting smoke or other particle effects
package particle

import (
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
)

type Particle struct {
	X, Y   float64
	Vx, Vy float64

	Age      float64
	Lifetime float64

	DriftTimer  float64
	DriftTarget float64

	Scale      float64
	StartScale float64
	EndScale   float64

	Alpha          float64
	StartAlpha     float64
	EndAlpha       float64
	FadeInDuration float64

	Rotation      float64
	RotationSpeed float64

	Sprite *ebiten.Image
}

func (p *Particle) Update(dt float64, turbulence float64, waveVx float64) {
	p.Age += dt

	// slowly choose a new horizontal direction
	p.DriftTimer -= dt

	if p.DriftTimer <= 0 {
		p.DriftTimer = randRange(0.3, 0.8)
		p.DriftTarget = randRange(-1.0, 1.0)
	}

	// randomized horizontal drift
	randomVx := p.DriftTarget * turbulence
	// incorporate wave
	targetVx := waveVx + randomVx

	// smoothly move toward the new drift direction
	p.Vx += (targetVx - p.Vx) * 2.0 * dt

	p.X += p.Vx * dt
	p.Y += p.Vy * dt

	p.Rotation += p.RotationSpeed * dt

	// normalize lifetime 0 to 1
	t := p.Age / p.Lifetime
	if t > 1 {
		t = 1
	}

	// grow and fade
	p.Scale = lerp(p.StartScale, p.EndScale, easeOut(t))

	// particles may fade in first before starting their regular fade out behavior
	fadeInT := 1.0
	if p.FadeInDuration > 0 {
		fadeInT = min(p.Age/p.FadeInDuration, 1.0)
	}
	fadeOutT := t
	fadeOutAlpha := lerp(p.StartAlpha, p.EndAlpha, easeIn(fadeOutT))
	p.Alpha = fadeOutAlpha * easeOut(fadeInT)
}

func (p *Particle) IsDead() bool {
	return p.Age >= p.Lifetime
}

func (p *Particle) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if p.Sprite == nil || p.Alpha <= 0 {
		return
	}

	bounds := p.Sprite.Bounds()

	w := float64(bounds.Dx())
	h := float64(bounds.Dy())

	op := &ebiten.DrawImageOptions{}

	// move the sprite's origin to its center
	op.GeoM.Translate(-w/2, -h/2)

	// scale and rotate around the center
	scale := p.Scale + config.GameScale
	op.GeoM.Scale(scale, scale)
	op.GeoM.Rotate(p.Rotation)

	drawX := (p.X - offsetX) * config.GameScale
	drawY := (p.Y - offsetY) * config.GameScale

	op.GeoM.Translate(drawX, drawY)

	op.ColorScale.ScaleAlpha(float32(p.Alpha))

	screen.DrawImage(p.Sprite, op)
}

func randRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func easeIn(t float64) float64 {
	return t * t
}

func easeOut(t float64) float64 {
	return 1 - (1-t)*(1-t)
}
