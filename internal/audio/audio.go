package audio

import (
	"bytes"
	"errors"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/webbben/2d-game-engine/internal/config"
)

const sampleRate = 44100

var audioContext = audio.NewContext(sampleRate)

type Sound struct {
	srcPath string
	player  *audio.Player
	volume  float64
	//data   []byte
}

func (s *Sound) SetVolume(v float64) {
	if s.player == nil {
		panic("no sound player found when setting volume")
	}
	s.player.SetVolume(s.volume)
}

func (s *Sound) Play() {
	if s.player == nil {
		panic("no sound player found")
	}
	err := s.player.Rewind()
	if err != nil {
		panic("failed to rewind sound: " + err.Error())
	}
	s.player.Play()
}

// for loading mp3
func NewSound(relAudioPath string, volume float64) (Sound, error) {
	if relAudioPath == "" {
		return Sound{}, errors.New("no relative audio path given")
	}
	srcPath := config.ResolveAudioPath(relAudioPath)

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return Sound{}, err
	}

	s, err := mp3.DecodeF32(bytes.NewReader(data))
	if err != nil {
		return Sound{}, err
	}

	player, err := audioContext.NewPlayerF32(s)
	if err != nil {
		return Sound{}, err
	}

	player.SetVolume(volume)

	sound := Sound{
		srcPath: srcPath,
		player:  player,
	}

	return sound, nil
}
