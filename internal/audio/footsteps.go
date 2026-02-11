package audio

import (
	"fmt"
)

type StepType string

const (
	STEP_DEFAULT StepType = "default"
	STEP_WOOD    StepType = "wood"
	STEP_STONE   StepType = "stone"
	STEP_GRASS   StepType = "grass"
	STEP_FOREST  StepType = "forest"
	STEP_SAND    StepType = "sand"
	STEP_SNOW    StepType = "snow"
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

// volumeFactor should be 0 to 1
func (sfx *FootstepSFX) Step(stepType StepType, volumeFactor float64) {
	switch stepType {
	case STEP_DEFAULT:
		sfx.stepDefault(volumeFactor)
	case STEP_WOOD:
		sfx.stepWood(volumeFactor)
	case STEP_STONE:
		sfx.stepStone(volumeFactor)
	case STEP_GRASS:
		sfx.stepGrass(volumeFactor)
	case STEP_FOREST:
		sfx.stepForest(volumeFactor)
	case STEP_SAND:
		sfx.stepSand(volumeFactor)
	case STEP_SNOW:
		sfx.stepSnow(volumeFactor)
	default:
		panic("unrecognized step type")
	}
}

func (sfx *FootstepSFX) stepDefault(volumeFactor float64) {
	if sfx.stepDefaultSound == nil {
		// TODO: disabling this panic for now, since I'm planning to redo the audio system into a centralized audio manager/player
		// panic("default step sound not loaded")
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepDefaultSound)
	sfx.stepDefaultSound[sfx.index].PlayVolumeAdjusted(volumeFactor)
}

func (sfx *FootstepSFX) stepWood(volumeFactor float64) {
	if sfx.stepWoodSound == nil {
		sfx.stepDefault(volumeFactor)
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepWoodSound)
	sfx.stepWoodSound[sfx.index].PlayVolumeAdjusted(volumeFactor)
}

func (sfx *FootstepSFX) stepStone(volumeFactor float64) {
	if sfx.stepStoneSound == nil {
		sfx.stepDefault(volumeFactor)
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepStoneSound)
	sfx.stepStoneSound[sfx.index].PlayVolumeAdjusted(volumeFactor)
}

func (sfx *FootstepSFX) stepGrass(volumeFactor float64) {
	if sfx.stepGrassSound == nil {
		sfx.stepDefault(volumeFactor)
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepGrassSound)
	sfx.stepGrassSound[sfx.index].PlayVolumeAdjusted(volumeFactor)
}

func (sfx *FootstepSFX) stepForest(volumeFactor float64) {
	if sfx.stepForestSound == nil {
		sfx.stepDefault(volumeFactor)
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepForestSound)
	sfx.stepForestSound[sfx.index].PlayVolumeAdjusted(volumeFactor)
}

func (sfx *FootstepSFX) stepSand(volumeFactor float64) {
	if sfx.stepSandSound == nil {
		sfx.stepDefault(volumeFactor)
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepSandSound)
	sfx.stepSandSound[sfx.index].PlayVolumeAdjusted(volumeFactor)
}

func (sfx *FootstepSFX) stepSnow(volumeFactor float64) {
	if sfx.stepSnowSound == nil {
		sfx.stepDefault(volumeFactor)
		return
	}
	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(sfx.stepSnowSound)
	sfx.stepSnowSound[sfx.index].PlayVolumeAdjusted(volumeFactor)
}
