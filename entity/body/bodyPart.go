package body

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
)

// represents either the head, body, eyes, or hair of an entity.
//
// Defines the animation patterns for each body part, so this is required to be defined for each entity.
// The actual body part definitions (which tiles to show for hair, eyes, etc) are defined by the TilesetSrc and start indices, and can be set
// using the set functions.
type BodyPartSet struct {
	sourceSet bool            // indicates if a source has been set yet (tilesetSrc, etc)
	PartSrc   SelectedPartDef // tileset and image source definitions

	// TODO do we need TilesetSrc etc here? it seems they are present in PartSrc...

	TilesetSrc                     string
	RStart, LStart, UStart, DStart int
	FlipRForL                      bool
	None                           bool // if set, this body part set is effectively turned off and not drawn or involved in any animations
	stretchX, stretchY             int

	// animation definitions

	animIndex          int  // index or "step" of the animation we are currently on
	reachedLastFrame   bool // used to detect when an animation has finished (if all sets are at last frame, entire animation is done)
	WalkAnimation      Animation
	RunAnimation       Animation
	SlashAnimation     Animation
	BackslashAnimation Animation
	HasUp              bool // if true, this set has an "up" direction animation. some don't since they will be covered by the body (such as eyes)

	img *ebiten.Image `json:"-"`
}

func (bps BodyPartSet) animationDebugString(anim string, dir byte) string {
	if bps.None {
		return fmt.Sprintf("[%s] NONE", bps.TilesetSrc)
	}
	if !bps.HasUp && dir == 'U' {
		return fmt.Sprintf("[%s] No Up", bps.TilesetSrc)
	}

	s := fmt.Sprintf("[%s] animIndex: %v lastframe: %v strX: %v strY: %v", bps.TilesetSrc, bps.animIndex, bps.reachedLastFrame, bps.stretchX, bps.stretchY)

	switch anim {
	case ANIM_WALK:
		s += "\n  " + bps.WalkAnimation.debugString()
	case ANIM_RUN:
		s += "\n  " + bps.RunAnimation.debugString()
	case ANIM_SLASH:
		s += "\n  " + bps.SlashAnimation.debugString()
	case ANIM_BACKSLASH:
		s += "\n  " + bps.BackslashAnimation.debugString()
	}

	return s
}

func (bps BodyPartSet) validate() {
	if bps.None {
		return
	}
	fmt.Println(bps.TilesetSrc)
	bps.WalkAnimation.validate()
	bps.RunAnimation.validate()
	bps.SlashAnimation.validate()
}

// for no body part
var NONE BodyPartSet = BodyPartSet{None: true}

func (bps *BodyPartSet) unsetAllImages() {
	bps.WalkAnimation.reset()
	bps.RunAnimation.reset()
	bps.SlashAnimation.reset()
	bps.img = nil
}

func (bps *BodyPartSet) setImageSource(def SelectedPartDef) {
	bps.TilesetSrc = def.TilesetSrc
	bps.LStart = def.LStart
	bps.RStart = def.RStart
	bps.UStart = def.UStart
	bps.DStart = def.DStart
	bps.FlipRForL = def.FlipRForL
	bps.None = def.None

	// set this so the part can be reloaded later
	bps.PartSrc = def

	bps.sourceSet = true

	bps.load()
}

