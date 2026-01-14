package body

import (
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
	AnimShield    = "shield"
)

func validateAnimation(anim string) {
	switch anim {
	case AnimIdle, AnimWalk, AnimRun, AnimSlash, AnimBackslash, AnimShield:
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
	Name         string
	Skip         bool            // if true, this animation does not get defined
	L            []*ebiten.Image `json:"-"`
	R            []*ebiten.Image `json:"-"`
	U            []*ebiten.Image `json:"-"`
	D            []*ebiten.Image `json:"-"`
	StepsOffsetY []int           // Only set by the body set
}

type AnimationParams struct {
	Name                                      string
	Skip                                      bool
	TilesetSrc                                string
	TilesLeft, TilesRight, TilesUp, TilesDown []int // indices of each frame of the animation
	AuxLeft, AuxRight, AuxUp, AuxDown         []int // if aux influeces this animation, define the aux frames here
	StepsOffsetY                              []int
}

func (ap AnimationParams) IsEqual(other AnimationParams) bool {
	if ap.Name != other.Name {
		return false
	}
	if ap.Skip != other.Skip {
		return false
	}
	if ap.TilesetSrc != other.TilesetSrc {
		return false
	}
	slicesEqual := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for i, v := range a {
			if b[i] != v {
				return false
			}
		}
		return true
	}

	if !slicesEqual(ap.TilesLeft, other.TilesLeft) {
		return false
	}
	if !slicesEqual(ap.TilesRight, other.TilesRight) {
		return false
	}
	if !slicesEqual(ap.TilesUp, other.TilesUp) {
		return false
	}
	if !slicesEqual(ap.TilesDown, other.TilesDown) {
		return false
	}
	if !slicesEqual(ap.AuxLeft, other.AuxLeft) {
		return false
	}
	if !slicesEqual(ap.AuxRight, other.AuxRight) {
		return false
	}
	if !slicesEqual(ap.AuxUp, other.AuxUp) {
		return false
	}
	if !slicesEqual(ap.AuxDown, other.AuxDown) {
		return false
	}
	if !slicesEqual(ap.StepsOffsetY, other.StepsOffsetY) {
		return false
	}

	return true
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

	// confirm all directions are the same length
	if (len(a.L)+len(a.R)+len(a.D))/3 != len(a.L) {
		// we leave out Up, since some sets don't have an Up direction for their animations.
		logz.Panicln(a.Name, "animation directions don't appear to be equal in length")
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
			logz.Panicln(a.Name, "past last index")
		}
		return a.L[animationIndex]
	case 'R':
		if len(a.R) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.R) {
			logz.Panicln(a.Name, "past last index")
		}
		return a.R[animationIndex]
	case 'U':
		if len(a.U) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.U) {
			logz.Panicln(a.Name, "past last index")
		}
		return a.U[animationIndex]
	case 'D':
		if len(a.D) == 0 {
			logz.Panicf("%s: no frames?", a.Name)
		}
		if animationIndex >= len(a.D) {
			logz.Panicln(a.Name, "past last index")
		}
		return a.D[animationIndex]
	}
	panic("unrecognized direction: " + string(dir))
}

// load loads the frames of an animation, given the animation params and other loading options like flipping, stretch, etc.
//
// flipRL: set to true to flip and reuse Right frames for the Left direction
func (a *Animation) load(params AnimationParams, aux, hasUp, flipRL bool, stretchX, stretchY int) {
	if params.Name == "" {
		panic("name is empty")
	}
	a.StepsOffsetY = params.StepsOffsetY
	a.Skip = params.Skip
	a.Name = params.Name
	if a.Skip {
		return
	}

	if a.Name == "" {
		panic("animation has no name")
	}

	l := params.TilesLeft
	r := params.TilesRight
	u := params.TilesUp
	d := params.TilesDown

	if aux {
		// only use different frames for aux if they are defined
		if len(params.AuxLeft) != 0 {
			l = params.AuxLeft
		}
		if len(params.AuxRight) != 0 {
			r = params.AuxRight
		}
		if len(params.AuxUp) != 0 {
			u = params.AuxUp
		}
		if len(params.AuxDown) != 0 {
			d = params.AuxDown
		}
	}

	// can flip Right flames to be reused for Left frames
	if flipRL {
		a.L = getAnimationFrames(params.TilesetSrc, r, true, stretchX, stretchY)
	} else {
		a.L = getAnimationFrames(params.TilesetSrc, l, false, stretchX, stretchY)
	}
	a.R = getAnimationFrames(params.TilesetSrc, r, false, stretchX, stretchY)
	if hasUp {
		a.U = getAnimationFrames(params.TilesetSrc, u, false, stretchX, stretchY)
	} else {
		a.U = []*ebiten.Image{}
	}
	a.D = getAnimationFrames(params.TilesetSrc, d, false, stretchX, stretchY)
}

func getAnimationFrames(tilesetSrc string, indices []int, flip bool, stretchX, stretchY int) []*ebiten.Image {
	if tilesetSrc == "" {
		panic("no tilesetSrc passed")
	}
	frames := []*ebiten.Image{}

	if len(indices) == 0 {
		panic("no indices given... must have at least one frame for an animation.")
	}
	for _, i := range indices {
		if i == -1 {
			// indicates a skip frame
			frames = append(frames, nil)
			continue
		}
		img := loadFrameImg(tilesetSrc, i, flip, stretchX, stretchY)
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
