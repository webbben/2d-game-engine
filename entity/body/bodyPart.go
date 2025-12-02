package body

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/internal/logz"
)

// BodyPartSet represents either the head, body, eyes, or hair of an entity.
//
// Defines the animation patterns for each body part, so this is required to be defined for each entity.
// The actual body part definitions (which tiles to show for hair, eyes, etc) are defined by the TilesetSrc and start indices, and can be set
// using the set functions.
type BodyPartSet struct {
	Name        string
	sourceSet   bool            // indicates if a source has been set yet (tilesetSrc, etc)
	PartSrc     SelectedPartDef // tileset and image source definitions
	IsRemovable bool            // if true, this body part set can be removed or hidden (i.e. have None set to true).

	// animation definitions

	animIndex          int  // index or "step" of the animation we are currently on
	reachedLastFrame   bool // used to detect when an animation has finished (if all sets are at last frame, entire animation is done)
	IdleAnimation      Animation
	WalkAnimation      Animation
	RunAnimation       Animation
	SlashAnimation     Animation
	BackslashAnimation Animation
	HasUp              bool // if true, this set has an "up" direction animation. some don't since they will be covered by the body (such as eyes)

	img *ebiten.Image `json:"-"`
}

func (bps BodyPartSet) HasLoaded() bool {
	return bps.sourceSet
}

type BodyPartSetParams struct {
	IsBody          bool // if true, this body part set will be treated as the main body set. this allows things like StepsOffsetY to be used.
	HasUp           bool // if true, this set has animation frames for "up". some may not, since they might be covered up (e.g. the eyes set)
	IsRemovable     bool // if true, this set can be removed or hidden from rendering
	WalkParams      AnimationParams
	RunParams       AnimationParams
	SlashParams     AnimationParams
	BackslashParams AnimationParams
	IdleParams      AnimationParams
	Name            string
}

func NewBodyPartSet(params BodyPartSetParams) BodyPartSet {
	if !params.IsBody {
		if len(params.WalkParams.StepsOffsetY) != 0 {
			logz.Panic("non-body sets should not define a stepsOffsetY; that is only for the main body set to define.")
		}
		if len(params.RunParams.StepsOffsetY) != 0 {
			logz.Panic("non-body sets should not define a stepsOffsetY; that is only for the main body set to define.")
		}
		if len(params.SlashParams.StepsOffsetY) != 0 {
			logz.Panic("non-body sets should not define a stepsOffsetY; that is only for the main body set to define.")
		}
		if len(params.BackslashParams.StepsOffsetY) != 0 {
			logz.Panic("non-body sets should not define a stepsOffsetY; that is only for the main body set to define.")
		}
	}
	if params.IsBody && params.IsRemovable {
		panic("body set cannot be removed")
	}
	if params.Name == "" {
		panic("must set name for bodyPartSet (for debugging purposes)")
	}
	bps := BodyPartSet{
		WalkAnimation:      NewAnimation(params.WalkParams),
		RunAnimation:       NewAnimation(params.RunParams),
		SlashAnimation:     NewAnimation(params.SlashParams),
		BackslashAnimation: NewAnimation(params.BackslashParams),
		IdleAnimation:      NewAnimation(params.IdleParams),
		HasUp:              params.HasUp,
		IsRemovable:        params.IsRemovable,
		Name:               params.Name,
		// all parts start off as being "none"/disabled. a partSrc can be added later.
		PartSrc: SelectedPartDef{None: true},
	}

	return bps
}

func (bps BodyPartSet) animationDebugString(anim string, dir byte) string {
	if bps.PartSrc.None {
		return fmt.Sprintf("[%s] NONE", bps.PartSrc.TilesetSrc)
	}
	if !bps.HasUp && dir == 'U' {
		return fmt.Sprintf("[%s] No Up", bps.PartSrc.TilesetSrc)
	}

	s := fmt.Sprintf("[%s] animIndex: %v lastframe: %v", bps.PartSrc.TilesetSrc, bps.animIndex, bps.reachedLastFrame)

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
	if bps.PartSrc.None {
		return
	}
	if !bps.sourceSet {
		panic("source not set!")
	}
	if bps.Name == "" {
		panic("no name set")
	}
	fmt.Println(bps.PartSrc.TilesetSrc)
	bps.WalkAnimation.validate()
	bps.RunAnimation.validate()
	bps.SlashAnimation.validate()
	bps.IdleAnimation.validate()
}

func (bps *BodyPartSet) unsetAllImages() {
	bps.WalkAnimation.reset()
	bps.RunAnimation.reset()
	bps.SlashAnimation.reset()
	bps.IdleAnimation.reset()
	bps.img = nil
	logz.Println(bps.Name, "unsetting all images")
}

func (bps *BodyPartSet) setImageSource(def SelectedPartDef, stretchX, stretchY int) {
	bps.PartSrc = def
	bps.sourceSet = true
	bps.load(stretchX, stretchY)
}

