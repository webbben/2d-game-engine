package particle

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

// the rest of the game doesn't currently use "real" tick duration to track actual time elapse,
// so we will just approximate it here since we've already baked "td" into all the update functions.
// In the future, we could consider moving more generally towards a tick duration based update system
// where update logic is based on how much time elapsed since last tick - helps if there were moments of lag
// or something, to make sure things are still progressing as expected.
const tickDuration = 1.0 / 60.0

type Emitter struct {
	// Particles that have been emitted
	particles []*Particle

	x, y float64

	wavePhase float64

	params defs.EmitterParams

	// Available particle sprites
	sprites []*ebiten.Image

	accumulator float64
}

func NewEmitter(params defs.EmitterParams) *Emitter {
	sprites := make([]*ebiten.Image, 0)
	for _, c := range params.PixelColors {
		pixel := ebiten.NewImage(1, 1)
		pixel.Fill(c)
		sprites = append(sprites, pixel)
	}

	if len(sprites) == 0 {
		logz.Panic("emitter has no sprites!")
	}

	return &Emitter{
		params:    params,
		particles: make([]*Particle, 0),
		sprites:   sprites,
		wavePhase: rand.Float64() * 2 * math.Pi, // randomize initial phase so no two emitters have same exact wave shape
	}
}

func (e *Emitter) Update(posX, posY float64) {
	e.x = posX
	e.y = posY

	dt := tickDuration
	e.emit(dt)

	// advance the shared wave
	if e.params.Wave.Enabled {
		e.wavePhase += e.params.Wave.Speed * dt
	}

	// iterate backwards so dead particles can be removed without affecting indices we haven't processed yet
	for i := len(e.particles) - 1; i >= 0; i-- {
		p := e.particles[i]

		waveVx := e.getWaveVelocity(p)

		p.Update(dt, e.params.Turbulence, waveVx)
		if p.IsDead() {
			// replace with last particle since ordering doesn't matter (and it's more efficient than cutting out middle element)
			last := len(e.particles) - 1
			e.particles[i] = e.particles[last]
			e.particles = e.particles[:last]
		}
	}
}

func (e *Emitter) emit(dt float64) {
	e.accumulator += e.params.SpawnRate * dt

	for e.accumulator >= 1.0 {
		e.accumulator -= 1.0
		e.spawnParticle()
	}
}

func (e *Emitter) spawnParticle() {
	if len(e.sprites) == 0 {
		logz.Panicln("Emitter", "can't spawn particle; no sprites exist!")
	}
	sprite := e.sprites[rand.Intn(len(e.sprites))]

	p := Particle{
		Sprite: sprite,

		X: e.x,
		Y: e.y,

		Vx: randRange(e.params.VelXMin, e.params.VelXMax),
		Vy: randRange(e.params.VelYMin, e.params.VelYMax),

		Lifetime: randRange(e.params.LifetimeMin, e.params.LifetimeMax),

		StartScale: e.params.StartScale,
		EndScale:   e.params.EndScale,
		StartAlpha: e.params.StartAlpha,
		EndAlpha:   e.params.EndAlpha,

		FadeInDuration: randRange(e.params.FadeInDurationMin, e.params.FadeInDurationMax),

		Scale: e.params.StartScale,
		Alpha: 0, // start at 0 since we fade in too

		Rotation:      randRange(0, math.Pi*2),
		RotationSpeed: randRange(e.params.RotationSpeedMin, e.params.RotationSpeedMax),
		DriftTarget:   randRange(-1, 1),
		DriftTimer:    randRange(0.3, 0.8),
	}

	e.particles = append(e.particles, &p)
}

func (e *Emitter) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	for _, p := range e.particles {
		p.Draw(screen, offsetX, offsetY)
	}
}

func (e *Emitter) getWaveVelocity(p *Particle) float64 {
	wave := e.params.Wave

	if !wave.Enabled {
		return 0
	}

	// convert particle position into a coordinate relative to the emitter.
	// this means the wave is based on how high the particle is above the source.
	relativeY := p.Y - e.y

	// calculate the wave at this height and point in time
	phase := e.wavePhase + relativeY*wave.Frequency

	return math.Sin(phase) * wave.Amplitude * wave.Strength
}
