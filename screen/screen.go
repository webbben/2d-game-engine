// Package screen defines a template for implementing screens in the game.
package screen

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/audio"
	"github.com/webbben/2d-game-engine/data/datamanager"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/pubsub"
	"github.com/webbben/2d-game-engine/ui/overlay"
	"github.com/webbben/2d-game-engine/ui/popup"
)

type Managers struct {
	Dataman        *datamanager.DataManager
	EventBus       *pubsub.EventBus
	OverlayManager *overlay.OverlayManager
	PopupManager   *popup.Manager
	Audioman       *audio.AudioManager
}

type Screen interface {
	GetID() defs.ScreenID
	// Load is for preparing the screen for use.
	// TODO: should we pass parameters or something? I'm wondering about how we use screens that are used with variable data.
	// for example, a chest inventory screen; it will need to know a specific chest state (not yet implemented) to open and load data for.
	// maybe params could be 'any', and then we can pre-define params for different screen types, like a "chestParams" which includes a chestID.
	// then, the screen that loads can confirm it got the right params type, or panic if not.
	Load(mgmt Managers, gameCtx defs.GameContext, params any)
	Update()
	OnOpen() // runs before the screen shows; for things like refreshing data if it could be out of date
	Draw(screen *ebiten.Image)
	// IsDone tells the code that triggered this screen that it can move on now.
	IsDone() bool

	// clones the screen to ensure no pointers to the original are passed. this prevents changes to screens being reflected back into the screen manager version.
	Clone() Screen
}

// LoadScreenParams is just for convenience to pass params through functions; includes the pointers you need for loading a screen.
type LoadScreenParams struct {
	OverlayManager *overlay.OverlayManager
	PopupManager   *popup.Manager
	GameCtx        defs.GameContext
}

type ScreenManager struct {
	screens map[defs.ScreenID]Screen
}

func NewScreenManager() *ScreenManager {
	sm := ScreenManager{
		screens: make(map[defs.ScreenID]Screen),
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

func (sm ScreenManager) GetScreen(id defs.ScreenID) Screen {
	if id == "" {
		panic("id was empty")
	}

	s, exists := sm.screens[id]
	if !exists {
		logz.Panicln("ScreenManager", "tried to get screen, but id doesn't exist:", id)
	}

	// clone it to attempt to prevent changes being reflected to "source" screen stored here in the screen manager.
	// Note that this clone probably (depending on screen implementation) just does a "shallow copy", so if there are
	// reference fields (slices, maps, pointers, etc) in the screen then those could still get affected.
	// In such a case, probably a good idea to have the Load function of the screen reset things to their expected initial states.
	return s.Clone()
}

// ScreenViewer provides the basic setup needed to run most screens;
// things like overlay managers, popup managers, and the minimal logic needed for updates and drawing control.
type ScreenViewer struct {
	lastUpdateTick int64 // used for detecting if OnOpen should be called
	scr            Screen
	om             *overlay.OverlayManager
	popupMgr       *popup.Manager
}

func NewScreenViewer(scr Screen, dataman *datamanager.DataManager, eventBus *pubsub.EventBus, audioman *audio.AudioManager, gameCtx defs.GameContext, params any) ScreenViewer {
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
	if audioman == nil {
		panic("audioman was nil")
	}

	popupMgr := popup.NewPopupManager()
	om := overlay.OverlayManager{}
	sv := ScreenViewer{
		scr:      scr,
		om:       &om,
		popupMgr: &popupMgr,
	}

	mgmt := Managers{
		Dataman:        dataman,
		EventBus:       eventBus,
		OverlayManager: &om,
		PopupManager:   &popupMgr,
		Audioman:       audioman,
	}

	sv.scr.Load(mgmt, gameCtx, params)

	return sv
}

func (sv ScreenViewer) IsDone() bool {
	return sv.scr.IsDone()
}

func (sv *ScreenViewer) Update() {
	if sv.scr == nil {
		logz.Panic("screen was nil!")
	}
	if ebiten.Tick()-sv.lastUpdateTick > 1 {
		sv.scr.OnOpen()
	}
	sv.lastUpdateTick = ebiten.Tick()

	// NOTE: I don't stop updates even if IsDone is true, since in some cases screens might stay active until something else decides to end it.
	if sv.popupMgr.IsPopupActive() {
		sv.popupMgr.Update()
		return
	}

	sv.scr.Update()
}

func (sv *ScreenViewer) Draw(screen *ebiten.Image) {
	if ebiten.Tick()-sv.lastUpdateTick > 1 {
		// don't allow draw if update hasn't run recently, since it might need to run OnOpen
		return
	}

	sv.scr.Draw(screen)
	sv.popupMgr.Draw(screen)
	sv.om.Draw(screen)
}
