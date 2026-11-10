package audio

import (
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
)

const globalMaxConcurrentPlaybacks = 100

type PlaybackID uint64

type activePlayback struct {
	player     *audio.Player
	loop       bool
	baseVolume float64

	// for fade in

	startVolumeFactor  float64
	targetVolumeFactor float64
	fadeStartTime      time.Time
	fadeDuration       time.Duration
}

func (pb *activePlayback) setVolume(vol float64) {
	pb.targetVolumeFactor = vol

	if pb.fadeDuration <= 0 || time.Since(pb.fadeStartTime) >= pb.fadeDuration {
		// no fade is active, so immediately snap to target volume
		pb.startVolumeFactor = vol
		pb.player.SetVolume(pb.baseVolume * vol)
	}
}

func (pb *activePlayback) updateVolume() {
	if pb.fadeDuration <= 0 {
		return
	}

	progress := float64(time.Since(pb.fadeStartTime)) / float64(pb.fadeDuration)
	progress = min(progress, 1.0)

	currentVolume := pb.startVolumeFactor + (pb.targetVolumeFactor-pb.startVolumeFactor)*progress

	pb.player.SetVolume(pb.baseVolume * currentVolume)
}

type AudioManager struct {
	Library map[defs.SoundID]*Sound

	playbackMu    sync.Mutex
	activePlayers map[PlaybackID]*activePlayback
	nextID        PlaybackID
}

func NewAudioManager() *AudioManager {
	return &AudioManager{
		Library:       make(map[defs.SoundID]*Sound),
		activePlayers: make(map[PlaybackID]*activePlayback),
	}
}

func (am *AudioManager) LoadSound(id defs.SoundID, relPath string, vol float64) {
	if _, exists := am.Library[id]; exists {
		logz.Panicln("AudioManager", "sound effect already exists:", id)
	}
	sound, err := NewSound(relPath, vol)
	if err != nil {
		logz.Panicln("AudioManager", "failed to load sound:", err)
	}

	am.Library[id] = &sound
}

type PlaySoundParams struct {
	Loop           bool
	FadeInDuration time.Duration
}

// PlaySound spawns a new playback instance for the given sound ID, and returns the playback ID for that instance.
// If PlaybackID is 0, that means no playback instance was created (perhaps because there are too many concurrent instances already).
func (am *AudioManager) PlaySound(id defs.SoundID, vol float64, params PlaySoundParams) PlaybackID {
	if id == "" {
		panic("id was empty")
	}
	if vol < 0 {
		panic("vol was negative")
	}

	am.playbackMu.Lock()
	defer am.playbackMu.Unlock()

	if len(am.activePlayers) >= globalMaxConcurrentPlaybacks {
		logz.Warnln("AudioManager", "PlaySound aborted; already at maximum concurrent playback instances:", globalMaxConcurrentPlaybacks)
		return 0
	}

	if vol > 1 {
		vol = 1
	}

	sound, exists := am.Library[id]
	if !exists {
		logz.Panicln("AudioManager", "tried to play sound that doesn't exist:", id)
	}

	player, err := sound.newPlayer(params.Loop)
	if err != nil {
		logz.Panicln("AudioManager", "an error occurred when trying to play sound:", id, err)
	}

	am.nextID++
	playbackID := am.nextID

	pb := &activePlayback{
		player:             player,
		loop:               params.Loop,
		baseVolume:         sound.baseVolume,
		startVolumeFactor:  0,
		targetVolumeFactor: vol,
		fadeStartTime:      time.Now(),
		fadeDuration:       params.FadeInDuration,
	}

	if pb.fadeDuration <= 0 {
		pb.startVolumeFactor = vol
		pb.player.SetVolume(sound.baseVolume * vol)
	} else {
		player.SetVolume(0)
	}

	am.activePlayers[playbackID] = pb

	player.Play()

	return playbackID
}

func (am *AudioManager) SetVolumeFactor(id PlaybackID, factor float64) bool {
	if factor < 0 || factor > 1 {
		logz.Panic("factor must be between 0 and 1")
	}

	am.playbackMu.Lock()
	defer am.playbackMu.Unlock()

	playback, exists := am.activePlayers[id]
	if !exists {
		return false
	}

	playback.setVolume(factor)

	return true
}

func (am *AudioManager) Stop(id PlaybackID) bool {
	am.playbackMu.Lock()
	defer am.playbackMu.Unlock()

	playback, exists := am.activePlayers[id]
	if !exists {
		return false
	}

	playback.player.Pause()
	_ = playback.player.Close()

	delete(am.activePlayers, id)

	return true
}

// Update handles maintenance of the audiomanager.
// Should only be called by game engine internals - so, games should never be calling this.
func (am *AudioManager) Update() {
	am.playbackMu.Lock()
	defer am.playbackMu.Unlock()

	for id, playback := range am.activePlayers {
		playback.updateVolume()

		if playback.loop {
			continue
		}

		if !playback.player.IsPlaying() {
			_ = playback.player.Close()
			delete(am.activePlayers, id)
		}
	}
}

func (am *AudioManager) Close() {
	// Note: we don't bother locking here because this probably won't run unless the game is shutting down anyway.
	// and besides, Stop gets the lock each time. most likely safe.
	for id := range am.activePlayers {
		am.Stop(id)
	}
}
