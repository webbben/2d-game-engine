package body

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/internal/logz"
)

// BodyPartSet represents an individual (visual) part of an entity, such as the body, arms, equiped items, etc.
// It is essentially a collection of animations.
type BodyPartSet struct {
	Name        string
	sourceSet   bool                 // indicates if a source has been set yet (tilesetSrc, etc)
	PartSrc     defs.SelectedPartDef // tileset and image source definitions
	IsRemovable bool                 // if true, this body part set can be removed or hidden (i.e. have None set to true).

	// animation definitions

	animIndex          int       // index or "step" of the animation we are currently on
	reachedLastFrame   bool      // used to detect when an animation has finished (if all sets are at last frame, entire animation is done)
	IdleAnimation      Animation `json:"-"`
	WalkAnimation      Animation `json:"-"`
	RunAnimation       Animation `json:"-"`
	SlashAnimation     Animation `json:"-"`
	BackslashAnimation Animation `json:"-"`
	ShieldAnimation    Animation `json:"-"`
	HasUp              bool      // if true, this set has an "up" direction animation. some don't since they will be covered by the body (such as eyes)

	img *ebiten.Image `json:"-"`
}

func (bps BodyPartSet) HasLoaded() bool {
	return bps.sourceSet
}

type BodyPartSetParams struct {
	IsBody      bool // if true, this body part set will be treated as the main body set. this allows things like StepsOffsetY to be used.
	HasUp       bool // if true, this set has animation frames for "up". some may not, since they might be covered up (e.g. the eyes set)
	IsRemovable bool // if true, this set can be removed or hidden from rendering
	Name        string
}

// NewBodyPartSet creates a new body part (e.g. arms, body, equiped head, etc) to which a part def (i.e. the actual data for animations, etc) can be set.
// a BodyPartSet is essentially just the "slot" where a part definition can be placed. Contains some high level rules, like if the part has an "up" direction, is removable, etc.
func NewBodyPartSet(params BodyPartSetParams) BodyPartSet {
	if params.IsBody && params.IsRemovable {
		panic("body set cannot be removed")
	}
	if params.Name == "" {
		panic("must set name for bodyPartSet (for debugging purposes)")
	}

	bps := BodyPartSet{
		HasUp:       params.HasUp,
		IsRemovable: params.IsRemovable,
		Name:        params.Name,
		// all parts start off as being "none"/disabled. a partSrc can be added later.
		PartSrc: defs.SelectedPartDef{None: true},
	}

	return bps
}

func (bps BodyPartSet) animationDebugString() string {
	if bps.PartSrc.None {
		return fmt.Sprintf("[%s] NONE", bps.Name)
	}
	if !bps.HasUp {
		return fmt.Sprintf("[%s] No Up", bps.Name)
	}

	s := fmt.Sprintf("[%s] animIndex: %v lastframe: %v", bps.Name, bps.animIndex, bps.reachedLastFrame)

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
	bps.WalkAnimation.validate()
	bps.RunAnimation.validate()
	bps.SlashAnimation.validate()
	bps.BackslashAnimation.validate()
	bps.IdleAnimation.validate()
	bps.ShieldAnimation.validate()

	// animation validation
	if bps.animIndex > 20 {
		// animation index is oddly high; is there a bug in detecting the end of an animation?
		// Note: if we want to support really long animations that have 20+ frames, just increase this upper threshold number
		logz.Println(bps.Name, bps.animationDebugString())
		logz.Panicln(bps.Name, "anim index is oddly high (>20). either we have an animation with a lot of frames, or something is going wrong with anim index.")
	}
}

func (bps *BodyPartSet) unsetAllImages() {
	bps.WalkAnimation.reset()
	bps.RunAnimation.reset()
	bps.SlashAnimation.reset()
	bps.BackslashAnimation.reset()
	bps.IdleAnimation.reset()
	bps.ShieldAnimation.reset()
	bps.img = nil
	logz.Println(bps.Name, "unsetting all images")
}

func (bps *BodyPartSet) setImageSource(def defs.SelectedPartDef, stretchX, stretchY int, aux bool) {
	bps.PartSrc = def
	bps.sourceSet = true
	bps.load(stretchX, stretchY, aux)
}