func (set *BodyPartSet) load() {
	set.WalkAnimation.reset()
	set.RunAnimation.reset()
	set.SlashAnimation.reset()

	if set.None {
		return
	}

	if !set.sourceSet {
		panic("source not set before attempting to load")
	}
	if set.TilesetSrc == "" {
		panic("no TilesetSrc set in BodyPartSet")
	}

	set.WalkAnimation.Name = fmt.Sprintf("%s/walk", set.TilesetSrc)
	set.RunAnimation.Name = fmt.Sprintf("%s/run", set.TilesetSrc)
	set.SlashAnimation.Name = fmt.Sprintf("%s/slash", set.TilesetSrc)
	set.BackslashAnimation.Name = fmt.Sprintf("%s/backslash", set.TilesetSrc)

	set.WalkAnimation.loadFrames(set.TilesetSrc, set.RStart, set.LStart, set.UStart, set.DStart, set.stretchX, set.stretchY, set.FlipRForL, set.HasUp)
	set.RunAnimation.loadFrames(set.TilesetSrc, set.RStart, set.LStart, set.UStart, set.DStart, set.stretchX, set.stretchY, set.FlipRForL, set.HasUp)
	set.SlashAnimation.loadFrames(set.TilesetSrc, set.RStart, set.LStart, set.UStart, set.DStart, set.stretchX, set.stretchY, set.FlipRForL, set.HasUp)
	set.BackslashAnimation.loadFrames(set.TilesetSrc, set.RStart, set.LStart, set.UStart, set.DStart, set.stretchX, set.stretchY, set.FlipRForL, set.HasUp)
}

func (set *BodyPartSet) setCurrentFrame(dir byte, animationName string) {
	if set.None {
		set.img = nil
		return
	}
	if dir == 'U' && !set.HasUp {
		set.img = nil
		return
	}

	switch animationName {
	case ANIM_WALK:
		set.img = set.WalkAnimation.getFrame(dir, set.animIndex)
	case ANIM_RUN:
		set.img = set.RunAnimation.getFrame(dir, set.animIndex)
	case ANIM_SLASH:
		set.img = set.SlashAnimation.getFrame(dir, set.animIndex)
	case ANIM_BACKSLASH:
		set.img = set.BackslashAnimation.getFrame(dir, set.animIndex)
	case "":
		set.img = set.WalkAnimation.getFrame(dir, 0)
	default:
		panic("unrecognized animation name: " + animationName)
	}
}

func (set BodyPartSet) getCurrentYOffset(animationName string) int {
	switch animationName {
	case ANIM_WALK:
		if len(set.WalkAnimation.StepsOffsetY) > 0 {
			return set.WalkAnimation.StepsOffsetY[set.animIndex]
		}
	case ANIM_RUN:
		if len(set.RunAnimation.StepsOffsetY) > 0 {
			return set.RunAnimation.StepsOffsetY[set.animIndex]
		}
	case ANIM_SLASH:
		if len(set.SlashAnimation.StepsOffsetY) > 0 {
			return set.SlashAnimation.StepsOffsetY[set.animIndex]
		}
	case ANIM_BACKSLASH:
		if len(set.BackslashAnimation.StepsOffsetY) > 0 {
			return set.BackslashAnimation.StepsOffsetY[set.animIndex]
		}
	}

	return 0
}

func (set *BodyPartSet) nextFrame(animationName string) {
	if set.None {
		return
	}
	if animationName == "" {
		logz.Panic("called nextFrame on empty animation")
	}

	set.animIndex++
	set.reachedLastFrame = false
	numSteps := 0
	switch animationName {
	case ANIM_WALK:
		numSteps = len(set.WalkAnimation.TileSteps)
	case ANIM_RUN:
		numSteps = len(set.RunAnimation.TileSteps)
	case ANIM_SLASH:
		numSteps = len(set.SlashAnimation.TileSteps)
	case ANIM_BACKSLASH:
		numSteps = len(set.BackslashAnimation.TileSteps)
	default:
		logz.Panicln(set.TilesetSrc, "nextFrame: animation name has no registered animation sequence:", animationName)
	}

	if numSteps == 0 {
		set.animIndex = 0
		set.reachedLastFrame = true
	} else {
		if set.animIndex >= numSteps {
			set.reachedLastFrame = true
			set.animIndex = numSteps - 1
		}
	}

	if set.animIndex < 0 {
		logz.Panicf("nextFrame: somehow animIndex became negative")
	}
}
