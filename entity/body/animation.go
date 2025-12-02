package body

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

const (
	AnimIdle      = "idle"
	AnimWalk      = "walk"
	AnimRun       = "run"
	AnimSlash     = "slash"
	AnimBackslash = "backslash"
)

func validateAnimation(anim string) {
	switch anim {
	case AnimIdle, AnimWalk, AnimRun, AnimSlash, AnimBackslash:
		return
	case "":
		panic("animation name is empty (this is not supported; for 'no animation' use the idle animation)")
	default:
		panic("unrecognized animation: " + anim)
	}
}

// IsAttacking determines if the body is currently doing an attack animation
func (eb EntityBodySet) IsAttacking() bool {
	switch eb.animation {
	case AnimSlash:
		return true
	case AnimBackslash:
		return true
	default:
		return false
	}
}

// IsMoving determines if the body is currently running, walking, or doing a purposeful movement animation
func (eb EntityBodySet) IsMoving() bool {
	switch eb.animation {
	case AnimWalk:
		return true
	case AnimRun:
		return true
	default:
		return false
	}
}

// Animation defines an animation for a specific bodyPartSet: which frames to show in which order, etc.
type Animation struct {
	Name string
	Skip bool            // if true, this animation does not get defined
	L    []*ebiten.Image `json:"-"`
	R    []*ebiten.Image `json:"-"`
	U    []*ebiten.Image `json:"-"`
	D    []*ebiten.Image `json:"-"`

	// (initial) frames for each direction, for when an aux is equiped. (Generally only used for the arms set).
	// currently, these are only used for the arms set and the equipedBody set.
	leftAux, rightAux, upAux, downAux *ebiten.Image

	// defines how many ID "steps" from the origin (for a given direction) to get to each animation frame.
	// therefore, it's assumed that each direction of an animation has the same relative pattern of "tile steps".
	// this just makes it easier programmatically; if the pattern were to change, it would be a pain in the ass to
	// have to redo each number for each direction. with this system, you just have to keep the pattern and the start index correct.
	TileSteps    []int
	StepsOffsetY []int // Only set by the body set
}

type AnimationParams struct {
	Skip         bool
	TileSteps    []int
	StepsOffsetY []int
}

func NewAnimation(params AnimationParams) Animation {
	if len(params.StepsOffsetY) != 0 {
		if len(params.StepsOffsetY) != len(params.TileSteps) {
			panic("if stepsOffsetY is defined, it should be the same length as tileSteps")
		}
	}
	a := Animation{
		TileSteps:    params.TileSteps,
		StepsOffsetY: params.StepsOffsetY,
		Skip:         params.Skip,
	}

	return a
}

func (a Animation) debugString() string {
	return fmt.Sprintf("Name: %s Skip: %v tileStepsLen: %v", a.Name, a.Skip, len(a.TileSteps))
}

func (a Animation) validate() {
	if a.Skip {
		return
	}

	if a.Name == "" {
		panic("no animation name")
	}

	// even if an animation only has a single frame (i.e. no TileSteps) it should still have a single frame for each direction
	if len(a.L) == 0 {
		logz.Panicln(a.Name, "left animation is empty")
	}
	if len(a.R) == 0 {
		logz.Panicln(a.Name, "right animation is empty")
	}
	// Up might be empty, if set is !hasUp.  TODO add this as a property to Animation so we can verify here?

	if len(a.D) == 0 {
		logz.Panicln(a.Name, "down animation is empty")
	}

	// no animation defined; this just shows a single frame
	if len(a.TileSteps) == 0 {
		return
	}

	if len(a.L) != len(a.TileSteps) {
		panic(a.Name + ": animation frame count doesn't match TileSteps")
	}
	if len(a.R) != len(a.TileSteps) {
		panic(a.Name + ": animation frame count doesn't match TileSteps")
	}
	if len(a.U) != len(a.TileSteps) {
		panic(a.Name + ": animation frame count doesn't match TileSteps")
	}
	if len(a.D) != len(a.TileSteps) {
		panic(a.Name + ": animation frame count doesn't match TileSteps")
	}
}

func (a *Animation) reset() {
	a.L = make([]*ebiten.Image, 0)
	a.R = make([]*ebiten.Image, 0)
	a.U = make([]*ebiten.Image, 0)
	a.D = make([]*ebiten.Image, 0)
	a.leftAux = nil
	a.rightAux = nil
	a.upAux = nil
	a.downAux = nil
}

func (a Animation) getFrame(dir byte, animationIndex int, aux bool) *ebiten.Image {
	if a.Skip {
		return nil
	}
	if animationIndex < 0 {
		logz.Panicf("animation index is negative? %v", animationIndex)
	}

	// check if we should show the aux frame:
	// aux is set, and (if tileSteps is defined) the current step is 0. If there are no steps, then we also allow it.
	auxCondition := aux && (len(a.TileSteps) == 0 || (len(a.TileSteps) > 0 && a.TileSteps[animationIndex] == 0))

	switch dir {
	case 'L':
		if len(a.L) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.L) {
			panic("past last index")
		}
		if auxCondition && a.leftAux != nil {
			return a.leftAux
		}
		return a.L[animationIndex]
	case 'R':
		if len(a.R) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.R) {
			panic("past last index")
		}
		if auxCondition && a.rightAux != nil {
			return a.rightAux
		}
		return a.R[animationIndex]
	case 'U':
		if len(a.U) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.U) {
			panic("past last index")
		}
		if auxCondition && a.upAux != nil {
			return a.upAux
		}
		return a.U[animationIndex]
	case 'D':
		if len(a.D) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.D) {
			panic("past last index")
		}
		if auxCondition && a.downAux != nil {
			return a.downAux
		}
		return a.D[animationIndex]
	}
	panic("unrecognized direction: " + string(dir))
}

