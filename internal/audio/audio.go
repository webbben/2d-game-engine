package audio

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

var audioContext = audio.NewContext(sampleRate)

func LoadSound(data []byte) (*audio.Player, error) {
	src, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode sound data: %w", err)
	}
	p, err := audioContext.NewPlayer(src)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio player: %w", err)
	}
	return p, nil
}
