package audio

import (
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/internal/logz"
)

type AudioManager struct {
	SFXLibrary map[defs.SoundID]*Sound
	BGMLibrary map[defs.SoundID]*Sound
}

func NewAudioManager() *AudioManager {
	return &AudioManager{
		SFXLibrary: make(map[defs.SoundID]*Sound),
		BGMLibrary: make(map[defs.SoundID]*Sound),
	}
}

func (am *AudioManager) LoadSFX(id defs.SoundID, relPath string, vol float64) {
	if _, exists := am.SFXLibrary[id]; exists {
		logz.Panicln("AudioManager", "sound effect already exists:", id)
	}
	sound, err := NewSound(relPath, vol)
	if err != nil {
		logz.Panicln("AudioManager", "failed to load sound:", err)
	}

	am.SFXLibrary[id] = &sound
}

func (am *AudioManager) LoadBGM(id defs.SoundID, relPath string, vol float64) {
	if _, exists := am.BGMLibrary[id]; exists {
		logz.Panicln("AudioManager", "bgm already exists:", id)
	}
	sound, err := NewSound(relPath, vol)
	if err != nil {
		logz.Panicln("AudioManager", "failed to load sound:", err)
	}

	am.BGMLibrary[id] = &sound
}

func (am AudioManager) PlaySFX(id defs.SoundID, vol float64) {
	if id == "" {
		panic("id was empty")
	}

	sound, exists := am.SFXLibrary[id]
	if !exists {
		logz.Panicln("AudioManager", "tried to play sound that doesn't exist:", id)
	}

	sound.PlayVolumeAdjusted(vol)
}