func (a *Animation) loadFrames(tilesetSrc string, rStart, lStart, uStart, dStart, stretchX, stretchY int, flip, hasUp bool, auxStep int) {
	if a.Skip {
		return
	}

	a.leftAux = nil
	a.rightAux = nil
	a.upAux = nil
	a.downAux = nil

	if flip {
		a.L = getAnimationFrames(tilesetSrc, rStart, a.TileSteps, true, stretchX, stretchY)
		if auxStep != 0 {
			a.leftAux = loadFrameImg(tilesetSrc, rStart+auxStep, true, stretchX, stretchY)
		}
	} else {
		a.L = getAnimationFrames(tilesetSrc, lStart, a.TileSteps, false, stretchX, stretchY)
		if auxStep != 0 {
			a.leftAux = loadFrameImg(tilesetSrc, lStart+auxStep, false, stretchX, stretchY)
		}
	}
	a.R = getAnimationFrames(tilesetSrc, rStart, a.TileSteps, false, stretchX, stretchY)
	if auxStep != 0 {
		a.rightAux = loadFrameImg(tilesetSrc, rStart+auxStep, false, stretchX, stretchY)
	}
	if hasUp {
		a.U = getAnimationFrames(tilesetSrc, uStart, a.TileSteps, false, stretchX, stretchY)
		if auxStep != 0 {
			a.upAux = loadFrameImg(tilesetSrc, uStart+auxStep, false, stretchX, stretchY)
		}
	}
	a.D = getAnimationFrames(tilesetSrc, dStart, a.TileSteps, false, stretchX, stretchY)
	if auxStep != 0 {
		a.downAux = loadFrameImg(tilesetSrc, dStart+auxStep, false, stretchX, stretchY)
	}
}

func getAnimationFrames(tilesetSrc string, startIndex int, indexSteps []int, flip bool, stretchX, stretchY int) []*ebiten.Image {
	if tilesetSrc == "" {
		panic("no tilesetSrc passed")
	}
	frames := []*ebiten.Image{}

	if len(indexSteps) == 0 {
		// no animation defined; just use the start tile
		img := loadFrameImg(tilesetSrc, startIndex, flip, stretchX, stretchY)
		frames = append(frames, img)
	}
	for _, step := range indexSteps {
		if step == -1 {
			// indicates a skip frame
			frames = append(frames, nil)
			continue
		}
		img := loadFrameImg(tilesetSrc, startIndex+step, flip, stretchX, stretchY)
		frames = append(frames, img)
	}
	return frames
}

func loadFrameImg(tilesetSrc string, index int, flip bool, stretchX, stretchY int) *ebiten.Image {
	img := tiled.GetTileImage(tilesetSrc, index)
	if flip {
		img = rendering.FlipHoriz(img)
	}
	if stretchX != 0 || stretchY != 0 {
		img = stretchImage(img, stretchX, stretchY)
	}
	return img
}

// stretches the image while keeping it in its same original frame size (centered within)
func stretchImage(img *ebiten.Image, stretchX, stretchY int) *ebiten.Image {
	if stretchX == 0 && stretchY == 0 {
		panic("no stretch set")
	}

	originalBounds := img.Bounds()

	stretchedImage := img
	if stretchX > 0 {
		stretchedImage = rendering.StretchMiddle(stretchedImage)
	}
	stretchedImage = rendering.StretchImage(stretchedImage, 0, stretchY)
	stretchedBounds := stretchedImage.Bounds()

	x := (originalBounds.Dx() / 2) - (stretchedBounds.Dx() / 2)
	y := (originalBounds.Dy() / 2) - (stretchedBounds.Dy() / 2)

	newImg := ebiten.NewImage(originalBounds.Dx(), originalBounds.Dy())
	rendering.DrawImage(newImg, stretchedImage, float64(x), float64(y), 0)

	return newImg
}

type damageFlickerFX struct {
	tickDuration int
	tickCount    int
	show         bool // if true, damage flicker effect will run
	red          bool // if true, the flicker is on the red step. otherwise it's on the white step
}

func (dfx *damageFlickerFX) update() {
	if !dfx.show {
		return
	}
	if dfx.tickCount > dfx.tickDuration {
		dfx.show = false
	}
	dfx.tickCount++
	if dfx.tickCount%5 == 0 {
		dfx.red = !dfx.red
	}
}

func (eb *EntityBodySet) SetDamageFlicker(tickDuration int) {
	if tickDuration <= 0 {
		panic("invalid tick duration")
	}
	eb.dmgFlicker = damageFlickerFX{
		tickDuration: tickDuration,
		show:         true,
		red:          true,
	}
}