func (set *BodyPartSet) load(stretchX, stretchY int) {
	set.unsetAllImages()
	logz.Println(set.Name, "loading bodyPartSet")

	if set.PartSrc.None {
		return
	}
	// leaving this below the above None check, since it makes it easier to define a None set without having to actually do the load process.
	if !set.sourceSet {
		panic("source not set before attempting to load")
	}
	if set.PartSrc.TilesetSrc == "" {
		logz.Panicln(set.Name, "no TilesetSrc set in BodyPartSet")
	}

	set.WalkAnimation.Name = fmt.Sprintf("%s/walk", set.PartSrc.TilesetSrc)
	set.RunAnimation.Name = fmt.Sprintf("%s/run", set.PartSrc.TilesetSrc)
	set.SlashAnimation.Name = fmt.Sprintf("%s/slash", set.PartSrc.TilesetSrc)
	set.BackslashAnimation.Name = fmt.Sprintf("%s/backslash", set.PartSrc.TilesetSrc)
	set.IdleAnimation.Name = fmt.Sprintf("%s/idle", set.PartSrc.TilesetSrc)

	set.WalkAnimation.loadFrames(set.PartSrc.TilesetSrc, set.PartSrc.RStart, set.PartSrc.LStart, set.PartSrc.UStart, set.PartSrc.DStart, stretchX, stretchY, set.PartSrc.FlipRForL, set.HasUp, set.PartSrc.AuxFirstFrameStep)
	set.RunAnimation.loadFrames(set.PartSrc.TilesetSrc, set.PartSrc.RStart, set.PartSrc.LStart, set.PartSrc.UStart, set.PartSrc.DStart, stretchX, stretchY, set.PartSrc.FlipRForL, set.HasUp, set.PartSrc.AuxFirstFrameStep)
	set.SlashAnimation.loadFrames(set.PartSrc.TilesetSrc, set.PartSrc.RStart, set.PartSrc.LStart, set.PartSrc.UStart, set.PartSrc.DStart, stretchX, stretchY, set.PartSrc.FlipRForL, set.HasUp, set.PartSrc.AuxFirstFrameStep)
	set.BackslashAnimation.loadFrames(set.PartSrc.TilesetSrc, set.PartSrc.RStart, set.PartSrc.LStart, set.PartSrc.UStart, set.PartSrc.DStart, stretchX, stretchY, set.PartSrc.FlipRForL, set.HasUp, set.PartSrc.AuxFirstFrameStep)
	set.IdleAnimation.loadFrames(set.PartSrc.TilesetSrc, set.PartSrc.RStart, set.PartSrc.LStart, set.PartSrc.UStart, set.PartSrc.DStart, stretchX, stretchY, set.PartSrc.FlipRForL, set.HasUp, set.PartSrc.AuxFirstFrameStep)

	set.validate()
}

func (set *BodyPartSet) setCurrentFrame(dir byte, animationName string, aux bool) {
	if animationName == "" {
		panic("animation is unset")
	}
	if set.PartSrc.None {
		set.img = nil
		return
	}
	if dir == 'U' && !set.HasUp {
		set.img = nil
		return
	}

	switch animationName {
	case ANIM_WALK:
		set.img = set.WalkAnimation.getFrame(dir, set.animIndex, aux)
	case ANIM_RUN:
		set.img = set.RunAnimation.getFrame(dir, set.animIndex, aux)
	case ANIM_SLASH:
		set.img = set.SlashAnimation.getFrame(dir, set.animIndex, aux)
	case ANIM_BACKSLASH:
		set.img = set.BackslashAnimation.getFrame(dir, set.animIndex, aux)
	case ANIM_IDLE:
		set.img = set.IdleAnimation.getFrame(dir, set.animIndex, aux)
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
	case ANIM_IDLE:
		if len(set.IdleAnimation.StepsOffsetY) > 0 {
			return set.IdleAnimation.StepsOffsetY[set.animIndex]
		}
	}

	return 0
}

func (set *BodyPartSet) nextFrame(animationName string) {
	if set.PartSrc.None {
		return
	}
	if !set.sourceSet {
		panic("source not set!")
	}
	if animationName == "" {
		logz.Panic("called nextFrame on empty animation. should this be the idle animation?")
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
	case ANIM_IDLE:
		numSteps = len(set.IdleAnimation.TileSteps)
	default:
		logz.Panicln(set.PartSrc.TilesetSrc, "nextFrame: animation name has no registered animation sequence:", animationName)
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

// Remove completely removes the definition and images of the body part (clears PartSrc and all animation frames).
// should be used when actually removing an item from the entity's body.
func (set *BodyPartSet) Remove() {
	if !set.IsRemovable {
		logz.Panic("set is not removable!")
	}
	set.setImageSource(SelectedPartDef{None: true}, 0, 0)
}

// Hide hides the body part (without actually clearing PartSrc).
// basically meant for toggling on and off a part from rendering, such as when a weapon is sheathed or unsheathed.
func (set *BodyPartSet) Hide() {
	if !set.IsRemovable {
		logz.Panic("set is not removable!")
	}
	set.PartSrc.None = true
	set.setImageSource(set.PartSrc, 0, 0)
}
