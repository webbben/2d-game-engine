// Package audio
package audio

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/webbben/2d-game-engine/config"
)

// Note: If audio is playing faster or higher pitched than expected, most likely the file you're playing
// does not match this sample rate. Consider resampling it to match this sample rate.
const sampleRate = 44100

var audioContext = audio.NewContext(sampleRate)

// A Sound represents an individual audio file that can be played in the game engine.
// There should only be one single Sound for each individual source of audio data.
type Sound struct {
	srcPath    string
	baseVolume float64
	data       []byte
}

// creates a new player, but does not actually start playing the sound
func (s *Sound) newPlayer(loop bool) (*audio.Player, error) {
	var player *audio.Player

	if loop {
		stream := audio.NewInfiniteLoopF32(bytes.NewReader(s.data), int64(len(s.data)))
		var err error
		player, err = audioContext.NewPlayerF32(stream)
		if err != nil {
			return nil, err
		}
	} else {
		player = audioContext.NewPlayerF32FromBytes(s.data)
	}

	return player, nil
}

// NewSound loads and decodes an audio file
func NewSound(relAudioPath string, volume float64) (Sound, error) {
	if relAudioPath == "" {
		return Sound{}, errors.New("no relative audio path given")
	}
	if volume < 0 || volume > 1 {
		return Sound{}, errors.New("volume must be between 0 and 1")
	}

	srcPath := config.ResolveAudioPath(relAudioPath)

	data, err := os.ReadFile(srcPath)
	if err != nil {
		return Sound{}, err
	}

	var stream io.ReadSeeker

	ext := filepath.Ext(srcPath)

	switch ext {
	case ".mp3":
		stream, err = mp3.DecodeF32(bytes.NewReader(data))
	case ".wav":
		stream, err = wav.DecodeF32(bytes.NewReader(data))
	default:
		return Sound{}, fmt.Errorf("unsupported audio format: %s", ext)
	}

	if err != nil {
		return Sound{}, err
	}

	audioData, err := io.ReadAll(stream)
	if err != nil {
		return Sound{}, err
	}

	sound := Sound{
		baseVolume: volume,
		srcPath:    srcPath,
		data:       audioData,
	}

	return sound, nil
}
