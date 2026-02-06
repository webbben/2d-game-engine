package cmd

import (
	"github.com/webbben/2d-game-engine/entity"
	"github.com/webbben/2d-game-engine/internal/audio"
)

func getFootstepSFX() entity.AudioProps {
	// TODO: make centralized "sound manager" that handles loading and playing sounds. instead of each entity loading its own stuff, they will
	// just keep IDs and all sounds are managed from one place
	footstepSFX := entity.AudioProps{
		FootstepSFX: audio.FootstepSFX{
			StepDefaultSrc: []string{
				"sfx/footsteps/footstep_stone_01_A.mp3",
				"sfx/footsteps/footstep_stone_01_B.mp3",
			},
			StepWoodSrc: []string{
				"sfx/footsteps/footstep_wood_01_A.mp3",
				"sfx/footsteps/footstep_wood_01_B.mp3",
			},
			StepGrassSrc: []string{
				"sfx/footsteps/footstep_grass_01_A.mp3",
				"sfx/footsteps/footstep_grass_01_B.mp3",
			},
		},
	}

	return footstepSFX
}