func (set *BodyPartSet) load(stretchX, stretchY int, aux bool) {
	set.unsetAllImages()
	logz.Println(set.Name, "loading bodyPartSet")

	set.animIndex = 0

	if set.PartSrc.None {
		return
	}
	// leaving this below the above None check, since it makes it easier to define a None set without having to actually do the load process.
	if !set.sourceSet {
		panic("source not set before attempting to load")
	}

	set.WalkAnimation.Name = fmt.Sprintf("%s/walk", set.Name)
	set.RunAnimation.Name = fmt.Sprintf("%s/run", set.Name)
	set.SlashAnimation.Name = fmt.Sprintf("%s/slash", set.Name)
	set.BackslashAnimation.Name = fmt.Sprintf("%s/backslash", set.Name)
	set.IdleAnimation.Name = fmt.Sprintf("%s/idle", set.Name)
	set.ShieldAnimation.Name = fmt.Sprintf("%s/shield", set.Name)
	set.PartSrc.IdleAnimation.Name = "idle"

	set.WalkAnimation.load(set.PartSrc.WalkAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
	set.RunAnimation.load(set.PartSrc.RunAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
	set.SlashAnimation.load(set.PartSrc.SlashAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
	set.BackslashAnimation.load(set.PartSrc.BackslashAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
	set.ShieldAnimation.load(set.PartSrc.ShieldAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)
	set.IdleAnimation.load(set.PartSrc.IdleAnimation, aux, set.HasUp, set.PartSrc.FlipRForL, stretchX, stretchY)

	set.validate()
}

func (set *BodyPartSet) setCurrentFrame(dir byte, animationName string) {
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
	case AnimWalk:
		set.img = set.WalkAnimation.getFrame(dir, set.animIndex)
	case AnimRun:
		set.img = set.RunAnimation.getFrame(dir, set.animIndex)
	case AnimSlash:
		set.img = set.SlashAnimation.getFrame(dir, set.animIndex)
	case AnimBackslash:
		set.img = set.BackslashAnimation.getFrame(dir, set.animIndex)
	case AnimIdle:
		set.img = set.IdleAnimation.getFrame(dir, set.animIndex)
	case AnimShield:
		set.img = set.ShieldAnimation.getFrame(dir, set.animIndex)
	default:
		panic("unrecognized animation name: " + animationName)
	}
}

func (set BodyPartSet) getCurrentYOffset(animationName string, direction byte) int {
	switch animationName {
	case AnimWalk:
		return set.WalkAnimation.GetOffsetY(direction, set.animIndex)
	case AnimRun:
		return set.RunAnimation.GetOffsetY(direction, set.animIndex)
	case AnimSlash:
		return set.SlashAnimation.GetOffsetY(direction, set.animIndex)
	case AnimBackslash:
		return set.BackslashAnimation.GetOffsetY(direction, set.animIndex)
	case AnimIdle:
		return set.IdleAnimation.GetOffsetY(direction, set.animIndex)
	case AnimShield:
		return set.ShieldAnimation.GetOffsetY(direction, set.animIndex)
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

	set.reachedLastFrame = false
	numSteps := 0
	// For now, we assume all directions of an animation have the same length. if this changes, we need to redo this logic
	switch animationName {
	case AnimWalk:
		// this body part skips this animation
		if set.WalkAnimation.Skip {
			// if we skip, mark this body part as having reached the last frame so that the animation doesn't get hung up waiting on this one
			set.reachedLastFrame = true
			return
		}
		numSteps = len(set.WalkAnimation.L)
	case AnimRun:
		if set.RunAnimation.Skip {
			set.reachedLastFrame = true
			return
		}
		numSteps = len(set.RunAnimation.L)
	case AnimSlash:
		if set.SlashAnimation.Skip {
			set.reachedLastFrame = true
			return
		}
		numSteps = len(set.SlashAnimation.L)
	case AnimBackslash:
		if set.BackslashAnimation.Skip {
			set.reachedLastFrame = true
			return
		}
		numSteps = len(set.BackslashAnimation.L)
	case AnimIdle:
		if set.IdleAnimation.Skip {
			set.reachedLastFrame = true
			return
		}
		numSteps = len(set.IdleAnimation.L)
	case AnimShield:
		if set.ShieldAnimation.Skip {
			set.reachedLastFrame = true
			return
		}
		numSteps = len(set.ShieldAnimation.L)
	default:
		logz.Panicln(set.Name, "nextFrame: animation name has no registered animation sequence:", animationName)
	}

	// do below the above switch, so that if the animation is skipped we don't keep incrementing animIndex
	set.animIndex++

	if numSteps == 0 {
		logz.Panicln(set.Name, "anim: ", animationName, "num steps is somehow 0")
		// set.animIndex = 0
		// set.reachedLastFrame = true
	}
	// ensure we don't go past the last frame - and mark this body part as done with the animation, if it has.
	if set.animIndex >= numSteps {
		set.reachedLastFrame = true
		set.animIndex = numSteps - 1
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
	set.setImageSource(defs.SelectedPartDef{None: true}, 0, 0, false)

	// sanity checks
	if !set.PartSrc.None {
		panic("removed body part set, but None is false")
	}
	if set.img != nil {
		panic("removed body part set, but img is not nil")
	}
}

// Hide hides the body part (without actually clearing PartSrc).
// basically meant for toggling on and off a part from rendering, such as when a weapon is sheathed or unsheathed.
func (set *BodyPartSet) Hide() {
	if !set.IsRemovable {
		logz.Panic("set is not removable!")
	}
	set.PartSrc.None = true
	set.setImageSource(set.PartSrc, 0, 0, false)

	// sanity checks
	if !set.PartSrc.None {
		panic("hid body part set, but None is false")
	}
	if set.img != nil {
		panic("hid removed body part set, but img is not nil")
	}
}
