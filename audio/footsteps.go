package audio

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

type StepType string

const (
	StepDefault StepType = "default"
	StepWood    StepType = "wood"
	StepStone   StepType = "stone"
	StepGrass   StepType = "grass"
	StepForest  StepType = "forest"
	StepSand    StepType = "sand"
	StepSnow    StepType = "snow"
)

type FootstepSFX struct {
	Def                defs.FootstepSFXDef
	TicksUntilNextPlay int // current remaining ticks until next play
	TickDelay          int // number of ticks to wait between each play
	index              int

	Volume float64

	AudioMgr *AudioManager
}

type FootstepSFXParams struct {
	Def           defs.FootstepSFXDef
	TickDelay     int
	DefaultVolume float64 // between 0 and 1
}

// NewFootstepSFX returns a FootstepSFX player that handles footstep SFX as an entity walks
func NewFootstepSFX(params FootstepSFXParams, audioMgr *AudioManager) FootstepSFX {
	if params.TickDelay == 0 {
		params.TickDelay = 20
	}
	if params.DefaultVolume <= 0 || params.DefaultVolume > 1 {
		logz.Panicln("NewFootstepSFX", "given default volume is invalid:", params.DefaultVolume)
	}

	sfx := FootstepSFX{
		Def:       params.Def,
		TickDelay: params.TickDelay,
		Volume:    params.DefaultVolume,

		AudioMgr: audioMgr,
	}

	return sfx
}

// Step plays the step sound effect for the given type, and at the given volume. volumeFactor should be 0 to 1
func (sfx *FootstepSFX) Step(stepType StepType, volumeFactor float64) {
	if sfx.AudioMgr == nil {
		panic("audioMgr was nil")
	}
	if volumeFactor < 0 || volumeFactor > 1 {
		logz.Panicln("Step SFX", "volume factor was invalid:", volumeFactor)
	}

	var soundIDs []defs.SoundID
	switch stepType {
	case StepDefault:
		soundIDs = sfx.Def.StepDefaultIDs
	case StepWood:
		soundIDs = sfx.Def.StepWoodIDs
	case StepStone:
		soundIDs = sfx.Def.StepStoneIDs
	case StepGrass:
		soundIDs = sfx.Def.StepGrassIDs
	case StepForest:
		soundIDs = sfx.Def.StepForestIDs
	case StepSand:
		soundIDs = sfx.Def.StepSandIDs
	case StepSnow:
		soundIDs = sfx.Def.StepSnowIDs
	default:
		logz.Panicln("FootstepSFX", "unrecognized step type:", stepType)
	}

	if len(soundIDs) == 0 {
		// whatever we grabbed was undefined; let's try default instead
		soundIDs = sfx.Def.StepDefaultIDs
		if len(soundIDs) == 0 {
			panic("no default step IDs")
		}
	}

	sfx.TicksUntilNextPlay = sfx.TickDelay
	sfx.index = (sfx.index + 1) % len(soundIDs)
	soundID := soundIDs[sfx.index]
	sfx.AudioMgr.PlaySFX(soundID, volumeFactor)
}
