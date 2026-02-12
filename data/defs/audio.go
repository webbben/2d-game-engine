package defs

type (
	SoundID          string
	FootstepSFXDefID string
)

// FootstepSFXDef defines all the sound IDs for SFX that play while walking or running
type FootstepSFXDef struct {
	ID             FootstepSFXDefID
	StepDefaultIDs []SoundID
	StepWoodIDs    []SoundID
	StepStoneIDs   []SoundID
	StepGrassIDs   []SoundID
	StepForestIDs  []SoundID
	StepSandIDs    []SoundID
	StepSnowIDs    []SoundID
}
