// Package screen defines a template for implementing screens in the game.
package screen

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/definitions"
	"github.com/webbben/2d-game-engine/internal/logz"
)

type ScreenID string

type Screen interface {
	GetID() ScreenID
	// Load is for preparing the screen for use.
	// The definition manager effectively gives read and write access to all states, so this should suffice.
	Load(defMgr *definitions.DefinitionManager)
	Update()
	Draw(screen *ebiten.Image)
	// IsDone tells the code that triggered this screen that it can move on now.
	IsDone() bool
}

type ScreenManager struct {
	screens map[ScreenID]Screen
}

func (sm *ScreenManager) LoadScreen(s Screen) {
	id := s.GetID()
	if id == "" {
		panic("id was empty")
	}

	if _, exists := sm.screens[id]; exists {
		logz.Panicln("ScreenManager", "tried to load screen, but id already exists:", id)
	}

	sm.screens[id] = s
}

func (sm ScreenManager) GetScreen(id ScreenID) Screen {
	if id == "" {
		panic("id was empty")
	}

	s, exists := sm.screens[id]
	if !exists {
		logz.Panicln("ScreenManager", "tried to get screen, but id doesn't exist:", id)
	}

	return s
}
