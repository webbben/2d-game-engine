package audio

import (
	"fmt"
)

type FootstepSFX struct {
	StepDefaultSrc   []string
	stepDefaultSound []*Sound
	StepWoodSrc      []string
	stepWoodSound    []*Sound
	StepStoneSrc     []string
	stepStoneSound   []*Sound
	StepGrassSrc     []string
	stepGrassSound   []*Sound
	StepForestSrc    []string
	stepForestSound  []*Sound
	StepSandSrc      []string
	stepSandSound    []*Sound
	StepSnowSrc      []string
	stepSnowSound    []*Sound

	TicksUntilNextPlay int // current remaining ticks until next play
	TickDelay          int // number of ticks to wait between each play
	index              int

	Volume float64
}

// Load all the source audio files and create their players
func (sfx *FootstepSFX) Load() {
	if sfx.TickDelay == 0 {
		sfx.TickDelay = 20
	}
	if sfx.Volume == 0 {
		sfx.Volume = 0.2
	}

	for _, src := range sfx.StepDefaultSrc {
		sound, err := NewSound(src, sfx.Volume)
		if err != nil {
			panic(fmt.Errorf("failed to load sound: %w", err))
		}
		sfx.stepDefaultSound = append(sfx.stepDefaultSound, &sound)
	}

	for _, src := range sfx.StepWoodSrc {
		sound, err := NewSound(src, sfx.Volume)
		if err != nil {
			panic(fmt.Errorf("failed to load sound: %w", err))
		}
		sfx.stepWoodSound = append(sfx.stepWoodSound, &sound)
	}

	for _, src := range sfx.StepStoneSrc {
		sound, err := NewSound(src, sfx.Volume)
		if err != nil {
			panic(fmt.Errorf("failed to load sound: %w", err))
		}
		sfx.stepStoneSound = append(sfx.stepStoneSound, &sound)
	}

	for _, src := range sfx.StepGrassSrc {
		sound, err := NewSound(src, sfx.Volume)
		if err != nil {
			panic(fmt.Errorf("failed to load sound: %w", err))
		}
		sfx.stepGrassSound = append(sfx.stepGrassSound, &sound)
	}

	for _, src := range sfx.StepForestSrc {
		sound, err := NewSound(src, sfx.Volume)
		if err != nil {
			panic(fmt.Errorf("failed to load sound: %w", err))
		}
		sfx.stepForestSound = append(sfx.stepForestSound, &sound)
	}

	for _, src := range sfx.StepSandSrc {
		sound, err := NewSound(src, sfx.Volume)
		if err != nil {
			panic(fmt.Errorf("failed to load sound: %w", err))
		}
		sfx.stepSandSound = append(sfx.stepSandSound, &sound)
	}

	for _, src := range sfx.StepSnowSrc {
		sound, err := NewSound(src, sfx.Volume)
		if err != nil {
			panic(fmt.Errorf("failed to load sound: %w", err))
		}
		sfx.stepSnowSound = append(sfx.stepSnowSound, &sound)
	}
}

func (sfx *FootstepSFX) StepDefault() {
	if sfx.stepDefaultSound == nil {
		panic("default step sound not loaded")
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepDefaultSound)
	sfx.stepDefaultSound[sfx.index].Play()
}

func (sfx *FootstepSFX) StepWood() {
	if sfx.stepWoodSound == nil {
		sfx.StepDefault()
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepWoodSound)
	sfx.stepWoodSound[sfx.index].Play()
}

func (sfx *FootstepSFX) StepStone() {
	if sfx.stepStoneSound == nil {
		sfx.StepDefault()
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepStoneSound)
	sfx.stepStoneSound[sfx.index].Play()
}

func (sfx *FootstepSFX) StepGrass() {
	if sfx.stepGrassSound == nil {
		sfx.StepDefault()
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepGrassSound)
	sfx.stepGrassSound[sfx.index].Play()
}

func (sfx *FootstepSFX) StepForest() {
	if sfx.stepForestSound == nil {
		sfx.StepDefault()
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepForestSound)
	sfx.stepForestSound[sfx.index].Play()
}

func (sfx *FootstepSFX) StepSand() {
	if sfx.stepSandSound == nil {
		sfx.StepDefault()
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepSandSound)
	sfx.stepSandSound[sfx.index].Play()
}

func (sfx *FootstepSFX) StepSnow() {
	if sfx.stepSnowSound == nil {
		sfx.StepDefault()
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepSnowSound)
	sfx.stepSnowSound[sfx.index].Play()
}
