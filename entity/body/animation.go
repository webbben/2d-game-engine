package body

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
	"github.com/webbben/2d-game-engine/internal/rendering"
	"github.com/webbben/2d-game-engine/internal/tiled"
)

const (
	ANIM_WALK      = "walk"
	ANIM_RUN       = "run"
	ANIM_SLASH     = "slash"
	ANIM_BACKSLASH = "backslash"
)

// if the body is currently doing an attack animation
func (eb EntityBodySet) IsAttacking() bool {
	switch eb.animation {
	case ANIM_SLASH:
		return true
	case ANIM_BACKSLASH:
		return true
	default:
		return false
	}
}

// if the body is currently running, walking, or doing a purposeful movement animation
func (eb EntityBodySet) IsMoving() bool {
	switch eb.animation {
	case ANIM_WALK:
		return true
	case ANIM_RUN:
		return true
	default:
		return false
	}
}

type Animation struct {
	Name         string
	Skip         bool            // if true, this animation does not get defined
	L            []*ebiten.Image `json:"-"`
	R            []*ebiten.Image `json:"-"`
	U            []*ebiten.Image `json:"-"`
	D            []*ebiten.Image `json:"-"`
	TileSteps    []int
	StepsOffsetY []int
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
}

func (a Animation) getFrame(dir byte, animationIndex int) *ebiten.Image {
	if a.Skip {
		return nil
	}
	if animationIndex < 0 {
		logz.Panicf("animation index is negative? %v", animationIndex)
	}

	switch dir {
	case 'L':
		if len(a.L) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.L) {
			panic("past last index")
		}
		return a.L[animationIndex]
	case 'R':
		if len(a.R) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.R) {
			panic("past last index")
		}
		return a.R[animationIndex]
	case 'U':
		if len(a.U) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.U) {
			panic("past last index")
		}
		return a.U[animationIndex]
	case 'D':
		if len(a.D) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.D) {
			panic("past last index")
		}
		return a.D[animationIndex]
	}
	panic("unrecognized direction: " + string(dir))
}

func (a *Animation) loadFrames(tilesetSrc string, rStart, lStart, uStart, dStart, stretchX, stretchY int, flip, hasUp bool) {
	if a.Skip {
		return
	}

	if flip {
		a.L = getAnimationFrames(tilesetSrc, rStart, a.TileSteps, true, stretchX, stretchY)
	} else {
		a.L = getAnimationFrames(tilesetSrc, lStart, a.TileSteps, false, stretchX, stretchY)
	}
	a.R = getAnimationFrames(tilesetSrc, rStart, a.TileSteps, false, stretchX, stretchY)
	if hasUp {
		a.U = getAnimationFrames(tilesetSrc, uStart, a.TileSteps, false, stretchX, stretchY)
	}
	a.D = getAnimationFrames(tilesetSrc, dStart, a.TileSteps, false, stretchX, stretchY)
}

func getAnimationFrames(tilesetSrc string, startIndex int, indexSteps []int, flip bool, stretchX, stretchY int) []*ebiten.Image {
	if tilesetSrc == "" {
		panic("no tilesetSrc passed")
	}
	frames := []*ebiten.Image{}

	if len(indexSteps) == 0 {
		// no animation defined; just use the start tile
		img := tiled.GetTileImage(tilesetSrc, startIndex)
		if flip {
			img = rendering.FlipHoriz(img)
		}
		if stretchX != 0 || stretchY != 0 {
			img = stretchImage(img, stretchX, stretchY)
		}
		frames = append(frames, img)
	}
	for _, step := range indexSteps {
		if step == -1 {
			// indicates a skip frame
			frames = append(frames, nil)
			continue
		}
		img := tiled.GetTileImage(tilesetSrc, startIndex+step)
		if flip {
			img = rendering.FlipHoriz(img)
		}
		if stretchX != 0 || stretchY != 0 {
			img = stretchImage(img, stretchX, stretchY)
		}
		frames = append(frames, img)
	}
	return frames
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
