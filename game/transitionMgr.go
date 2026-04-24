package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/config"
	"github.com/webbben/2d-game-engine/cutscene"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/logz"
	"github.com/webbben/2d-game-engine/screen"
)

// TransitionManager handles transitioning from one screen to another, and loading content asynchronously during the transition.
// There are a few likely use cases for transitions:
//
// 1) Loading a new map when a player moved into it
//
// 2) Transitioning into some kind of simple cutscene that just shows text, or some kind of animation (via a Screen).
//
// ... but, generally, anytime you want to cut away from the game temporarily to do something, and possibly show a screen during that time.
type TransitionManager struct {
	TransitionInProgress            bool // if set, there is a transition in progress
	ShowingLoadingScreen            bool // if true, there is a loading screen that is showing in between transitions; during this time, don't show anything else.
	OpenTransition, CloseTransition defs.Transition
	LoadingScreen                   screen.Screen
	loadingScreenViewer             screen.ScreenViewer
	LoadingProgress                 float64 // progress, updated by loadFunction, to indicate the percentage of progress made.
	LoadingComplete                 bool    // set by loadFunction when loading has finished
	loadFunc                        func(ctx defs.GameContext)
	runLoadFuncSync                 bool // if true, the load function will NOT be launched in a separate go routine; it will just run in the same process synchronously.
}

func (g *Game) StartCustomLoadScreen(scrID defs.ScreenID, open, close defs.Transition, loadFunction func(ctx defs.GameContext)) {
	logz.Println("TransitionManager", "Starting custom load process")
	if g.TransitionManager.TransitionInProgress {
		logz.Panicln("TransitionInProgress", "tried to start transition, but one is already in progress...")
	}

	// block player changes so they can't accidentally enter the same map twice
	if g.World != nil {
		g.World.BlockPlayerChanges = true
	}

	// reset loading flags and start up the loading process
	g.TransitionManager.LoadingProgress = 0
	g.TransitionManager.LoadingComplete = false
	g.TransitionManager.TransitionInProgress = true

	g.TransitionManager.loadFunc = loadFunction
	g.TransitionManager.runLoadFuncSync = false

	scr := g.ScreenManager.GetScreen(scrID)
	g.TransitionManager.LoadingScreen = scr
	g.TransitionManager.loadingScreenViewer = screen.NewScreenViewer(scr, g.Dataman, g.EventBus, g.AudioManager, g, nil)

	g.TransitionManager.OpenTransition = open
	g.TransitionManager.CloseTransition = close
}

func (g *Game) StartLoadScreen(loadFunction func(ctx defs.GameContext)) {
	logz.Println("TransitionManager", "Starting load process")
	if g.TransitionManager.TransitionInProgress {
		logz.Panicln("TransitionInProgress", "tried to start transition, but one is already in progress...")
	}

	// block player changes so they can't accidentally enter the same map twice
	if g.World != nil {
		g.World.BlockPlayerChanges = true
	}

	// reset loading flags and start up the loading process
	g.TransitionManager.LoadingProgress = 0
	g.TransitionManager.LoadingComplete = false
	g.TransitionManager.TransitionInProgress = true

	g.TransitionManager.loadFunc = loadFunction
	g.TransitionManager.runLoadFuncSync = false

	scr := g.ScreenManager.GetScreen(config.DefaultLoadingScreen)
	g.TransitionManager.LoadingScreen = scr
	g.TransitionManager.loadingScreenViewer = screen.NewScreenViewer(scr, g.Dataman, g.EventBus, g.AudioManager, g, nil)

	g.TransitionManager.OpenTransition = cutscene.NewFadeToBlackTransition(0.9)
	g.TransitionManager.CloseTransition = cutscene.NewFadeFromBlackTransition(0.9)
}

func (tm *TransitionManager) runLoadFunction(ctx defs.GameContext) {
	tm.loadFunc(ctx)
	tm.LoadingComplete = true
	// we leave it up to the load screen to actually notice this, and set Loading to false.
	// the reason is, we want to allow it to do things like fade out transitions, instead of immediately cutting the load screen from view.
	logz.Println("TransitionManager", "Loading function completed.")
}

// StartSyncTransition is for starting a "lightweight" transition - a transition that has no screen, and is only for lightweight setup.
// It's workload runs SYNCHRONOUSLY, so ensure that nothing within expects to await events or anything that requires main Update processing.
// Think of a simple fade out and back in. If you pass a setup function, it should be something that can execute immediately and in within the Update loop without causing hang.
// For heavier loading that could be take time, use StartLoadScreen so that the loading can be done in a separate go routine and a screen can show during the wait.
func (g *Game) StartSyncTransition(open, close defs.Transition, lightWeightSetup func(ctx defs.GameContext)) {
	logz.Println("TransitionManager", "Starting basic sync transition")
	if g.TransitionManager.TransitionInProgress {
		logz.Panicln("TransitionInProgress", "tried to start transition, but one is already in progress...")
	}

	// block player changes so they can't accidentally enter the same map twice
	if g.World != nil {
		g.World.BlockPlayerChanges = true
	}

	// reset loading flags and start up the loading process
	g.TransitionManager.LoadingComplete = true // initialize this to true, since we aren't waiting on any loading function
	g.TransitionManager.TransitionInProgress = true
	g.TransitionManager.OpenTransition = open
	g.TransitionManager.CloseTransition = close

	g.TransitionManager.loadFunc = lightWeightSetup
	g.TransitionManager.runLoadFuncSync = true
}

func (g Game) GetLoadingStatus() (complete bool, progress float64) {
	return g.TransitionManager.LoadingComplete, g.TransitionManager.LoadingProgress
}

func (tm *TransitionManager) Update(gameCtx defs.GameContext) {
	tm.ShowingLoadingScreen = false

	// show opening transition until its complete
	if !tm.OpenTransition.IsDone() {
		tm.OpenTransition.Update()
		if tm.OpenTransition.IsDone() {
			tm.ShowingLoadingScreen = true
			// start the loading process now that the opening transition is done;
			// we don't want to start it until this point, because it could mess with the game struct which could lead to problems
			// if the game world is still showing (e.g. loading a new map while an existing map is still showing)
			if tm.loadFunc != nil {
				if tm.runLoadFuncSync {
					logz.Println("TransitionManager", "running load function sync")
					tm.runLoadFunction(gameCtx)
				} else {
					logz.Println("TransitionManager", "running load function async")
					go tm.runLoadFunction(gameCtx)
				}
			} else {
				tm.LoadingComplete = true
			}
		}
		return
	}

	// show the loading screen while loading is still working or the screen isn't finished yet
	if !tm.LoadingComplete || !tm.loadingScreenViewer.IsDone() {
		tm.ShowingLoadingScreen = true
		tm.loadingScreenViewer.Update()
		return
	}

	// once loading is done, show the closing transition
	if !tm.CloseTransition.IsDone() {
		tm.CloseTransition.Update()
		return
	}

	// everything is fully finished
	tm.TransitionInProgress = false
	logz.Println("TransitionManager", "Transition finished")
}

func (tm *TransitionManager) Draw(screen *ebiten.Image) {
	if !tm.OpenTransition.IsDone() {
		tm.OpenTransition.Draw(screen)
		return
	}
	if tm.ShowingLoadingScreen {
		tm.loadingScreenViewer.Draw(screen)
		return
	}
	if !tm.LoadingComplete {
		logz.Panicln("TransitionManager", "why are we not showing a loading screen, but loading isn't completed yet?")
	}

	if !tm.CloseTransition.IsDone() {
		tm.CloseTransition.Draw(screen)
	}
}
