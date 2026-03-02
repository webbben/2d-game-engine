// Package screen defines a template for implementing screens in the game.
package screen

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/ui/popup"
)

type ScreenID string

type Screen interface {
	GetID() ScreenID
	// Load is for preparing the screen for use.
	// TODO: should we pass parameters or something? I'm wondering about how we use screens that are used with variable data.
	// for example, a chest inventory screen; it will need to know a specific chest state (not yet implemented) to open and load data for.
	// maybe params could be 'any', and then we can pre-define params for different screen types, like a "chestParams" which includes a chestID.
	// then, the screen that loads can confirm it got the right params type, or panic if not.
	Load(dataman *datamanager.DataManager, eventBus *pubsub.EventBus, om *overlay.OverlayManager, popupMgr *popup.Manager, gameCtx defs.GameContext)
	Update()
	Draw(screen *ebiten.Image)
	// IsDone tells the code that triggered this screen that it can move on now.
	IsDone() bool
}

type ScreenManager struct {
	screens map[ScreenID]Screen
}

func NewScreenManager() *ScreenManager {
	sm := ScreenManager{
		screens: make(map[ScreenID]Screen),
	}

	return &sm
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

// ScreenViewer provides the basic setup needed to run most screens;
// things like overlay managers, popup managers, and the minimal logic needed for updates and drawing control.
type ScreenViewer struct {
	scr      Screen
	om       *overlay.OverlayManager
	popupMgr *popup.Manager
}

func NewScreenViewer(scr Screen, dataman *datamanager.DataManager, eventBus *pubsub.EventBus, gameCtx defs.GameContext) ScreenViewer {
	if scr == nil {
		panic("screen was nil")
	}
	if dataman == nil {
		panic("dataman was nil")
	}
	if eventBus == nil {
		panic("eventbus was nil")
	}
	if gameCtx == nil {
		panic("gameCtx was nil")
	}
	popupMgr := popup.NewPopupManager()
	om := overlay.OverlayManager{}
	sv := ScreenViewer{
		scr:      scr,
		om:       &om,
		popupMgr: &popupMgr,
	}

	sv.scr.Load(dataman, eventBus, &om, &popupMgr, gameCtx)

	return sv
}

func (sv ScreenViewer) IsDone() bool {
	return sv.scr.IsDone()
}

func (sv *ScreenViewer) Update() {
	if sv.scr.IsDone() {
		return
	}

	if sv.popupMgr.IsPopupActive() {
		sv.popupMgr.Update()
		return
	}

	sv.scr.Update()
}

func (sv *ScreenViewer) Draw(screen *ebiten.Image) {
	sv.scr.Draw(screen)

	sv.popupMgr.Draw(screen)

	sv.om.Draw(screen)
}
