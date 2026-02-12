package cmd

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/internal/audio"
)

const (
	SFXFootstepStone01A defs.SoundID = "footstep_stone_01_A"
	SFXFootstepStone01B defs.SoundID = "footstep_stone_01_B"
	SFXFootstepWood01A  defs.SoundID = "footstep_wood_01_A"
	SFXFootstepWood01B  defs.SoundID = "footstep_wood_01_B"
	SFXFootstepGrass01A defs.SoundID = "footstep_grass_01_A"
	SFXFootstepGrass01B defs.SoundID = "footstep_grass_01_B"

	SFXMetalGateOpen01 defs.SoundID = "metal_gate_open_01"

	DefaultFootstepSFXID defs.FootstepSFXDefID = "default_footstep_sfx"
)

func LoadAllSoundEffects(audioMgr *audio.AudioManager) {
	fileMapping := map[defs.SoundID]string{
		SFXFootstepStone01A: "sfx/footsteps/footstep_stone_01_A.mp3",
		SFXFootstepStone01B: "sfx/footsteps/footstep_stone_01_B.mp3",
		SFXFootstepWood01A:  "sfx/footsteps/footstep_wood_01_A.mp3",
		SFXFootstepWood01B:  "sfx/footsteps/footstep_wood_01_B.mp3",
		SFXFootstepGrass01A: "sfx/footsteps/footstep_grass_01_A.mp3",
		SFXFootstepGrass01B: "sfx/footsteps/footstep_grass_01_B.mp3",

		// Doors, gates, etc
		SFXMetalGateOpen01: "sfx/door/gdc_mix_prison_gate_open_01.mp3",
	}

	for soundID, path := range fileMapping {
		audioMgr.LoadSFX(soundID, path, 1)
	}
}

func GetFootstepSFXDefs() []defs.FootstepSFXDef {
	defs := []defs.FootstepSFXDef{
		{
			ID:             DefaultFootstepSFXID,
			StepDefaultIDs: []defs.SoundID{SFXFootstepStone01A, SFXFootstepStone01B},
			StepWoodIDs:    []defs.SoundID{SFXFootstepWood01A, SFXFootstepWood01B},
			StepGrassIDs:   []defs.SoundID{SFXFootstepGrass01A, SFXFootstepGrass01B},
		},
	}

	return defs
}
